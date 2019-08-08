package teak

import (
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	echo "github.com/labstack/echo/v4"
)

//auth
var authenticator Authenticator
var authorizer Authorizer

//M - map of string to any
type M map[string]interface{}

//SM - map of string to string
type SM map[string]string

//AuthLevel - authorization of an user
type AuthLevel int

const (
	//Super - super user
	Super AuthLevel = iota

	//Admin - application admin
	Admin

	//Normal - normal user
	Normal

	//Monitor - readonly user
	Monitor

	//Public - no authentication required
	Public
)

//UserState - state of the user account
type UserState string

//Verfied - user account is verified by the user
var Verfied UserState = "verified"

//Active - user is active
var Active UserState = "active"

//Disabled - user account is disabled by an admin
var Disabled UserState = "disabled"

//Flagged - user account is flagged by a user
var Flagged UserState = "flagged"

//User - represents an user
type User struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Auth        AuthLevel `json:"auth"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Title       string    `json:"title"`
	FullName    string    `json:"fullName"`
	State       UserState `json:"state"`
	VerID       string    `json:"verID"`
	PwdExpiry   time.Time `json:"pwdExpiry"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	ModifiedBy  string    `json:"modifiedBy"`
	VerfiedDate time.Time `json:"verified"`
	Props       SM        `json:"props"`
}

//Group - group of users
type Group struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

func (a AuthLevel) String() string {
	switch a {
	case Super:
		return "Super"
	case Admin:
		return "Admin"
	case Normal:
		return "Normal"
	case Monitor:
		return "Monitor"
	case Public:
		return "Public"

	}
	return "Unknown"
}

//UserStorage - interface representing strategy to store and manage user
//information
type UserStorage interface {
	//CreateUser - creates user in database
	CreateUser(user *User) (err error)

	//UpdateUser - updates user in database
	UpdateUser(user *User) (err error)

	//DeleteUser - deletes user with given user ID
	DeleteUser(userID string) (err error)

	//GetUser - gets details of the user corresponding to ID
	GetUser(userID string) (user *User, err error)

	//GetAllUsers - gets all users based on offset and limit
	GetUsers(offset int,
		limit int,
		filter *Filter) (users []*User, err error)

	//GetCount - gives the number of user selected by given filter
	GetCount(filter *Filter) (count int, err error)

	//GetUsersWithCount - gives a list of users paged with total count
	GetUsersWithCount(offset int,
		limit int,
		filter *Filter) (total int, users []*User, err error)

	//ResetPassword - sets password of a unauthenticated user
	ResetPassword(userID, oldPwd, newPwd string) (err error)

	//SetPassword - sets password of a already authenticated user, old password
	//is not required
	SetPassword(userID, newPwd string) (err error)

	//ValidateUser - validates user ID and password
	ValidateUser(userID, password string) (err error)

	//GetUserAuthLevel - gets user authorization level
	GetUserAuthLevel(userID string) (level AuthLevel, err error)

	//CreateSuperUser - creates the first super user for the application
	CreateSuperUser(user *User, password string) (err error)

	//SetUserState - sets state of an user account
	SetUserState(userID string, state UserState) (err error)

	//VerifyUser - sets state of an user account to verified based on userID
	//and verification ID
	VerifyUser(userID, verID string) (err error)

	//UpdateProfile - updates user details - this should be used when user
	//logged in is updating own user account
	UpdateProfile(user *User) (err error)
}

//Authenticator - a function that is used to authenticate an user. The function
//takes map of parameters contents of which will differ based on actual function
//used
type Authenticator func(params map[string]interface{}) (*User, error)

//Authorizer - a function that will be used authorize an user
type Authorizer func(userID string) (AuthLevel, error)

//NoOpAuthenticator - authenticator that does not do anything
func NoOpAuthenticator(params map[string]interface{}) (*User, error) {
	return nil, nil
}

//NoOpAuthorizer - authorizer that does not do anything
func NoOpAuthorizer(userID string) (AuthLevel, error) {
	return Public, nil
}

func dummyAuthenticator(params map[string]interface{}) (
	user *User, err error) {
	user = nil
	err = errors.New("No valid authenticator found")
	return user, err
}

func dummyAuthorizer(userID string) (role AuthLevel, err error) {
	err = errors.New("No valid authorizer found")
	return role, err
}

//GetUserIDPassword - gets userID and password from parameter app, if not
//available a error is returned
func GetUserIDPassword(params map[string]interface{}) (
	userID string, password string, err error) {
	var aok, bok bool
	userID, aok = params["userID"].(string)
	//UserID is the SHA1 hash of the userID provided
	if aok {
		userID = Hash(userID)
	}
	password, bok = params["password"].(string)
	if !aok || !bok {
		err = errors.New("Authorization, Invalid credentials provided")
	}
	return userID, password, err
}

//GetEndpoints - Export app security related APIs
func getAuthEndpoints() []*Endpoint {
	return []*Endpoint{
		&Endpoint{
			Method:   echo.POST,
			URL:      "login",
			Category: "security",
			Func:     login,
			Access:   Public,
			Comment:  "Login to application",
		},
	}
}

func login(ctx echo.Context) (err error) {
	defer func() {
		LogError("Net:Sec:API", err)
	}()
	msg := "Login successful"
	status := http.StatusOK
	var data map[string]interface{}
	userID := ""
	name := "" //user name is used for auditing
	creds := make(map[string]string)
	err = ctx.Bind(&creds)
	if err == nil {
		var user *User
		userID = creds["userID"]
		name = userID
		user, err = DoLogin(userID, creds["password"])
		if err == nil {
			if user.State == Active {
				token := jwt.New(jwt.SigningMethodHS256)
				claims := token.Claims.(jwt.MapClaims)
				name = user.FirstName + " " + user.LastName
				claims["userID"] = user.ID
				claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix()
				claims["access"] = user.Auth
				claims["userName"] = name
				claims["userType"] = "normal"
				var signed string
				key := GetJWTKey()
				signed, err = token.SignedString(key)
				if err == nil {
					data = make(map[string]interface{})
					data["token"] = signed
					data["user"] = user
				} else {
					msg = "Failed to sign token"
					status = http.StatusInternalServerError
				}
			} else {
				data = make(map[string]interface{})
				data["state"] = user.State
				msg = "User is not active"
				status = http.StatusUnauthorized
				err = errors.New(msg)
			}
		} else {
			msg = "Login failed"
			status = http.StatusUnauthorized
		}
	} else {
		msg = "Failed to read credentials from request"
		status = http.StatusBadRequest
	}
	//SHA1 encoded to avoid storing email in db
	ctx.Set("userID", Hash(userID))
	ctx.Set("userName", name)
	AuditedSend(ctx, &Result{
		Status: status,
		Op:     "login",
		Msg:    msg,
		OK:     err == nil,
		Data:   data,
		Err:    ErrString(err),
	})
	return LogError("Net:Sec:API", err)
}
