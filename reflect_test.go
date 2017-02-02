package jsonschema

import (
	"net"
	"net/url"
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

type ProtoEnum int32

func (ProtoEnum) EnumDescriptor() ([]byte, []int) { return []byte(nil), []int{0} }

const (
	Unset ProtoEnum = iota
	Great
)

type TestUser struct {
	SomeBaseType
	nonExported

	ID      int                    `json:"id"`
	Name    string                 `json:"name"`
	Friends []int                  `json:"friends,omitempty"`
	Tags    map[string]interface{} `json:"tags,omitempty"`

	TestFlag       bool
	IgnoredCounter int `json:"-"`

	// Tests for RFC draft-wright-json-schema-validation-00, section 7.3
	BirthDate time.Time `json:"birth_date,omitempty"`
	Website   url.URL   `json:"website,omitempty"`
	IPAddress net.IP    `json:"network_address,omitempty"`

	// Tests for RFC draft-wright-json-schema-hyperschema-00, section 4
	Photo []byte `json:"photo,omitempty"`

	// Tests for jsonpb enum support
	Feeling ProtoEnum `json:"feeling,omitempty"`
}

// TestSchemaGeneration checks if schema generated correctly:
// - fields marked with "-" are ignored
// - non-exported fields are ignored
// - anonymous fields are expanded
func TestSchemaGeneration(t *testing.T) {
	s := Reflect(&TestUser{})

	expectedProperties := map[string]string{
		"id":                       "integer",
		"name":                     "string",
		"friends":                  "array",
		"tags":                     "object",
		"birth_date":               "string",
		"TestFlag":                 "boolean",
		"some_base_property":       "integer",
		"grand":                    "#/definitions/GrandfatherType",
		"SomeUntaggedBaseProperty": "boolean",
		"website":                  "string",
		"network_address":          "string",
		"photo":                    "string",
		"feeling":                  "",
	}

	props := s.Definitions["TestUser"].Properties
	for defKey, prop := range props {
		typeOrRef, ok := expectedProperties[defKey]
		if !ok {
			t.Fatalf("unexpected property '%s'", defKey)
		}
		if prop.Type != "" && prop.Type != typeOrRef {
			t.Fatalf("expected property type '%s', got '%s' for property '%s'", typeOrRef, prop.Type, defKey)
		} else if prop.Ref != "" && prop.Ref != typeOrRef {
			t.Fatalf("expected reference to '%s', got '%s' for property '%s'", typeOrRef, prop.Ref, defKey)
		}

		if prop.Media != nil {
			if prop.Type != "string" {
				t.Fatalf("expected property type 'string' due to existence of 'media' property, got '%s'", prop.Type)
			}

			// Technically this is case insensitive and could be a handful of
			// other encoding types per RFC 2046 section 6.1, but this code
			// naively assumes byte slices will encode to base64 strings.
			if prop.Media.BinaryEncoding != "base64" {
				t.Fatalf("expected 'base64' binary encoding, got '%s'", prop.Media.BinaryEncoding)
			}
		}

		if defKey == "feeling" {
			if prop.OneOf == nil {
				t.Fatal("expected 'oneOf' for 'feeling', got nil")
			}

			if prop.OneOf[0].Type != "string" || prop.OneOf[1].Type != "integer" {
				t.Fatalf("expected oneOf 'string' or 'integer', got %+v %+v", prop.OneOf[0], prop.OneOf[1])
			}
		}
	}

	for defKey := range expectedProperties {
		if _, ok := props[defKey]; !ok {
			t.Fatalf("expected property missing '%s'", defKey)
		}
	}
}
