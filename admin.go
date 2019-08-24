package teak

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"gopkg.in/urfave/cli.v1"
)

func getAdminEndpoints() []*Endpoint {
	return []*Endpoint{
		&Endpoint{
			Method:   echo.GET,
			URL:      "event",
			Access:   Admin,
			Category: "administration",
			Func:     getEvents,
			Comment:  "Fetch all the events",
		},
		&Endpoint{
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
	var total int
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
			TotalCount: total,
			Data:       events,
		},
		Err: ErrString(err),
	})
	return LogError("t.admin.events", err)
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
	return LogError("t.net.ping", err)
}

//getAdminCommands - gives commands related to HTTP networking
func getAdminCommands() []*cli.Command {
	return []*cli.Command{
		dataStorage.Wrap(createUserCmd()),
		dataStorage.Wrap(setupCmd()),
		dataStorage.Wrap(resetCmd()),
		dataStorage.Wrap(overridePasswordCmd()),
		testEMail(),
	}
}

func createUserCmd() *cli.Command {
	return &cli.Command{
		Name:  "create-super",
		Usage: "Create a super user",
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
		},
		Action: func(ctx *cli.Context) (err error) {
			ag := NewArgGetter(ctx)
			id := ag.GetRequiredString("id")
			email := ag.GetRequiredString("email")
			first := ag.GetRequiredString("first")
			last := ag.GetRequiredString("last")
			if err = ag.Err; err == nil {
				one := AskPassword("Password")
				two := AskPassword("Confirm")
				if one == two {
					user := User{
						ID:         id,
						Email:      email,
						Auth:       Super,
						FirstName:  first,
						LastName:   last,
						FullName:   first + " " + last,
						CreatedAt:  time.Now(),
						ModifiedAt: time.Now(),
						Props:      SM{},
						PwdExpiry:  time.Now().AddDate(1, 0, 0),
						State:      Active,
					}
					err = userStorage.CreateUser(&user)
					if err != nil {
						//wrap
						return err
					}
					err = userStorage.SetPassword(id, one)
				}
			}
			return err
		},
	}
}

func initCmd() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize application",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "reinit",
				Usage: "Remove everything and reinitialize [Not supported yet]",
			},
		},
		Action: func(ctx *cli.Context) (err error) {
			return err
		},
	}
}

func setupCmd() *cli.Command {
	return &cli.Command{
		Name:  "setup",
		Usage: "Sets up the application",
		Flags: []cli.Flag{
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
					user, err = DoLogin(superID, superPW)
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
					err = vapp.Setup()
				}
			} else {
				err = errors.New("V App not properly initialized")
			}
			return LogError("t.app.admin", err)
		},
	}
}

func resetCmd() *cli.Command {
	return &cli.Command{
		Name:  "reset",
		Usage: "Resets the application",
		Flags: []cli.Flag{
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
					user, err = DoLogin(superID, superPW)
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
					err = vapp.Reset()
				}
			} else {
				err = errors.New("V App not properly initialized")
			}
			return LogError("t.app.admin", err)
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
				user, err = DoLogin(superID, superPW)
				if err != nil {
					err = fmt.Errorf("Failed to authenticate super user: %v",
						err)
					return err
				}
				if user.Auth != Super {
					err = errors.New("User forcing reset is not a super user")
				}
				err = userStorage.SetPassword(id, pw)
				if err == nil {
					Info("t.app.admin",
						"Password for %s successfully reset", id)
				}
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
