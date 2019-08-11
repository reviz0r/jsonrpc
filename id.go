package jsonrpc

import (
	"encoding/json"
	"reflect"
	"strconv"
)

type id struct {
	num int64
	str string

	isString bool
}

func (i id) String() string {
	if i.isString {
		return strconv.Quote(i.str)
	}

	return strconv.FormatInt(i.num, 10)
}

// MarshalJSON implements json.Marshaler
func (i id) MarshalJSON() ([]byte, error) {
	if i.num == 0 && i.str == "" {
		return []byte("null"), nil
	}

	if i.isString {
		return json.Marshal(i.str)
	}

	return json.Marshal(i.num)
}

// UnmarshalJSON implements json.Unmarshaler
func (i *id) UnmarshalJSON(data []byte) error {
	idType := reflect.TypeOf(i).Elem()

	var valString string
	if err := json.Unmarshal(data, &valString); err == nil {
		*i = id{str: valString, isString: true}
		return nil
	}

	var valNumber json.Number
	if err := json.Unmarshal(data, &valNumber); err == nil {
		valInt, err := valNumber.Int64()
		if err != nil {
			return unmarshalIDError("float", idType)
		}
		*i = id{num: valInt, isString: false}
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch val := value.(type) {
	case []interface{}:
		return unmarshalIDError("array", idType)
	case map[string]interface{}:
		return unmarshalIDError("object", idType)
	default:
		return unmarshalIDError(reflect.TypeOf(val).String(), idType)
	}
}

func unmarshalIDError(Value string, Type reflect.Type) *json.UnmarshalTypeError {
	return &json.UnmarshalTypeError{Value: Value, Type: Type, Struct: "request", Field: "id"}
}
