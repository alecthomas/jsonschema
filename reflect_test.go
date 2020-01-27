package jsonschema

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type GrandfatherType struct {
	FamilyName string `json:"family_name" jsonschema:"required"`
}

type SomeBaseType struct {
	SomeBaseProperty     int `json:"some_base_property"`
	SomeBasePropertyYaml int `yaml:"some_base_property_yaml"`
	// The jsonschema required tag is nonsensical for private and ignored properties.
	// Their presence here tests that the fields *will not* be required in the output
	// schema, even if they are tagged required.
	somePrivateBaseProperty   string          `json:"i_am_private" jsonschema:"required"`
	SomeIgnoredBaseProperty   string          `json:"-" jsonschema:"required"`
	SomeSchemaIgnoredProperty string          `jsonschema:"-,required"`
	Grandfather               GrandfatherType `json:"grand"`

	SomeUntaggedBaseProperty           bool `jsonschema:"required"`
	someUnexportedUntaggedBaseProperty bool
}

type MapType map[string]interface{}

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
	MapType

	ID      int                    `json:"id" jsonschema:"required"`
	Name    string                 `json:"name" jsonschema:"required,minLength=1,maxLength=20,pattern=.*,description=this is a property,title=the name,example=joe,example=lucy,default=alex"`
	Friends []int                  `json:"friends,omitempty" jsonschema_description:"list of IDs, omitted when empty"`
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
	Age     int       `json:"age" jsonschema:"minimum=18,maximum=120,exclusiveMaximum=true,exclusiveMinimum=true"`
	Email   string    `json:"email" jsonschema:"format=email"`

	// Test for "extras" support
	Baz string `jsonschema_extras:"foo=bar,hello=world"`
}

type CustomTime time.Time

type CustomTypeField struct {
	CreatedAt CustomTime
}

type RootOneOf struct {
	Field1 string      `json:"field1" jsonschema:"oneof_required=group1"`
	Field2 string      `json:"field2" jsonschema:"oneof_required=group2"`
	Field3 interface{} `json:"field3" jsonschema:"oneof_type=string;array"`
	Field4 string      `json:"field4" jsonschema:"oneof_required=group1"`
	Field5 ChildOneOf  `json:"child"`
}

type ChildOneOf struct {
	Child1 string      `json:"child1" jsonschema:"oneof_required=group1"`
	Child2 string      `json:"child2" jsonschema:"oneof_required=group2"`
	Child3 interface{} `json:"child3" jsonschema:"oneof_required=group2,oneof_type=string;array"`
	Child4 string      `json:"child4" jsonschema:"oneof_required=group1"`
}

func TestSchemaGeneration(t *testing.T) {
	tests := []struct {
		typ       interface{}
		reflector *Reflector
		fixture   string
	}{
		{&RootOneOf{}, &Reflector{RequiredFromJSONSchemaTags: true}, "fixtures/oneof.json"},
		{&TestUser{}, &Reflector{}, "fixtures/defaults.json"},
		{&TestUser{}, &Reflector{AllowAdditionalProperties: true}, "fixtures/allow_additional_props.json"},
		{&TestUser{}, &Reflector{RequiredFromJSONSchemaTags: true}, "fixtures/required_from_jsontags.json"},
		{&TestUser{}, &Reflector{ExpandedStruct: true}, "fixtures/defaults_expanded_toplevel.json"},
		{&TestUser{}, &Reflector{IgnoredTypes: []interface{}{GrandfatherType{}}}, "fixtures/ignore_type.json"},
		{&CustomTypeField{}, &Reflector{
			TypeMapper: func(i reflect.Type) *Type {
				if i == reflect.TypeOf(CustomTime{}) {
					return &Type{
						Type:   "string",
						Format: "date-time",
					}
				}
				return nil
			},
		}, "fixtures/custom_type.json"},
	}

	for _, tt := range tests {
		name := strings.TrimSuffix(filepath.Base(tt.fixture), ".json")
		t.Run(name, func(t *testing.T) {
			f, err := ioutil.ReadFile(tt.fixture)
			require.NoError(t, err)

			actualSchema := tt.reflector.Reflect(tt.typ)
			expectedSchema := &Schema{}

			err = json.Unmarshal(f, expectedSchema)
			require.NoError(t, err)

			expectedJSON, _ := json.MarshalIndent(expectedSchema, "", "  ")
			actualJSON, _ := json.MarshalIndent(actualSchema, "", "  ")
			require.Equal(t, string(expectedJSON), string(actualJSON))
		})
	}
}

func TestBaselineUnmarshal(t *testing.T) {
	expectedJSON, err := ioutil.ReadFile("fixtures/defaults.json")
	require.NoError(t, err)

	reflector := &Reflector{}
	actualSchema := reflector.Reflect(&TestUser{})

	actualJSON, _ := json.MarshalIndent(actualSchema, "", "  ")

	require.Equal(t, strings.Replace(string(expectedJSON), `\/`, "/", -1), string(actualJSON))
}
