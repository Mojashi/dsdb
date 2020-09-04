package database

import "sync"

type Table struct {
	Ds interface{} //must be poitner
	//Funcs map[string]interface{}
	//PtrFuncs map[string]interface{}

	mutex *sync.Mutex
}
