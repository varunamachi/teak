package teak

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"gopkg.in/urfave/cli.v1"
)

func getAdminEndpoints() []*Endpoint {
	return []*Endpoint{
		{
			Method:   echo.GET,
			URL:      "event",
			Access:   Admin,
			Category: "administration",
			Func:     getEvents,
			Comment:  "Fetch all the events",
		},
		{
			Method:   echo.GET,
			URL:      "ping",
			Access:   Public,
			Category: "app",
			Func:     ping,
			Comment:  "Ping the server",
		},
	}
}

func getEvents(ctx echo.Context) (err error) {
	status, msg := DefMS("Fetch events")
	var events []*Event
	var total int64
	offset, limit, has := GetOffsetLimit(ctx)
	var filter Filter
	err = LoadJSONFromArgs(ctx, "filter", &filter)
	if err == nil && has {
		total, events, err = GetAuditor().GetEvents(
			offset, limit, &filter)
		if err != nil {
			msg = "Could not retrieve event info from database"
			status = http.StatusInternalServerError
		}
	} else {
		if err == nil {
			err = errors.New("Could not get Offset and Limit arguments")
		}
		msg = "Could not find required parameter"
		status = http.StatusBadRequest
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     "events_fetch",
		Msg:    msg,
		OK:     err == nil,
		Data: CountList{
			TotalCount: int64(total),
			Data:       events,
		},
		Err: ErrString(err),
	})
	return LogError("t.app", err)
}

func ping(ctx echo.Context) (err error) {
	session, _ := RetrieveSessionInfo(ctx)
	err = SendAndAuditOnErr(ctx, &Result{
		Status: http.StatusOK,
		Op:     "ping",
		Msg:    "ping",
		OK:     err == nil,
		Data:   session,
		Err:    ErrString(err),
	})
	return LogError("t.app", err)
}

//getAdminCommands - gives commands related to HTTP networking
func getAdminCommands() []*cli.Command {
	return []*cli.Command{
		initCmd(),
		destroyCmd(),
		setupCmd(),
		resetCmd(),
		isInitCmd(),
		userCmd(),
	}
}

func initCmd() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize application",
		Flags: dataStorage.WithFlags(
			cli.StringFlag{
				Name:  "super-id",
				Usage: "Unique ID of the admin",
			},
			cli.StringFlag{
				Name:  "email",
				Usage: "Email of the admin user",
			},
			cli.StringFlag{
				Name:  "first",
				Usage: "First name of the admin",
			},
			cli.StringFlag{
				Name:  "last",
				Usage: "Last name of the admin",
			},
		),
		Action: func(ctx *cli.Context) (err error) {
			ag := NewArgGetter(ctx)
			id := ag.GetRequiredString("super-id")
			email := ag.GetRequiredString("email")
			first := ag.GetRequiredString("first")
			last := ag.GetRequiredString("last")
			if err = ag.Err; err != nil {
				return err
			}
			one := AskPassword("Password")
			two := AskPassword("Confirm")
			if one != two {
				err = Error("t.app",
					"Initial super user password does not match")
				return err
			}
			user := User{
				ID:        id,
				Email:     email,
				Auth:      Super,
				FirstName: first,
				LastName:  last,
				Props: M{
					"initial": "true",
				},
				PwdExpiry: time.Now().AddDate(1, 0, 0),
			}
			// UpdateUserInfo(&user)
			user.State = Active
			err = GetStore().Init(context.TODO(), &user, one, M{})
			if err == nil {
				Info("t.app", "App setup successful")
			}
			return err
		},
	}
}

func destroyCmd() *cli.Command {
	return &cli.Command{
		Name:  "destroy",
		Usage: "Destroy application, data source etc",
		Flags: dataStorage.WithFlags(
			cli.StringFlag{
				Name:  "super-id",
				Usage: "Unique ID of the admin",
			},
			cli.BoolFlag{
				Name:   "force",
				Usage:  "Force destroy a corrupted database",
				Hidden: true,
			},
		),
		Action: func(ctx *cli.Context) (err error) {
			init, err := GetStore().IsInitialized(context.TODO())
			if err != nil {
				return err
			}
			force := ctx.Bool("force")
			if !force && init {
				ag := NewArgGetter(ctx)
				superID := ag.GetRequiredString("super-id")
				if err = ag.Err; err != nil {
					err = Error("t.store",
						"Initial super user password does not match")
					return err
				}
				superPW := AskPassword("Password")
				var user *User
				user, err = DoLogin(context.TODO(), superID, superPW)
				if err != nil {
					err = fmt.Errorf("Failed to authenticate super user: %v",
						err)
					return err
				}
				if user.Auth != Super {
					err = errors.New("Only super user can destroy the app")
					return err
				}
			}
			err = GetStore().Destroy(context.TODO())
			if err == nil {
				Info("t.app", "App storage destroyed")
			}
			return err
		},
	}
}

