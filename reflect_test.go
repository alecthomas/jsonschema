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

	"github.com/go-test/deep"
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

	SomeUntaggedBaseProperty           bool `jsonschema:"required,example=true,example=OK,default=false"`
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

	// Tests for jsonpb enum support and a default value that is not valid
	Feeling ProtoEnum `json:"feeling,omitempty"`
	Age     int       `json:"age" jsonschema:"minimum=18,maximum=120,exclusiveMaximum=true,exclusiveMinimum=true,default=1or2,example=3,example=four"`
	Email   string    `json:"email" jsonschema:"format=email"`
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
		name := strings.TrimSuffix(filepath.Base(tt.fixture), ".json")
		t.Run(name, func(t *testing.T) {
			f, err := ioutil.ReadFile(tt.fixture)
			if err != nil {
				t.Errorf("ioutil.ReadAll(%s): %s", tt.fixture, err)
				return
			}

			actualSchema := tt.reflector.Reflect(&TestUser{})
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
				t.Errorf("unexpected diff: %+v\nreflector %+v wanted schema %s, got %s", deep.Equal(expectedSchema, actualSchema), tt.reflector, f, actualJSON)
			}
		})
	}
}
