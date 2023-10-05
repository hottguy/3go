package cfg

import (
	"encoding/json"
	"os"
)

type Object map[string]any
type Array []any

func GetInstance(filepath string) Object {
	if b, err := os.ReadFile(filepath); err != nil {
		panic(err)
	} else {
		c := Object{}
		if err := json.Unmarshal(b, &c); err != nil {
			panic(err)
		} else {
			return c
		}
	}
}

func (c Object) String() string {
	if b, err := json.MarshalIndent(c, "", "  "); err != nil {
		panic(err)
	} else {
		return string(b)
	}
}

////////////////////////////////////////////////////////////////////////////////

func (a Array) GetString(idx int) string {
	return a[idx].(string)
}

func (a Array) GetInt(idx int) int64 {
	return int64(a[idx].(float64))
}

func (a Array) GetFloat(idx int) float64 {
	return a[idx].(float64)
}

func (a Array) GetBool(idx int) bool {
	return a[idx].(bool)
}

func (a Array) GetArray(idx int) Array {
	return a[idx].([]any)
}

func (a Array) GetObject(idx int) Object {
	return a[idx].(map[string]any)
}

////////////////////////////////////////////////////////////////////////////////

func (c Object) GetString(key string) string {
	return c[key].(string)
}

func (c Object) GetInt(key string) int64 {
	return int64(c[key].(float64))
}

func (c Object) GetFloat(key string) float64 {
	return c[key].(float64)
}

func (c Object) GetBool(key string) bool {
	return c[key].(bool)
}

func (c Object) GetArray(key string) Array {
	return c[key].([]any)
}

func (c Object) GetObject(key string) Object {
	return c[key].(map[string]any)
}
