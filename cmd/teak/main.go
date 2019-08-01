package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/varunamachi/teak"
)

func main() {
	// app := mongo.NewDefaultApp(
	// 	"teak",
	// 	teak.Version{
	// 		Major: 0,
	// 		Minor: 0,
	// 		Patch: 1,
	// 	},
	// 	"v0",
	// 	"Default teak app",
	// )
	// app.Exec(os.Args)

	teak.Traverse(&teak.User{}, 2, func(
		path string,
		parent *reflect.Value,
		value *reflect.Value) bool {
		fmt.Println(path, value.String(), value.Kind())
		if value.Type() == reflect.TypeOf(time.Time{}) {
			fmt.Println("Found time")
			return false
		}
		if parent.Kind() == reflect.Struct {
		}
		return true
	})
}
