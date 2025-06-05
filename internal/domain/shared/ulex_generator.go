package shared

import (
	ulid "github.com/oklog/ulid/v2"
)

type ULIDsGenerator struct{}

func NewULIDsGenerator() *ULIDsGenerator {
	return &ULIDsGenerator{}
}

func (g *ULIDsGenerator) Generate() string {
	return ulid.Make().String()
}
