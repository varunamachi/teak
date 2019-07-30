package main

import (
	"os"

	"github.com/varunamachi/teak"
	"github.com/varunamachi/teak/mongo"
)

func main() {
	// store := mongo.NewStorage()
	// userStorage := mongo.NewUserStorage()
	// authtr := teak.DefaultAuthenticator
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
	app.Exec(os.Args)
}
