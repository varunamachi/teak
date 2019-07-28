package pg

import (
	"github.com/varunamachi/teak"
	"gopkg.in/urfave/cli.v1"
)

//pgFlags - flags to get postgres connection options
var pgFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "db-host",
		Value: "localhost",
		Usage: "Address of the host running postgres",
	},
	cli.IntFlag{
		Name:  "db-port",
		Value: 27017,
		Usage: "Port on which postgres is listening",
	},
	cli.StringFlag{
		Name:  "db-name",
		Value: "",
		Usage: "Database name",
	},
	cli.StringFlag{
		Name:  "db-user",
		Value: "",
		Usage: "Postgres user name",
	},
	cli.StringFlag{
		Name:  "db-pass",
		Value: "",
		Usage: "Postgres password for connection",
	},
}

func requireMongo(ctx *cli.Context) (err error) {
	ag := teak.NewArgGetter(ctx)
	dbHost := ag.GetRequiredString("db-host")
	dbPort := ag.GetRequiredInt("db-port")
	dbName := ag.GetRequiredString("db-name")
	dbUser := ag.GetRequiredString("db-user")
	dbPassword := ""
	if len(dbUser) != 0 {
		dbPassword = ag.GetRequiredSecret("db-pass")
	}
	if err = ag.Err; err == nil {
		err = ConnectWithOpts(&ConnOpts{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			DBName:   dbName,
		})
	}
	if err != nil {
		teak.LogFatal("t.ds.postgres", err)
	}
	return err
}
