package dynamodb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/AdRoll/goamz/dynamodb/dynamizer"
)

func (t *Table) BatchGetDocument(keys []*Key, consistentRead bool, v interface{}) (error, []error) {
	numKeys := len(keys)

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("v must be a slice with the same length as keys"), nil
	} else if rv.Len() != numKeys {
		return fmt.Errorf("v must be a slice with the same length as keys"), nil
	}

	q := NewDynamoBatchGetQuery(t)
	for _, key := range keys {
		if err := q.AddKey(key); err != nil {
			return err, nil
		}
	}

	if consistentRead {
		q.SetConsistentRead(consistentRead)
	}

	jsonResponse, err := t.Server.queryServer(target("BatchGetItem"), q)
	if err != nil {
		return err, nil
	}

	// Deserialize from []byte to JSON.
	var response DynamoBatchGetResponse
	err = json.Unmarshal(jsonResponse, &response)
	if err != nil {
		return err, nil
	}

	// DynamoDB doesn't return the items in any particular order, but we promise
	// callers that we will. So we build a map of key to resposne to match up
	// inputs to return values.
	//
	// N.B. The map is of type Key - not *Key - so that equality is based on the
	// hash and range key values, not the pointer location.
	responses := make(map[Key]dynamizer.DynamoItem)
	for _, item := range response.Responses[t.Name] {
		key, err := t.getKeyFromItem(item)
		if err != nil {
			return err, nil
		}
		t.deleteKeyFromItem(item)
		responses[key] = item
	}

	// Handle unprocessed keys. We return a special error code so that the
	// caller can decide how to handle the partial result. This allows callers
	// to utilize the responses we do have available right away.
	unprocessed := make(map[Key]bool)
	if r, ok := response.UnprocessedKeys[t.Name]; ok {
		for _, item := range r.Keys {
			key, err := t.getKeyFromItem(item)
			if err != nil {
				return err, nil
			}
			unprocessed[key] = true
		}
	}

	// Package the final response maintaining the original ordering as specified
	// by the caller.
	errs := make([]error, numKeys)
	for i, key := range keys {
		if item, ok := responses[*key]; ok {
			errs[i] = dynamizer.FromDynamo(item, rv.Index(i))
		} else if _, ok := unprocessed[*key]; ok {
			errs[i] = ErrNotProcessed
		} else {
			errs[i] = ErrNotFound
		}
	}

	return nil, errs
}

func (t *Table) BatchPutDocument(keys []*Key, v interface{}) (error, []error) {
	numKeys := len(keys)

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("v must be a slice with the same length as keys"), nil
	} else if rv.Len() != numKeys {
		return fmt.Errorf("v must be a slice with the same length as keys"), nil
	}

	q := NewDynamoBatchPutQuery(t)
	for i, key := range keys {
		item, err := dynamizer.ToDynamo(rv.Index(i).Interface())
		if err != nil {
			return err, nil
		}
		if err := q.AddItem(key, item); err != nil {
			return err, nil
		}
	}

	jsonResponse, err := t.Server.queryServer(target("BatchWriteItem"), q)
	if err != nil {
		return err, nil
	}

	// Deserialize from []byte to JSON.
	var response DynamoBatchPutResponse
	err = json.Unmarshal(jsonResponse, &response)
	if err != nil {
		return err, nil
	}

	// Handle unprocessed items. We return a special error code so that the
	// caller can decide how to handle the partial result. This allows callers
	// to move on from successful writes immediately.
	unprocessed := make(map[Key]bool)
	if r, ok := response.UnprocessedItems[t.Name]; ok {
		for _, item := range r {
			key, err := t.getKeyFromItem(item.PutRequest.Item)
			if err != nil {
				return err, nil
			}
			unprocessed[key] = true
		}
	}

	// Package the final response maintaining the original ordering as specified
	// by the caller.
	errs := make([]error, numKeys)
	for i, key := range keys {
		if _, ok := unprocessed[*key]; ok {
			errs[i] = ErrNotProcessed
		}
	}

	return nil, errs
}
