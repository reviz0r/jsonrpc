package jsonrpc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type ID struct {
	str      string
	isQuoted bool
}

func IntID(num int) ID {
	return ID{str: strconv.Itoa(num), isQuoted: false}
}

func StringID(str string) ID {
	return ID{str: str, isQuoted: true}
}

func (i ID) String() string {
	return i.str
}

func (i ID) IsZero() bool {
	return i.str == "" && i.isQuoted == false
}

// MarshalJSON implements json.Marshaler
func (i ID) MarshalJSON() ([]byte, error) {
	if i.IsZero() {
		return json.RawMessage("null"), nil
	}

	if i.isQuoted {
		quoted := strconv.Quote(i.str)
		return json.RawMessage(quoted), nil
	}

	return json.RawMessage(i.str), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (i *ID) UnmarshalJSON(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("jsonrpc: unmarshal ID type: %w", err)
	}

	switch val := value.(type) {
	case float64:
		if !isInteger(val) {
			return unmarshalIDError("float")
		}

		*i = ID{str: strconv.Itoa(int(val)), isQuoted: false}
		return nil
	case string:
		*i = ID{str: val, isQuoted: true}
		return nil

	case bool:
		return unmarshalIDError("bool")
	case []any:
		return unmarshalIDError("array")
	case map[string]any:
		return unmarshalIDError("object")
	default:
		return unmarshalIDError(reflect.TypeOf(val).String())
	}
}

func isInteger(val float64) bool {
	return val == float64(int(val))
}

func unmarshalIDError(value string) *json.UnmarshalTypeError {
	idType := reflect.TypeFor[*ID]().Elem()
	return &json.UnmarshalTypeError{Value: value, Type: idType}
}
