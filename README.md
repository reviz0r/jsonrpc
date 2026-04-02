# JSON-RPC 2.0

Simple idiomatic package with simple api for implement
[JSON-RPC 2.0](https://www.jsonrpc.org/specification) server.
Package based on standard library and __don't used__ empty interfaces and `reflect` package.
Ready for use with `go mod`.

## Install

```
$ go get -u github.com/reviz0r/jsonrpc/v2
```

## Usage

In common case you must write function with signature
`func[Params, Result any](ctx context.Context, params Params) (Result, error)`
and register it. That's all, folks!

```golang
// Params of your method
type GreetingParams struct {
	Name string `json:"name"`
}

// Result of your method
type GreetingResult struct {
	Greeting string `json:"greeting"`
}

// Your method
func Greeting(ctx context.Context, params *GreetingParams) (*GreetingResult, error) {
	var result GreetingResult

	log.Printf("incoming request with id %s", jsonrpc.RequestID(ctx))

	if params.Name == "" {
		params.Name = "stranger"
	}

	result.Greeting = fmt.Sprintf("Hello, %s", params.Name)

	return &result, nil
}

func main() {
	server := jsonrpc.NewServer()
	server.RegisterMethod("greeting",
		jsonrpc.CreateMethod(jsonrpc.FuncMethod[*GreetingParams, *GreetingResult](Greeting)))

	http.Handle("/rpc", server)
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
