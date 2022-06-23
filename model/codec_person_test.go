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
			want:   "{id=00000000-0000-0000-0000-000000000000 person={name= name1=<nil> name2=<nil> age=0 size=<nil> weight=<nil> id=00000000-0000-0000-0000-000000000000 sid=<nil> pid=<nil> url=<nil> time=0001-01-01 00:00:00 +0000 UTC date=<nil> date2=<nil> ok=false ok1=<nil> ok2=<nil> fruit=[] fruit2=[] fruit3=<nil>}}",
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
