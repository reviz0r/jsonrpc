package jsonrpc

import "sync"

type RequestIDGenerator interface {
	Generate() ID
}

type IntGenerator struct {
	m sync.Mutex

	i int
}

func NewIntGenerator() *IntGenerator {
	return new(IntGenerator)
}

func (g *IntGenerator) Generate() ID {
	g.m.Lock()
	defer g.m.Unlock()

	g.i++
	return IntID(g.i)
}
