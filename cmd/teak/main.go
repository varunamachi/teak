package main

import (
	"os"

	"github.com/varunamachi/teak"
	"github.com/varunamachi/teak/mongo"
)

func main() {
	app := mongo.NewDefaultApp(
		"teak",
		teak.Version{
			Major: 0,
			Minor: 0,
			Patch: 1,
		},
		"v0",
		"Default teak app",
	)
	app.Commands = append(
		app.Commands,
		*teak.GetStore().Wrap(teak.GetServiceStartCmd(teak.Serve)))
	app.Exec(os.Args)

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
}
