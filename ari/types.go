package ari

import (
	"fmt"
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

func (p Configuration) GetFloat64(name string) (value float64, err error)  {
	v, ok := p[name]
	if !ok {
		err = fmt.Errorf("no field %s", name)
		return value, err
	}
	switch tp:=v.(type) {
	case float64:
		value = tp
	default:
		err = fmt.Errorf("no float64 %s", name)
	}
	return value, err
}

// GetInt picks Int value of `name`
// it just casts int64 value to int
func (p Configuration) GetInt(name string) (value int, err error)  {
	var v float64
	v, err = p.GetFloat64(name)
	if err != nil {
		return value, err
	}
	value = int(v)
	return value, nil
}

// GetString picks string value of `name`
// it just converts []byte value to string
func (p Configuration) GetString(name string) (value string, err error)  {
	var v interface{}
	var ok bool
	v, exists := p[name]
	if exists {
		value, ok = v.(string)
	}
	if !exists || !ok {
		return value, fmt.Errorf("no string field %s", name)
	}
	return value, err
}

