package dynamodb_test

import (
	"github.com/alimoeeny/goamz/aws"
	"github.com/alimoeeny/goamz/dynamodb"
	simplejson "github.com/bitly/go-simplejson"
	"os"
	"testing"
)

var (
	AWS_KEY    = os.Getenv("AWS_ACCESS_KEY_ID")
	AWS_SECRET = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

func TestEmptyQuery(t *testing.T) {
	q := NewEmptyQuery()
	queryString := q.String()
	expectedString := "{}"

	if expectedString != queryString {
		t.Fatalf("Unexpected Query String : %s\n", queryString)
	}

}

func TestAddWriteRequestItems(t *testing.T) {
	auth := &aws.Auth{AccessKey: AWS_KEY, SecretKey: AWS_SECRET}
	server := Server{*auth, aws.USEast}
	primary := NewStringAttribute("WidgetFoo", "")
	secondary := NewNumericAttribute("Created", "")
	key := PrimaryKey{primary, secondary}
	table := server.NewTable("FooData", key)

	primary2 := NewStringAttribute("TestHashKey", "")
	secondary2 := NewNumericAttribute("TestRangeKey", "")
	key2 := PrimaryKey{primary2, secondary2}
	table2 := server.NewTable("TestTable", key2)

	q := NewEmptyQuery()

	attribute1 := NewNumericAttribute("testing", "4")
	attribute2 := NewNumericAttribute("testingbatch", "2111")
	attribute3 := NewStringAttribute("testingstrbatch", "mystr")
	item1 := []Attribute{*attribute1, *attribute2, *attribute3}

	attribute4 := NewNumericAttribute("testing", "444")
	attribute5 := NewNumericAttribute("testingbatch", "93748249272")
	attribute6 := NewStringAttribute("testingstrbatch", "myotherstr")
	item2 := []Attribute{*attribute4, *attribute5, *attribute6}

	attributeDel1 := NewStringAttribute("TestHashKeyDel", "DelKey")
	attributeDel2 := NewNumericAttribute("TestRangeKeyDel", "7777777")
	itemDel := []Attribute{*attributeDel1, *attributeDel2}

	attributeTest1 := NewStringAttribute("TestHashKey", "MyKey")
	attributeTest2 := NewNumericAttribute("TestRangeKey", "0193820384293")
	itemTest := []Attribute{*attributeTest1, *attributeTest2}

	tableItems := map[*Table]map[string][][]Attribute{}
	actionItems := make(map[string][][]Attribute)
	actionItems["Put"] = [][]Attribute{item1, item2}
	actionItems["Delete"] = [][]Attribute{itemDel}
	tableItems[table] = actionItems

	actionItems2 := make(map[string][][]Attribute)
	actionItems2["Put"] = [][]Attribute{itemTest}
	tableItems[table2] = actionItems2

	q.AddWriteRequestItems(tableItems)

	desiredString := "\"RequestItems\":{\"FooData\":[{\"PutRequest\":{\"Item\":{\"testing\":{\"N\":\"4\"},\"testingbatch\":{\"N\":\"2111\"},\"testingstrbatch\":{\"S\":\"mystr\"}}}},{\"PutRequest\":{\"Item\":{\"testing\":{\"N\":\"444\"},\"testingbatch\":{\"N\":\"93748249272\"},\"testingstrbatch\":{\"S\":\"myotherstr\"}}}},{\"DeleteRequest\":{\"Key\":{\"TestHashKeyDel\":{\"S\":\"DelKey\"},\"TestRangeKeyDel\":{\"N\":\"7777777\"}}}}],\"TestTable\":[{\"PutRequest\":{\"Item\":{\"TestHashKey\":{\"S\":\"MyKey\"},\"TestRangeKey\":{\"N\":\"0193820384293\"}}}}]}"
	queryString := q.buffer.String()

	if queryString != desiredString {
		t.Fatalf("Unexpected Query String : %s\n", queryString)
	}
}

func TestGetItemQuery(t *testing.T) {
	auth := &aws.Auth{AccessKey: AWS_KEY, SecretKey: AWS_SECRET}
	server := Server{*auth, aws.USEast}
	primary := NewStringAttribute("WidgetFoo", "")
	secondary := NewNumericAttribute("Created", "")
	key := PrimaryKey{primary, secondary}
	table := server.NewTable("FooData", key)

	q := NewQuery(table)
	q.AddKey(table, &Key{HashKey: "test"})

	queryString := []byte(q.String())

	json, err := simplejson.NewJson(queryString)

	if err != nil {
		t.Logf("JSON err : %s\n", err)
		t.Fatalf("Invalid JSON : %s\n", queryString)
	}

	tableName := json.Get("TableName").MustString()

	if tableName != "FooData" {
		t.Fatalf("Expected tableName to be sites was : %s", tableName)
	}

	keyMap, err := json.Get("Key").Map()

	if err != nil {
		t.Fatalf("Expected a Key")
	}

	hashRangeKey := keyMap["Created"]

	if hashRangeKey == nil {
		t.Fatalf("Expected a HashKeyElement found : %s", keyMap)
	}

	if v, ok := hashRangeKey.(map[string]interface{}); ok {
		if val, ok := v["S"].(string); ok {
			if val != "test" {
				t.Fatalf("Expected HashKeyElement to have the value 'test' found : %s", val)
			}
		}
	} else {
		t.Fatalf("HashRangeKeyt had the wrong type found : %s", hashRangeKey)
	}
}

/*func TestUpdateQuery(t *testing.T) {
	auth := &aws.Auth{AccessKey: AWS_KEY, SecretKey: AWS_SECRET}
	server := Server{*auth, aws.USEast}
	primary := NewStringAttribute("WidgetFoo", "")
	secondary := NewNumericAttribute("Created", "")
	key := PrimaryKey{primary, secondary}
	table := server.NewTable("FooData", key)

	countAttribute := NewNumericAttribute("count", "4")
	attributes := []Attribute{*countAttribute}

	q := NewQuery(table)
	q.AddKey(table, &Key{HashKey: "1:test", RangeKey: "1234374638364"})
	q.AddUpdates(attributes, "ADD")

	queryString := []byte(q.String())

	json, err := simplejson.NewJson(queryString)

	if err != nil {
		t.Logf("JSON err : %s\n", err)
		t.Fatalf("Invalid JSON : %s\n", queryString)
	}

	tableName := json.Get("TableName").MustString()

	if tableName != "sites" {
		t.Fatalf("Expected tableName to be sites was : %s", tableName)
	}

	keyMap, err := json.Get("Key").Map()

	if err != nil {
		t.Fatalf("Expected a Key")
	}

	hashRangeKey := keyMap["HashKeyElement"]

	if hashRangeKey == nil {
		t.Fatalf("Expected a HashKeyElement found : %s", keyMap)
	}

	rangeKey := keyMap["RangeKeyElement"]

	if rangeKey == nil {
		t.Fatalf("Expected a RangeKeyElement found : %s", keyMap)
	}

}*/
