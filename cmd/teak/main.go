package main

import (
	"fmt"
	"reflect"

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

	teak.Traverse(&teak.User{}, func(
		path string,
		parent *reflect.Value,
		value *reflect.Value) {
		fmt.Println(path, value.String(), value.Kind())
	})
}
