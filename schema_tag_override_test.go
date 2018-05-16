package jsonschema

import (
	"reflect"
	"testing"
)

type Human struct {
	Name   string
	Sex    string
	age    int
	secret Alien
}

type Alien struct {
	Blerp   string
	bloop   float64
	molting bool
}

func TestSchemaTagOverrideSet(t *testing.T) {
	sto := GetSchemaTagOverride()
	ok := sto.Set(Human{}, "Name", "required")

	if ok != nil {
		t.Errorf("failed to set field %s due to %s", "Name", ok)
	}
}

func TestSchemaTagOverrideSetPointer(t *testing.T) {
	sto := GetSchemaTagOverride()
	ok := sto.Set(&Human{}, "Name", "required")

	if ok == nil {
		t.Error("failed to return error when given non-struct as target struct")
	}
}

func TestSchemaTagOverrideSetErrorForInvalidField(t *testing.T) {
	sto := GetSchemaTagOverride()
	ok := sto.Set(Human{}, "name", "required")

	if ok == nil {
		t.Errorf("was able to set field %s even though it doesn't exist on target struct", "name")
	}
}

func TestGetSchemaTagOverrideGet(t *testing.T) {
	sto := GetSchemaTagOverride()
	sto.Set(Human{}, "Name", "required")

	tag := sto.Get(reflect.TypeOf(Human{}), "Name")

	if tag != "required" {
		t.Errorf("did not receive expected tag override from Name field, got %s instead", tag)
	}
}

func TestGetSchemaTagOverrideGetNonExistentField(t *testing.T) {
	sto := GetSchemaTagOverride()
	sto.Set(Human{}, "Name", "required")

	tag := sto.Get(reflect.TypeOf(Human{}), "bloop")

	if tag != "" {
		t.Errorf("tag should be empty string, got %s instead", tag)
	}
}

func TestGetSchemaTagOverrideGetMultipleFields(t *testing.T) {
	sto := GetSchemaTagOverride()
	sto.Set(Human{}, "Name", "required")
	sto.Set(Human{}, "Sex", "required,enum=male|female|neither|both")

	nameTag := sto.Get(reflect.TypeOf(Human{}), "Name")
	sexTag := sto.Get(reflect.TypeOf(Human{}), "Sex")

	if nameTag != "required" && sexTag != "required,enum=male|female|neither|both" {
		t.Errorf("did not receive expected tag overrides, got %s and %s", nameTag, sexTag)
	}
}

func TestGetSchemaTagOverrideGetMultipleKeys(t *testing.T) {
	sto := GetSchemaTagOverride()
	sto.Set(Human{}, "Name", "required")
	sto.Set(Alien{}, "bloop", "required,enum=3.1415926535897932384626")

	nameTag := sto.Get(reflect.TypeOf(Human{}), "Name")
	bloopTag := sto.Get(reflect.TypeOf(Human{}), "bloop")

	if nameTag != "required" && bloopTag != "required,enum=3.1415926535897932384626" {
		t.Errorf("did not receive expected tag overrides, got %s and %s", nameTag, bloopTag)
	}
}
