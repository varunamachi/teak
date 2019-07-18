package teak

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	echo "github.com/labstack/echo/v4"
)

//GetString - retrieves property with key from context
func GetString(ctx echo.Context, key string) (value string) {
	ui := ctx.Get(key)
	userID, ok := ui.(string)
	if ok {
		return userID
	}
	return ""
}

//AuditedSend - sends result as JSON while logging it as event. The event data
//is same as the data present in the result
func AuditedSend(ctx echo.Context, res *Result) (err error) {
	err = ctx.JSON(res.Status, res)
	LogEvent(
		res.Op,
		GetString(ctx, "userID"),
		GetString(ctx, "userName"),
		res.OK,
		res.Err,
		res.Data)
	return err
}

//AuditedSendSecret - Sends result to client and logs everything other than the
//secret data field
func AuditedSendSecret(ctx echo.Context, res *Result) (err error) {
	err = ctx.JSON(res.Status, res)
	LogEvent(
		res.Op,
		GetString(ctx, "userID"),
		GetString(ctx, "userName"),
		res.OK,
		res.Err,
		nil)
	return err
}

//AuditedSendX - sends result as JSON while logging it as event. This method
//logs event data which is seperate from result data
func AuditedSendX(ctx echo.Context, data interface{}, res *Result) (err error) {
	err = ctx.JSON(res.Status, res)
	LogEvent(
		res.Op,
		GetString(ctx, "userID"),
		GetString(ctx, "userName"),
		res.OK,
		res.Err,
		data)
	return err
}

//SendAndAuditOnErr - sends the result to client and puts an audit record in
//audit log if the result is error OR sending failed
func SendAndAuditOnErr(ctx echo.Context, res *Result) (err error) {
	err = ctx.JSON(res.Status, res)
	if len(res.Err) != 0 || err != nil {
		estr := res.Err
		if err != nil {
			estr = err.Error()
		}
		LogEvent(
			res.Op,
			GetString(ctx, "userID"),
			GetString(ctx, "userName"),
			false,
			estr,
			res.Data)
	}
	return err
}

//LoadJSONFromArgs - decodes argument identified by 'param' to JSON and
//unmarshals it into the given 'out' structure
func LoadJSONFromArgs(ctx echo.Context, param string, out interface{}) (
	err error) {
	val := ctx.QueryParam(param)
	if len(val) != 0 {
		var decoded string
		decoded, err = url.PathUnescape(val)
		if err == nil {
			err = json.Unmarshal([]byte(decoded), out)
		}
	}
	return LogError("Net:Utils", err)
}

//Merge - merges multple endpoint arrays
func Merge(epss ...[]*Endpoint) (eps []*Endpoint) {
	eps = make([]*Endpoint, 0, 100)
	for _, es := range epss {
		eps = append(eps, es...)
	}
	return eps
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
