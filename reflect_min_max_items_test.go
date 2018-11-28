package jsonschema

func (st *SliceTestType) MinItems() int {
	return 2
}

func (st *SliceTestType) MaxItems() int {
	return 2
}

type SliceTestType struct {
	TestSlice []string `json:"testSlice"`
}
