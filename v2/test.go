package v2

type Hardware struct {
	Brand  string `json:"brand" jsonschema:"required,notEmpty"`
	Memory int    `json:"memory" jsonschema:"required"`
}
