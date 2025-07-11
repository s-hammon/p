package p

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name       string
		a, b, want any
	}{
		{"int 2nd", 0, 1, 1},
		{"int 1st", 2, 1, 2},
		{"string 2nd", "", "hello, world", "hello, world"},
		{"string 2nd", "hello, world", "sekai, konnichiha", "hello, world"},
		{"bool 2nd", false, true, true},
		{"bool 1st", true, false, true},
		{"bool 0th", false, false, false},
		{"float 2nd", 0.0, 4.20, 4.20},
		{"float 1st", 4.20, 6.9, 4.20},
		{"slice 2nd", []int(nil), []int{4, 2, 0}, []int{4, 2, 0}},
		{"slice 1st", []int{6, 9}, []int{4, 2, 0}, []int{6, 9}},
		{
			"struct 2nd",
			struct {
				A int
				B string
			}{},
			struct {
				A int
				B string
			}{420, "blaze it"},
			struct {
				A int
				B string
			}{420, "blaze it"},
		},
		{
			"struct 1st",
			struct {
				A int
				B string
			}{420, "blaze it"},
			struct {
				A int
				B string
			}{69, "nice"},
			struct {
				A int
				B string
			}{420, "blaze it"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Coalesce(tt.a, tt.b)
			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("want '%v', got '%v'", tt.want, got)
			}
		})
	}
}

func TestStringDeserialize(t *testing.T) {
	tests := []struct {
		name  string
		props []string
		want  map[string]string
	}{
		{
			"key-value pair",
			[]string{"key=value"},
			map[string]string{"key": "value"},
		},
		{
			"multiple key-value pair",
			[]string{
				"user=name",
				"pass=word",
				"8=D",
			},
			map[string]string{
				"user": "name",
				"pass": "word",
				"8":    "D",
			},
		},
		{
			"missing key/value",
			[]string{
				"this=works",
				"nokey=",
				"=noval",
			},
			map[string]string{"this": "works"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringDeserialize(tt.props)
			assert.Equal(t, tt.want, got)
		})
	}
}
