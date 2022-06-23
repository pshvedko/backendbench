//go:generate go run ./stringer Person

package model

import (
	"github.com/google/uuid"
	"net/url"
	"time"
)

type Person struct {
	Name   string         `json:"name"`
	Name1  *string        `json:"name1"`
	Name2  *string        `json:"name2"`
	Age    int            `json:"age"`
	Size   *int           `json:"size"`
	Weight *int           `json:"weight"`
	ID     uuid.UUID      `json:"id"`
	SID    *uuid.UUID     `json:"sid"`
	PID    *uuid.UUID     `json:"pid"`
	URL    *url.URL       `json:"url"`
	Time   time.Time      `json:"time"`
	Date1  *time.Time     `json:"date"`
	Date2  *time.Time     `json:"date2"`
	Ok     bool           `json:"ok"`
	Ok1    *bool          `json:"ok1"`
	Ok2    *bool          `json:"ok2"`
	Fruit  []string       `json:"fruit"`
	Fruit2 []uint         `json:"fruit2"`
	Fruit3 *[0][]*url.URL `json:"fruit3"`
}
