package jsonschema

import "reflect"

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
		reflect.StructField{Type: reflect.TypeOf(Desktop{})},
	}
}
