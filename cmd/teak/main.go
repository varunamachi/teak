package main

import (
	"os"

	_ "github.com/lib/pq"
	"github.com/varunamachi/teak"
	"github.com/varunamachi/teak/pg"
)

func main() {
	app := pg.NewDefaultApp(
		"teak",
		teak.Version{
			Major: 0,
			Minor: 0,
			Patch: 1,
		},
		0,
		"Default teak app",
	)
	app.Commands = append(
		app.Commands,
		*teak.GetStore().Wrap(teak.GetServiceStartCmd(teak.Serve)))
	app.Exec(os.Args)

}
