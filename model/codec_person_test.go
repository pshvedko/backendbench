package model

import (
	"github.com/google/uuid"
	"net/url"
	"testing"
	"time"
)

func TestCodecPerson_String(t *testing.T) {
	type fields struct {
		ID     uuid.UUID
		Person Person
	}
	yes, no := true, false
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "",
			fields: fields{
				ID: uuid.UUID{},
				Person: Person{
					Name:    "",
					Name1:   nil,
					Name2:   nil,
					Age:     0,
					Size:    nil,
					Weight:  new(int),
					ID:      uuid.UUID{},
					SID:     nil,
					PID:     new(uuid.UUID),
					URL:     new(url.URL),
					Time:    time.Time{},
					Date1:   nil,
					Date2:   nil,
					Ok:      false,
					Ok1:     &yes,
					Ok2:     &no,
					Fruit:   nil,
					Fruit2:  nil,
					Fruit3:  nil,
					Handler: nil,
				},
			},
			want: "{id=00000000-0000-0000-0000-000000000000 person={name= name1=<nil> name2=<nil> age=0 size=<nil> weight=<nil> id=00000000-0000-0000-0000-000000000000 sid=<nil> pid=<nil> url=<nil> time=0001-01-01 00:00:00 +0000 UTC date=<nil> date2=<nil> ok=false ok1=<nil> ok2=<nil> fruit=[] fruit2=[] fruit3=<nil>}}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &CodecPerson{
				ID:     tt.fields.ID,
				Person: tt.fields.Person,
			}
			if got := obj.String(); got != tt.want {
				t.Errorf("\n\t\t got = %v,\n\t\twant = %v", got, tt.want)
			}
		})
	}
}
