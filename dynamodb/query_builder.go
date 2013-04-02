package dynamodb

import (
	"bytes"
	"fmt"
	)

type Query struct {
	buffer *bytes.Buffer
}

func NewEmptyQuery() *Query {
	q := &Query{new(bytes.Buffer)}
	q.buffer.WriteString("")
	return q
}

func NewQuery(t *Table) *Query {	
	q := &Query{new(bytes.Buffer)}
	q.addTable(t)
	return q
}

// This way of specifing the key is used when doing a Get.
func (q *Query) AddKey(t *Table, hashKey string, rangeKey string) {
	b := q.buffer
	k := t.Key

	addComma(b)

	b.WriteString(quote("Key"))
	b.WriteString(":")

	b.WriteString("{")
	b.WriteString(quote("HashKeyElement"))
	b.WriteString(":")

	b.WriteString("{")
	b.WriteString(quote(k.KeyAttribute.Type))
	b.WriteString(":")
	b.WriteString(quote(hashKey))	

	b.WriteString("}")
	
	if k.HasRange() {
		b.WriteString(",")
		b.WriteString(quote("RangeKeyElement"))
		b.WriteString(":")

		b.WriteString("{")
		b.WriteString(quote(k.RangeAttribute.Type))
		b.WriteString(":")
		b.WriteString(quote(rangeKey))
		b.WriteString("}")
	}

	b.WriteString("}")
}

func (q *Query) AddAttributesToGet(attributes []string){
	if len(attributes) == 0 {
		return 
	}
	b := q.buffer
	addComma(b)

	b.WriteString(quote("AttributesToGet"))
	b.WriteString(":")
	
	b.WriteString("[")
	
	for index, val := range attributes {
		if index > 0 {
			b.WriteString(",")
		}
		b.WriteString(quote(val))
	}

	b.WriteString("]")	
}

func(q *Query) ConsistentRead(c bool){
	if c == true {
		b := q.buffer
		addComma(b)
		
		b.WriteString(quote("ConsistentRead"))
		b.WriteString(":")
		b.WriteString("true")
	}
}

// The primary key must be included in attributes.
func (q *Query) AddItem(attributes []Attribute) {
	b := q.buffer

	addComma(b)

	b.WriteString(quote("Item"))
	b.WriteString(":")
	
	attributeList(b, attributes)
}

func (q *Query) AddUpdates(attributes []Attribute, action string){
	b := q.buffer

	addComma(b)

	b.WriteString(quote("AttributeUpdates"))
	b.WriteString(":")

	b.WriteString("{")
	for index, a := range attributes {
		if index > 0 {
			b.WriteString(",")
		}

		b.WriteString(quote(a.Name))
		b.WriteString(":")
		b.WriteString("{")
		b.WriteString(quote("Value"))
		b.WriteString(":")
		b.WriteString("{")
		b.WriteString(quote(a.Type))
		b.WriteString(":")
		b.WriteString(quote(a.Value))
		b.WriteString("}")
		b.WriteString(",")
		b.WriteString(quote("Action"))
		b.WriteString(":")
		b.WriteString(quote(action))
		b.WriteString("}")
	}
	b.WriteString("}")	
}

func (q *Query) AddExpected(attributes []Attribute){
	b := q.buffer
	addComma(b)

	b.WriteString(quote("Expected"))
	b.WriteString(":")
	b.WriteString("{")

	for index, a := range attributes {
		if index > 0 {
			b.WriteString(",")
		}

		b.WriteString(quote(a.Name))
		b.WriteString(":")
		
		b.WriteString("{")
		b.WriteString(quote("Value"))
		b.WriteString(":")
		b.WriteString("{")
		b.WriteString(quote(a.Type))
		b.WriteString(":")
		b.WriteString(quote(a.Value))
		b.WriteString("}")
		b.WriteString("}")
	}

	b.WriteString("}")
}

func attributeList(b *bytes.Buffer, attributes []Attribute){
	b.WriteString("{")
	for index, a := range attributes {
		if index > 0 {
			b.WriteString(",")
		}
		
		b.WriteString(quote(a.Name))
		b.WriteString(":")
		
		b.WriteString("{")
		b.WriteString(quote(a.Type))
		b.WriteString(":")
		b.WriteString(quote(a.Value))
		b.WriteString("}")
	}
	b.WriteString("}")
}

func (q *Query) addTable(t *Table) {
	q.buffer.WriteString(keyValue("TableName", t.Name))
}

func quote(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

func addComma(b *bytes.Buffer) {
	if b.Len() != 0 {
		b.WriteString(",")
	}
}

func (q *Query) String() string {
	qs := fmt.Sprintf("{%s}", q.buffer.String())
	return qs
}
