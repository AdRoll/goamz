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
	rv := reflect.ValueOf(v)
	t := rv.Type()
	var buf bytes.Buffer
	switch t.Kind() {
	case reflect.Struct:
		marshalStruct(&buf, rv, true)
	case reflect.Map:
		marshalMap(&buf, rv, true)
	default:
		return nil, errors.New("top level object must be a struct or map")
	}
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
	case reflect.Float32, reflect.Float64:
		marshalFloat(buf, v)
	case reflect.String:
		marshalString(buf, v)
	case reflect.Interface:
		marshalInterface(buf, v)
	case reflect.Struct:
		marshalStruct(buf, v, false)
	case reflect.Map:
		marshalMap(buf, v, false)
	case reflect.Slice:
		marshalSlice(buf, v)
	case reflect.Array:
		marshalArray(buf, v)
	case reflect.Ptr, reflect.Uintptr:
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

func marshalInterface(buf *bytes.Buffer, v reflect.Value) {
	if v.IsNil() {
		buf.WriteString(`{"NULL":"true"}`)
		return
	}
	marshal(buf, v.Elem())
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

func marshalStruct(buf *bytes.Buffer, v reflect.Value, outer bool) {
	if outer {
		buf.WriteByte('{')
	} else {
		buf.WriteString(`{"M":{`)
	}
	t := v.Type()
	n := t.NumField()
	if n > 0 {
		first := true
		for i := 0; i < n; i++ {
			f := t.Field(i)
			if first {
				first = false
			} else {
				buf.WriteByte(',')
			}
			buf.WriteString(fmt.Sprintf(`"%s":`, f.Name))
			marshal(buf, v.FieldByIndex(f.Index))
		}
	}
	if outer {
		buf.WriteByte('}')
	} else {
		buf.WriteString(`}}`)
	}
}

func marshalMap(buf *bytes.Buffer, v reflect.Value, outer bool) {
	if v.IsNil() {
		if outer {
			panic(errors.New("outer map is nil"))
		}
		buf.WriteString(`{"NULL":"true"}`)
		return
	}
	if outer {
		buf.WriteByte('{')
	} else {
		buf.WriteString(`{"M":{`)
	}
	first := true
	for _, k := range v.MapKeys() {
		if first {
			first = false
		} else {
			buf.WriteString(`,`)
		}
		// Get the key as a string.
		var s string
		buf.WriteString(fmt.Sprintf(`"%s":`, k.Convert(reflect.TypeOf(s)).String()))
		marshal(buf, v.MapIndex(k))
	}
	if outer {
		buf.WriteByte('}')
	} else {
		buf.WriteString(`}}`)
	}
}

func marshalSlice(buf *bytes.Buffer, v reflect.Value) {
	if v.IsNil() {
		buf.WriteString(`{"NULL":"true"}`)
		return
	}
	marshalArray(buf, v)
}

func marshalArray(buf *bytes.Buffer, v reflect.Value) {
	buf.WriteString(`{"L":[`)
	n := v.Len()
	for i := 0; i < n; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		marshal(buf, v.Index(i))
	}
	buf.WriteString(`]}`)
}
