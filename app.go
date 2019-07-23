package teak

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/urfave/cli.v1"
)

//App - the application itself
type App struct {
	cli.App
	Modules          []*Module `json:"modules"`
	NetOptions       Options   `json:"netOptions"`
	IsService        bool      `json:"isService"`
	RequiresMongo    bool      `json:"requiresMongo"`
	RequiredPostgres bool      `json:"requiresPostgres"`
}

//ModuleConfigFunc Signature used by functions that are used to configure a
//module. Some config callbacks include - initialize, setup, reset etc
type ModuleConfigFunc func(app *App) (err error)

//Module - represents an application module
type Module struct {
	Name         string              `json:"name"`
	Description  string              `json:"desc"`
	Endpoints    []*Endpoint         `json:"endpoints"`
	ItemHandlers []StoredItemHandler `json:"itemHandlers"`
	Commands     []cli.Command
	Initialize   ModuleConfigFunc
	Setup        ModuleConfigFunc
	Reset        ModuleConfigFunc
}

//FromAppDir - gives a absolute path from a path relative to
//app directory
func (app *App) FromAppDir(relPath string) (abs string) {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("APPDATA")
	}
	return filepath.Join(home, "."+app.Name, relPath)
}

//AddModule - registers a module with the app
func (app *App) AddModule(module *Module) {
	app.Modules = append(app.Modules, module)
}

//Exec - runs the applications
func (app *App) Exec(args []string) (err error) {
	// if app.IsService {
	// 	AddEndpoints(getUserManagementEndpoints()...)
	// }
	// if app.RequiresMongo {
	// 	app.Commands = append(app.Commands, GetCommands(app)...)
	// }

	for _, module := range app.Modules {
		if module.Initialize != nil {
			err = module.Initialize(app)
			if err != nil {
				Error("App", "Failed to initialize module %s",
					module.Name)
				break
			}
		}
		if module.Commands != nil {
			app.Commands = append(app.Commands, module.Commands...)
		}
		for _, fc := range module.ItemHandlers {
			siHandlers[fc.DataType()] = fc
		}
		if app.IsService {
			AddEndpoints(module.Endpoints...)
		}
	}
	if err == nil {
		InitServer(app.NetOptions)
		err = app.Run(args)
	}
	return err
}

//NewWebApp - creates a new web application with default options
func NewWebApp(
	name string,
	appVersion Version,
	apiVersion string,
	authors []cli.Author,
	requiresMongo bool, desc string) (app *App) {

	//@TODO take these decisions in data store impl
	// var store vsec.UserStorage
	// var auditor vevt.EventAuditor
	// store = &vuman.MongoStorage{}
	// auditor = &vevt.MongoAuditor{}
	// authr := vuman.MongoAuthenticator
	// if !requiresMongo {
	// 	store = &vuman.PGStorage{}
	// 	auditor = &vevt.PGAuditor{}
	// }
	// vuman.SetStorageStrategy(store)
	// vevt.SetEventAuditor(auditor)
	InitLogger(LoggerConfig{
		Logger:      NewDirectLogger(),
		LogConsole:  true,
		FilterLevel: TraceLevel,
	})

	LoadConfig(name)
	app = &App{
		IsService:     true,
		RequiresMongo: true,
		App: cli.App{
			Name:      name,
			Commands:  make([]cli.Command, 0, 100),
			Version:   appVersion.String(),
			Authors:   authors,
			Usage:     desc,
			ErrWriter: ioutil.Discard,
			Metadata:  map[string]interface{}{},
		},
		NetOptions: Options{
			RootName:      "",
			APIVersion:    apiVersion,
			Authenticator: dummyAuthenticator,
			Authorizer:    nil,
		},
		Modules: make([]*Module, 0, 10),
	}
	app.Metadata["vapp"] = app
	return app
}

//NewSimpleApp - an app that is not a service and does not use mongodb
func NewSimpleApp(
	name string,
	appVersion Version,
	apiVersion string,
	authors []cli.Author,
	requiresMongo bool,
	desc string) (app *App) {
	InitLogger(LoggerConfig{
		Logger:      NewDirectLogger(),
		LogConsole:  true,
		FilterLevel: TraceLevel,
	})
	//@TODO take these decisions in data store impl
	// var store vsec.UserStorage
	// var auditor vevt.EventAuditor
	// store = &vuman.MongoStorage{}
	// auditor = &vevt.MongoAuditor{}
	// if !requiresMongo {
	// 	store = &vuman.PGStorage{}
	// 	auditor = &vevt.PGAuditor{}
	// }
	// vuman.SetStorageStrategy(store)
	// vevt.SetEventAuditor(auditor)
	LoadConfig(name)
	app = &App{
		IsService:     false,
		RequiresMongo: requiresMongo,
		App: cli.App{
			Name:      name,
			Commands:  make([]cli.Command, 0, 100),
			Version:   appVersion.String(),
			Authors:   authors,
			Usage:     desc,
			ErrWriter: ioutil.Discard,
			Metadata:  map[string]interface{}{},
		},
		Modules: make([]*Module, 0, 10),
	}
	app.Metadata["vapp"] = app
	return app
}

//Setup - sets up the application and the registered module. This is not
//initialization and needs to be called when app/module configuration changes.
//This is the place where mongoDB indices are expected to be created.
func (app *App) Setup() (err error) {
	err = dataStorage.Setup(nil)
	if err != nil {
		LogErrorX("t.app.setup", "Failed to setup data storage", err)
		return err
	}
	Info("t.app.setup", "Data storage setup succesful")

	for _, module := range app.Modules {
		if module.Setup != nil {
			err = module.Setup(app)
			if err != nil {
				Error("t.app.setup", "Failed to set module %s up",
					module.Name)
				break
			}
			Info("t.app.setup", "Configured module %s", module.Name)
		}
	}
	if err == nil {
		Info("t.app.setup", "Application setup complete")
	}
	return err
}

//Reset - resets the application and module configuration and data.
//USE WITH CAUTION
func (app *App) Reset() (err error) {
	err = dataStorage.Reset()
	if err != nil {
		LogErrorX("t.app.reset", "Failed to reset app", err)
		return err
	}
	for _, module := range app.Modules {
		if module.Reset != nil {
			err = module.Reset(app)
			if err != nil {
				Error("t.app.reset", "Failed to reset module %s",
					module.Name)
				break
			}
			Info("t.app.reset", "Reset module %s succesfully", module.Name)
		}
	}
	if err == nil {
		Info("t.app.setup", "Application reset complete")
	}
	return err
}

//NewAppWithOptions - creates app with non default options
func NewAppWithOptions( /*****/ ) (app *App) {
	return app
}
