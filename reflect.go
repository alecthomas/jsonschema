// Package jsonschema uses reflection to generate JSON Schemas from Go types [1].
//
// If json tags are present on struct fields, they will be used to infer
// property names and if a property is required (omitempty is present).
//
// [1] http://json-schema.org/latest/json-schema-validation.html
package jsonschema

import (
	"encoding/json"
	"net"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// Version is the JSON Schema version.
// If extending JSON Schema with custom values use a custom URI.
// RFC draft-wright-json-schema-00, section 6
var Version = "http://json-schema.org/draft-04/schema#"

// Schema is the root schema.
// RFC draft-wright-json-schema-00, section 4.5
type Schema struct {
	*Type
	Definitions Definitions `json:"definitions,omitempty"`
}

// Type represents a JSON Schema object type.
type Type struct {
	// RFC draft-wright-json-schema-00
	Version string `json:"$schema,omitempty"` // section 6.1
	Ref     string `json:"$ref,omitempty"`    // section 7
	// RFC draft-wright-json-schema-validation-00, section 5
	MultipleOf           int              `json:"multipleOf,omitempty"`           // section 5.1
	Maximum              int              `json:"maximum,omitempty"`              // section 5.2
	ExclusiveMaximum     bool             `json:"exclusiveMaximum,omitempty"`     // section 5.3
	Minimum              int              `json:"minimum,omitempty"`              // section 5.4
	ExclusiveMinimum     bool             `json:"exclusiveMinimum,omitempty"`     // section 5.5
	MaxLength            int              `json:"maxLength,omitempty"`            // section 5.6
	MinLength            int              `json:"minLength,omitempty"`            // section 5.7
	Pattern              string           `json:"pattern,omitempty"`              // section 5.8
	AdditionalItems      *Type            `json:"additionalItems,omitempty"`      // section 5.9
	Items                *Type            `json:"items,omitempty"`                // section 5.9
	MaxItems             int              `json:"maxItems,omitempty"`             // section 5.10
	MinItems             int              `json:"minItems,omitempty"`             // section 5.11
	UniqueItems          bool             `json:"uniqueItems,omitempty"`          // section 5.12
	MaxProperties        int              `json:"maxProperties,omitempty"`        // section 5.13
	MinProperties        int              `json:"minProperties,omitempty"`        // section 5.14
	Required             []string         `json:"required,omitempty"`             // section 5.15
	Properties           map[string]*Type `json:"properties,omitempty"`           // section 5.16
	PatternProperties    map[string]*Type `json:"patternProperties,omitempty"`    // section 5.17
	AdditionalProperties json.RawMessage  `json:"additionalProperties,omitempty"` // section 5.18
	Dependencies         map[string]*Type `json:"dependencies,omitempty"`         // section 5.19
	Enum                 []interface{}    `json:"enum,omitempty"`                 // section 5.20
	Type                 string           `json:"type,omitempty"`                 // section 5.21
	AllOf                map[string]*Type `json:"allOf,omitempty"`                // section 5.22
	AnyOf                map[string]*Type `json:"anyOf,omitempty"`                // section 5.23
	OneOf                map[string]*Type `json:"oneOf,omitempty"`                // section 5.24
	Not                  map[string]*Type `json:"not,omitempty"`                  // section 5.25
	Definitions          Definitions      `json:"definitions,omitempty"`          // section 5.26
	// RFC draft-wright-json-schema-validation-00, section 6, 7
	Title       string      `json:"title,omitempty"`       // section 6.1
	Description string      `json:"description,omitempty"` // section 6.1
	Default     interface{} `json:"default,omitempty"`     // section 6.2
	Format      string      `json:"format,omitempty"`      // section 7
	// RFC draft-wright-json-schema-hyperschema-00, section 4
	Media          *Type  `json:"media,omitempty"`          // section 4.3
	BinaryEncoding string `json:"binaryEncoding,omitempty"` // section 4.3
}

// Reflect reflects to Schema from a value.
func Reflect(v interface{}) *Schema {
	return ReflectFromType(reflect.TypeOf(v))
}

// ReflectFromType generates root schema
func ReflectFromType(t reflect.Type) *Schema {
	definitions := Definitions{}
	s := &Schema{
		Type:        reflectTypeToSchema(definitions, t),
		Definitions: definitions,
	}
	return s
}

// Definitions hold schema definitions.
// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.26
// RFC draft-wright-json-schema-validation-00, section 5.26
type Definitions map[string]*Type

// Avaialble Go defined types for JSON Schema Validation.
// RFC draft-wright-json-schema-validation-00, section 7.3
var (
	timeType = reflect.TypeOf(time.Time{}) // date-time RFC section 7.3.1
	ipType   = reflect.TypeOf(net.IP{})    // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	uriType  = reflect.TypeOf(url.URL{})   // uri RFC section 7.3.6
)

// Byte slices will be encoded as base64
var byteSliceType = reflect.TypeOf([]byte(nil))

func reflectTypeToSchema(definitions Definitions, t reflect.Type) *Type {
	// Already added to definitions?
	if _, ok := definitions[t.Name()]; ok {
		return &Type{Ref: "#/definitions/" + t.Name()}
	}

	// Defined format types for JSON Schema Validation
	// RFC draft-wright-json-schema-validation-00, section 7.3
	// TODO email RFC section 7.3.2, hostname RFC section 7.3.3, uriref RFC section 7.3.7
	switch t {
	case ipType:
		// TODO differentiate ipv4 and ipv6 RFC section 7.3.4, 7.3.5
		return &Type{Type: "string", Format: "ipv4"} // ipv4 RFC section 7.3.4
	}

	switch t.Kind() {
	case reflect.Struct:
		switch t {
		case timeType: // date-time RFC section 7.3.1
			return &Type{Type: "string", Format: "date-time"}
		case uriType: // uri RFC section 7.3.6
			return &Type{Type: "string", Format: "uri"}
		default:
			return reflectStruct(definitions, t)
		}

	case reflect.Map:
		rt := &Type{
			Type: "object",
			PatternProperties: map[string]*Type{
				".*": reflectTypeToSchema(definitions, t.Elem()),
			},
		}
		delete(rt.PatternProperties, "additionalProperties")
		return rt

	case reflect.Slice:
		switch t {
		case byteSliceType:
			return &Type{
				Type:  "string",
				Media: &Type{BinaryEncoding: "base64"},
			}
		default:
			return &Type{
				Type:  "array",
				Items: reflectTypeToSchema(definitions, t.Elem()),
			}
		}

	case reflect.Interface:
		return &Type{
			Type:                 "object",
			AdditionalProperties: []byte("true"),
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Type{Type: "integer"}

	case reflect.Float32, reflect.Float64:
		return &Type{Type: "number"}

	case reflect.Bool:
		return &Type{Type: "boolean"}

	case reflect.String:
		return &Type{Type: "string"}

	case reflect.Ptr:
		return reflectTypeToSchema(definitions, t.Elem())
	}
	panic("unsupported type " + t.String())
}

// Refects a struct to a JSON Schema type.
func reflectStruct(definitions Definitions, t reflect.Type) *Type {
	st := &Type{
		Type:                 "object",
		Properties:           map[string]*Type{},
		AdditionalProperties: []byte("false"),
	}
	definitions[t.Name()] = st
	reflectStructFields(st, definitions, t)

	return &Type{
		Version: Version,
		Ref:     "#/definitions/" + t.Name(),
	}
}

func reflectStructFields(st *Type, definitions Definitions, t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// anonymous and exported type should be processed recursively
		// current type should inherit properties of anonymous one
		if f.Anonymous && f.PkgPath == "" {
			reflectStructFields(st, definitions, f.Type)
			continue
		}

		name, required := reflectFieldName(f)
		if name == "" {
			continue
		}
		st.Properties[name] = reflectTypeToSchema(definitions, f.Type)
		if required {
			st.Required = append(st.Required, name)
		}
	}
}

func reflectFieldName(f reflect.StructField) (string, bool) {
	if f.PkgPath != "" { // unexported field, ignore it
		return "", false
	}
	parts := strings.Split(f.Tag.Get("json"), ",")
	if parts[0] == "-" {
		return "", false
	}

	name := f.Name
	required := true

	if parts[0] != "" {
		name = parts[0]
	}

	if len(parts) > 1 && parts[1] == "omitempty" {
		required = false
	}
	return name, required
}
