package ari

import (
	"fmt"
	"golang.org/x/tools/container/intsets"
	"errors"
)

type Configuration map[string] interface{}

func (p Configuration) get(name string) (value interface{}, err error)  {
	v, ok := p[name]
	if !ok {
		err = fmt.Errorf("no field %s", name)
		return value, err
	}
	value = v
	return value, err
}

// GetBool picks bool value of `name`
// it converts int64, int value to bool (if value is 0 or 1)
func (p Configuration) GetBool(name string) (value bool, err error)  {
	var v interface{}
	v, err = p.get(name)
	if err != nil {
		return value, err
	}
	switch tp:=v.(type) {
	case bool:
		value = tp
	case int64, int:
		var i int
		if j64, ok := v.(int64); ok {
			i = int(j64)
		}else {
			i = v.(int)
		}
		if i != 0 && i != 1 {
			err = fmt.Errorf("no bool `%s`", name)
		}else{
			if i == 0 {
				value = false
			}
			value = true
		}
	default:
		err = fmt.Errorf("no bool `%s`", name)
	}
	return value, err
}

func (p Configuration) GetInt64(name string) (value int64, err error)  {
	v, ok := p[name]
	if !ok {
		err = fmt.Errorf("no field %s", name)
		return value, err
	}
	switch tp:=v.(type) {
	case int64:
		value = tp
	default:
		err = fmt.Errorf("no int64 %s", name)
	}
	return value, err
}

// GetInt picks Int value of `name`
// it just casts int64 value to int
func (p Configuration) GetInt(name string) (value int, err error)  {
	var v int64
	v, err = p.GetInt64(name)
	if v >= int64(intsets.MaxInt) {
		err = errors.New("int overflow")
		return value, err
	}
	if err != nil {
		return value, err
	}
	value = int(v)
	return value, err
}

// GetString picks string value of `name`
// it just converts []byte value to string
func (p Configuration) GetString(name string) (value string, err error)  {
	var v []byte
	v, err = p.GetBytes(name)
	if err != nil {
		return value, err
	}
	value = string(v)
	return value, err
}

// GetBytes picks []bytes value of `name`
func (p Configuration) GetBytes(name string) (value []byte, err error)  {
	v, ok := p[name]
	if !ok {
		err = fmt.Errorf("no field %s", name)
		return value, err
	}
	switch tp:=v.(type) {
	case []byte:
		value = tp
	default:
		err = fmt.Errorf("no []byte %s", name)
	}
	return value, err
}