func isInitCmd() *cli.Command {
	return &cli.Command{
		Name:  "is-init",
		Usage: "Check if storage is initialized",
		Flags: dataStorage.WithFlags(),
		Action: func(ctx *cli.Context) (err error) {
			yes, err := GetStore().IsInitialized(context.TODO())
			if err != nil {
				Error("t.app", "Failed to check app init state")
			} else if yes {
				Info("t.app", "App is initialized")
			} else {
				Info("t.app", "App not initialized")
			}
			return err
		},
	}
}

func setupCmd() *cli.Command {
	return &cli.Command{
		Name:  "setup",
		Usage: "Sets up the application",
		Flags: dataStorage.WithFlags(
			cli.StringFlag{
				Name:  "super-id",
				Usage: "Super user ID",
			},
			cli.StringFlag{
				Name:  "super-pw",
				Usage: "Super user password",
			},
		),
		Action: func(ctx *cli.Context) (err error) {
			vapp := GetAppReference(ctx)
			var user *User
			if vapp != nil {
				ag := NewArgGetter(ctx)
				superID := ag.GetRequiredString("super-id")
				superPW := ag.GetOptionalString("super-pw")
				if err = ag.Err; err == nil {
					defer func() {
						suname := superID
						if user != nil {
							suname = user.FirstName +
								" " + user.LastName
						}
						LogEvent(
							"app_setup",
							superID,
							suname,
							err != nil,
							ErrString(err),
							nil)
					}()
					if len(superPW) == 0 {
						superPW = AskPassword("Super-user Password")
					}
					user, err = DoLogin(context.TODO(), superID, superPW)
					if err != nil {
						err = fmt.Errorf(
							"Failed to authenticate super user: %v",
							err)
						return err
					}
					if user.Auth != Super {
						err = errors.New(
							"User forcing reset is not a super user")
					}
					err = vapp.Setup(context.TODO())
				}
			} else {
				err = errors.New("V App not properly initialized")
			}
			return LogError("t.app", err)
		},
	}
}

func resetCmd() *cli.Command {
	return &cli.Command{
		Name:  "reset",
		Usage: "Resets the application",
		Flags: dataStorage.WithFlags(
			cli.StringFlag{
				Name:  "super-id",
				Usage: "Super user ID",
			},
			cli.StringFlag{
				Name:  "super-pw",
				Usage: "Super user password",
			},
		),
		Action: func(ctx *cli.Context) (err error) {
			vapp := GetAppReference(ctx)
			var user *User
			if vapp != nil {
				ag := NewArgGetter(ctx)
				superID := ag.GetRequiredString("super-id")
				superPW := ag.GetOptionalString("super-pw")
				if err = ag.Err; err == nil {
					defer func() {
						suname := superID
						if user != nil {
							suname = user.FirstName +
								" " + user.LastName
						}
						LogEvent(
							"app_reset",
							superID,
							suname,
							err != nil,
							ErrString(err),
							nil)
					}()
					if len(superPW) == 0 {
						superPW = AskPassword("Super-user Password")
					}
					user, err = DoLogin(context.TODO(), superID, superPW)
					if err != nil {
						err = fmt.Errorf(
							"Failed to authenticate super user: %v",
							err)
						return err
					}
					if user.Auth != Super {
						err = errors.New(
							"User forcing reset is not a super user")
					}
					err = vapp.Reset(context.TODO())
				}
			} else {
				err = errors.New("V App not properly initialized")
			}
			return LogError("t.app", err)
		},
	}
}

func userCmd() *cli.Command {
	return &cli.Command{
		Name:  "uman",
		Usage: "Commands for user management",
		Subcommands: []cli.Command{
			*createUserCmd(),
			*overridePasswordCmd(),
			*testEMail(),
			*testLoginCmd(),
		},
	}
}

func createUserCmd() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a user",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "id",
				Usage: "Unique ID of the user",
			},
			cli.StringFlag{
				Name:  "email",
				Usage: "Email of the password",
			},
			cli.StringFlag{
				Name:  "first",
				Usage: "First name of the user",
			},
			cli.StringFlag{
				Name:  "last",
				Usage: "Last name of the user",
			},
			cli.StringFlag{
				Name: "role",
				Usage: "Role of the user, one of: " +
					"'super', 'admin', 'normal', 'monitor'",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			ag := NewArgGetter(ctx)
			id := ag.GetRequiredString("id")
			email := ag.GetRequiredString("email")
			first := ag.GetRequiredString("first")
			last := ag.GetRequiredString("last")
			roleStr := ag.GetRequiredString("role")
			if err = ag.Err; err == nil {
				one := AskPassword("Password")
				two := AskPassword("Confirm")
				if one == two {
					user := User{
						ID:        id,
						Email:     email,
						Auth:      toRole(roleStr),
						FirstName: first,
						LastName:  last,
						Props:     M{},
						PwdExpiry: time.Now().AddDate(1, 0, 0),
					}
					// UpdateUserInfo(&user)
					user.State = Active
					idHash, err := userStorage.CreateUser(
						context.TODO(),
						&user)
					if err != nil {
						//wrap
						return err
					}
					err = userStorage.SetPassword(context.TODO(), idHash, one)
					if err == nil {
						Info("t.uman", "User %s created successfully", id)
					}
				}
			}
			return err
		},
	}
}

