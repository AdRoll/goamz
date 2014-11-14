package entity

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"reflect"
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

func testMarshal(t *testing.T, v interface{}, expectedString string) {
	b, err := Marshal(v)
	if err != nil {
		t.Errorf("Expected %s, got error '%s'", expectedString, err.Error())
	}
	actualString := string(b)
	// Since JSON is unordered, we can't do a simple string compare. Instead, we
	// deserialize into maps, and do a recursive comparison of the elements.
	var actual, expected map[string]interface{}
	if err := json.Unmarshal(b, &actual); err != nil {
		t.Errorf("Got error '%s' unmarshalling %s", err.Error(), actualString)
	}
	var buf bytes.Buffer
	buf.WriteString(expectedString)
	if err := json.Unmarshal(buf.Bytes(), &expected); err != nil {
		t.Errorf("Got error '%s' unmarshalling %s", err.Error(), expectedString)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %s, got %s", expectedString, actualString)
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

type simpleStruct struct {
	Int    int
	String string
}

type complexStruct struct {
	Int    int
	String string
	Simple simpleStruct
}

func TestMarshalStruct(t *testing.T) {
	simple := simpleStruct{4, "this is a string"}
	testMarshal(t, simple, `{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}`)
	complex := complexStruct{11, "blah", simple}
	testMarshal(t, complex, `{"M":{"Int":{"N":"11"},"String":{"S":"blah"},"Simple":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}}}`)
}

func TestMarshalMap(t *testing.T) {
	m1 := map[string]interface{}{
		"Int":    4,
		"String": "this is a string"}
	testMarshal(t, m1, `{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}`)
	m2 := map[string]interface{}{
		"Map": map[string]interface{}{
			"Int":    4,
			"String": "this is a string"},
		"Nil": nil}
	testMarshal(t, m2, `{"M":{"Map":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}},"Nil":{"NULL":"true"}}}`)
}

func TestMarshalPtr(t *testing.T) {
	// Pointers aren't supported, this should be an error (and not a panic).
	m := make([]int, 1)
	ptr1 := *(*uintptr)(unsafe.Pointer(&m))
	testMarshalError(t, ptr1, errors.New("the type uintptr is not supported"))
	ptr2 := *(**int)(unsafe.Pointer(&m))
	testMarshalError(t, ptr2, errors.New("the type ptr is not supported"))
}
