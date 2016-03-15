package jsonschema

import (
	"testing"
	"time"
)

type GrandfatherType struct {
	FamilyName string `json:"family_name"`
}

type SomeBaseType struct {
	SomeBaseProperty        int             `json:"some_base_property"`
	somePrivateBaseProperty string          `json:"i_am_private"`
	SomeIgnoredBaseProperty string          `json:"-"`
	Grandfather             GrandfatherType `json:"grand"`

	SomeUntaggedBaseProperty           bool
	someUnexportedUntaggedBaseProperty bool
}

type nonExported struct {
	PublicNonExported  int
	privateNonExported int
}

type TestUser struct {
	SomeBaseType
	nonExported

	ID        int                    `json:"id"`
	Name      string                 `json:"name"`
	Friends   []int                  `json:"friends,omitempty"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
	BirthDate time.Time              `json:"birth_date,omitempty"`

	TestFlag       bool
	IgnoredCounter int `json:"-"`
}

// TestSchemaGeneration checks if schema generated correctly:
// - fields marked with "-" are ignored
// - non-exported fields are ignored
// - anonymous fields are expanded
func TestSchemaGeneration(t *testing.T) {
	s := Reflect(&TestUser{})

	expectedProperties := map[string]bool{
		"id":                       true,
		"name":                     true,
		"friends":                  true,
		"tags":                     true,
		"birth_date":               true,
		"TestFlag":                 true,
		"some_base_property":       true,
		"grand":                    true,
		"SomeUntaggedBaseProperty": true,
	}

	props := s.Definitions["TestUser"].Properties

	for defKey := range props {
		if _, ok := expectedProperties[defKey]; !ok {
			t.Fatalf("unexpected property '%s'", defKey)
		}
	}

	for defKey := range expectedProperties {
		if _, ok := props[defKey]; !ok {
			t.Fatalf("expected property missing '%s'", defKey)
		}
	}
}
