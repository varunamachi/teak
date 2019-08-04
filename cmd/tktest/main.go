package main

import (
	"fmt"
	"reflect"

	_ "github.com/lib/pq"
	"github.com/varunamachi/teak"
)

func main() {
	teak.Walk(&teak.User{}, &teak.WalkConfig{
		MaxDepth:         3,
		IgnoreContainers: true,
		FieldNameRetriever: func(field *reflect.StructField) string {
			jt := field.Tag.Get("json")
			if jt != "" {
				return jt
			}
			return field.Name
		},
		Visitor: func(state *teak.WalkerState) bool {
			fmt.Println(
				state.Path,
				state.Name,
				state.Current.String(),
				state.Current.Kind())
			return true
		},
	})

	fmt.Println("-- -- -- -- -- -- -- -- -- --")

	mp := teak.ToFlatMap(make([]teak.User, 2), "json")
	teak.DumpJSON(mp)
}
