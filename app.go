package teak

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/urfave/cli.v1"
)

//ModuleConfigFunc Signature used by functions that are used to configure a
//module. Some config callbacks include - initialize, setup, reset etc
type ModuleConfigFunc func(gtx context.Context, app *App) (err error)

//Module - represents an application module
type Module struct {
	Name         string              `json:"name" db:"name"`
	Description  string              `json:"desc" db:"desc"`
	Endpoints    []*Endpoint         `json:"endpoints" db:"endpoints"`
	ItemHandlers []StoredItemHandler `json:"itemHandlers" db:"item_handlers"`
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
	apiVersion int
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

func addInitializer(
	gtx context.Context,
	cmd *cli.Command,
	module *Module,
	app *App) {
	req := func(ctx *cli.Context) error {
		if module.Initialize != nil {
			err := module.Initialize(gtx, app)
			if err != nil {
				Error("App", "Failed to initialize module %s",
					module.Name)
			}
		}
		return nil
	}
	if cmd.Before == nil {
		cmd.Before = req
	} else {
		otherBefore := cmd.Before
		cmd.Before = func(ctx *cli.Context) (err error) {
			err = otherBefore(ctx)
			if err == nil {
				err = req(ctx)
			}
			return err
		}
	}
}

//Exec - runs the applications
func (app *App) Exec(gtx context.Context, args []string) (err error) {

	for _, module := range app.modules {
		if module.Commands != nil {
			for _, cmd := range module.Commands {
				addInitializer(gtx, cmd, module, app)
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
	apiVersion int,
	desc string,
	authtr Authenticator,
	authzr Authorizer,
	uStorage UserStorage,
	genStorage DataStorage) (app *App) {

	dataStorage = genStorage
	authenticator = authtr
	authorizer = authzr
	userStorage = uStorage
	// if err := dataStorage.Init(); err != nil {
	// 	Fatal("t.app.dataStore", "Failed to initilize application store")
	// }
	InitLogger(LoggerConfig{
		Logger:      NewDirectLogger(),
		LogConsole:  true,
		FilterLevel: InfoLevel,
	})

	LoadConfig(name)

	app = &App{
		App: cli.App{
			Name:     name,
			Commands: make([]cli.Command, 0, 100),
			Version:  appVersion.String(),
			Authors: []cli.Author{
				{
					Name: "The " + name + " team",
				},
			},
			Usage:     desc,
			ErrWriter: ioutil.Discard,
			Metadata:  map[string]interface{}{},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "log-level",
					Value: "info",
					Usage: "Give log level, one of: 'trace', 'debug', " +
						"'info', 'warn', 'error'",
				},
			},
			Before: func(ctx *cli.Context) error {
				ag := NewArgGetter(ctx)
				logLevel := ag.GetOptionalString("log-level")
				if logLevel != "" {
					switch logLevel {
					case "trace":
						SetLevel(TraceLevel)
					case "debug":
						SetLevel(DebugLevel)
					case "info":
						SetLevel(InfoLevel)
					case "warn":
						SetLevel(WarnLevel)
					case "error":
						SetLevel(ErrorLevel)
					}
				}
				return nil
			},
		},
		apiRoot:    "",
		apiVersion: apiVersion,
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
		Setup: func(gtx context.Context, app *App) error {
			// return dataStorage.Init()
			return nil
		},
		Initialize: func(gtx context.Context, app *App) error {
			return dataStorage.Init(gtx, nil)
		},
		Reset: func(gtx context.Context, app *App) error {
			return dataStorage.Reset(gtx)
		},
	})
	return app
}

// Init - initializes the application and the registered module. This needs to
// be called when app/module configuration changes.
// For example: This is the place where mongoDB indices are expected to
// be created.
func (app *App) Init(
	gtx context.Context, admin *User, adminPass string, param M) (err error) {
	err = GetStore().Setup(context.TODO(), admin, adminPass, M{})
	if err != nil {
		return err
	}
	for _, module := range app.modules {
		if module.Initialize != nil {
			err = module.Initialize(gtx, app)
			if err != nil {
				Error("t.app.init", "Failed to initialize %s",
					module.Name)
				break
			}
			Info("t.app.init", "Initialized module %s", module.Name)
		}
	}
	return err
}

// Setup - Setup the application for the first time
func (app *App) Setup(gtx context.Context) (err error) {
	defer func() {
		if err != nil {
			LogErrorX("t.app.setup", "Failed to setup data storage", err)
		}
	}()
	Info("t.app.setup", "Data storage setup succesful")

	for _, module := range app.modules {
		if module.Setup != nil {
			err = module.Setup(gtx, app)
			if err != nil {
				Error("t.app.setup", "Failed to set module %s up",
					module.Name)
				break
			}
			Info("t.app.setup", "Setup module %s", module.Name)
		}
	}
	if err == nil {
		Info("t.app.setup", "Application setup complete")
	}
	return err
}

//Reset - resets the application and module configuration and data.
//USE WITH CAUTION
func (app *App) Reset(gtx context.Context) (err error) {
	defer func() {
		if err != nil {
			LogErrorX("t.app.reset", "Failed to reset app", err)
		}
	}()
	for _, module := range app.modules {
		if module.Reset != nil {
			err = module.Reset(gtx, app)
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
