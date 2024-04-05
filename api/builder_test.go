package api

import (
	"encoding/json"
	"fmt"
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

type HasMessage interface {
	Message() string
}

type FriendRequest struct {
	Friend string `json:"friend"`
	Note   string `json:"note,omitempty"`
}

func (ft FriendRequest) Message() string {
	return fmt.Sprintf("%s requests to be friends: %s", ft.Friend, ft.Note)
}

type PointIncrease struct {
	Points int `json:"points"`
}

func (pi PointIncrease) Message() string {
	return fmt.Sprintf("%d points were added to your account!", pi.Points)
}

type EmbeddedPointer struct {
	Okay string `json:"okay"`
}

type Embedded struct {
	Fine string `json:"fine"`
}

type MessageList struct {
	*EmbeddedPointer
	Embedded
	Messages []HasMessage `json:"messages,omitempty"`
}

type MessageListParams struct {
	ID UUID `json:"id"`
}

func TestComplexScenarios(t *testing.T) {
	b := NewBuilder()

	SetFullSchema[HasMessage](b, &Schema{
		OneOf: []Schema{
			*SchemaRef("FriendRequest"),
			*SchemaRef("PointIncrease"),
		},
	})
	AddSchema[FriendRequest](b)
	AddSchema[PointIncrease](b)
	AddSchema[HasMessage](b)
	AddSchema[MessageList](b)

	path := Path{}
	path.Description = "Get messages"
	path.Get = &Operation{
		Summary: "Get messages by ID",
		Responses: Responses{
			"200": &Response{
				Content: Contents{
					ContentTypeJSON: &MediaType{
						Schema: SchemaRef("MessageList"),
					},
				},
			},
		},
	}
	path.Get.AddParameters(b, ParameterInPath, reflect.TypeOf(MessageListParams{}))
	b.AddPath("/messages/{id}", &path)

	b.Document = Document{
		OpenAPI: "3.0.0",
		Info:    Info{Title: "Test"},
	}

	doc, _ := json.Marshal(b.Build())

	assert.Equal(t, string(doc), `{"openapi":"3.0.0","info":{"title":"Test","version":""},"paths":{"/messages/{id}":{"description":"Get messages","get":{"summary":"Get messages by ID","parameters":[{"name":"id","in":"path","required":true,"schema":{"$ref":"#/components/schemas/UUID"}}],"responses":{"200":{"description":"","content":{"application/json":{"schema":{"$ref":"#/components/schemas/MessageList"}}}}}}}},"components":{"schemas":{"FriendRequest":{"type":"object","required":["friend"],"properties":{"friend":{"type":"string"},"note":{"type":"string"}},"additionalProperties":false},"HasMessage":{"oneOf":[{"$ref":"#/components/schemas/FriendRequest"},{"$ref":"#/components/schemas/PointIncrease"}]},"MessageList":{"type":"object","required":["okay","fine"],"properties":{"fine":{"type":"string"},"messages":{"type":"array","items":{"$ref":"#/components/schemas/HasMessage"}},"okay":{"type":"string"}},"additionalProperties":false},"PointIncrease":{"type":"object","required":["points"],"properties":{"points":{"type":"integer"}},"additionalProperties":false},"UUID":{"type":"string","description":"A universally unique identifier","format":"uuid"}}}}`)
}
