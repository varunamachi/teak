package mg

import (
	"context"

	"github.com/varunamachi/teak"
	"gopkg.in/urfave/cli.v1"
)

//NewDefaultApp - creates a new app with MongoDB based storage providers
func NewDefaultApp(
	name string,
	appVersion teak.Version,
	apiVersion int,
	desc string) *teak.App {
	return teak.NewApp(
		name,
		appVersion,
		apiVersion,
		desc,
		teak.DefaultAuthenticator,
		nil,
		NewUserStorage(),
		NewStorage(),
	)
}

//mongoFlags - flags to get mongo connection options
var mongoFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "mongo-url",
		Value:  "localhost:27017",
		Usage:  "Address and port running mongodb instance",
		EnvVar: "MONGO_URL",
	},
	cli.StringFlag{
		Name:   "mongo-user",
		Value:  "",
		Usage:  "Mongodb user name",
		EnvVar: "MONGO_USER",
	},
	cli.StringFlag{
		Name:   "mongo-pass",
		Value:  "",
		Usage:  "Mongodb password for connection",
		EnvVar: "MONGO_PASS",
	},
}

func requireMongo(gtx context.Context, ctx *cli.Context) (err error) {
	ag := teak.NewArgGetter(ctx)
	var opts ConnOpts
	if !teak.GetConfig("mongo.opts", &opts) {
		url := ag.GetRequiredString("mongo-url")
		opts.URLs = []string{url}
		opts.User = ag.GetOptionalString("mongo-user")
		if len(opts.User) != 0 {
			opts.Password = ag.GetRequiredSecret("mongo-pass")
		}
	} else {
		teak.Info("t.mongo", "Read mongo options from app config")
	}
	err = Connect(gtx, &opts)
	if err != nil {
		err = teak.LogErrorX("t.pg",
			"Failed to open MongoDB connection to '%s'", err, opts.String())
		return err
	}
	teak.Info("t.mongo", "Connected to mongoDB server at %s", opts.String())
	return err
}
