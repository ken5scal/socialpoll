package main

import (
	"sync"
	"net/http"
)

var (
	varsLock sync.RWMutex
	vars map[*http.Request]map[string]interface{}
)

// Creates a map to hold data for designated request in memory
func OpenVars(r *http.Request) {
	varsLock.Lock()
	if vars == nil {
		vars = map[*http.Request]map[string]interface{}{}
	}
	vars[r] = map[string]interface{}{}
	varsLock.Unlock()
}

// Release memory after finishing request
func CloseVars(r *http.Request) {
	varsLock.Lock()
	delete(vars, r)
	varsLock.Unlock()
}

func GetVar(r * http.Request, key string) interface {} {
	varsLock.RLock()
	value := vars[r][key]
	varsLock.RUnlock()
	return value
}

func SetVar(r *http.Request, key string, value interface{}) {
	varsLock.Lock()
	vars[r][key] = value
	varsLock.Unlock()
}

func withVars(fn http.HandlerFunc) http.HandlerFunc  {
	return func(w http.ResponseWriter, r * http.Request) {
		OpenVars(r)
		defer CloseVars(r)
		fn(w, r)
	}
}

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(w, r)
	}
}