package entity

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestUnmarshalString(t *testing.T) {
	testUnmarshal(t, "this is a string", `{"S":"this is a string"}`)
}

func TestUnmarshalNumber(t *testing.T) {
	testUnmarshal(t, 3.12234234, `{"N":"3.12234234"}`)
}

func TestUnmarshalBool(t *testing.T) {
	testUnmarshal(t, true, `{"BOOL":"true"}`)
}

func TestUnmarshalNull(t *testing.T) {
	testUnmarshal(t, nil, `{"NULL":"true"}`)
}

func TestUnmarshalMap(t *testing.T) {
	testUnmarshal(t, map[string]interface{}{}, `{"M":{}}`)
	testUnmarshal(t, map[string]interface{}{"id": "test"}, `{"M":{"id":{"S":"test"}}}`)
	testUnmarshal(t, map[string]interface{}{"id": nil}, `{"M":{"id":{"NULL":"true"}}}`)
	testUnmarshal(t, map[string]interface{}{"id": float64(12), "data": map[string]interface{}{}}, `{"M":{"id":{"N":"12"},"data":{"M":{}}}}`)
}

func TestUnmarshalList(t *testing.T) {
	testUnmarshal(t, []interface{}{}, `{"L":[]}`)
	testUnmarshal(t, []interface{}{float64(1), 3.14, nil}, `{"L":[{"N":"1"},{"N":"3.14"},{"NULL":"true"}]}`)
	testUnmarshal(t, []interface{}{true, false, "this is a string"}, `{"L":[{"BOOL":"true"},{"BOOL":"false"},{"S":"this is a string"}]}`)
}

func testUnmarshal(t *testing.T, expected interface{}, s string) {
	var data map[string]interface{}
	var buf bytes.Buffer
	buf.WriteString(s)
	err := json.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
	actual := unmarshal(data)
	compareObjects(t, expected, actual)
}

// Data that gets passed into entity.Unmarshal is a pure JSON object which gets
// wrapped into the Dynamo format. Test that here.
func TestUnmarshal(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString(`{"Attr1":{"S":"Attr1Val"},"TestHashKey":{"S":"NewHashKeyVal"}}`)
	var actual interface{}
	err := Unmarshal(buf.Bytes(), &actual)
	if err != nil {
		t.Error(err)
	}
	expected := map[string]interface{}{
		"Attr1":       "Attr1Val",
		"TestHashKey": "NewHashKeyVal"}
	compareObjects(t, expected, actual)
}

func compareObjects(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		ab, _ := json.Marshal(actual)
		eb, _ := json.Marshal(expected)
		t.Errorf("Expected %v, got %s", string(eb), string(ab))
	}
}
