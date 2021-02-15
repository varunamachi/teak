package teak

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	echo "github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
)

//User management
var userStorage UserStorage

func getUserManagementEndpoints() []*Endpoint {
	return []*Endpoint{
		{
			Method:   echo.POST,
			URL:      "uman/user",
			Access:   Admin,
			Category: "user management",
			Func:     createUser,
			Comment:  "Create an user",
		},
		{
			Method:   echo.PUT,
			URL:      "uman/user",
			Access:   Admin,
			Category: "user management",
			Func:     updateUser,
			Comment:  "Update an user",
		},
		{
			Method:   echo.DELETE,
			URL:      "uman/user/:userID",
			Access:   Admin,
			Category: "user management",
			Func:     deleteUser,
			Comment:  "Delete an user",
		},
		{
			Method:   echo.GET,
			URL:      "uman/user/:userID",
			Access:   Monitor,
			Category: "user management",
			Func:     getUser,
			Comment:  "Get info about an user",
		},
		{
			Method:   echo.GET,
			URL:      "uman/user",
			Access:   Monitor,
			Category: "user management",
			Func:     getUsers,
			Comment:  "Get list of user & their details",
		},
		{
			Method:   echo.POST,
			URL:      "uman/user/password",
			Access:   Admin,
			Category: "user management",
			Func:     setPassword,
			Comment:  "Set password for an user",
		},
		{
			Method:   echo.PUT,
			URL:      "uman/user/password",
			Access:   Monitor,
			Category: "user management",
			Func:     resetPassword,
			Comment:  "Reset password",
		},
		{
			Method:   echo.POST,
			URL:      "uman/user/self",
			Access:   Public,
			Category: "user management",
			Func:     registerUser,
			Comment:  "Registration for new user",
		},
		{
			Method:   echo.POST,
			URL:      "uman/user/verify/:userID/:verID",
			Access:   Public,
			Category: "user management",
			Func:     verify,
			Comment:  "Verify a registered account",
		},
		{
			Method:   echo.PUT,
			URL:      "/uman/user/self",
			Access:   Public,
			Category: "user management",
			Func:     updateProfile,
		},
	}
}

//SendVerificationMail - send mail with a link to user verification based on
//user email
func SendVerificationMail(user *User) (err error) {
	content := "Hi!,\n Verify your account by clicking on " +
		"below link\n" + getVerificationLink(user)
	subject := "Verification for Sparrow"
	var emailKey string
	if GetConfig("emailKey", &emailKey) {
		var email string
		email, err = DecryptStr(emailKey, user.Email)
		if err == nil {
			err = SendEmail(email, subject, content)
		}
	} else {
		err = errors.New("Failed read EMail configuration")
	}
	// fmt.Println(content)
	return LogError("t.uman", err)
}

//DefaultAuthenticator - authenticator that uses applications UserStorage to
//authenticate a user
func DefaultAuthenticator(gtx context.Context, params map[string]interface{}) (
	user *User, err error) {
	userID, password, err := GetUserIDPassword(params)
	if err == nil {
		err = GetUserStorage().ValidateUser(gtx, userID, password)
		if err == nil {
			user, err = GetUserStorage().GetUser(gtx, userID)
		}
	}
	return user, LogError("t.auth.default", err)
}

func getVerificationLink(user *User) (link string) {
	name := user.FirstName + " " + user.LastName
	if name == "" {
		name = user.ID
	}
	//@MAYBE use a template
	var host string
	if !GetConfig("hostAddress", &host) {
		host = "http://localhost:4200"
	}
	link = host + "/" + "verify?" +
		"verifyID=" + user.VerID +
		"&userID=" + url.PathEscape(user.ID)
	return link
}

//SetUserStorage - sets the user storage strategu
func SetUserStorage(storage UserStorage) {
	userStorage = storage
}

//GetUserStorage - get the configured user storage
func GetUserStorage() UserStorage {
	return userStorage
}

//UpdateUserInfo - updates common user fields
func UpdateUserInfo(user *User) (err error) {
	if len(user.ID) == 0 {
		// @TODO - store hash of user ID
		user.ID = Hash(user.Email)
	} else {
		user.ID = Hash(user.ID)
	}
	user.VerID = uuid.NewV4().String()
	user.CreatedAt = time.Now()
	user.ModifiedAt = time.Now()
	// user.State = Disabled
	user.FullName = user.FirstName + " " + user.LastName
	// @TODO create a key retrieving strategy -- local | remote etc
	var emailKey string
	if GetConfig("emailKey", &emailKey) {
		user.Email, err = EncryptStr(emailKey, user.Email)
	} else {
		err = errors.New("Failed to read email configuration")
	}
	return err
}

