package entity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

func Marshal(v interface{}) (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			b = nil
			err = r.(error)
		}
	}()
	var buf bytes.Buffer
	marshal(&buf, reflect.ValueOf(v))
	return buf.Bytes(), nil
}

func marshal(buf *bytes.Buffer, v reflect.Value) {
	t := v.Type()
	switch t.Kind() {
	case reflect.Bool:
		marshalBool(buf, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		marshalInt(buf, v)
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		marshalFloat(buf, v)
	case reflect.String:
		marshalString(buf, v)
	case reflect.Struct:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		fallthrough
	case reflect.Interface, reflect.Ptr, reflect.Uintptr:
		fallthrough
	default:
		panic(errors.New(fmt.Sprintf(`the type %s is not supported`, t.Kind())))
	}
}

func marshalBool(buf *bytes.Buffer, v reflect.Value) {
	if v.Bool() {
		buf.WriteString(`{"BOOL":true}`)
	} else {
		buf.WriteString(`{"BOOL":false}`)
	}
}

func marshalString(buf *bytes.Buffer, v reflect.Value) {
	b, err := json.Marshal(v.String())
	if err != nil {
		panic(err)
	}
	buf.WriteString(fmt.Sprintf(`{"S":%s}`, string(b)))
}

func marshalInt(buf *bytes.Buffer, v reflect.Value) {
	buf.WriteString(fmt.Sprintf(`{"N":"%d"}`, v.Int()))
}

func marshalFloat(buf *bytes.Buffer, v reflect.Value) {
	b, err := json.Marshal(v.Float())
	if err != nil {
		panic(err)
	}
	buf.WriteString(fmt.Sprintf(`{"N":"%s"}`, string(b)))
}