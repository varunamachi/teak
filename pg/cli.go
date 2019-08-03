package pg

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

//pgFlags - flags to get postgres connection options
var pgFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "pg-host",
		Value:  "localhost",
		Usage:  "Address of the host running postgres",
		EnvVar: "PG_HOST",
	},
	cli.IntFlag{
		Name:   "pg-port",
		Value:  5432,
		Usage:  "Port on which postgres is listening",
		EnvVar: "PG_PORT",
	},
	cli.StringFlag{
		Name:   "pg-db",
		Value:  "",
		Usage:  "Database name",
		EnvVar: "PG_DB",
	},
	cli.StringFlag{
		Name:   "pg-user",
		Value:  "",
		Usage:  "Postgres user name",
		EnvVar: "PG_USER",
	},
	cli.StringFlag{
		Name:   "pg-pass",
		Value:  "",
		Usage:  "Postgres password for connection",
		EnvVar: "PG_PASS",
	},
}

func requirePostgres(ctx *cli.Context) (err error) {
	defer teak.LogErrorX("t.pg", "Failed to initialize postgres", err)
	ag := teak.NewArgGetter(ctx)
	var opts ConnOpts
	if !teak.GetConfig("postgres.opts", &opts) {
		opts.Host = ag.GetRequiredString("pg-host")
		opts.Port = ag.GetRequiredInt("pg-port")
		opts.User = ag.GetRequiredString("pg-user")
		opts.DBName = ag.GetRequiredString("pg-user")
		opts.Password = ag.GetRequiredSecret("pg-pass")
	} else {
		teak.Info("t.pg", "Read postgresql options from app config")
		opts.Host = ag.GetStringOr("pg-host", opts.Host)
		opts.Port = ag.GetIntOr("pg-port", opts.Port)
		opts.User = ag.GetStringOr("pg-user", opts.User)
		opts.DBName = ag.GetStringOr("pg-db", opts.User)
		opts.Password = ag.GetSecretOr("pg-pass", opts.Password)
	}
	err = ConnectWithOpts(&opts)

	if err = db.Ping(); err != nil {
		teak.LogFatal("t.pg", err)
	} else {
		teak.Info("t.pg", "Connected to postgres server at %s:%d",
			opts.Host, opts.Port)
	}
	return err
}
