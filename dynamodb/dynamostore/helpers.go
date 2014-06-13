package dynamostore

import (
	"encoding/base64"
	"fmt"
	"github.com/flowhealth/goamz/dynamodb"
	"strconv"
	"time"
)

const NullString = "NULL"

func MakeAttrNotFoundErr(attr string) error {
	return fmt.Errorf("DeSerialization error: attribute %s not found")
}

func MakeAttrInvalidErr(attr, value string) error {
	return fmt.Errorf("DeSerialization error: attribute %s has unexpected value: %s", attr, value)
}

func MakeStringAttr(name string, value string) dynamodb.Attribute {
	if value == "" {
		return *dynamodb.NewStringAttribute(name, NullString)
	} else {
		return *dynamodb.NewStringAttribute(name, value)
	}
}

const (
	dynamoBoolTrue  = "1"
	dynamoBoolFalse = "0"
)

func MakeBoolAttr(name string, value bool) dynamodb.Attribute {
	var converted string
	if value == true {
		converted = dynamoBoolTrue
	} else {
		converted = dynamoBoolFalse
	}
	return *dynamodb.NewNumericAttribute(name, converted)
}

func MakeInt32Attr(name string, value int32) dynamodb.Attribute {
	return *dynamodb.NewNumericAttribute(name, strconv.FormatInt(int64(value), 10))
}

func MakeBinaryAttr(name string, value []byte) dynamodb.Attribute {
	b64val := base64.StdEncoding.EncodeToString(value)
	return *dynamodb.NewBinaryAttribute(name, b64val)
}

func MakeTimeTimeAttr(name string, value time.Time) dynamodb.Attribute {
	return *dynamodb.NewNumericAttribute(name, strconv.FormatInt(value.Unix(), 10))
}

func GetBinaryAttr(name string, attrs map[string]*dynamodb.Attribute) ([]byte, error) {
	if val, ok := attrs[name]; !ok {
		return nil, MakeAttrNotFoundErr(name)
	} else {
		if val.Value == NullString {
			return []byte{}, nil
		} else {
			if binVal, err := base64.StdEncoding.DecodeString(val.Value); err != nil {
				return binVal, nil
			} else {
				return nil, err
			}
		}
	}
}

func GetStringAttr(name string, attrs map[string]*dynamodb.Attribute) (string, error) {
	if val, ok := attrs[name]; !ok {
		return "", MakeAttrNotFoundErr(name)
	} else {
		if val.Value == NullString {
			return "", nil
		} else {
			return val.Value, nil
		}
	}
}

func GetBoolAttr(name string, attrs map[string]*dynamodb.Attribute) (bool, error) {
	if val, ok := attrs[name]; !ok {
		return false, MakeAttrNotFoundErr(name)
	} else {
		if val.Value == dynamoBoolTrue {
			return true, nil
		} else if val.Value == dynamoBoolFalse {
			return false, nil
		} else {
			return false, MakeAttrInvalidErr(name, val.Value)
		}
	}
}

func GetInt32Attr(name string, attrs map[string]*dynamodb.Attribute) (v int32, err error) {
	var v64 int64
	if val, ok := attrs[name]; !ok {
		err = MakeAttrNotFoundErr(name)
		return
	} else {
		if v64, err = strconv.ParseInt(val.Value, 10, 32); err != nil {
			err = MakeAttrInvalidErr(name, val.Value)
		} else {
			return int32(v64), nil
		}
		return
	}
}

func GetTimeTimeAttr(name string, attrs map[string]*dynamodb.Attribute) (t time.Time, err error) {
	var timestamp int64
	if val, ok := attrs[name]; !ok {
		err = MakeAttrNotFoundErr(name)
		return
	} else {
		if timestamp, err = strconv.ParseInt(val.Value, 10, 64); err != nil {
			err = MakeAttrInvalidErr(name, val.Value)
		} else {
			t = time.Unix(timestamp, 0)
		}
		return
	}
}
