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

func itemMap(v interface{}) interface{} {
	return map[string]interface{}{"Item": v}
}

type item struct {
	Item interface{}
}

func itemStruct(v interface{}) interface{}  {
	return map[string]interface{}{"Item": v}
}

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
	testMarshal(t, itemMap(true), `{"Item":{"BOOL":true}}`)
	testMarshal(t, itemStruct(true), `{"Item":{"BOOL":true}}`)
	testMarshal(t, itemMap(false), `{"Item":{"BOOL":false}}`)
	testMarshal(t, itemStruct(false), `{"Item":{"BOOL":false}}`)
}

func TestMarshalInt(t *testing.T) {
	testMarshal(t, itemMap(12), `{"Item":{"N":"12"}}`)
	testMarshal(t, itemStruct(12), `{"Item":{"N":"12"}}`)
	testMarshal(t, itemMap(-2), `{"Item":{"N":"-2"}}`)
	testMarshal(t, itemStruct(-2), `{"Item":{"N":"-2"}}`)
	testMarshal(t, itemMap(math.Pow(2, 53)-1), `{"Item":{"N":"9.007199254740991e+15"}}`)
	testMarshal(t, itemStruct(math.Pow(2, 53)-1), `{"Item":{"N":"9.007199254740991e+15"}}`)
}

func TestMarshalFloat(t *testing.T) {
	testMarshal(t, itemMap(3.14), `{"Item":{"N":"3.14"}}`)
	testMarshal(t, itemStruct(3.14), `{"Item":{"N":"3.14"}}`)
	var f32 float32 = -99.99
	testMarshal(t, itemMap(f32), `{"Item":{"N":"-99.98999786376953"}}`)
	testMarshal(t, itemStruct(f32), `{"Item":{"N":"-99.98999786376953"}}`)
	var f64 float64 = math.MaxFloat32 + 1
	testMarshal(t, itemStruct(f64), `{"Item":{"N":"3.4028234663852886e+38"}}`)
	testMarshal(t, itemMap(f64), `{"Item":{"N":"3.4028234663852886e+38"}}`)
}

func TestMarshalString(t *testing.T) {
	testMarshal(t, itemMap("this is a string"), `{"Item":{"S":"this is a string"}}`)
	testMarshal(t, itemStruct("this is a string"), `{"Item":{"S":"this is a string"}}`)
	testMarshal(t, itemMap(`"this is a string"`), `{"Item":{"S":"\"this is a string\""}}`)
	testMarshal(t, itemStruct(`"this is a string"`), `{"Item":{"S":"\"this is a string\""}}`)
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
	testMarshal(t, itemMap(simple), `{"Item":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}}`)
	testMarshal(t, itemStruct(simple), `{"Item":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}}`)
	complex := complexStruct{11, "blah", simple}
	testMarshal(t, itemMap(complex), `{"Item":{"M":{"Int":{"N":"11"},"String":{"S":"blah"},"Simple":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}}}}`)
	testMarshal(t, itemStruct(complex), `{"Item":{"M":{"Int":{"N":"11"},"String":{"S":"blah"},"Simple":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}}}}`)
}

func TestMarshalMap(t *testing.T) {
	m1 := map[string]interface{}{
		"Int":    4,
		"String": "this is a string"}
	testMarshal(t, itemMap(m1), `{"Item":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}}`)
	testMarshal(t, itemStruct(m1), `{"Item":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}}}`)
	m2 := map[string]interface{}{
		"Map": map[string]interface{}{
			"Int":    4,
			"String": "this is a string"},
		"Nil": nil}
	testMarshal(t, itemMap(m2), `{"Item":{"M":{"Map":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}},"Nil":{"NULL":"true"}}}}`)
	testMarshal(t, itemStruct(m2), `{"Item":{"M":{"Map":{"M":{"Int":{"N":"4"},"String":{"S":"this is a string"}}},"Nil":{"NULL":"true"}}}}`)
}

func TestMarshalArray(t *testing.T) {
	var a [5]interface{}
	a[0] = 3.14
	a[1] = -2
	a[2] = true
	a[3] = "and a string!"
	a[4] = nil
	testMarshal(t, itemMap(a), `{"Item":{"L":[{"N":"3.14"},{"N":"-2"},{"BOOL":true},{"S":"and a string!"},{"NULL":"true"}]}}`)
	testMarshal(t, itemStruct(a), `{"Item":{"L":[{"N":"3.14"},{"N":"-2"},{"BOOL":true},{"S":"and a string!"},{"NULL":"true"}]}}`)
	// Test again treating the array as a slice.
	var s []interface{} = a[:]
	testMarshal(t, itemMap(s), `{"Item":{"L":[{"N":"3.14"},{"N":"-2"},{"BOOL":true},{"S":"and a string!"},{"NULL":"true"}]}}`)
	testMarshal(t, itemStruct(s), `{"Item":{"L":[{"N":"3.14"},{"N":"-2"},{"BOOL":true},{"S":"and a string!"},{"NULL":"true"}]}}`)
}

func TestMarshalSlice(t *testing.T) {
	s := []interface{}{3.14, -2, true, "and a string!", nil}
	testMarshal(t, itemMap(s), `{"Item":{"L":[{"N":"3.14"},{"N":"-2"},{"BOOL":true},{"S":"and a string!"},{"NULL":"true"}]}}`)
	testMarshal(t, itemStruct(s), `{"Item":{"L":[{"N":"3.14"},{"N":"-2"},{"BOOL":true},{"S":"and a string!"},{"NULL":"true"}]}}`)
}
func TestMarshalPtr(t *testing.T) {
	// Pointers aren't supported, this should be an error (and not a panic).
	m := make([]int, 1)
	ptr1 := *(*uintptr)(unsafe.Pointer(&m))
	testMarshalError(t, itemMap(ptr1), errors.New("the type uintptr is not supported"))
	ptr2 := *(**int)(unsafe.Pointer(&m))
	testMarshalError(t, itemStruct(ptr2), errors.New("the type ptr is not supported"))
}