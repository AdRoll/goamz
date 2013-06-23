package dynamodb

import (
	"errors"
	"fmt"
  simplejson "github.com/bitly/go-simplejson"
)

func (t *Table) Query(hashKey string, attributeComparisons []AttributeComparison) ([]map[string]*Attribute, error) {
	q := NewQuery(t)
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

  b.WriteString("}")
  
  q.AddKeyConditions(attributeComparisons)
	jsonResponse, err := t.Server.queryServer(target("Query"), q)
	if err != nil { return nil, err	}

	json, err := simplejson.NewJson(jsonResponse)
	if err != nil { return nil, err	}

  itemCount, err := json.Get("Count").Int()
	if err != nil {
		message := fmt.Sprintf("Unexpected response %s", jsonResponse)
		return nil, errors.New(message)
	}

  results := make([]map[string]*Attribute, itemCount)

  for i, _ := range results {
  	item, err := json.Get("Items").GetIndex(i).Map()
  	if err != nil {
  		message := fmt.Sprintf("Unexpected response %s", jsonResponse)
  		return nil, errors.New(message)
  	}
    results[i] = parseAttributes(item)
  }
	return results, nil
}
