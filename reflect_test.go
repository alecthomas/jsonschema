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
	"bytes"
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

	SecretNumber int    `json:"secret_number,omitempty" jsonschema:"enum=9|30|28|52"`
	Sex          string `json:"sex,omitempty" jsonschema:"enum=male|female|neither|whatever|other|not applicable"`
	SecretFloatNumber float64  `json:"secret_float_number,omitempty" jsonschema:"enum=9.1|30.2|28.4|52.9"`
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

			actualJSON, err := json.Marshal(actualSchema)
			if err != nil {
				t.Errorf("json.MarshalIndent(%v, \"\", \"  \"): %v", actualJSON, err)
				return
			}
			actualJSON = sanitizeExpectedJson(actualJSON)
			cleanExpectedJSON := sanitizeExpectedJson(f)

			if !bytes.Equal(cleanExpectedJSON,actualJSON) {

				t.Errorf("reflector %+v wanted schema %s, got %s", tt.reflector, cleanExpectedJSON, actualJSON)
			}
		})
	}
}

func sanitizeExpectedJson(expectedJSON []byte) []byte {
	var js interface{}
	json.Unmarshal(expectedJSON, &js)
	clean, _ := json.Marshal(js)
	return clean
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
	Experience     StringOrNull `json:"experience" jsonschema:"minLength=1"`
	Language       StringOrNull `json:"language" jsonschema:"required,pattern=\\S+"`
	HardwareChoice Hardware     `json:"hardware"`
}

type StringOrNull struct {
	String string
	IsNull bool
}

type Hardware struct {
	Brand  string `json:"brand" jsonschema:"required,notEmpty"`
	Memory int    `json:"memory" jsonschema:"required"`
}

type Laptop struct {
	Brand           string `json:"brand" jsonschema:"pattern=^(apple|lenovo|dell)$"`
	NeedTouchScreen bool   `json:"need_touchscreen"`
}

type Desktop struct {
	FormFactor   string `json:"form_factor" jsonschema:"pattern=^(standard|micro|mini|nano)"`
	NeedKeyboard bool   `json:"need_keyboard"`
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

type Application struct {
	Type string `json:"type" jsonschema:"required"`
}

type ApplicationValidation struct {
	Type string `json:"type" jsonschema:"enum=web"`
}

type WebApp struct {
	Browser string `json:"browser"`
}

type MobileApp struct {
	Device string `json:"device"`
}

func (app Application) IfThenElse() SchemaCondition {
	conditionField, _ := reflect.TypeOf(ApplicationValidation{}).FieldByName("Type")
	return SchemaCondition{
		If: conditionField,
		Then: WebApp{},
		Else: MobileApp{},
	}
}

var ifThenElseSchemaGenerationTests = []struct {
	reflector *Reflector
	fixture   string
}{
	{&Reflector{}, "fixtures/if_then_else.json"},
}

func TestIfThenElseSchemaGeneration(t *testing.T) {
	for _, tt := range ifThenElseSchemaGenerationTests {
		name := strings.TrimSuffix(filepath.Base(tt.fixture), ".json")
		t.Run(name, func(t *testing.T) {
			f, err := ioutil.ReadFile(tt.fixture)
			if err != nil {
				t.Errorf("ioutil.ReadAll(%s): %s", tt.fixture, err)
				return
			}

			actualSchema := tt.reflector.Reflect(Application{})
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
