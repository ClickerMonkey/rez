package db

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
)

// An interface that covers a connection or transaction
type Queryable interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var _ Queryable = &sql.Tx{}
var _ Queryable = &sql.DB{}
var _ Queryable = &sql.Conn{}

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
	// The named input variables
	In map[string]any
	// The output variables
	Out map[string]any
	// The position input variables
	InArgs []any
	// The position output variables
	OutArgs []any
	// Creates a row to be populated. This may be necessary for complex
	// types with pointers, slices, maps, etc.
	Create func() R
	// Where to place the rows received. The presense of this field
	// indicates that rows should be requested and parsed.
	Results func(result R, index int) bool
	// Where to place a single result row. The presense of this field
	// indicates that a row should be requested and parsed.
	Result *R
	// How many rows were affected on the last query run.
	RowsAffected int64
	// The ID of the last inserted record on the last query run.
	LastID int64
	// The columns returned from a query
	Columns []*sql.ColumnType

	handlers     []func(row reflect.Value) any
	columnsQuery string
	stmt         *sql.Stmt
	stmtQuery    string
}

var _ Runnable = &Query[string]{}

// Preps the query for receiving rows with the given columns
func (q *Query[R]) SetColumns(columns []*sql.ColumnType) {
	q.Columns = columns
	q.handlers = make([]func(row reflect.Value) any, len(columns))

	var rowInstance R
	rootType := reflect.TypeOf(rowInstance)
	for rootType.Kind() == reflect.Pointer {
		rootType = rootType.Elem()
	}

	if rootType.Kind() == reflect.Struct {
		columnMap := make(map[string]int, len(columns))
		for i, c := range columns {
			columnMap[strings.ToLower(c.Name())] = i
		}

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
						q.handlers[i] = q.handlerForIndex(indexes)
					}
				}
			}
		}

		iterateStruct(rootType, []int{})
	}
	if rootType.Kind() == reflect.Map {
		for i, c := range columns {
			q.handlers[i] = q.handlerForMapColumn(c)
		}
	}
}

func (q Query[R]) handlerForIndex(indexes []int) func(row reflect.Value) any {
	return func(row reflect.Value) any {
		return row.Elem().FieldByIndex(indexes).Interface()
	}
}

func (q Query[R]) handlerForMapColumn(col *sql.ColumnType) func(row reflect.Value) any {
	mapIndex := reflect.ValueOf(col.Name())
	mapValueType := col.ScanType()
	mapValueTypePtr := reflect.PointerTo(mapValueType)
	nullable, _ := col.Nullable()

	return func(row reflect.Value) any {
		val := reflect.New(mapValueType)
		if nullable {
			ptr := reflect.New(mapValueTypePtr)
			ptr.Elem().Set(val)
			val = ptr
		}

		row.Elem().SetMapIndex(mapIndex, val)

		return val.Interface()
	}
}

// Gets the references to values in the given row based on the
// columns prepped.
func (q Query[R]) GetValues(row *R) []any {
	r := reflect.ValueOf(row)
	kind := r.Elem().Kind()
	columns := make([]any, len(q.handlers))

	if kind != reflect.Struct && kind != reflect.Map {
		columns[0] = row
	} else {
		for i, handler := range q.handlers {
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
	args := make([]any, 0, len(q.In)+len(q.Out)+len(q.InArgs)+len(q.OutArgs))
	if len(q.InArgs) > 0 {
		args = append(args, q.InArgs...)
	}
	if len(q.OutArgs) > 0 {
		for i := range q.OutArgs {
			args = append(args, sql.Out{Dest: q.OutArgs[i]})
		}
	}
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
func (q *Query[R]) Run(ctx context.Context, able Queryable) error {
	if q.Prepared && (q.stmt == nil || q.stmtQuery != q.Query) {
		stmt, err := able.PrepareContext(ctx, q.Query)
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
			rows, err = able.QueryContext(ctx, q.Query, q.GetArgs()...)
		}
		if err != nil {
			return err
		}
		defer rows.Close()

		if q.handlers == nil || q.columnsQuery != q.Query {
			cols, err := rows.ColumnTypes()
			if err != nil {
				return err
			}
			q.SetColumns(cols)
			q.columnsQuery = q.Query
		}

		if q.Results != nil {
			rowIndex := 0
			for rows.Next() {
				row := q.NewRow()
				err = rows.Scan(q.GetValues(row)...)
				if err != nil {
					return err
				}
				if !q.Results(*row, rowIndex) {
					break
				}
				rowIndex++
			}
		} else {
			if rows.Next() {
				row := q.NewRow()
				err = rows.Scan(q.GetValues(row)...)
				if err != nil {
					return err
				}
				*q.Result = *row
			} else {
				q.Result = nil
			}
		}
	} else {
		var result sql.Result
		var err error
		if q.stmt != nil {
			result, err = q.stmt.ExecContext(ctx, q.GetArgs()...)
		} else {
			result, err = able.ExecContext(ctx, q.Query, q.GetArgs()...)
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
