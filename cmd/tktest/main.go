package main

import (
	"os"

	_ "github.com/lib/pq"
	"github.com/varunamachi/teak"
	"github.com/varunamachi/teak/pg"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	// fmt.Printf("%20s%10s%10s%20s\n", "NAME", "VALUE", "KIND", "PATH")
	// teak.Walk(&teak.User{}, &teak.WalkConfig{
	// 	MaxDepth:         3,
	// 	IgnoreContainers: true,
	// 	FieldNameRetriever: func(field *reflect.StructField) string {
	// 		jt := field.Tag.Get("json")
	// 		if jt != "" {
	// 			return jt
	// 		}
	// 		return field.Name
	// 	},
	// 	Visitor: func(state *teak.WalkerState) bool {
	// 		name := ""
	// 		if state.Field != nil {
	// 			name = state.Field.Tag.Get("json")
	// 		}
	// 		fmt.Printf("%20s%10s%10s%20s\n",
	// 			name,
	// 			state.Current.Interface(),
	// 			state.Current.Kind().String(),
	// 			state.Path)
	// 		return true
	// 	},
	// })

	// fmt.Println("-- -- -- -- -- -- -- -- -- --")

	// mp := teak.ToFlatMap(make([]teak.User, 2), "json")
	// teak.DumpJSON(mp)

	app := pg.NewDefaultApp(
		"tktest",
		teak.Version{
			Major: 0,
			Minor: 0,
			Patch: 1,
		},
		0,
		"Teak experimentation app",
	)
	app.Commands = append(
		app.Commands,
		cli.Command{
			Name: "test-create",
			Action: func(ctx *cli.Context) {
				teak.GetStore().Create("teakUser", &teak.User{})
			},
		},
		*teak.GetStore().Wrap(teak.GetServiceStartCmd(teak.Serve)))
	app.Exec(os.Args)

}
