# Go JSON Schema Reflection

[![Build Status](https://travis-ci.org/alecthomas/jsonschema.png)](https://travis-ci.org/alecthomas/jsonschema)
[![Gitter chat](https://badges.gitter.im/alecthomas.png)](https://gitter.im/alecthomas/Lobby)
[![Go Report Card](https://goreportcard.com/badge/github.com/alecthomas/jsonschema)](https://goreportcard.com/report/github.com/alecthomas/jsonschema)
[![GoDoc](https://godoc.org/github.com/alecthomas/jsonschema?status.svg)](https://godoc.org/github.com/alecthomas/jsonschema)

This package can be used to generate [JSON Schemas](http://json-schema.org/latest/json-schema-validation.html) from Go types through reflection.

- Supports arbitrarily complex types, including `interface{}`, maps, slices, etc.
- Supports json-schema features such as minLength, maxLength, pattern, format, etc.
- Supports simple string and numeric enums.
- Supports custom property fields via the `jsonschema_extras` struct tag.

## Example

The following Go type:

```go
type TestUser struct {
  ID            int                    `json:"id"`
  Name          string                 `json:"name" jsonschema:"title=the name,description=The name of a friend,example=joe,example=lucy,default=alex"`
  Friends       []int                  `json:"friends,omitempty" jsonschema_description:"The list of IDs, omitted when empty"`
  Tags          map[string]interface{} `json:"tags,omitempty" jsonschema_extras:"a=b,foo=bar,foo=bar1"`
  BirthDate     time.Time              `json:"birth_date,omitempty" jsonschema:"oneof_required=date"`
  YearOfBirth   string                 `json:"year_of_birth,omitempty" jsonschema:"oneof_required=year"`
  Metadata      interface{}            `json:"metadata,omitempty" jsonschema:"oneof_type=string;array"`
  FavColor      string                 `json:"fav_color,omitempty" jsonschema:"enum=red,enum=green,enum=blue"`
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
        "metadata": {
          "oneOf": [
            {
              "type": "string"
            },
            {
              "type": "array"
            }
          ]
        },
        "birth_date": {
          "type": "string",
          "format": "date-time"
        },
        "friends": {
          "type": "array",
          "items": {
            "type": "integer"
          },
          "description": "The list of IDs, omitted when empty"
        },
        "id": {
          "type": "integer"
        },
        "name": {
          "type": "string",
          "title": "the name",
          "description": "The name of a friend",
          "default": "alex",
          "examples": [
            "joe",
            "lucy"
          ]
        },
        "tags": {
          "type": "object",
          "patternProperties": {
            ".*": {
              "additionalProperties": true
            }
          },
          "a": "b",
          "foo": [
            "bar",
            "bar1"
          ]
        },
        "fav_color": {
          "type": "string",
          "enum": [
            "red",
            "green",
            "blue"
          ]
        }
      },
      "additionalProperties": false,
      "required": ["id", "name"],
      "oneOf": [
        {
          "required": [
            "birth_date"
          ],
          "title": "date"
        },
        {
          "required": [
            "year_of_birth"
          ],
          "title": "year"
        }
      ]
    }
  }
}
```
## Configurable behaviour

The behaviour of the schema generator can be altered with parameters when a `jsonschema.Reflector`
instance is created.

### ExpandedStruct

If set to ```true```, makes the top level struct not to reference itself in the definitions. But type passed should be a struct type.

eg.

```go
type GrandfatherType struct {
	FamilyName string `json:"family_name" jsonschema:"required"`
}

type SomeBaseType struct {
	SomeBaseProperty int `json:"some_base_property"`
	// The jsonschema required tag is nonsensical for private and ignored properties.
	// Their presence here tests that the fields *will not* be required in the output
	// schema, even if they are tagged required.
	somePrivateBaseProperty            string `json:"i_am_private" jsonschema:"required"`
	SomeIgnoredBaseProperty            string `json:"-" jsonschema:"required"`
	SomeSchemaIgnoredProperty          string `jsonschema:"-,required"`
	SomeUntaggedBaseProperty           bool   `jsonschema:"required"`
	someUnexportedUntaggedBaseProperty bool
	Grandfather                        GrandfatherType `json:"grand"`
}
```

will output:

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
