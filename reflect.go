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

var (
	timeType = reflect.TypeOf(time.Time{})
	ipType   = reflect.TypeOf(net.IP{})
	urlType  = reflect.TypeOf(url.URL{})
)

type Type struct {
	Type                 string           `json:"type,omitempty"`
	Format               string           `json:"format,omitempty"`
	Items                *Type            `json:"items,omitempty"`
	Properties           map[string]*Type `json:"properties,omitempty"`
	PatternProperties    map[string]*Type `json:"patternProperties,omitempty"`
	AdditionalProperties json.RawMessage  `json:"additionalProperties,omitempty"`
	Ref                  string           `json:"$ref,omitempty"`
	Required             []string         `json:"required,omitempty"`
	MaxLength            int              `json:"maxLength,omitempty"`
	MinLength            int              `json:"minLength,omitempty"`
	Pattern              string           `json:"pattern,omitempty"`
	Enum                 []interface{}    `json:"enum,omitempty"`
	Default              interface{}      `json:"default,omitempty"`
	Title                string           `json:"title,omitempty"`
	Description          string           `json:"description,omitempty"`
}

type Schema struct {
	*Type
	Definitions Definitions `json:"definitions,omitempty"`
}

// Reflect a Schema from a value.
func Reflect(v interface{}) *Schema {
	return ReflectFromType(reflect.TypeOf(v))
}

func ReflectFromType(t reflect.Type) *Schema {
	definitions := Definitions{}
	s := &Schema{
		Type:        reflectTypeToSchema(definitions, t),
		Definitions: definitions,
	}
	return s
}

type Definitions map[string]*Type

func reflectTypeToSchema(definitions Definitions, t reflect.Type) *Type {
	if _, ok := definitions[t.Name()]; ok {
		return &Type{Ref: "#/definitions/" + t.Name()}
	}

	switch t.Kind() {
	case reflect.Struct:
		switch t {
		case timeType:
			return &Type{Type: "string", Format: "date-time"}

		case ipType:
			return &Type{Type: "string", Format: "ipv4"}

		case urlType:
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
		return &Type{
			Type:  "array",
			Items: reflectTypeToSchema(definitions, t.Elem()),
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

func reflectStruct(definitions Definitions, t reflect.Type) *Type {
	st := &Type{
		Type:                 "object",
		Properties:           map[string]*Type{},
		AdditionalProperties: []byte("false"),
	}
	definitions[t.Name()] = st
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, required := reflectFieldName(f)
		st.Properties[name] = reflectTypeToSchema(definitions, f.Type)
		if required {
			st.Required = append(st.Required, name)
		}
	}
	return &Type{Ref: "#/definitions/" + t.Name()}
}

func reflectFieldName(f reflect.StructField) (string, bool) {
	name := f.Name
	required := true
	parts := strings.Split(f.Tag.Get("json"), ",")
	if parts[0] != "" {
		name = parts[0]
	}
	if len(parts) > 1 && parts[1] == "omitempty" {
		required = false
	}
	return name, required
}
