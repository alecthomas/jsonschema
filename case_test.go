package jsonschema

type ExampleCase struct {
	Type string `json:"type" jsonschema:"required;enum="bool|int|string"`
}

type IntPayload struct {
	Payload int `json:"payload"`
}

type StringPayload struct {
	Payload string `json:"payload"`
}

type BoolPayload struct {
	Payload bool `json:"payload"`
}

func (ex ExampleCase) Case() SchemaSwitch {
	cases := make(map[string]interface{})
	cases["bool"] = BoolPayload{}
	cases["int"] = IntPayload{}
	cases["string"] = StringPayload{}

	return SchemaSwitch{
		ByField: "type",
		Cases:   cases,
	}
}
