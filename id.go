package jsonrpc

import (
	"encoding/json"
	"reflect"
	"strconv"
)

type id struct {
	str      string
	isQuoted bool
}

func IntID(num int) id {
	return id{str: strconv.Itoa(num), isQuoted: false}
}

func StringID(str string) id {
	return id{str: str, isQuoted: true}
}

func (i id) String() string {
	return i.str
}

func (i id) IsZero() bool {
	return i.str == "" && i.isQuoted == false
}

// MarshalJSON implements json.Marshaler
func (i id) MarshalJSON() ([]byte, error) {
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
func (i *id) UnmarshalJSON(data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch val := value.(type) {
	case float64:
		if !isInteger(val) {
			return unmarshalIDError("float")
		}

		*i = id{str: strconv.Itoa(int(val)), isQuoted: false}
		return nil
	case string:
		*i = id{str: val, isQuoted: true}
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

func unmarshalIDError(Value string) *json.UnmarshalTypeError {
	idType := reflect.TypeFor[*id]().Elem()
	return &json.UnmarshalTypeError{Value: Value, Type: idType}
}
