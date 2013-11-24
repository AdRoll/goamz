package dynamodb

import simplejson "github.com/bitly/go-simplejson"
import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

type BatchGetItem struct {
	Server *Server
	Keys   map[*Table][]Key
}

type BatchWriteItem struct {
	Server      *Server
	ItemActions map[*Table]map[string][][]Attribute
}

func (t *Table) BatchGetItems(keys []Key) *BatchGetItem {
	batchGetItem := &BatchGetItem{t.Server, make(map[*Table][]Key)}

	batchGetItem.Keys[t] = keys
	return batchGetItem
}

func (t *Table) BatchWriteItems(itemActions map[string][][]Attribute) *BatchWriteItem {
	batchWriteItem := &BatchWriteItem{t.Server, make(map[*Table]map[string][][]Attribute)}

	batchWriteItem.ItemActions[t] = itemActions
	return batchWriteItem
}

func (batchGetItem *BatchGetItem) AddTable(t *Table, keys *[]Key) *BatchGetItem {
	batchGetItem.Keys[t] = *keys
	return batchGetItem
}

func (batchWriteItem *BatchWriteItem) AddTable(t *Table, itemActions *map[string][][]Attribute) *BatchWriteItem {
	batchWriteItem.ItemActions[t] = *itemActions
	return batchWriteItem
}

func (batchGetItem *BatchGetItem) Execute() (map[string][]map[string]*Attribute, error) {
	q := NewEmptyQuery()
	q.AddGetRequestItems(batchGetItem.Keys)

	jsonResponse, err := batchGetItem.Server.queryServer("DynamoDB_20120810.BatchGetItem", q)
	if err != nil {
		return nil, err
	}

	json, err := simplejson.NewJson(jsonResponse)

	if err != nil {
		return nil, err
	}

	results := make(map[string][]map[string]*Attribute)

	tables, err := json.Get("Responses").Map()
	if err != nil {
		message := fmt.Sprintf("Unexpected response %s", jsonResponse)
		return nil, errors.New(message)
	}

	for table, entries := range tables {
		var tableResult []map[string]*Attribute

		jsonEntriesArray, ok := entries.([]interface{})
		if !ok {
			message := fmt.Sprintf("Unexpected response %s", jsonResponse)
			return nil, errors.New(message)
		}

		for _, entry := range jsonEntriesArray {
			item, ok := entry.(map[string]interface{})
			if !ok {
				message := fmt.Sprintf("Unexpected response %s", jsonResponse)
				return nil, errors.New(message)
			}

			unmarshalledItem := parseAttributes(item)
			tableResult = append(tableResult, unmarshalledItem)
		}

		results[table] = tableResult
	}

	return results, nil
}

func (batchWriteItem *BatchWriteItem) Execute() (map[string]interface{}, error) {
	q := NewEmptyQuery()
	q.AddWriteRequestItems(batchWriteItem.ItemActions)

	jsonResponse, err := batchWriteItem.Server.queryServer("DynamoDB_20120810.BatchWriteItem", q)

	if err != nil {
		return nil, err
	}

	json, err := simplejson.NewJson(jsonResponse)

	if err != nil {
		return nil, err
	}

	unprocessed, err := json.Get("UnprocessedItems").Map()
	if err != nil {
		message := fmt.Sprintf("Unexpected response %s", jsonResponse)
		return nil, errors.New(message)
	}

	if len(unprocessed) == 0 {
		return nil, nil
	} else {
		return unprocessed, errors.New("One or more unprocessed items.")
	}

}

func (t *Table) GetItem(key *Key) (map[string]*Attribute, error) {
	q := NewQuery(t)
	q.AddKey(t, key)

	jsonResponse, err := t.Server.queryServer(target("GetItem"), q)
	if err != nil {
		return nil, err
	}

	json, err := simplejson.NewJson(jsonResponse)
	if err != nil {
		return nil, err
	}

	itemJson, ok := json.CheckGet("Item")
	if !ok {
		// We got an empty from amz. The item doesn't exist.
		return nil, ErrNotFound
	}

	item, err := itemJson.Map()
	if err != nil {
		message := fmt.Sprintf("Unexpected response %s", jsonResponse)
		return nil, errors.New(message)
	}

	return parseAttributes(item), nil

}

