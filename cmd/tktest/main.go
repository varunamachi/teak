package main

import (
	"fmt"
	"reflect"

	_ "github.com/lib/pq"
	"github.com/varunamachi/teak"
)

func main() {
	fmt.Printf("%20s%10s%10s%20s\n", "NAME", "VALUE", "KIND", "PATH")
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
			name := ""
			if state.Field != nil {
				name = state.Field.Tag.Get("json")
			}
			fmt.Printf("%20s%10s%10s%20s\n",
				name,
				state.Current.Interface(),
				state.Current.Kind().String(),
				state.Path)
			return true
		},
	})

	fmt.Println("-- -- -- -- -- -- -- -- -- --")

	// mp := teak.ToFlatMap(make([]teak.User, 2), "json")
	// teak.DumpJSON(mp)
}
