package api

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert := assert.New(t)

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

	assert.Equal(string(doc), `{"openapi":"3.0.0","info":{"title":"Test","version":""},"paths":{"/tasks/{id}":{"description":"Actions for tasks","get":{"summary":"Get task","parameters":[{"name":"id","in":"path","required":true,"schema":{"$ref":"#/components/schemas/UUID"}}],"responses":{"200":{"description":"","content":{"application/json":{"schema":{"type":"object","required":["id","name","done"],"properties":{"done":{"type":"boolean"},"doneAt":{"oneOf":[{"$ref":"#/components/schemas/Time"},{"type":"null"}]},"id":{"$ref":"#/components/schemas/UUID"},"name":{"type":"string"}},"additionalProperties":false}}}}}}}},"components":{"schemas":{"TaskResult":{"type":"object","required":["id","name","done"],"properties":{"done":{"type":"boolean"},"doneAt":{"oneOf":[{"$ref":"#/components/schemas/Time"},{"type":"null"}]},"id":{"$ref":"#/components/schemas/UUID"},"name":{"type":"string"}},"additionalProperties":false},"TestA":{"type":"object","required":["Name","FavoriteNumbers"],"properties":{"Age":{"type":"integer","nullable":true},"FavoriteNumbers":{"type":"array","items":{"type":"number"}},"Name":{"type":"string"}},"additionalProperties":false},"TestB":{"type":"object","required":["id","IntMap","AMap"],"properties":{"AMap":{"type":"object","additionalProperties":{"oneOf":[{"$ref":"#/components/schemas/TestA"},{"type":"null"}]}},"IntMap":{"type":"object","additionalProperties":{"type":"integer"}},"bool":{"type":"boolean"},"id":{"type":"integer"}},"additionalProperties":false},"Time":{"type":"string","format":"date"},"UUID":{"type":"string","description":"A universally unique identifier","format":"uuid"}}}}`)
}

type TestHasName struct{}

func (t TestHasName) APIName() string { return "TestHasNameAlias" }

func TestGetName(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("String", GetName(""))
	assert.Equal("String", GetName(typeOf[string]()))

	assert.Equal("TestHasNameAlias", GetName(TestHasName{}))
	assert.Equal("TestHasNameAlias", GetName(typeOf[TestHasName]()))
}

type TestHasEnum struct{}

func (t TestHasEnum) APIEnum() []any { return []any{"A", "B"} }

func TestGetEnum(t *testing.T) {
	assert := assert.New(t)

	assert.Equal([]any(nil), GetEnums(""))

	assert.Equal([]any{"A", "B"}, GetEnums(TestHasEnum{}))
	assert.Equal([]any{"A", "B"}, GetEnums(typeOf[TestHasEnum]()))
}

func TestGetNameQualified(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("GithubComClickerMonkeyRezApiOperation", GetNameQualified(Operation{}))
	assert.Equal("String", GetNameQualified(""))
}
