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

	ID       int                    `json:"id" jsonschema:"required"`
	Name     string                 `json:"name" jsonschema:"required,minLength=1,maxLength=20"`
	Nickname *string                `json:"nickname" jsonschema:"required,allowNull"`
	Friends  []int                  `json:"friends,omitempty"`
	Tags     map[string]interface{} `json:"tags,omitempty"`
	Keywords map[string]string      `json:"keywords,omitempty"`

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

	SecretNumber int  `json:"secret_number,omitempty" jsonschema:"enum=9|30|28|52"`
	Sex		string    `json:"sex,omitempty" jsonschema:"enum=male|female|neither|whatever|other|not applicable"`
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

			expectedSchema.Type = reconcileTypes(expectedSchema.Type)
			expectedSchema.Definitions = reconcileEnumTypes(expectedSchema.Definitions)

			if !reflect.DeepEqual(actualSchema, expectedSchema) {
				actualJSON, err := json.Marshal(actualSchema)
				if err != nil {
					t.Errorf("json.MarshalIndent(%v, \"\", \"  \"): %v", actualSchema, err)
					return
				}

				expectedJSON, err := json.Marshal(expectedSchema)
				if err != nil {
					t.Errorf("json.MarshalIndent(%v, \"\", \"  \"): %v", expectedJSON, err)
					return
				}

				t.Errorf("reflector %+v wanted schema %s, got %s", tt.reflector, expectedJSON, actualJSON)
			}
		})
	}
}

// The marshaling of json into interface{} results in a mismatch when we DeepEqual the expected/actual schemas for enum
// This coerces all float64 enums back to int
func reconcileTypes(t *Type) *Type {
	t.Definitions = reconcileEnumTypes(t.Definitions)
	t.Properties = convertEnum(t.Properties)

	return t
}

func reconcileEnumTypes(definitions Definitions) Definitions {
	mismatched := convertDefinitions(definitions)
	converted :=  convertEnum(mismatched)

	return Definitions(converted)
}

func convertDefinitions(definitions Definitions) map[string]*Type {
	return map[string]*Type(definitions)
}

func convertEnum(definitions map[string]*Type) map[string]*Type {
	for _, v := range definitions {
		if len(v.Definitions) > 0 {
			d := convertDefinitions(v.Definitions)
			v.Definitions = convertEnum(d)
		}

		if len(v.Properties) > 0 {
			v.Properties = convertEnum(v.Properties)
		}

		if len(v.Enum) > 0 && (reflect.TypeOf(v.Enum[0]).Kind() == reflect.Float64) {
			for idx, val := range v.Enum {
				v.Enum[idx] = int(val.(float64))
			}
		}
	}

	return definitions
}

type TestUserOneOf struct {
	Tester    Tester    `json:"tester" jsonschema:"required"`
	Developer Developer `json:"developer" jsonschema:"required"`
}

func (user TestUserOneOf) OneOf() []reflect.StructField {
	tester, _ := reflect.TypeOf(user).FieldByName("Tester")
	developer, _ := reflect.TypeOf(user).FieldByName("Developer")
	return []reflect.StructField{
		tester,
		developer,
	}
}

// Tester  struct
type Tester struct {
	Experience StringOrNull `json:"experience"`
}

// Developer  struct
type Developer struct {
	Experience StringOrNull `json:"experience" jsonschema:"minLength=1"`
	Language   StringOrNull `json:"language" jsonschema:"required,pattern=\\S+"`
	HardwareChoice Hardware  `json:"hardware"`
}

type StringOrNull struct {
	String string
	IsNull bool
}

type Hardware struct {
	Brand string `json:"brand" jsonschema:"required,notEmpty"`
	Memory int `json:"memory" jsonschema:"required"`
}

type Laptop struct {
	Brand string `json:"brand" jsonschema:"pattern=^(apple|lenovo|dell)$"`
	NeedTouchScreen bool `json:"need_touchscreen"`
}

type Desktop struct {
	FormFactor string `json:"form_factor" jsonschema:"pattern=^(standard|micro|mini|nano)"`
	NeedKeyboard bool `json:"need_keyboard"`
}

func (p StringOrNull) OneOf() []reflect.StructField {
	strings, _ := reflect.TypeOf(p).FieldByName("String")
	return []reflect.StructField{
		strings,
		reflect.StructField{Type: nil},
	}
}

func (h Hardware) AndOneOf() []reflect.StructField {
	return []reflect.StructField{
		reflect.StructField{Type: reflect.TypeOf(Laptop{})},
		reflect.StructField{Type: reflect.TypeOf(Hardware{})},
	}
}

var oneOfSchemaGenerationTests = []struct {
	reflector *Reflector
	fixture   string
}{
	{&Reflector{}, "fixtures/test_one_of_default.json"},
}
func TestOneOfSchemaGeneration(t *testing.T) {
	for _, tt := range oneOfSchemaGenerationTests {
		name := strings.TrimSuffix(filepath.Base(tt.fixture), ".json")
		t.Run(name, func(t *testing.T) {
			f, err := ioutil.ReadFile(tt.fixture)
			if err != nil {
				t.Errorf("ioutil.ReadAll(%s): %s", tt.fixture, err)
				return
			}

			actualSchema := tt.reflector.Reflect(TestUserOneOf{})
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

