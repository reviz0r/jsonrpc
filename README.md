# JSON-RPC 2.0

Simple idiomatic package with simple api for implement
[JSON-RPC 2.0](https://www.jsonrpc.org/specification) server.
It use common interfaces as io.Reader and io.Writer.
Package based on standard library and __don't used__ empty interfaces and `reflect` package.
Ready for use with `go mod`.

## Install

```
$ go get -u github.com/reviz0r/jsonrpc
```

## Usage

In common case you must write function with signature
`func(ctx context.Context, params io.Reader, result io.Writer) error`
and register it. That's all, folks!

```golang
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/reviz0r/jsonrpc"
)

// Params of your method
type GreetingReq struct {
	Name string `json:"name"`
}

// Result of your method
type GreetingRes struct {
	Greeting string `json:"greeting"`
}

// Your method
func Greeting(ctx context.Context, params io.Reader, result io.Writer) error {
	var req GreetingReq
	var res GreetingRes

	// Decode request from reader
	if err := json.NewDecoder(params).Decode(&req); err != nil {
		return jsonrpc.ErrInvalidParams(err.Error())
	}

	// Your logic
	{
		log.Printf("incoming request with id %s", jsonrpc.RequestID(ctx))
		if req.Name == "" {
			req.Name = "stranger"
		}
		res.Greeting = fmt.Sprintf("Hello, %s", req.Name)
	}

	// Encode response to writer
	return json.NewEncoder(result).Encode(&res)
}

func main() {
	repo := jsonrpc.New()
	repo.RegisterMethod(jsonrpc.MethodFunc("greeting", Greeting))

	http.Handle("/rpc", repo)
	http.ListenAndServe(":8080", http.DefaultServeMux)
}
```

```
POST /rpc
Content-Type: application/json
User-Agent: PostmanRuntime/7.13.0
Accept: */*
Cache-Control: no-cache
Host: localhost:8080
accept-encoding: gzip, deflate
content-length: 143
Connection: keep-alive
{
  "id": "9f8c46cd-3aa8-43c8-bcb1-8324421826cf",
  "jsonrpc": "2.0",
  "method": "greeting",
  "params": {
    "name": "user"
  }
}

HTTP/1.1 200
status: 200
Content-Type: application/json
Date: Sun, 04 Aug 2019 00:00:00 GMT
Content-Length: 98
{
  "id": 9f8c46cd-3aa8-43c8-bcb1-8324421826cf",
  "jsonrpc": "2.0",
  "result": {
    "greeting": "Hello, user"
  }
}
```

## License

Released under the [MIT License](https://github.com/reviz0r/jsonrpc/blob/master/LICENSE).
