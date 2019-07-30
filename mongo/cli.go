package mongo

import (
	"github.com/varunamachi/teak"
	"gopkg.in/urfave/cli.v1"
)

//NewDefaultApp - creates a new app with MongoDB based storage providers
func NewDefaultApp(
	name string,
	appVersion teak.Version,
	apiVersion string,
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
		Name:  "db-host",
		Value: "localhost",
		Usage: "Address of the host running mongodb",
	},
	cli.IntFlag{
		Name:  "db-port",
		Value: 27017,
		Usage: "Port on which Mongodb is listening",
	},
	cli.StringFlag{
		Name:  "db-user",
		Value: "",
		Usage: "Mongodb user name",
	},
	cli.StringFlag{
		Name:  "db-pass",
		Value: "",
		Usage: "Mongodb password for connection",
	},
}

func requireMongo(ctx *cli.Context) (err error) {
	ag := teak.NewArgGetter(ctx)
	dbHost := ag.GetRequiredString("db-host")
	dbPort := ag.GetRequiredInt("db-port")
	dbUser := ag.GetOptionalString("db-user")
	dbPassword := ""
	if len(dbUser) != 0 {
		dbPassword = ag.GetRequiredSecret("db-pass")
	}
	if err = ag.Err; err == nil {
		err = ConnectSingle(&ConnOpts{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
		})
	}
	if err != nil {
		teak.LogFatal("DB:Mongo", err)
	}
	return err
}

//MakeRequireMongo - makes ccommand to require information that is needed to
//connect to a mongodb instance
// func MakeRequireMongo(cmd *cli.Command) *cli.Command {

// }