func setRoleCmd() *cli.Command {
	return &cli.Command{
		Name:  "set-role",
		Usage: "Sets auth-level/role to a user",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "admin-id",
				Usage: "Super user ID",
			},
			cli.StringFlag{
				Name:  "admin-pw",
				Usage: "Super user password",
			},
			cli.StringFlag{
				Name:  "id",
				Usage: "Unique ID of the user",
			},
			cli.StringFlag{
				Name: "role",
				Usage: "Role of the user, one of: " +
					"'super', 'admin', 'normal', 'monitor'",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			ag := NewArgGetter(ctx)
			id := ag.GetRequiredString("id")
			roleStr := ag.GetRequiredString("role")
			adminID := ag.GetRequiredString("admin-id")
			adminPW := ag.GetOptionalString("admin-pw")
			var user *User
			if err = ag.Err; err == nil {
				defer func() {
					adminName := adminID
					if user != nil {
						adminName = user.FirstName +
							" " + user.LastName
					}
					LogEvent(
						"user_role_set",
						adminID,
						adminName,
						err != nil,
						ErrString(err),
						nil)
				}()
				if len(adminPW) == 0 {
					adminPW = AskPassword("Admin Password")
				}
				user, err = DoLogin(context.TODO(), adminID, adminPW)
				if err != nil {
					err = LogErrorX("t.app.admin",
						"Failed to authenticate admin user: %s",
						err,
						adminID)
					return err
				}
				if user.Auth != Super && user.Auth != Admin {
					err = Error("t.app.admin",
						"User '%s' is not an admin/super user",
						adminID)
					return err
				}
				err = GetUserStorage().SetAuthLevel(
					context.TODO(), id, toRole(roleStr))
			}
			return LogErrorX("t.app",
				"Failed to set role for user %s", err, id)
		},
	}
}

func overridePasswordCmd() *cli.Command {
	return &cli.Command{
		Name:  "force-pw-reset",
		Usage: "Forced password rest - super admin only",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "id",
				Usage: "Unique ID of the user",
			},
			cli.StringFlag{
				Name:  "password",
				Usage: "New password",
			},
			cli.StringFlag{
				Name:  "super-id",
				Usage: "Super user ID",
			},
			cli.StringFlag{
				Name:  "super-pw",
				Usage: "Super user password",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			ag := NewArgGetter(ctx)
			id := ag.GetRequiredString("id")
			pw := ag.GetOptionalString("password")
			superID := ag.GetRequiredString("super-id")
			superPW := ag.GetOptionalString("super-pw")
			defer func() {
				LogError("t.app.admin", err)
			}()
			var user *User
			if err = ag.Err; err == nil {
				defer func() {
					suname := superID
					if user != nil {
						suname = user.FirstName +
							" " + user.LastName
					}
					LogEvent(
						"setup app",
						superID,
						suname,
						err != nil,
						ErrString(err),
						nil)
				}()
				if len(pw) == 0 {
					pw = AskPassword("New Password")
				}
				if len(superPW) == 0 {
					superPW = AskPassword("Super-user Password")
				}
				user, err = DoLogin(context.TODO(), superID, superPW)
				if err != nil {
					err = fmt.Errorf("Failed to authenticate super user: %v",
						err)
					return err
				}
				if user.Auth != Super {
					err = errors.New("User forcing reset is not a super user")
				}
				err = userStorage.SetPassword(context.TODO(), id, pw)
				if err == nil {
					Info("t.app",
						"Password for %s successfully reset", id)
				}
			}
			return err
		},
	}
}

func testLoginCmd() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Test login",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "id",
				Usage: "user ID",
			},
			cli.StringFlag{
				Name:  "password",
				Usage: "User password",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			ag := NewArgGetter(ctx)
			id := ag.GetRequiredString("id")
			password := ag.GetOptionalString("password")
			var user *User
			if err = ag.Err; err == nil {
				if password == "" {
					password = AskPassword("Password")
				}
				user, err = DoLogin(context.TODO(), id, password)
				if err != nil {
					err = LogErrorX("t.app.admin", "Login failed", err)
					return err
				}
				Info("t.uman", "User details: ")
				DumpJSON(user)
			}
			return err
		},
	}
}

func testEMail() *cli.Command {
	return &cli.Command{
		Name:  "test-email",
		Usage: "Sends a test EMail",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "to",
				Usage: "EMail ID of the recipient",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			ag := NewArgGetter(ctx)
			to := ag.GetRequiredString("to")
			if err = ag.Err; err == nil {
				err = SendEmail(to, "test", "hello!")
			}
			return err
		},
	}
}

//GetAppReference - gets instance of teak.App which is stored inside
//cli.App.Metadata
func GetAppReference(ctx *cli.Context) (vapp *App) {
	metadata := ctx.App.Metadata
	fmt.Println(metadata)
	vi, found := metadata["teak"]
	if found {
		vapp, _ = vi.(*App)
	}
	return vapp
}

func toRole(roleStr string) AuthLevel {
	switch roleStr {
	case "super":
		return Super
	case "admin":
		return Admin
	case "normal":
		return Normal
	case "monitor":
		return Monitor
	case "public":
		return Public
	}
	return Monitor
}
