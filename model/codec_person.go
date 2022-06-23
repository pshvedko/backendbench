//go:generate go run ./stringer CodecPerson

package model

import "github.com/google/uuid"

type CodecPerson struct {
	ID     uuid.UUID `json:"id,omitempty"`
	Person Person    `json:"person,omitempty"`
}
