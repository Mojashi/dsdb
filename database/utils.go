package database

import (
	"fmt"
	"reflect"
)

func getKeys(mymap interface{}) []string {
	vs := reflect.ValueOf(mymap).MapKeys()
	keys := []string{}
	for _, v := range vs {
		keys = append(keys, fmt.Sprint(v))
	}
	return keys
}
