package workflow

import (
	"reflect"
	"runtime"
)

// ActivityName returns the name of the function that is passed in as an interface
func ActivityName(i interface{}) string {
	// get the value and type of the interface
	v := reflect.ValueOf(i)
	t := v.Type()
	// check if the interface is a function
	if t.Kind() == reflect.Func {
		// get the name of the function
		funcName := runtime.FuncForPC(v.Pointer()).Name()

		return funcName
	}
	// return an empty string if the interface is not a function
	return ""
}
