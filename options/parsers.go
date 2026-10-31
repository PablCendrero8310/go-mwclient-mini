package options

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

func toParams(opts any) url.Values {
	params := url.Values{}

	v := reflect.ValueOf(opts)

	if !v.IsValid() {
		return params
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return params
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return params
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() {
			continue
		}

		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")

		if name == "" || name == "-" || value.IsZero() {
			continue
		}

		switch value.Kind() {
		case reflect.Bool:
			params.Set(name, "")

		case reflect.Slice:
			if value.Len() == 0 {
				continue
			}

			values := make([]string, value.Len())

			for j := 0; j < value.Len(); j++ {
				values[j] = fmt.Sprint(value.Index(j).Interface())
			}

			params.Set(name, strings.Join(values, "|"))

		default:
			params.Set(name, fmt.Sprint(value.Interface()))
		}
	}

	return params
}

func (o EditOptions) ToParams() url.Values {
	return toParams(o)
}
