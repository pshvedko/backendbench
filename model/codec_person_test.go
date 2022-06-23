package model

import (
	"github.com/google/uuid"
	"testing"
)

func TestCodecPerson_String(t *testing.T) {
	type fields struct {
		ID     uuid.UUID
		Person Person
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name:   "",
			fields: fields{},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &CodecPerson{
				ID:     tt.fields.ID,
				Person: tt.fields.Person,
			}
			if got := obj.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
