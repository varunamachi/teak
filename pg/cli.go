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

func requirePostgres(ctx *cli.Context) (err error) {
	defer teak.LogErrorX("t.pg", "Failed to initialize postgres", err)

	ag := teak.NewArgGetter(ctx)
	var opts ConnOpts
	if !teak.GetConfig("postgres.opts", &opts) {
		opts.Host = ag.GetRequiredString("pg-host")
		opts.Port = ag.GetRequiredInt("pg-port")
		opts.User = ag.GetRequiredString("pg-user")
		opts.DBName = ag.GetRequiredString("pg-db")
		opts.Password = ag.GetRequiredSecret("pg-pass")
	} else {
		teak.Info("t.pg", "Read postgresql options from app config")
		opts.Host = ag.GetStringOr("pg-host", opts.Host)
		opts.Port = ag.GetIntOr("pg-port", opts.Port)
		opts.User = ag.GetStringOr("pg-user", opts.User)
		opts.DBName = ag.GetStringOr("pg-db", opts.DBName)
		opts.Password = ag.GetSecretOr("pg-pass", opts.Password)
	}
	defDB, err = ConnectWithOpts(&opts)
	if err != nil {
		err = teak.LogErrorX("t.pg", "Failed to open postgres connection", err)
		return err
	}
	err = defDB.Ping()
	if err != nil {
		err = teak.LogErrorX("t.pg", "Failed to ping postgres DB", err)
		return err
	}
	teak.Info("t.pg", "Connected to postgres server at %s:%d - to DB: %s",
		opts.Host, opts.Port, opts.DBName)
	return err
}
