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
	somePrivateBaseProperty   string          `json:"i_am_private" jsonschema:"required"`
	SomeIgnoredBaseProperty   string          `json:"-" jsonschema:"required"`
	SomeSchemaIgnoredProperty string          `jsonschema:"-,required"`
	Grandfather               GrandfatherType `json:"grand"`

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
	Name    string                 `json:"name" jsonschema:"required,minLength=1,maxLength=20"`
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
	Age     int       `json:"age" jsonschema:"minimum=18,maximum=120,exclusiveMaximum=true,exclusiveMinimum=true"`
	Email   string    `json:"email" jsonschema:"format=email"`
}

var schemaGenerationTests = []struct {
	reflector *Reflector
	structure interface{}
	fixture   string
}{
	{&Reflector{}, &TestUser{}, "defaults"},
	{&Reflector{AllowAdditionalProperties: true}, &TestUser{}, "allow_additional_props"},
	{&Reflector{RequiredFromJSONSchemaTags: true}, &TestUser{}, "required_from_jsontags"},
	{&Reflector{ExpandedStruct: true}, &TestUser{}, "defaults_expanded_toplevel"},
	{&Reflector{}, &Professional{}, "any_of"},
}

type Professional struct {
	developer *Developer
	qa        *QA
}

func (p Professional) AnyOf() []reflect.StructField {
	return AnyOfFieldsIn(p, "developer", "qa")
}

type Developer struct {
	Profession string `jsonschema:"pattern:^software developer$"`
	Languages  []string
}

type QA struct {
	Profession string `jsonschema:"pattern:^QA$"`
	Automation bool
}

func TestSchemaGeneration(t *testing.T) {
	for _, tt := range schemaGenerationTests {
		t.Run(tt.fixture, func(t *testing.T) {
			f, err := ioutil.ReadFile("fixtures/" + tt.fixture + ".json")
			if err != nil {
				t.Errorf("ioutil.ReadAll(%s): %s", tt.fixture, err)
				return
			}

			actualSchema := tt.reflector.Reflect(tt.structure)
			expectedSchema := &Schema{}

			if err := json.Unmarshal(f, expectedSchema); err != nil {
				t.Errorf("json.Unmarshal(%s, %v): %s", tt.fixture, expectedSchema, err)
				return
			}

			if !reflect.DeepEqual(actualSchema, expectedSchema) {
				actualJSON, err := json.MarshalIndent(actualSchema, "", "  ")
				if err != nil {
					t.Errorf("json.MarshalIndent(%v, \"\", \"  \"): %v", actualSchema, err)
					return
				}
				t.Errorf("reflector %+v wanted schema %s, got %s", tt.reflector, f, actualJSON)
			}
		})
	}
}
