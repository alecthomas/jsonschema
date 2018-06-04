package jsonschema

import (
	"reflect"

	"github.com/discovery-digital/jsonschema/v1"
	"github.com/discovery-digital/jsonschema/v2"
)

type TestUserPackage struct {
	Tester    TesterPackage    `json:"tester" jsonschema:"required"`
	Developer DeveloperPackage `json:"developer" jsonschema:"required"`
}

func (user TestUserPackage) OneOf() []reflect.StructField {
	tester, _ := reflect.TypeOf(user).FieldByName("Tester")
	developer, _ := reflect.TypeOf(user).FieldByName("Developer")
	return []reflect.StructField{
		tester,
		developer,
	}
}

// Tester  struct
type TesterPackage struct {
	Experience StringOrNull1 `json:"experience"`
}

// Developer  struct
type DeveloperPackage struct {
	Experience     StringOrNull1 `json:"experience" jsonschema:"minLength=1"`
	Language       StringOrNull1 `json:"language" jsonschema:"required,pattern=\\S+"`
	HardwareChoice v1.Hardware   `json:"hardware"`
	HardwareChoic  v2.Hardware   `json:"hardware1"`
}

type StringOrNull1 struct {
	String string
	IsNull bool
}

type DesktopPackage struct {
	FormFactor   string `json:"form_factor" jsonschema:"pattern=^(standard|micro|mini|nano)"`
	NeedKeyboard bool   `json:"need_keyboard"`
}

func (p StringOrNull1) OneOf() []reflect.StructField {
	strings, _ := reflect.TypeOf(p).FieldByName("String")
	return []reflect.StructField{
		strings,
		reflect.StructField{Type: nil},
	}
}
