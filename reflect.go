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
	"strconv"
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

// SchemaCondition holds data for if/then/else jsonschema statements
type SchemaCondition struct {
	If   reflect.StructField
	Then interface{}
	Else interface{}
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
	AllOf                []*Type          `json:"allOf,omitempty"`                // section 5.22
	AnyOf                []*Type          `json:"anyOf,omitempty"`                // section 5.23
	OneOf                []*Type          `json:"oneOf,omitempty"`                // section 5.24
	If                   *Type            `json:"if,omitempty"`
	Then                 *Type            `json:"then,omitempty"`
	Else                 *Type            `json:"else,omitempty"`
	Not                  *Type            `json:"not,omitempty"`         // section 5.25
	Definitions          Definitions      `json:"definitions,omitempty"` // section 5.26
	// RFC draft-wright-json-schema-validation-00, section 6, 7
	Title       string      `json:"title,omitempty"`       // section 6.1
	Description string      `json:"description,omitempty"` // section 6.1
	Default     interface{} `json:"default,omitempty"`     // section 6.2
	Format      string      `json:"format,omitempty"`      // section 7
	// RFC draft-wright-json-schema-hyperschema-00, section 4
	Media          *Type  `json:"media,omitempty"`          // section 4.3
	BinaryEncoding string `json:"binaryEncoding,omitempty"` // section 4.3
}

// Reflect reflects to Schema from a value using the default Reflector
func Reflect(v interface{}) *Schema {
	return ReflectFromType(reflect.TypeOf(v))
}

// ReflectFromType generates root schema using the default Reflector
func ReflectFromType(t reflect.Type) *Schema {
	r := &Reflector{}
	return r.ReflectFromType(t)
}

// A Reflector reflects values into a Schema.
type Reflector struct {
	// AllowAdditionalProperties will cause the Reflector to generate a schema
	// with additionalProperties to 'true' for all struct types. This means
	// the presence of additional keys in JSON objects will not cause validation
	// to fail. Note said additional keys will simply be dropped when the
	// validated JSON is unmarshaled.
	AllowAdditionalProperties bool

	// RequiredFromJSONSchemaTags will cause the Reflector to generate a schema
	// that requires any key tagged with `jsonschema:required`, overriding the
	// default of requiring any key *not* tagged with `json:,omitempty`.
	RequiredFromJSONSchemaTags bool

	// ExpandedStruct will cause the toplevel definitions of the schema not
	// be referenced itself to a definition.
	ExpandedStruct bool
}

// Reflect reflects to Schema from a value.
func (r *Reflector) Reflect(v interface{}) *Schema {
	return r.ReflectFromType(reflect.TypeOf(v))
}

// ReflectFromType generates root schema
func (r *Reflector) ReflectFromType(t reflect.Type) *Schema {
	definitions := Definitions{}
	if r.ExpandedStruct {
		st := &Type{
			Version:              Version,
			Type:                 "object",
			Properties:           map[string]*Type{},
			AdditionalProperties: []byte("false"),
		}
		if r.AllowAdditionalProperties {
			st.AdditionalProperties = []byte("true")
		}
		r.reflectStructFields(st, definitions, t)
		r.reflectStruct(definitions, t)
		delete(definitions, t.Name())
		return &Schema{Type: st, Definitions: definitions}
	}

	s := &Schema{
		Type:        r.reflectTypeToSchema(definitions, t),
		Definitions: definitions,
	}
	return s
}

// Definitions hold schema definitions.
// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.26
// RFC draft-wright-json-schema-validation-00, section 5.26
type Definitions map[string]*Type

// Available Go defined types for JSON Schema Validation.
// RFC draft-wright-json-schema-validation-00, section 7.3
var (
	timeType = reflect.TypeOf(time.Time{}) // date-time RFC section 7.3.1
	ipType   = reflect.TypeOf(net.IP{})    // ipv4 and ipv6 RFC section 7.3.4, 7.3.5
	uriType  = reflect.TypeOf(url.URL{})   // uri RFC section 7.3.6
)

// Byte slices will be encoded as base64
var byteSliceType = reflect.TypeOf([]byte(nil))

