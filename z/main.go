package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type Object map[string]interface{}
type Array []interface{}

func main() {
	x := map[string]interface{}{
		"yeeha":        "brohoho",
		"chinken":      "nunget",
		"beesechurger": 65,
		"phucking":     func(s string) string { return s },
	}
	xs, err := SanitizeObject(x)
	if err != nil {
		log.Fatalln(err)
	}
	b, err := json.Marshal(xs)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(b))
}

func SanitizeObject(object map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	for key, value := range object {
		switch value := value.(type) {
		case nil: // null
			output[key] = value
		case string: // string
			output[key] = value
		case int, int8, int16, int32, int64, uint, uint8,
			uint16, uint64, uintptr, float32, float64: // number
			output[key] = value
		case map[string]interface{}: // object
			v, err := SanitizeObject(value)
			if err != nil {
				return output, err
			}
			output[key] = v
		case []interface{}: // array
			v, err := SanitizeArray(value)
			if err != nil {
				return output, err
			}
			output[key] = v
		default:
			v, err := SanitizeInterface(value)
			if err != nil {
				return output, err
			}
			output[key] = v
		}
	}
	return output, nil
}

func SanitizeArray(array []interface{}) ([]interface{}, error) {
	var output []interface{}
	for _, item := range array {
		switch value := item.(type) {
		case nil: // null
			output = append(output, value)
		case string: // string
			output = append(output, value)
		case int, int8, int16, int32, int64, uint, uint8,
			uint16, uint64, uintptr, float32, float64: // number
			output = append(output, value)
		case map[string]interface{}: // object
			v, err := SanitizeObject(value)
			if err != nil {
				return output, err
			}
			output = append(output, v)
		case []interface{}: // array
			v, err := SanitizeArray(value)
			if err != nil {
				return output, err
			}
			output = append(output, v)
		default:
			v, err := SanitizeInterface(value)
			if err != nil {
				return output, err
			}
			output = append(output, v)
		}
	}
	return output, nil
}

func SanitizeInterface(v interface{}) (interface{}, error) {
	var output interface{}
	switch vv := reflect.ValueOf(v); vv.Kind() {
	case reflect.Array: // K
	case reflect.Chan: // K
	case reflect.Func: // ?
		return funcType(v), nil
	case reflect.Interface: // ?
	case reflect.Map: // K,V
	case reflect.Ptr: // K
	case reflect.Slice: // K
	case reflect.Struct: // K
	case reflect.Complex64, reflect.Complex128: // unsupported
		return output, fmt.Errorf("unsupported type: complex number")
	case reflect.UnsafePointer: // unsupported
		return output, fmt.Errorf("unsupported type: unsafe.Pointer")
	case reflect.Invalid: // unsupported
		return output, fmt.Errorf("unsupported type: reflect.Invalid")
	default:
		output = v
	}
	return output, nil
}

func funcType(f interface{}) string {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return "<not a function>"
	}
	buf := strings.Builder{}
	buf.WriteString("func(")
	for i := 0; i < t.NumIn(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(t.In(i).String())
	}
	buf.WriteString(")")
	if numOut := t.NumOut(); numOut > 0 {
		if numOut > 1 {
			buf.WriteString(" (")
		} else {
			buf.WriteString(" ")
		}
		for i := 0; i < t.NumOut(); i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(t.Out(i).String())
		}
		if numOut > 1 {
			buf.WriteString(")")
		}
	}
	return buf.String()
}
