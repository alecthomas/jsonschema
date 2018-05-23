package jsonschema

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

type testSet struct {
	reflector *Reflector
	fixture   string
	actual    interface{}
}

var schemaGenerationTests = []testSet{
	{&Reflector{}, "fixtures/defaults.json", TestUser{}},
	{&Reflector{AllowAdditionalProperties: true}, "fixtures/allow_additional_props.json", TestUser{}},
	{&Reflector{RequiredFromJSONSchemaTags: true}, "fixtures/required_from_jsontags.json", TestUser{}},
	{&Reflector{ExpandedStruct: true}, "fixtures/defaults_expanded_toplevel.json", TestUser{}},
	{&Reflector{}, "fixtures/test_one_of_default.json", TestUserOneOf{}},
	{&Reflector{}, "fixtures/test_package.json", TestUserPackage{}},
	{&Reflector{}, "fixtures/if_then_else.json", Application{}},
}

func TestSchemaGeneration(t *testing.T) {
	for _, tt := range schemaGenerationTests {
		runTests(t, tt)
	}
}
func TestOverrides(t *testing.T) {
	override := GetSchemaTagOverride()
	override.Set(Hardware{}, "Brand", "enum=microsoft|apple|lenovo|dell")

	test := testSet{
		reflector: &Reflector{Overrides: override},
		fixture:   "fixtures/override_jsonschema_tag.json",
		actual:    TestUserOneOf{},
	}

	runTests(t, test)
}

func runTests(t *testing.T, tt testSet) {
	name := strings.TrimSuffix(filepath.Base(tt.fixture), ".json")
	t.Run(name, func(t *testing.T) {
		f, err := ioutil.ReadFile(tt.fixture)
		if err != nil {
			t.Errorf("ioutil.ReadAll(%s): %s", tt.fixture, err)
			return
		}

		actualSchema := tt.reflector.Reflect(tt.actual)

		actualJSON, err := json.Marshal(actualSchema)
		if err != nil {
			t.Errorf("json.MarshalIndent(%v, \"\", \"  \"): %v", actualJSON, err)
			return
		}
		actualJSON = sanitizeExpectedJson(actualJSON)
		cleanExpectedJSON := sanitizeExpectedJson(f)

		if !bytes.Equal(cleanExpectedJSON, actualJSON) {

			t.Errorf("reflector %+v wanted schema %s, got %s", tt.reflector, cleanExpectedJSON, actualJSON)
		}
	})
}

func sanitizeExpectedJson(expectedJSON []byte) []byte {
	var js interface{}
	json.Unmarshal(expectedJSON, &js)
	clean, _ := json.Marshal(js)
	return clean
}