// Go code generated from protobuf enum types should fulfil this interface.
type protoEnum interface {
	EnumDescriptor() ([]byte, []int)
}

// Implement AndOneOf() when oneOf is used to factor out common parts of subschema
// {
//  "type": "number",
//  "oneOf": [
//    { "multipleOf": 5 },
//    { "multipleOf": 3 }
//  ]
//}
type andOneOf interface {
	AndOneOf() []reflect.StructField
}

// Implement OneOf() when oneOf is exclusive
// {
//  "oneOf": [
//    { "type": "number", "multipleOf": 5 },
//    { "type": "number", "multipleOf": 3 }
//  ]
// }
type oneOf interface {
	OneOf() []reflect.StructField
}

//Implement IfThenElse() when condition needs to be used
// {
//    "if": { "properties": { "power": { "minimum": 9000 } } },
//    "then": { "required": [ "disbelief" ] },
//    "else": { "required": [ "confidence" ] }
// }
type ifThenElse interface {
	IfThenElse() SchemaCondition
}

var protoEnumType = reflect.TypeOf((*protoEnum)(nil)).Elem()
var andOneOfType = reflect.TypeOf((*andOneOf)(nil)).Elem()
var oneOfType = reflect.TypeOf((*oneOf)(nil)).Elem()
var ifThenElseType = reflect.TypeOf((*ifThenElse)(nil)).Elem()

func (r *Reflector) reflectTypeToSchema(definitions Definitions, t reflect.Type) (schema *Type) {
	// Already added to definitions?
	if _, ok := definitions[t.Name()]; ok {
		return &Type{Ref: "#/definitions/" + t.Name()}
	}

	// jsonpb will marshal protobuf enum options as either strings or integers.
	// It will unmarshal either.
	if t.Implements(protoEnumType) {
		return &Type{OneOf: []*Type{
			{Type: "string"},
			{Type: "integer"},
		}}
	}

	// Return only oneOf array when OneOf() is implemented
	if t.Implements(oneOfType) {
		s := reflect.New(t).Interface().(oneOf).OneOf()
		return &Type{OneOf: r.getOneOfList(definitions, s)}
	}

	// Append oneOf array to existing non-object type when AndOneOf() is implemented
	defer func() {
		if t.Kind() != reflect.Struct {
			if t.Implements(andOneOfType) {
				s := reflect.New(t).Interface().(andOneOf).AndOneOf()
				schema.OneOf = r.getOneOfList(definitions, s)
			}
		}
	}()

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
			return r.reflectStruct(definitions, t)
		}

	case reflect.Map:
		rt := &Type{
			Type:              "object",
			PatternProperties: nil,
		}

		// map[...]interface{} should allow any child type. If another value type is specified,
		// It should be added to the object properties spec.
		if t.Elem().Kind() != reflect.Interface {
			rt.PatternProperties = map[string]*Type{
				".*": r.reflectTypeToSchema(definitions, t.Elem()),
			}
			delete(rt.PatternProperties, "additionalProperties")
		}

		return rt

	case reflect.Slice, reflect.Array:
		returnType := &Type{}
		if t.Kind() == reflect.Array {
			returnType.MinItems = t.Len()
			returnType.MaxItems = returnType.MinItems
		}
		switch t {
		case byteSliceType:
			returnType.Type = "string"
			returnType.Media = &Type{BinaryEncoding: "base64"}
			return returnType
		default:
			returnType.Type = "array"
			returnType.Items = r.reflectTypeToSchema(definitions, t.Elem())
			return returnType
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
		return r.reflectTypeToSchema(definitions, t.Elem())
	}
	panic("unsupported type " + t.String())
}

// Refects a struct to a JSON Schema type.
func (r *Reflector) reflectStruct(definitions Definitions, t reflect.Type) *Type {
	st := &Type{
		Type:                 "object",
		Properties:           map[string]*Type{},
		AdditionalProperties: []byte("false"),
	}
	if r.AllowAdditionalProperties {
		st.AdditionalProperties = []byte("true")
	}
	definitions[t.Name()] = st
	r.reflectStructFields(st, definitions, t)
	if t.Implements(ifThenElseType) {
		condition := reflect.New(t).Interface().(ifThenElse).IfThenElse()
		r.reflectCondition(definitions, condition, st)
	}

	return &Type{
		Version: Version,
		Ref:     "#/definitions/" + t.Name(),
	}
}