func (t *Table) PutItem(hashKey string, rangeKey string, attributes []Attribute) (bool, error) {

	if len(attributes) == 0 {
		return false, errors.New("At least one attribute is required.")
	}

	q := NewQuery(t)

	keys := t.Key.Clone(hashKey, rangeKey)
	attributes = append(attributes, keys...)

	q.AddItem(attributes)
	//Ali - debug
	fmt.Println("++++++++++++++++++++++++++++++++++++++")
	fmt.Println("q:", q)
	fmt.Println("q-attributes:", attributes)
	fmt.Println("++++++++++++++++++++++++++++++++++++++")

	jsonResponse, err := t.Server.queryServer(target("PutItem"), q)

	//ALI
	fmt.Printf("AMZ response: %s\n", string(jsonResponse))
	var amzResponse map[string]interface{}
	err = json.Unmarshal(jsonResponse, &amzResponse)
	if err != nil {
		log.Println("Error Processing Amazon response as JSON:", err)
	} else {
		if amzResponse["ConsumedCapacityUnits"] != "" {
			r := map[string]float64{t.Name: amzResponse["ConsumedCapacityUnits"].(float64)}
			t.Server.CapacityChannel <- r
			fmt.Printf("Reported AMZ response: %v\n", r)
		}
	}
	if strings.Index(strings.ToLower(string(jsonResponse)), "exception") > -1 {
		resp := make(map[string]string)
		err := json.Unmarshal(jsonResponse, &resp)
		if err != nil {
			return false, err
		}
		log.Printf("An exception happened: %s - %s \n", resp["__type"], resp["message"])
		log.Println("resp:", resp)
		log.Println("hashkey:", hashKey)
		log.Println("rangekey:", rangeKey)
		log.Println("attributes:", attributes)
		return false, errors.New(resp["message"])
	}

	if err != nil {
		return false, err
	}

	json, err := simplejson.NewJson(jsonResponse)

	if err != nil {
		return false, err
	}

	units, _ := json.CheckGet("ConsumedCapacityUnits")

	if units == nil {
		message := fmt.Sprintf("Unexpected response %s", jsonResponse)
		return false, errors.New(message)
	}

	return true, nil
}

func (t *Table) DeleteItem(key *Key) (bool, error) {

	q := NewQuery(t)
	q.AddKey(t, key)

	//Ali - debug
	fmt.Println("++++++++++++++++++++++++++++++++++++++")
	fmt.Println("q:", q)
	fmt.Println("++++++++++++++++++++++++++++++++++++++")

	jsonResponse, err := t.Server.queryServer(target("DeleteItem"), q)

	//ALI
	fmt.Printf("AMZ response: %s\n", string(jsonResponse))

	if err != nil {
		return false, err
	}

	json, err := simplejson.NewJson(jsonResponse)

	if err != nil {
		return false, err
	}

	units, _ := json.CheckGet("ConsumedCapacityUnits")

	if units == nil {
		message := fmt.Sprintf("Unexpected response %s", jsonResponse)
		return false, errors.New(message)
	}

	return true, nil

}

func (t *Table) AddAttributes(key *Key, attributes []Attribute) (bool, error) {
	return t.modifyAttributes(key, attributes, "ADD")
}

func (t *Table) UpdateAttributes(key *Key, attributes []Attribute) (bool, error) {
	return t.modifyAttributes(key, attributes, "PUT")
}

func (t *Table) DeleteAttributes(key *Key, attributes []Attribute) (bool, error) {
	return t.modifyAttributes(key, attributes, "DELETE")
}

func (t *Table) modifyAttributes(key *Key, attributes []Attribute, action string) (bool, error) {

	if len(attributes) == 0 {
		return false, errors.New("At least one attribute is required.")
	}

	q := NewQuery(t)
	q.AddKey(t, key)
	q.AddUpdates(attributes, action)

	jsonResponse, err := t.Server.queryServer(target("UpdateItem"), q)

	if err != nil {
		return false, err
	}

	_, err = simplejson.NewJson(jsonResponse)
	if err != nil {
		return false, err
	}

	return true, nil
}

func parseAttributes(s map[string]interface{}) map[string]*Attribute {
	results := map[string]*Attribute{}

	for key, value := range s {
		if v, ok := value.(map[string]interface{}); ok {
			if val, ok := v[TYPE_STRING].(string); ok {
				results[key] = &Attribute{
					Type:  TYPE_STRING,
					Name:  key,
					Value: val,
				}
			} else if val, ok := v[TYPE_NUMBER].(string); ok {
				results[key] = &Attribute{
					Type:  TYPE_NUMBER,
					Name:  key,
					Value: val,
				}
			} else if val, ok := v[TYPE_BINARY].(string); ok {
				results[key] = &Attribute{
					Type:  TYPE_BINARY,
					Name:  key,
					Value: val,
				}
			} else if vals, ok := v[TYPE_STRING_SET].([]interface{}); ok {
				arry := make([]string, len(vals))
				for i, ivalue := range vals {
					if val, ok := ivalue.(string); ok {
						arry[i] = val
					}
				}
				results[key] = &Attribute{
					Type:      TYPE_STRING_SET,
					Name:      key,
					SetValues: arry,
				}
			} else if vals, ok := v[TYPE_NUMBER_SET].([]interface{}); ok {
				arry := make([]string, len(vals))
				for i, ivalue := range vals {
					if val, ok := ivalue.(string); ok {
						arry[i] = val
					}
				}
				results[key] = &Attribute{
					Type:      TYPE_NUMBER_SET,
					Name:      key,
					SetValues: arry,
				}
			} else if vals, ok := v[TYPE_BINARY_SET].([]interface{}); ok {
				arry := make([]string, len(vals))
				for i, ivalue := range vals {
					if val, ok := ivalue.(string); ok {
						arry[i] = val
					}
				}
				results[key] = &Attribute{
					Type:      TYPE_BINARY_SET,
					Name:      key,
					SetValues: arry,
				}
			}
		} else {
			log.Printf("type assertion to map[string] interface{} failed for : %s\n ", value)
		}

	}

	return results
}
