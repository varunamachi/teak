package mongo

import (
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
		Name:   "mongo-host",
		Value:  "localhost",
		Usage:  "Address of the host running mongodb",
		EnvVar: "MONGO_HOST",
	},
	cli.IntFlag{
		Name:   "mongo-port",
		Value:  27017,
		Usage:  "Port on which Mongodb is listening",
		EnvVar: "MONGO_PORT",
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

func requireMongo(ctx *cli.Context) (err error) {
	defer teak.LogErrorX("t.mongo", "Failed to initialize mongoDB", err)
	ag := teak.NewArgGetter(ctx)
	var opts ConnOpts
	if !teak.GetConfig("mongo.opts", &opts) {
		opts.Host = ag.GetRequiredString("mongo-host")
		opts.Port = ag.GetRequiredInt("mongo-port")
		opts.User = ag.GetOptionalString("mongo-user")
		if len(opts.User) != 0 {
			opts.Password = ag.GetRequiredSecret("mongo-pass")
		}
	} else {
		teak.Info("t.mongo", "Read mongo options from app config")
		opts.Host = ag.GetStringOr("mongo-host", opts.Host)
		opts.Port = ag.GetIntOr("mongo-port", opts.Port)
		opts.User = ag.GetStringOr("mongo-user", opts.User)
		opts.Password = ag.GetSecretOr("mongo-pass", opts.Password)
	}
	err = ConnectSingle(&opts)
	if err != nil {
		teak.LogFatal("t.pg", err)
	} else {
		teak.Info("t.mongo", "Connected to mongoDB server at %s:%d",
			opts.Host, opts.Port)
	}
	return err
}
