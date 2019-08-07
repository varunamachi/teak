package main

import (
	"fmt"
	"os"
	"reflect"

	_ "github.com/lib/pq"
	"github.com/varunamachi/teak"
	"github.com/varunamachi/teak/pg"
	"gopkg.in/urfave/cli.v1"
)

func commands() []*cli.Command {
	return []*cli.Command{
		teak.GetStore().Wrap(&cli.Command{
			Name:        "ds",
			Description: "Test data source related functionality",
			Subcommands: []cli.Command{
				cli.Command{
					Name:        "create",
					Description: "Test creation",
					Action: func(ctx *cli.Context) {
						teak.GetStore().Create("teakUser", &teak.User{})
					},
				},
			},
		}),
		&cli.Command{
			Name:        "walk",
			Description: "Test struct walk",
			Action: func(ctx *cli.Context) {
				fmt.Printf("%20s%10s%20s\n",
					"NAME", "KIND", "PATH")
				teak.Walk(&teak.User{}, &teak.WalkConfig{
					MaxDepth:         3,
					IgnoreContainers: true,
					FieldNameRetriever: func(
						field *reflect.StructField) string {
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
						fmt.Printf("%20s%10s%20s\n",
							name,
							state.Current.Kind().String(),
							state.Path)
						return true
					},
				})
			},
		},
		&cli.Command{
			Name:        "flat-map",
			Description: "Test flat map",
			Action: func(ctx *cli.Context) {
				mp := teak.ToFlatMap(make([]teak.User, 2), "json")
				teak.DumpJSON(mp)
			},
		},
	}
}

func main() {
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
	app.AddModule(&teak.Module{
		Name:        "test",
		Description: "Module for functional testing",
		Commands:    commands(),
	})

	app.Exec(os.Args)

}
