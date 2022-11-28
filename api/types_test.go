package api

import (
	"encoding/json"
	"testing"
)

func TestReference(t *testing.T) {
	path := Ref[Path]("Test")
	json, _ := json.Marshal(path)

	if string(json) != `{"$ref":"#/paths/Test"}` {
		t.Errorf("Ref serialization error, got: %s", json)
	}
}
