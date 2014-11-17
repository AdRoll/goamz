package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

// Unmarshal parses the DynamoDB-encoded data and stores the result in the value
// pointed to by v.
func Unmarshal(data []byte, v *interface{}) error {
	var src map[string]interface{}
	err := json.Unmarshal(data, &src)
	if err != nil {
		return err
	}
	return UnmarshalMap(src, v)
}

// Unmarshal parses the DynamoDB-encoded data which has been unmarshalled into
// a map[string]inteface{} and stores the result in the value pointed to by v.
//
// This method is provided as an optimization, since the expected use involves
// first JSON-decoding the data.
func UnmarshalMap(src map[string]interface{}, v *interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(runtime.Error); ok {
				err = e
			} else if s, ok := r.(string); ok {
				err = errors.New(s)
			} else {
				err = r.(error)
			}
		}
	}()
	*v = unmarshalMap(map[string]interface{}{dynamoMap: src})
	return nil
}

func unmarshal(src map[string]interface{}) interface{} {
	if _, ok := src[dynamoString]; ok {
		return unmarshalString(src)
	} else if _, ok := src[dynamoNumber]; ok {
		return unmarshalNumber(src)
	} else if _, ok := src[dynamoBool]; ok {
		return unmarshalBool(src)
	} else if _, ok := src[dynamoNull]; ok {
		return unmarshalNull(src)
	} else if _, ok := src[dynamoMap]; ok {
		return unmarshalMap(src)
	} else if _, ok := src[dynamoList]; ok {
		return unmarshalList(src)
	}

	// If we make it here the data type is unsupported.
	for k, _ := range src {
		panic(errors.New(fmt.Sprintf(`The type %s is not supported`, k)))
	}

	// If we make it here no data type was provided.
	panic(errors.New(`Invalid format`))
}

func unmarshalString(src map[string]interface{}) interface{} {
	return src[dynamoString].(string)
}

func unmarshalNumber(src map[string]interface{}) interface{} {
	f, err := strconv.ParseFloat(src[dynamoNumber].(string), 64)
	if err != nil {
		panic(err)
	}
	return f
}

func unmarshalBool(src map[string]interface{}) interface{} {
	return src[dynamoBool].(string) == "true"
}

func unmarshalNull(src map[string]interface{}) interface{} {
	return nil
}

func unmarshalMap(src map[string]interface{}) interface{} {
	m := src[dynamoMap].(map[string]interface{})
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = unmarshal(v.(map[string]interface{}))
	}
	return result
}

func unmarshalList(src map[string]interface{}) interface{} {
	l := src[dynamoList].([]interface{})
	result := make([]interface{}, len(l))
	for index, v := range l {
		result[index] = unmarshal(v.(map[string]interface{}))
	}
	return result
}
