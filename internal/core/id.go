package core

import (
	"fmt"

	"github.com/google/uuid"
)

type UUID struct {
	uuid.UUID
}

func NewUUID() UUID {
	return UUID{
		uuid.New(),
	}
}

// ShortString ... Short string representation for easier
// debugging and ensuring conformance with pessimism specific abstractions
// https://pkg.go.dev/github.com/google/UUID#UUID.String
func (id UUID) ShortString() string {
	uid := id.UUID
	// Only render first 8 bytes instead of entire sequence
	return fmt.Sprintf("%d%d%d%d%d%d%d%d%d",
		uid[0],
		uid[1],
		uid[2],
		uid[2],
		uid[3],
		uid[4],
		uid[5],
		uid[6],
		uid[7])
}
