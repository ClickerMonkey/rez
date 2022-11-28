package db

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
)

// A type that can receive rows from a query.
type QueryResult[R any] interface {
	Add(result R) bool
}

// A QueryResult for a slice.
type QuerySlice[R any] struct {
	Slice *[]R
	Max   int
}

var _ QueryResult[int] = &QuerySlice[int]{}

func (st *QuerySlice[R]) Add(result R) bool {
	if st.Max > 0 && len(*st.Slice) >= st.Max {
		return false
	}
	*st.Slice = append(*st.Slice, result)
	return true
}

// A query to run against a database. For stored procedures the
// query is the stored procedure name. If Prepared is true the
// query will generated a prepared statement on the first run.
// If Results or Result is given those values are populated,
// otherwise the query is executed and RowsAffected and LastID
// could be updated.
type Query[R any] struct {
	// The query, sql statements, or the stored procedure to query/execute.
	Query string
	// If the query will be called multiple times setting this to true
	// will generate a reusable prepared statement that will be used for
	// subsequent runs.
	Prepared bool
	// The input variables
	In map[string]any
	// The output variables
	Out map[string]any
	// Creates a row to be populated. This may be necessary for complex
	// types with pointers, slices, maps, etc.
	Create func() R
	// Where to place the rows received. The presense of this field
	// indicates that rows should be requested and parsed.
	Results QueryResult[R]
	// Where to place a single result row. The presense of this field
	// indicates that a row should be requested and parsed.
	Result *R
	// How many rows were affected on the last query run.
	RowsAffected int64
	// The ID of the last inserted record on the last query run.
	LastID int64

	columns      []func(row reflect.Value) any
	columnsQuery string
	stmt         *sql.Stmt
	stmtQuery    string
}

var _ Runnable = &Query[string]{}

// Preps the query for receiving rows with the given columns
func (q *Query[R]) SetColumns(columns []string) {
	columnMap := make(map[string]int, len(columns))
	for i, c := range columns {
		columnMap[strings.ToLower(c)] = i
	}

	q.columns = make([]func(row reflect.Value) any, len(columns))

	var iterateStruct func(s reflect.Type, fieldIndexes []int)

	iterateStruct = func(s reflect.Type, fieldIndexes []int) {
		for fieldIndex := 0; fieldIndex < s.NumField(); fieldIndex++ {
			field := s.Field(fieldIndex)
			indexes := append(fieldIndexes, fieldIndex)
			if field.Anonymous {
				iterateStruct(field.Type, indexes)
			} else {
				name := field.Name
				if db := field.Tag.Get("db"); db != "" {
					name = db
				}
				key := strings.ToLower(name)
				if i, exists := columnMap[key]; exists {
					q.columns[i] = func(row reflect.Value) any {
						return row.FieldByIndex(indexes).Interface()
					}
				}
			}
		}
	}

	var rowInstance R
	rootType := reflect.TypeOf(rowInstance)
	if rootType.Kind() == reflect.Struct {
		iterateStruct(rootType, []int{})
	}
}

// Gets the references to values in the given row based on the
// columns prepped.
func (q Query[R]) GetValues(row *R) []any {
	r := reflect.ValueOf(row)
	columns := make([]any, len(q.columns))

	if r.Elem().Kind() != reflect.Struct {
		columns[0] = row
	} else {
		for i, handler := range q.columns {
			if handler == nil {
				columns[i] = any(nil)
			} else {
				columns[i] = handler(r)
			}
		}
	}

	return columns
}

// Converts the in & out to arguments to pass to the query.
func (q *Query[R]) GetArgs() []any {
	args := make([]any, 0, len(q.In)+len(q.Out))
	if len(q.In) > 0 {
		for name := range q.In {
			args = append(args, sql.Named(name, q.In[name]))
		}
	}
	if len(q.Out) > 0 {
		for name := range q.Out {
			args = append(args, sql.Named(name, sql.Out{Dest: q.Out[name]}))
		}
	}
	return args
}

// Creates a new row for this query.
func (q *Query[R]) NewRow() *R {
	var value R
	if q.Create != nil {
		value = q.Create()
	}
	return &value
}

// Runs the query against the given context and connection.
func (q *Query[R]) Run(ctx context.Context, conn *sql.Conn) error {
	if q.Prepared && (q.stmt == nil || q.stmtQuery != q.Query) {
		stmt, err := conn.PrepareContext(ctx, q.Query)
		if err != nil {
			return err
		}
		q.stmt = stmt
		q.stmtQuery = q.Query
	}

	if q.Results != nil || q.Result != nil {
		var rows *sql.Rows
		var err error
		if q.stmt != nil {
			rows, err = q.stmt.QueryContext(ctx, q.GetArgs()...)
		} else {
			rows, err = conn.QueryContext(ctx, q.Query, q.GetArgs()...)
		}
		if err != nil {
			return err
		}
		defer rows.Close()

		if q.columns == nil || q.columnsQuery != q.Query {
			cols, err := rows.Columns()
			if err != nil {
				return err
			}
			q.SetColumns(cols)
			q.columnsQuery = q.Query
		}

		for rows.Next() {
			row := q.NewRow()
			err = rows.Scan(q.GetValues(row)...)
			if err != nil {
				return err
			}
			if q.Results != nil {
				if !q.Results.Add(*row) {
					break
				}
			} else {
				*q.Result = *row
				break
			}
		}
	} else {
		var result sql.Result
		var err error
		if q.stmt != nil {
			result, err = q.stmt.ExecContext(ctx, q.GetArgs()...)
		} else {
			result, err = conn.ExecContext(ctx, q.Query, q.GetArgs()...)
		}
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err == nil {
			q.RowsAffected = affected
		}
		lastID, err := result.LastInsertId()
		if err == nil {
			q.LastID = lastID
		}
	}
	return nil
}