func (r *Reflector) reflectCondition(definitions Definitions, sc SchemaCondition, t *Type) {
	conditionSchema := Type{}
	conditionSchema.structKeywordsFromTags(sc.If)

	condition := &Type{
		Properties: map[string]*Type{
			sc.If.Tag.Get("json"): &conditionSchema,
		},
	}

	t.If = condition
	if reflect.TypeOf(sc.Then) != nil {
		t.Then = r.reflectTypeToSchema(definitions, reflect.TypeOf(sc.Then))
	}
	if reflect.TypeOf(sc.Else) != nil {
		t.Else = r.reflectTypeToSchema(definitions, reflect.TypeOf(sc.Else))
	}
}

func (r *Reflector) reflectStructFields(st *Type, definitions Definitions, t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// anonymous and exported type should be processed recursively
		// current type should inherit properties of anonymous one
		if f.Anonymous && f.PkgPath == "" {
			r.reflectStructFields(st, definitions, f.Type)
			continue
		}

		name, required := r.reflectFieldName(f)
		if name == "" {
			continue
		}
		property := r.reflectTypeToSchema(definitions, f.Type)
		property.structKeywordsFromTags(f)
		st.Properties[name] = property
		if required {
			st.Required = append(st.Required, name)
		}

		// Append oneOf array to existing object type when AndOneOf() is implemented
		if t.Implements(andOneOfType) {
			s := reflect.New(t).Interface().(andOneOf).AndOneOf()
			st.OneOf = r.getOneOfList(definitions, s)
		}
	}
}

func (t *Type) structKeywordsFromTags(f reflect.StructField) {
	tags := strings.Split(f.Tag.Get("jsonschema"), ",")
	switch t.Type {
	case "string":
		t.stringKeywords(tags)
	case "number":
		t.floatKeywords(tags)
	case "integer":
		t.numbericKeywords(tags)
	case "array":
		t.arrayKeywords(tags)
	case "":
		t.stringKeywords(tags)
	}
}

// read struct tags for string type keywords
func (t *Type) stringKeywords(tags []string) {
	for _, tag := range tags {
		nameValue := strings.Split(tag, "=")
		if len(nameValue) == 2 {
			name, val := nameValue[0], nameValue[1]
			switch name {
			case "minLength":
				i, _ := strconv.Atoi(val)
				t.MinLength = i
			case "maxLength":
				i, _ := strconv.Atoi(val)
				t.MaxLength = i
			case "enum":
				enum := strings.Split(val, "|")
				s := make([]interface{}, len(enum))
				for k, v := range enum {
					s[k] = v
				}

				t.Enum = s
			case "format":
				switch val {
				case "date-time", "email", "hostname", "ipv4", "ipv6", "uri":
					t.Format = val
					break
				}
			case "pattern":
				t.Pattern = val
			}
		} else {
			name := nameValue[0]
			switch name {
			case "notEmpty":
				t.Pattern = "^\\S"
			case "allowNull":
				t.OneOf = []*Type{
					{Type: t.Type},
					{Type: "null"},
				}
				t.Type = ""
			}
		}
	}
}

// read struct tags for numberic type keywords
func (t *Type) numbericKeywords(tags []string) {
	for _, tag := range tags {
		nameValue := strings.Split(tag, "=")
		if len(nameValue) == 2 {
			name, val := nameValue[0], nameValue[1]
			switch name {
			case "multipleOf":
				i, _ := strconv.Atoi(val)
				t.MultipleOf = i
			case "minimum":
				i, _ := strconv.Atoi(val)
				t.Minimum = i
			case "maximum":
				i, _ := strconv.Atoi(val)
				t.Maximum = i
			case "exclusiveMaximum":
				b, _ := strconv.ParseBool(val)
				t.ExclusiveMaximum = b
			case "exclusiveMinimum":
				b, _ := strconv.ParseBool(val)
				t.ExclusiveMinimum = b
			case "enum":
				enum := strings.Split(val, "|")
				s := make([]interface{}, len(enum))
				for k, v := range enum {
					s[k], _ = strconv.Atoi(v)
				}
				t.Enum = s
			}
		} else {
			name := nameValue[0]
			switch name {
			case "allowNull":
				t.OneOf = []*Type{
					{Type: t.Type},
					{Type: "null"},
				}
				t.Type = ""
			}
		}
	}
}

