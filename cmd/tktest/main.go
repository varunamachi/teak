package main

import (
	"github.com/varunamachi/teak"
	"github.com/varunamachi/teak/pg"
)

func main() {
	// 	teak.Walk(&teak.User{}, &teak.WalkConfig{
	// 		MaxDepth:             3,
	// 		VisitContainerParent: false,
	// 		Visitor: func(
	// 			path string,
	// 			tag reflect.StructTag,
	// 			parent *reflect.Value,
	// 			value *reflect.Value) bool {
	// 			fmt.Println(path, tag, value.String(), value.Kind())
	// 			return true
	// 		},
	// 	})

	// mp := teak.ToFlatMap(make([]teak.User, 10))
	// teak.DumpJSON(mp)

	teak.DumpJSON(&pg.ConnOpts{})
}
