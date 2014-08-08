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
}

func TestJSONSchema(t *testing.T) {
	s := Reflect(&TestUser{})
	b, _ := json.MarshalIndent(s, "", "  ")
	fmt.Printf("%s\n", b)
}
