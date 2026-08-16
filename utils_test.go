package tools

import (
	"testing"
)

func TestJsonString(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	type address struct {
		City string `json:"city"`
	}
	type contact struct {
		Email  string `json:"email"`
		Phone  string `json:"phone"`
		secret string
	}
	type employee struct {
		person
		Address *address  `json:"address"`
		Contact *contact  `json:"contact,omitempty"`
	}
	// meta is a struct value field with omitempty, which encoding/json ignores:
	// omitempty only omits pointers, interfaces and zero-length basic values.
	type profile struct {
		Name string  `json:"name"`
		Meta contact `json:"meta,omitempty"`
	}
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{name: "nil", input: nil, want: ""},
		{name: "string", input: "hello", want: `"hello"`},
		{name: "int", input: 42, want: `42`},
		{name: "bool", input: true, want: `true`},
		{name: "slice", input: []int{1, 2, 3}, want: `[1,2,3]`},
		{name: "map", input: map[string]int{"a": 1, "b": 2}, want: `{"a":1,"b":2}`},
		{name: "struct", input: person{Name: "Tom", Age: 18}, want: `{"name":"Tom","age":18}`},
		{name: "struct pointer", input: &person{Name: "Tom", Age: 18}, want: `{"name":"Tom","age":18}`},
		{name: "struct nil pointer", input: (*person)(nil), want: `null`},
		{name: "struct slice", input: []person{{Name: "Tom", Age: 18}, {Name: "Jerry", Age: 20}}, want: `[{"name":"Tom","age":18},{"name":"Jerry","age":20}]`},
		{name: "struct unexported field", input: contact{Email: "a@b.c", Phone: "123", secret: "hidden"}, want: `{"email":"a@b.c","phone":"123"}`},
		{name: "struct omitempty zero", input: employee{person: person{Name: "Tom", Age: 18}, Address: &address{City: "Shanghai"}}, want: `{"name":"Tom","age":18,"address":{"city":"Shanghai"}}`},
		{name: "struct omitempty set", input: employee{person: person{Name: "Tom", Age: 18}, Address: &address{City: "Beijing"}, Contact: &contact{Email: "a@b.c"}}, want: `{"name":"Tom","age":18,"address":{"city":"Beijing"},"contact":{"email":"a@b.c","phone":""}}`},
		{name: "struct nil pointer field", input: employee{person: person{Name: "Tom", Age: 18}, Address: nil}, want: `{"name":"Tom","age":18,"address":null}`},
		{name: "struct omitempty no effect on struct value", input: profile{Name: "Tom"}, want: `{"name":"Tom","meta":{"email":"","phone":""}}`},
		{name: "nested", input: map[string]any{"user": person{Name: "Tom", Age: 18}}, want: `{"user":{"name":"Tom","age":18}}`},
		{name: "marshal error", input: make(chan int), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JsonString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("JsonString(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JsonString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
