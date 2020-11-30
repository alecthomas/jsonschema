package test

import (
	"encoding/json"
	//"fmt"
	"github.com/alecthomas/jsonschema"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path"
	"testing"
)

type Object struct {
	Embedded `json:",some_tag,does_not_matter"`
}

type Embedded struct {
	F1 string `json:"f1"`
	F2 string `json:"f2"`
}

func TestMissingEmbeddedFieldIssue65(t *testing.T) {

	expected, err := ioutil.ReadFile(path.Join("test_data", "expected.yaml"))
	assert.NoError(t, err)


	reflector := jsonschema.Reflector{

	}
	schema := reflector.Reflect(&Object{})
	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}
	assert.Equal(t, string(expected), string(jsonBytes))
}
