package entity

import (
	"errors"
	"math"
	"testing"
	"unsafe"
)

func testMarshalError(t *testing.T, v interface{}, expected error) {
	b, err := Marshal(v)
	if err == nil {
		t.Errorf("Expected error '%s', got %s", expected.Error(), string(b))
	}
	if err.Error() != expected.Error() {
		t.Errorf("Expected error '%s', got error '%s'", expected.Error(), err.Error())
	}
	if b != nil {
		t.Errorf("Expected bytes to be nil, got %s", string(b))
	}
}

func testMarshal(t *testing.T, v interface{}, expected string) {
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Expected %s, got error %s", expected, err.Error())
	}
	actual := string(b)
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestMarshalBool(t *testing.T) {
	testMarshal(t, true, `{"BOOL":true}`)
	testMarshal(t, false, `{"BOOL":false}`)
}

func TestMarshalInt(t *testing.T) {
	testMarshal(t, 12, `{"N":"12"}`)
	testMarshal(t, -2, `{"N":"-2"}`)
	testMarshal(t, math.Pow(2, 53)-1, `{"N":"9.007199254740991e+15"}`)
}

func TestMarshalFloat(t *testing.T) {
	testMarshal(t, 3.14, `{"N":"3.14"}`)
	testMarshal(t, -99.99, `{"N":"-99.99"}`)
}

func TestMarshalString(t *testing.T) {
	testMarshal(t, "this is a string", `{"S":"this is a string"}`)
	testMarshal(t, `"this is a string"`, `{"S":"\"this is a string\""}`)
}

func TestMarshalPtr(t *testing.T) {
	// Pointers aren't supported, this should be an error (and not a panic).
	m := make([]int, 1)
	ptr1 := *(*uintptr)(unsafe.Pointer(&m))
	testMarshalError(t, ptr1, errors.New("the type uintptr is not supported"))
	ptr2 := *(**int)(unsafe.Pointer(&m))
	testMarshalError(t, ptr2, errors.New("the type ptr is not supported"))
}