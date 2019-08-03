package jsonrpc

type errorCode int

const (
	parseError     errorCode = -32700
	invalidRequest errorCode = -32600
	methodNotFound errorCode = -32601
	invalidParams  errorCode = -32602
	internalError  errorCode = -32603
	serverError    errorCode = -32000
)

func (e errorCode) Int() int {
	return int(e)
}

func (e errorCode) String() string {
	messages := map[int]string{
		parseError.Int():     "Parse error",
		invalidRequest.Int(): "Invalid Request",
		methodNotFound.Int(): "Method not found",
		invalidParams.Int():  "Invalid params",
		internalError.Int():  "Internal error",
		serverError.Int():    "Server error",
	}

	message, exist := messages[e.Int()]
	if !exist {
		return ""
	}

	return message
}

// ErrParseError Invalid JSON was received by the server.
// An error occurred on the server while parsing the JSON text.
func ErrParseError(data interface{}) *Error {
	return &Error{
		Code:    parseError.Int(),
		Message: parseError.String(),
		Data:    data,
	}
}

// ErrInvalidRequest The JSON sent is not a valid Request object.
func ErrInvalidRequest(data interface{}) *Error {
	return &Error{
		Code:    invalidRequest.Int(),
		Message: invalidRequest.String(),
		Data:    data,
	}
}

// ErrMethodNotFound The method does not exist / is not available.
func ErrMethodNotFound(data interface{}) *Error {
	return &Error{
		Code:    methodNotFound.Int(),
		Message: methodNotFound.String(),
		Data:    data,
	}
}

// ErrInvalidParams Invalid method parameter(s).
func ErrInvalidParams(data interface{}) *Error {
	return &Error{
		Code:    invalidParams.Int(),
		Message: invalidParams.String(),
		Data:    data,
	}
}

// ErrInternalError Internal JSON-RPC error.
func ErrInternalError(data interface{}) *Error {
	return &Error{
		Code:    internalError.Int(),
		Message: internalError.String(),
		Data:    data,
	}
}
