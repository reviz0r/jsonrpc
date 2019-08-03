package jsonrpc

import (
	"context"
	"encoding/json"
	"io"
)

type SubstractPositionalRequest [2]int
type SubstractPositionalResponse int

func SubstractPositional(ctx context.Context, in io.Reader, out io.Writer) error {
	var req SubstractPositionalRequest
	var res SubstractPositionalResponse
	if err := json.NewDecoder(in).Decode(&req); err != nil {
		return ErrInvalidParams(err.Error())
	}

	// Your logic
	{
		res = SubstractPositionalResponse(req[0] - req[1])
	}

	return json.NewEncoder(out).Encode(&res)
}

type SubstractNamedRequest struct {
	Minuend, Subtrahend int
}

type SubstractNamedResponse int

func SubstractNamed(ctx context.Context, in io.Reader, out io.Writer) error {
	var req SubstractNamedRequest
	var res SubstractNamedResponse
	if err := json.NewDecoder(in).Decode(&req); err != nil {
		return ErrInvalidParams(err.Error())
	}

	// Your logic
	{
		res = SubstractNamedResponse(req.Minuend - req.Subtrahend)
	}

	return json.NewEncoder(out).Encode(&res)
}
