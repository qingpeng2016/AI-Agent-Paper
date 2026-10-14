package jsonx

import (
	"encoding/json"

	"gorm.io/datatypes"
)

func MustJSON(v any) datatypes.JSON {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON([]byte(`{"marshal_error":true}`))
	}
	return datatypes.JSON(b)
}

func MustJSONStep(m map[string]any) datatypes.JSON {
	return MustJSON(m)
}

func PtrString(s string) *string { return &s }
