package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type UUID string

func (u UUID) APITypeName() string {
	return "UUID"
}
func (u UUID) APIFullSchema() *Schema {
	return &Schema{
		Type:        DataTypeString,
		Description: "A universally unique identifier",
		Format:      "uuid",
	}
}

func TestBuild(t *testing.T) {
	b := NewBuilder()
	b.NullableIsOptional = true

	type TestA struct {
		Name            string
		Age             *int
		FavoriteNumbers []float32
	}
	type TestB struct {
		ID     int  `json:"id"`
		Bool   bool `json:"bool,omitempty"`
		IntMap map[string]int
		AMap   map[int]*TestA
	}

	values := []any{
		int(0),
		TestB{},
		TestA{},
	}

	for _, val := range values {
		b.AddSchema(reflect.TypeOf(val))
	}

	type TaskParams struct {
		ID UUID `json:"id"`
	}
	type TaskResult struct {
		ID     UUID       `json:"id"`
		Name   string     `json:"name"`
		Done   bool       `json:"done"`
		DoneAt *time.Time `json:"doneAt,omitempty"`
	}

	b.FullSchema[reflect.TypeOf(time.Time{})] = &Schema{
		Type:   DataTypeString,
		Format: "date",
	}

	path := Path{}
	path.Description = "Actions for tasks"
	path.Get = &Operation{
		Summary: "Get task",
		Responses: Responses{
			"200": &Response{
				Content: Contents{
					ContentTypeJSON: &MediaType{
						Schema: b.GetSchema(reflect.TypeOf(TaskResult{})),
					},
				},
			},
		},
	}
	path.Get.AddParameters(b, ParameterInPath, reflect.TypeOf(TaskParams{}))
	b.AddPath("/tasks/{id}", &path)

	b.Document = Document{
		OpenAPI: "3.0.0",
		Info:    Info{Title: "Test"},
	}
	doc, _ := json.Marshal(b.Build())
	fmt.Println(string(doc))
}
