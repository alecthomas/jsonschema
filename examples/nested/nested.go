package nested

// Pet defines the user's fury friend.
type Pet struct {
	// Name of the animal.
	Name string `json:"name" jsonschema:"title=Name"`
}

type (
	// Plant represents the plants the user might have and serves as a test
	// of structs inside a `type` set.
	Plant struct {
		Variant string `json:"variant" jsonschema:"title=Variant"` // This comment will be ignored
	}
)
