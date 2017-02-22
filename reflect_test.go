package jsonschema

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/url"
	"reflect"
	"testing"
	"time"
)

type GrandfatherType struct {
	FamilyName string `json:"family_name" jsonschema:"required"`
}

type SomeBaseType struct {
	SomeBaseProperty int `json:"some_base_property"`
	// The jsonschema required tag is nonsensical for private and ignored properties.
	// Their presence here tests that the fields *will not* be required in the output
	// schema, even if they are tagged required.
	somePrivateBaseProperty string          `json:"i_am_private" jsonschema:"required"`
	SomeIgnoredBaseProperty string          `json:"-" jsonschema:"required"`
	Grandfather             GrandfatherType `json:"grand"`

	SomeUntaggedBaseProperty           bool `jsonschema:"required"`
	someUnexportedUntaggedBaseProperty bool
}

type nonExported struct {
	PublicNonExported  int
	privateNonExported int
}

type ProtoEnum int32

func (ProtoEnum) EnumDescriptor() ([]byte, []int) { return []byte(nil), []int{0} }

const (
	Unset ProtoEnum = iota
	Great
)

type TestUser struct {
	SomeBaseType
	nonExported

	ID      int                    `json:"id" jsonschema:"required"`
	Name    string                 `json:"name" jsonschema:"required"`
	Friends []int                  `json:"friends,omitempty"`
	Tags    map[string]interface{} `json:"tags,omitempty"`

	TestFlag       bool
	IgnoredCounter int `json:"-"`

	// Tests for RFC draft-wright-json-schema-validation-00, section 7.3
	BirthDate time.Time `json:"birth_date,omitempty"`
	Website   url.URL   `json:"website,omitempty"`
	IPAddress net.IP    `json:"network_address,omitempty"`

	// Tests for RFC draft-wright-json-schema-hyperschema-00, section 4
	Photo []byte `json:"photo,omitempty" jsonschema:"required"`

	// Tests for jsonpb enum support
	Feeling ProtoEnum `json:"feeling,omitempty"`
}

var schemaGenerationTests = []struct {
	reflector *Reflector
	fixture   string
}{
	{&Reflector{}, "fixtures/defaults.json"},
	{&Reflector{AllowAdditionalProperties: true}, "fixtures/allow_additional_props.json"},
	{&Reflector{RequiredFromJSONSchemaTags: true}, "fixtures/required_from_jsontags.json"},
	{&Reflector{ExpandedStruct: true}, "fixtures/defaults_expanded_toplevel.json"},
}

func TestSchemaGeneration(t *testing.T) {
	for _, tt := range schemaGenerationTests {

		f, err := ioutil.ReadFile(tt.fixture)
		if err != nil {
			t.Errorf("ioutil.ReadAll(%s): %s", tt.fixture, err)
			continue
		}

		actualSchema := tt.reflector.Reflect(&TestUser{})
		expectedSchema := &Schema{}

		if err := json.Unmarshal(f, expectedSchema); err != nil {
			t.Errorf("json.Unmarshal(%s, %v): %s", tt.fixture, expectedSchema, err)
			continue
		}

		if !reflect.DeepEqual(actualSchema, expectedSchema) {
			actualJSON, err := json.MarshalIndent(actualSchema, "", "  ")
			if err != nil {
				t.Errorf("json.MarshalIndent(%v, \"\", \"  \"): %v", actualSchema, err)
				continue
			}
			t.Errorf("reflector %+v wanted schema %s, got %s", tt.reflector, f, actualJSON)
		}
	}
}
