package dynamodb_test

import (
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"testing"
	"fmt"
)

func TestBatchWriteItem(t *testing.T) {
	auth := &aws.Auth{AccessKey: AWS_KEY, SecretKey: AWS_SECRET}
	server := Server{*auth, aws.USEast}
	primary := NewStringAttribute("WidgetFoo", "")
	secondary := NewNumericAttribute("Created", "")
	key := PrimaryKey{primary, secondary}
	table := server.NewTable("FooData", key)

	attribute1 := NewStringAttribute("WidgetFoo", "21:42")
	attribute2 := NewNumericAttribute("Created", "1257894000012239402")
	attribute3 := NewNumericAttribute("WidgetId", "21")
	attribute4 := NewNumericAttribute("FooId", "42")
	attribute5 := NewNumericAttribute("Value", "4.28762")
	item1 := []Attribute{*attribute1, *attribute2, *attribute3, *attribute4, *attribute5}

	attribute6 := NewStringAttribute("WidgetFoo", "3:81")
	attribute7 := NewNumericAttribute("Created", "1257894000011856639")
	attribute8 := NewNumericAttribute("WidgetId", "3")
	attribute9 := NewNumericAttribute("FooId", "81")
	attribute10 := NewNumericAttribute("Value", "624.2")
	item2 := []Attribute{*attribute6, *attribute7, *attribute8, *attribute9, *attribute10}

	itemActions := map[string][][]Attribute{}
	itemActions["Put"] = [][]Attribute{item1, item2}

	bput := table.BatchWriteItems(itemActions)

	out, err := bput.Execute()

	if out != nil {
		t.Fatalf("Unexpected unprocessed items: %v", out)
	} else if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	item1 = []Attribute{*attribute1, *attribute2}
	item2 = []Attribute{*attribute6, *attribute7}

	itemActions2 := map[string][][]Attribute{}
	itemActions2["Delete"] = [][]Attribute{item1, item2}

	bdel := table.BatchWriteItems(itemActions2)

	out2, err2 := bdel.Execute()

	fmt.Println(err2)
	fmt.Println(out2)
}