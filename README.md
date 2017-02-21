# Go JSON Schema Reflection

This package can be used to generate [JSON Schemas](http://json-schema.org/latest/json-schema-validation.html) from Go types through reflection.

It supports arbitrarily complex types, including `interface{}`, maps, slices, etc.

## Example

The following Go type:

```go
type TestUser struct {
  ID        int                    `json:"id"`
  Name      string                 `json:"name"`
  Friends   []int                  `json:"friends,omitempty"`
  Tags      map[string]interface{} `json:"tags,omitempty"`
  BirthDate time.Time              `json:"birth_date,omitempty"`
}
```

Results in following JSON Schema:

```go
jsonschema.Reflect(&TestUser{})
```

```json
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "$ref": "#/definitions/TestUser",
  "definitions": {
    "TestUser": {
      "type": "object",
      "properties": {
        "birth_date": {
          "type": "string",
          "format": "date-time"
        },
        "friends": {
          "type": "array",
          "items": {
            "type": "integer"
          }
        },
        "id": {
          "type": "integer"
        },
        "name": {
          "type": "string"
        },
        "tags": {
          "type": "object",
          "patternProperties": {
            ".*": {
              "type": "object",
              "additionalProperties": true
            }
          }
        }
      },
      "additionalProperties": false,
      "required": ["id", "name"]
    }
  }
}
```
# flags
* ExpandedStruct
	If set, makes the top level struct not to reference itself in the definitions

```go
type GrandfatherType struct {
	FamilyName string `json:"family_name" jsonschema:"required"`
}

type SomeBaseType struct {
	SomeBaseProperty int `json:"some_base_property"`
	// The jsonschema required tag is nonsensical for private and ignored properties.
	// Their presence here tests that the fields *will not* be required in the output
	// schema, even if they are tagged required.
	somePrivateBaseProperty string          `json:"i_am_private" jsonschema:"required"`
	SomeIgnoredBaseProperty string          `json:"-" jsonschema:"required"`
	Grandfather             GrandfatherType `json:"grand"`

	SomeUntaggedBaseProperty           bool `jsonschema:"required"`
	someUnexportedUntaggedBaseProperty bool
}

r := jsonschema.Reflector{ExpandedStruct: true}
r.ReflectStruct(SomeBaseType{})
```

will result in

```json
{
	"$schema": "http://json-schema.org/draft-04/schema#",
	"required": [
		"some_base_property",
		"grand",
		"SomeUntaggedBaseProperty"
	],
	"properties": {
		"SomeUntaggedBaseProperty": {
			"type": "boolean"
		},
		"grand": {
			"$schema": "http://json-schema.org/draft-04/schema#",
			"$ref": "#/definitions/GrandfatherType"
		},
		"some_base_property": {
			"type": "integer"
		}
	},
	"type": "object",
	"definitions": {
		"GrandfatherType": {
			"required": [
				"family_name"
			],
			"properties": {
				"family_name": {
					"type": "string"
				}
			},
			"additionalProperties": false,
			"type": "object"
		}
	}
}
```