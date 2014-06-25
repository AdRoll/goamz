package dynamostore

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/flowhealth/goamz/dynamodb"
	"strconv"
	"strings"
	"time"
)

const (
	NullString = "NULL"
	timePrefix = "time_"
)

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

func MakeFloat64Attr(name string, value float64) dynamodb.Attribute {
	return *dynamodb.NewNumericAttribute(name, strconv.FormatFloat(value, 'f', -1, 64))
}

func MakeBinaryAttr(name string, value []byte) dynamodb.Attribute {
	b64val := base64.StdEncoding.EncodeToString(value)
	return *dynamodb.NewBinaryAttribute(name, b64val)
}

func MakeTimeTimeAttr(name string, value time.Time) dynamodb.Attribute {
	return *dynamodb.NewNumericAttribute(name, strconv.FormatInt(value.Unix(), 10))
}

func MakeTimeTimeNanoAttr(name string, value time.Time) dynamodb.Attribute {
	return *dynamodb.NewNumericAttribute(name, strconv.FormatInt(value.UnixNano(), 10))
}

func MakeTimePrefixedAttr(name string, value time.Time) dynamodb.Attribute {
	b := bytes.NewBufferString(timePrefix)
	b.WriteString(strconv.FormatInt(value.UnixNano(), 10))
	return *dynamodb.NewStringAttribute(name, b.String())
}

func GetBinaryAttr(name string, attrs map[string]*dynamodb.Attribute) ([]byte, error) {
	if val, ok := attrs[name]; !ok {
		return nil, MakeAttrNotFoundErr(name)
	} else {
		if val.Value == NullString {
			return []byte{}, nil
		} else {
			if binVal, err := base64.StdEncoding.DecodeString(val.Value); err != nil {
				return nil, err
			} else {
				return binVal, nil
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

func GetFloat64Attr(name string, attrs map[string]*dynamodb.Attribute) (v float64, err error) {
	var v64 float64
	if val, ok := attrs[name]; !ok {
		err = MakeAttrNotFoundErr(name)
		return
	} else {
		if v64, err = strconv.ParseFloat(val.Value, 64); err != nil {
			err = MakeAttrInvalidErr(name, val.Value)
		} else {
			return v64, nil
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

func GetTimeTimeNanoAttr(name string, attrs map[string]*dynamodb.Attribute) (t time.Time, err error) {
	var timestamp int64
	if val, ok := attrs[name]; !ok {
		err = MakeAttrNotFoundErr(name)
		return
	} else {
		if timestamp, err = strconv.ParseInt(val.Value, 10, 64); err != nil {
			err = MakeAttrInvalidErr(name, val.Value)
		} else {
			t = time.UnixNano(0, timestamp)
		}
		return
	}
}

func GetTimePrefixedAttr(name string, attrs map[string]*dynamodb.Attribute) (t time.Time, err error) {
	var timestamp int64
	if val, ok := attrs[name]; !ok {
		err = MakeAttrNotFoundErr(name)
		return
	} else {
		if timestamp, err = strconv.ParseInt(strings.TrimPrefix(val.Value, timePrefix), 10, 64); err != nil {
			err = MakeAttrInvalidErr(name, val.Value)
		} else {
			t = time.Unix(timestamp/1e9, timestamp%1e9)
		}
		return
	}
}

func IsTimePrefixedAttr(attr *dynamodb.Attribute) bool {
	return strings.LastIndex(attr.Name, timePrefix) == 0
}
