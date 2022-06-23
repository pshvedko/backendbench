package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pshvedko/backendbench/model"
	"log"
	"net/url"
	"reflect"
	"strings"
	"time"
)

func xx() {
	var no bool
	var name string
	var id uuid.UUID
	var t time.Time
	var size int
	log.Println(toString(model.Person{
		Name:   "123",
		Name2:  &name,
		Age:    123,
		Size:   &size,
		Weight: nil,
		ID:     uuid.New(),
		PID:    &id,
		URL: &url.URL{
			Scheme:   "ftp",
			Host:     "localhost:8080",
			Path:     "/user/1",
			RawQuery: "a=b",
			Fragment: "",
		},
		Time:   time.Now(),
		Date1:  &t,
		Ok2:    &no,
		Fruit2: []uint{1, 2, 3},
	}, "json", "nil", '"', ':'))
}

type builder struct {
	strings.Builder
}

func toString(v interface{}, k, n string, c, e byte) string {
	b := builder{}
	t := reflect.TypeOf(v)
	f := reflect.ValueOf(v)
	switch t.Kind() {
	case reflect.Struct:
		b.WriteByte('{')
		for i := 0; i < t.NumField(); i++ {
			if i > 0 {
				b.WriteByte(',')
				b.WriteByte(' ')
			}
			b.WriteByte(c)
			b.WriteString(t.Field(i).Tag.Get(k))
			b.WriteByte(c)
			b.WriteByte(e)
			p := f.Field(i)
			b.write(p, k, n, c, e)
		}
		b.WriteByte('}')
	default:
		b.write(f, k, n, c, e)
	}
	return b.String()
}

func (b *builder) write(p reflect.Value, k, n string, c, e byte) {
	switch p.Kind() {
	case reflect.Slice, reflect.Array:
		if !p.IsZero() {
			_, ok := p.Interface().(fmt.Stringer)
			if !ok {
				b.WriteByte('[')
				for j := 0; j < p.Len(); j++ {
					if j > 0 {
						b.WriteByte(',')
						b.WriteByte(' ')
					}
					b.WriteString(toString(p.Index(j).Interface(), k, n, c, e))
				}
				b.WriteByte(']')
			}
		}
		fallthrough
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Interface, reflect.Pointer, reflect.UnsafePointer:
		if p.IsZero() {
			b.WriteString(n)
			return
		}
	}
	switch x := p.Interface().(type) {
	case *time.Time:
		b.WriteString(fmt.Sprintf("%q", x.Format(time.RFC3339)))
	case time.Time:
		b.WriteString(fmt.Sprintf("%q", x.Format(time.RFC3339)))
	case fmt.Stringer:
		b.WriteString(fmt.Sprintf("%q", x.String()))
	case string:
		b.WriteString(fmt.Sprintf("%q", x))
	case *string:
		b.WriteString(fmt.Sprintf("%q", *x))
	case nil, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
		b.WriteString(fmt.Sprint(x))
	case *bool:
		b.WriteString(fmt.Sprint(*x))
	case *int:
		b.WriteString(fmt.Sprint(*x))
	case *int8:
		b.WriteString(fmt.Sprint(*x))
	case *int16:
		b.WriteString(fmt.Sprint(*x))
	case *int32:
		b.WriteString(fmt.Sprint(*x))
	case *int64:
		b.WriteString(fmt.Sprint(*x))
	}
}
