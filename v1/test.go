package v1

import "reflect"

type Hardware struct {
	Brand  string `json:"brand" jsonschema:"required,notEmpty"`
	Memory int    `json:"memory" jsonschema:"required"`
}

func (h Hardware) AndOneOf() []reflect.StructField {
	return []reflect.StructField{
		reflect.StructField{Type: reflect.TypeOf(Laptop{})},
		reflect.StructField{Type: reflect.TypeOf(PC{})},
	}
}

type Laptop struct {
	Brand           string `json:"brand" jsonschema:"pattern=^(apple|lenovo|dell)$"`
	NeedTouchScreen bool   `json:"need_touchscreen"`
}

type PC struct {
	Brand           string `json:"brand" jsonschema:"pattern=^(apple|lenovo|dell)$"`
	NeedTouchScreen bool   `json:"need_touchscreen"`
}
