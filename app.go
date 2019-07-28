package teak

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/urfave/cli.v1"
)

//ModuleConfigFunc Signature used by functions that are used to configure a
//module. Some config callbacks include - initialize, setup, reset etc
type ModuleConfigFunc func(app *App) (err error)

//Module - represents an application module
type Module struct {
	Name         string              `json:"name"`
	Description  string              `json:"desc"`
	Endpoints    []*Endpoint         `json:"endpoints"`
	ItemHandlers []StoredItemHandler `json:"itemHandlers"`
	Commands     []*cli.Command
	Initialize   ModuleConfigFunc
	Setup        ModuleConfigFunc
	Reset        ModuleConfigFunc
}

//App - the application itself
type App struct {
	cli.App
	modules    []*Module
	apiRoot    string
	apiVersion string
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
	app.modules = append(app.modules, module)
}

//Exec - runs the applications
func (app *App) Exec(args []string) (err error) {
	for _, module := range app.modules {
		if module.Initialize != nil {
			err = module.Initialize(app)
			if err != nil {
				Error("App", "Failed to initialize module %s",
					module.Name)
				break
			}
		}
		if module.Commands != nil {
			for _, cmd := range module.Commands {
				app.Commands = append(app.Commands, *cmd)
			}
		}
		for _, fc := range module.ItemHandlers {
			siHandlers[fc.DataType()] = fc
		}
		AddEndpoints(module.Endpoints...)
	}
	if err == nil {
		InitServer(app.apiRoot, app.apiVersion)
		err = app.Run(args)
	}
	return err
}

//NewApp - creates a new application with default options
func NewApp(
	name string,
	appVersion Version,
	apiVersion string,
	authtr Authenticator,
	authzr Authorizer,
	storage DataStorage,
	desc string) (app *App) {

	dataStorage = storage
	authenticator = authtr
	authorizer = authzr
	if err := dataStorage.Init(); err != nil {
		Fatal("t.app.dataStore", "Failed to initilize application store")
	}
	InitLogger(LoggerConfig{
		Logger:      NewDirectLogger(),
		LogConsole:  true,
		FilterLevel: TraceLevel,
	})
	LoadConfig(name)

	app = &App{
		App: cli.App{
			Name:     name,
			Commands: make([]cli.Command, 0, 100),
			Version:  appVersion.String(),
			Authors: []cli.Author{
				cli.Author{
					Name: "The " + name + " team",
				},
			},
			Usage:     desc,
			ErrWriter: ioutil.Discard,
			Metadata:  map[string]interface{}{},
		},
		apiRoot:    "",
		apiVersion: "v0", //TODO - Whoever registers should tell
		modules:    make([]*Module, 0, 10),
	}
	app.Metadata["teak"] = app
	app.modules = append(app.modules, &Module{
		Name:        "Core",
		Description: "teak Core module",
		Endpoints: MergeEnpoints(
			getUserManagementEndpoints(),
			getDataEndpoints(),
			getAdminEndpoints(),
		),
		Commands: MergeCommands(
			getAdminCommands(),
		),
		ItemHandlers: []StoredItemHandler{
			&UserHandler{},
		},
	})
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

	for _, module := range app.modules {
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
	for _, module := range app.modules {
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
