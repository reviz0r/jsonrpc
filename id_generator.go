package jsonrpc

import (
	"sync"

	"github.com/google/uuid"
)

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

type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return new(UUIDGenerator)
}

func (g *UUIDGenerator) Generate() ID {
	uid := uuid.New()
	return StringID(uid.String())
}
