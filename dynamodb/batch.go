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
	var response DynamoBatchResponse
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
	m := make(map[Key]dynamizer.DynamoItem)
	for _, item := range response.Responses[t.Name] {
		key, err := t.getKeyFromItem(item)
		if err != nil {
			return err, nil
		}
		t.deleteKeyFromItem(item)
		m[key] = item
	}

	// TODO: Handle unprocessed keys. Simplest method may be to return a
	// special error code, so that the caller can decide how to handle the
	// partial result. This would allow callers to utilize the responses that
	// were returned immediately.

	errs := make([]error, numKeys)
	for i, key := range keys {
		if item, ok := m[*key]; ok {
			errs[i] = dynamizer.FromDynamo(item, rv.Index(i))
		} else {
			errs[i] = ErrNotFound
		}
	}

	return nil, errs
}
