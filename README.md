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

Results in this JSON Schema:

```json
{
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
      "required": [
        "id",
        "name"
      ]
    }
  }
}
```
