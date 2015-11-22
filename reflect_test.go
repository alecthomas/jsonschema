package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type TestUser struct {
	ID        int                    `json:"id"`
	Name      string                 `json:"name"`
	Friends   []int                  `json:"friends,omitempty"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
	BirthDate time.Time              `json:"birth_date,omitempty"`

	TestFlag       bool
	IgnoredCounter int `json:"-"`
}

func TestJSONSchema(t *testing.T) {
	s := Reflect(&TestUser{})
	b, _ := json.MarshalIndent(s, "", "  ")
	fmt.Printf("%s\n", b)
}

// TestIgnoredProperties checks if fields with "-" tag or no json tag are ignored
func TestIgnoredProperties(t *testing.T) {
	s := Reflect(&TestUser{})

	expectedProperties := map[string]bool{
		"id":         true,
		"name":       true,
		"friends":    true,
		"tags":       true,
		"birth_date": true,
	}

	props := s.Definitions["TestUser"].Properties

	for defKey := range props {
		if _, ok := expectedProperties[defKey]; !ok {
			t.Fatalf("unexpected property %s", defKey)
		}
	}
}
