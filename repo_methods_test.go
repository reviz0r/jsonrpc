package jsonrpc

import (
	"context"
	"encoding/json"
	"io"
)

type SubtractPositionalRequest [2]int
type SubtractPositionalResponse int

func SubtractPositional(ctx context.Context, in io.Reader, out io.Writer) error {
	var req SubtractPositionalRequest
	var res SubtractPositionalResponse
	if err := json.NewDecoder(in).Decode(&req); err != nil {
		return ErrInvalidParams(err.Error())
	}

	// Your logic
	{
		res = SubtractPositionalResponse(req[0] - req[1])
	}

	return json.NewEncoder(out).Encode(&res)
}

type SubtractNamedRequest struct {
	Minuend, Subtrahend int
}

type SubtractNamedResponse int

func SubtractNamed(ctx context.Context, in io.Reader, out io.Writer) error {
	var req SubtractNamedRequest
	var res SubtractNamedResponse
	if err := json.NewDecoder(in).Decode(&req); err != nil {
		return ErrInvalidParams(err.Error())
	}

	// Your logic
	{
		res = SubtractNamedResponse(req.Minuend - req.Subtrahend)
	}

	return json.NewEncoder(out).Encode(&res)
}