// read struct tags for float type keywords
func (t *Type) floatKeywords(tags []string) {
	for _, tag := range tags {
		nameValue := strings.Split(tag, "=")
		if len(nameValue) == 2 {
			name, val := nameValue[0], nameValue[1]
			switch name {
			case "enum":
				enum := strings.Split(val, "|")
				s := make([]interface{}, len(enum))
				for k, v := range enum {
					s[k], _ = strconv.ParseFloat(v, 64)
				}
				t.Enum = s
			}
		} else {
			name := nameValue[0]
			switch name {
			case "allowNull":
				t.OneOf = []*Type{
					{Type: t.Type},
					{Type: "null"},
				}
				t.Type = ""
			}
		}
	}
}

// read struct tags for object type keyworks
// func (t *Type) objectKeywords(tags []string) {
//     for _, tag := range tags{
//         nameValue := strings.Split(tag, "=")
//         name, val := nameValue[0], nameValue[1]
//         switch name{
//             case "dependencies":
//                 t.Dependencies = val
//                 break;
//             case "patternProperties":
//                 t.PatternProperties = val
//                 break;
//         }
//     }
// }

// read struct tags for array type keyworks
func (t *Type) arrayKeywords(tags []string) {
	for _, tag := range tags {
		nameValue := strings.Split(tag, "=")
		if len(nameValue) == 2 {
			name, val := nameValue[0], nameValue[1]
			switch name {
			case "minItems":
				i, _ := strconv.Atoi(val)
				t.MinItems = i
			case "maxItems":
				i, _ := strconv.Atoi(val)
				t.MaxItems = i
			case "uniqueItems":
				t.UniqueItems = true
			}
		}
	}
}

func requiredFromJSONTags(tags []string) bool {
	if ignoredByJSONTags(tags) {
		return false
	}

	for _, tag := range tags[1:] {
		if tag == "omitempty" {
			return false
		}
	}
	return true
}

func requiredFromJSONSchemaTags(tags []string) bool {
	if ignoredByJSONSchemaTags(tags) {
		return false
	}
	for _, tag := range tags {
		if tag == "required" {
			return true
		}
	}
	return false
}

func ignoredByJSONTags(tags []string) bool {
	return tags[0] == "-"
}

func ignoredByJSONSchemaTags(tags []string) bool {
	return tags[0] == "-"
}

func (r *Reflector) reflectFieldName(f reflect.StructField) (string, bool) {
	if f.PkgPath != "" { // unexported field, ignore it
		return "", false
	}

	jsonTags := strings.Split(f.Tag.Get("json"), ",")

	if ignoredByJSONTags(jsonTags) {
		return "", false
	}

	jsonSchemaTags := strings.Split(f.Tag.Get("jsonschema"), ",")
	if ignoredByJSONSchemaTags(jsonSchemaTags) {
		return "", false
	}

	name := f.Name
	required := requiredFromJSONTags(jsonTags)

	if r.RequiredFromJSONSchemaTags {
		required = requiredFromJSONSchemaTags(jsonSchemaTags)
	}

	if jsonTags[0] != "" {
		name = jsonTags[0]
	}

	return name, required
}

func (r *Reflector) getOneOfList(definitions Definitions, s []reflect.StructField) []*Type {
	oneOfList := make([]*Type, 0)
	for _, oneType := range s {
		if oneType.Type == nil {
			oneOfList = append(oneOfList, &Type{Type: "null"})
		} else {
			oneOfList = append(oneOfList, r.reflectTypeToSchema(definitions, oneType.Type))
		}
	}

	return oneOfList
}