func createUser(ctx echo.Context) (err error) {
	status, msg := DefMS("Create User")
	var user User
	err = ctx.Bind(&user)
	if err == nil {
		user.Props = M{
			"creationMode": "admin",
		}
		_, err = userStorage.CreateUser(ctx.Request().Context(), &user)
		if err != nil {
			msg = "Failed to create user in database"
			status = http.StatusInternalServerError
		} else {
			err = SendVerificationMail(&user)
			// fmt.Println(getVerificationLink(&user))
			if err != nil {
				msg = "Failed to send verification email"
				status = http.StatusInternalServerError
			}
		}
	} else {
		status = http.StatusBadRequest
		msg = "User information given is malformed"
	}
	err = AuditedSendX(ctx, user, &Result{
		Status: status,
		Op:     "user_create",
		Msg:    msg,
		OK:     err == nil,
		Data:   nil,
		Err:    ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func registerUser(ctx echo.Context) (err error) {
	status, msg := DefMS("Register User")
	// var user User
	upw := struct {
		User     User   `json:"user"`
		Password string `json:"password"`
	}{}
	err = ctx.Bind(&upw)
	if err == nil {
		upw.User.Auth = Normal
		idHash, err := userStorage.CreateUser(
			ctx.Request().Context(), &upw.User)
		if err != nil {
			msg = "Failed to register user in database"
			status = http.StatusInternalServerError
		} else {

			err = userStorage.SetPassword(
				ctx.Request().Context(), idHash, upw.Password)
			if err != nil {
				msg = "Failed to set password"
				status = http.StatusInternalServerError
			} else {
				err = SendVerificationMail(&upw.User)
				if err != nil {
					msg = "Failed to send verification email"
					status = http.StatusInternalServerError
				}
			}
		}
	} else {
		status = http.StatusBadRequest
		msg = "User information given is malformed"
	}
	err = AuditedSendX(ctx, upw.User, &Result{
		Status: status,
		Op:     "user_register",
		Msg:    msg,
		OK:     err == nil,
		Data: M{
			"user": upw.User,
		},
		Err: ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func updateUser(ctx echo.Context) (err error) {
	status, msg := DefMS("Update User")
	var user User
	err = ctx.Bind(&user)
	if err == nil {
		err = userStorage.UpdateUser(ctx.Request().Context(), &user)
		if err != nil {
			msg = "Failed to update user in database"
			status = http.StatusInternalServerError
		}
	} else {
		status = http.StatusBadRequest
		msg = "User information given is malformed"
	}
	err = AuditedSendX(ctx, user, &Result{
		Status: status,
		Op:     "user_update",
		Msg:    msg,
		OK:     err == nil,
		Data:   nil,
		Err:    ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func deleteUser(ctx echo.Context) (err error) {
	status, msg := DefMS("Delete User")
	userID := ctx.Param("userID")
	var user *User
	user, err = userStorage.GetUser(ctx.Request().Context(), userID)
	if err == nil {
		curID := GetString(ctx, "userID")
		if userID == curID {
			msg = "Can not delete own user account"
			status = http.StatusBadRequest
		} else if user.Auth == Super {
			msg = "Super account can not be deleted from web interface"
			status = http.StatusBadRequest
			err = errors.New(msg)
		} else {
			err = userStorage.DeleteUser(ctx.Request().Context(), userID)
			if err != nil {
				msg = "Failed to delete user from database"
				status = http.StatusInternalServerError
			}
		}
	} else {
		msg = "Invalid user ID is given for deletion"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = AuditedSend(ctx, &Result{
		Status: status,
		Op:     "user_remove",
		Msg:    msg,
		OK:     err == nil,
		Data: M{
			"id":   userID,
			"user": user,
		},
		Err: ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func getUser(ctx echo.Context) (err error) {
	status, msg := DefMS("Get User")
	userID := ctx.Param("userID")
	var user *User
	if len(userID) == 0 {
		user, err = userStorage.GetUser(ctx.Request().Context(), userID)
		if err != nil {
			msg = "Failed to retrieve user info from database"
			status = http.StatusInternalServerError
		}
	} else {
		msg = "Invalid user ID is given for retrieval"
		status = http.StatusBadRequest
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     "user_get",
		Msg:    msg,
		OK:     err == nil,
		Data:   user,
		Err:    ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func getUsers(ctx echo.Context) (err error) {
	status, msg := DefMS("Get Users")
	offset, limit, has := GetOffsetLimit(ctx)
	// us := GetFirstValidStr(ctx.Param("status"), string(Active))
	var users []*User
	var total int
	var filter Filter
	err = LoadJSONFromArgs(ctx, "filter", &filter)
	if has && err == nil {
		total, users, err = userStorage.GetUsersWithCount(
			ctx.Request().Context(), offset, limit, &filter)
		if err != nil {
			msg = "Failed to retrieve user info from database"
			status = http.StatusInternalServerError
		}
	} else if err != nil {
		msg = "Failed to decode filter"
		status = http.StatusBadRequest
	} else {
		msg = "Could not retrieve user list, offset/limit not found"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     "user_multi_fetch",
		Msg:    msg,
		OK:     err == nil,
		Data: CountList{
			TotalCount: int64(total),
			Data:       users,
		},
		Err: ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func setPassword(ctx echo.Context) (err error) {
	status, msg := DefMS("Set Password")
	pinfo := make(map[string]string)
	err = ctx.Bind(&pinfo)
	userID, ok1 := pinfo["userID"]
	password, ok2 := pinfo["password"]
	if err == nil && ok1 && ok2 {
		err = userStorage.SetPassword(ctx.Request().Context(), userID, password)
		if err != nil {
			msg = "Failed to set password in database"
			status = http.StatusInternalServerError
		}
	} else {
		status = http.StatusBadRequest
		msg = "Password information given is invalid, cannot set"
	}
	err = AuditedSendX(ctx, Hash(userID), &Result{
		Status: status,
		Op:     "user_password_set",
		Msg:    msg,
		OK:     err == nil,
		Data:   nil,
		Err:    ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func resetPassword(ctx echo.Context) (err error) {
	status, msg := DefMS("Set Password")
	pinfo := make(map[string]string)
	err = ctx.Bind(&pinfo)
	userID := GetString(ctx, "userID")
	oldPassword, ok2 := pinfo["oldPassword"]
	newPassword, ok3 := pinfo["newPassword"]
	if err == nil && ok2 && ok3 && len(userID) != 0 {
		err = userStorage.ResetPassword(
			ctx.Request().Context(), userID, oldPassword, newPassword)
		if err != nil {
			msg = "Failed to reset password in database"
			status = http.StatusInternalServerError
		}
	} else {
		status = http.StatusBadRequest
		msg = "Password information given is invalid, cannot reset"
	}
	err = AuditedSendX(ctx, Hash(userID), &Result{
		Status: status,
		Op:     "user_password_reset",
		Msg:    msg,
		OK:     err == nil,
		Data:   nil,
		Err:    ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func verify(ctx echo.Context) (err error) {
	status, msg := DefMS("Create Password")
	params := make(map[string]string)
	userID := ctx.Param("userID")
	verID := ctx.Param("verID")
	err = ctx.Bind(&params)
	if len(userID) > 0 && len(verID) > 0 && err == nil {
		err = userStorage.VerifyUser(ctx.Request().Context(), userID, verID)
		if err == nil {
			err = userStorage.SetPassword(
				ctx.Request().Context(), userID, params["password"])
			if err != nil {
				msg = "Failed to set password"
				status = http.StatusInternalServerError
			}
		} else {
			msg = "Failed to verify user"
			status = http.StatusInternalServerError
		}
	} else {
		status = http.StatusBadRequest
		msg = "Invalid information provided for creating password"
	}
	ctx.Set("userName", "N/A")
	hash := Hash(userID)
	err = AuditedSendX(ctx, hash, &Result{
		Status: status,
		Op:     "user_account_verify",
		Msg:    msg,
		OK:     err == nil,
		Data: M{
			"userID":         hash,
			"verificationID": verID,
		},
		Err: ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

func updateProfile(ctx echo.Context) (err error) {
	status, msg := DefMS("Update profile")
	var user User
	sessionUserID := GetString(ctx, "userID")
	err = ctx.Bind(&user)
	if err == nil && sessionUserID == user.ID {
		err = userStorage.UpdateProfile(ctx.Request().Context(), &user)
		if err != nil {
			msg = "Failed to update profile in database"
			status = http.StatusInternalServerError
		}
	} else {
		status = http.StatusBadRequest
		if err != nil {
			msg = "User information given is malformed"
		} else {
			msg = "Cannot update profile of another user"
		}
	}
	err = AuditedSendX(ctx, user, &Result{
		Status: status,
		Op:     "user_profile_update",
		Msg:    msg,
		OK:     err == nil,
		Data:   nil,
		Err:    ErrString(err),
	})
	return LogError("Sec:Hdl", err)
}

//UserHandler - CRUD support for User data type
type UserHandler struct{}

//DataType - type of data for which this handler is written
func (uh *UserHandler) DataType() string {
	return "teakUser"
}

//UniqueKeyField - gives the field which uniquely identifies the user
func (uh *UserHandler) UniqueKeyField() string {
	return "ID"
}

//GetKey - get the uniquely identifying key for the given item
func (uh *UserHandler) GetKey(item interface{}) interface{} {
	if user, ok := item.(User); ok {
		return user.ID
	}
	return ""
}

//SetModInfo - set the modifincation information for the data
func (uh *UserHandler) SetModInfo(item interface{}, at time.Time, by string) {
	if user, ok := item.(User); ok {
		user.ModifiedAt = at
		user.ModifiedBy = by
	}
}

//CreateInstance - create instance of the data type for which the handler is
//written
func (uh *UserHandler) CreateInstance(by string) interface{} {
	return &User{
		CreatedAt: time.Now(),
		CreatedBy: by,
	}
}

//PropNames - get prop names of Users
func (uh *UserHandler) PropNames() []string {
	return []string{
		"id",
		"email",
		"auth",
		"firstName",
		"lastName",
		"title",
		"fullName",
		"state",
		"verID",
		"pwdExpiry",
		"createdAt",
		"createdBy",
		"modifiedAt",
		"modifiedBy",
		"verified",
	}
}
