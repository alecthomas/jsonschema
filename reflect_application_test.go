package jsonschema

import "reflect"

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
		If:   conditionField,
		Then: WebApp{},
		Else: MobileApp{},
	}
}
