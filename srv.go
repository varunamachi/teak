package teak

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	uuid "github.com/satori/go.uuid"
)

//Endpoint - represents a REST endpoint with associated metadata
type Endpoint struct {
	Method   string      `json:"method"`
	URL      string      `json:"url"`
	Access   AuthLevel   `json:"access"`
	Category string      `json:"cateogry"`
	Route    *echo.Route `json:"route"`
	Comment  string      `json:"Comment"`
	Func     echo.HandlerFunc
}

//Result - result of an API call
type Result struct {
	Status int         `json:"status" bson:"status"`
	Op     string      `json:"op" bson:"op"`
	Msg    string      `json:"msg" bson:"msg"`
	OK     bool        `json:"ok" bson:"ok"`
	Err    string      `json:"error" bson:"error"`
	Data   interface{} `json:"data" bson:"data"`
}

//Options - options for initializing web APIs
type Options struct {
	RootName      string
	APIVersion    string
	Authenticator Authenticator
	Authorizer    Authorizer
}

//GetJWTKey - gives a unique JWT key
func GetJWTKey() []byte {
	if len(jwtKey) == 0 {
		jwtKey, _ = uuid.NewV4().MarshalBinary()
	}
	//@TODO - remove later - Just for test
	// jwtKey = []byte("sdhfjsdfhdskjghdfkjhgdf")
	return jwtKey
}

//Session - container for retrieving session & user information from JWT
type Session struct {
	UserID   string    `json:"userID"`
	UserName string    `json:"userName"`
	UserType string    `json:"userType"`
	Valid    bool      `json:"valid"`
	Role     AuthLevel `json:"role"`
}

func getAccessLevel(path string) (access AuthLevel, err error) {
	if len(path) > (accessPos+2) && path[accessPos] == 'r' {
		switch path[accessPos+1] {
		case '0':
			access = Super
		case '1':
			access = Admin
		case '2':
			access = Normal
		case '3':
			access = Monitor
		default:
			access = Public
			err = fmt.Errorf("Invalid authorized URL: %s", path)
		}
	}
	return access, err
}

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		var userRole, access AuthLevel
		access, err = getAccessLevel(ctx.Path())
		if err != nil {
			Error("Net", "URL: %s ERR: %v", ctx.Path(), err)
			err = &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  "Invalid URL",
				Internal: err,
			}
		}
		var userInfo Session
		userInfo, err = RetrieveSessionInfo(ctx)
		// fmt.Println(err)
		if err != nil {
			err = &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  "Invalid JWT toke found, does not have user info",
				Internal: err,
			}
			LogError("Net", err)
		}
		if access < userRole {
			err = &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  "",
				Internal: err,
			}
			return err
		}
		if err == nil {
			ctx.Set("userID", userInfo.UserID)
			ctx.Set("userName", userInfo.UserName)
			err = next(ctx)
		}
		return LogError("Net", err)
	}
}

//DoLogin - performs login using username and password
func DoLogin(userID string, password string) (*User, error) {
	//Check for password expiry and stuff
	params := make(map[string]interface{})
	params["userID"] = userID
	params["password"] = password
	user, err := authenticator(params)
	if err == nil && authorizer != nil {
		user.Auth, err = authorizer(user.ID)
	}
	return user, err
}

//GetToken - gets token from context or from header
func GetToken(ctx echo.Context) (token *jwt.Token, err error) {
	itk := ctx.Get("token")
	if itk != nil {
		var ok bool
		if token, ok = itk.(*jwt.Token); !ok {
			err = fmt.Errorf("Invalid token found in context")
		}
	} else {
		header := ctx.Request().Header.Get("Authorization")
		authSchemeLen := len("Bearer")
		if len(header) > authSchemeLen {
			tokStr := header[authSchemeLen+1:]
			keyFunc := func(t *jwt.Token) (interface{}, error) {
				return GetJWTKey(), nil
			}
			token = new(jwt.Token)
			token, err = jwt.Parse(tokStr, keyFunc)
		} else {
			err = fmt.Errorf("Unexpected auth scheme used to JWT")
		}
	}
	return token, err
}

//RetrieveSessionInfo - retrieves session information from JWT token
func RetrieveSessionInfo(ctx echo.Context) (uinfo Session, err error) {
	success := true
	var token *jwt.Token
	if token, err = GetToken(ctx); err == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			var access float64
			if uinfo.UserID, ok = claims["userID"].(string); !ok {
				Error("Net:Sec:API", "Invalid user ID in JWT")
				success = false
			}
			if uinfo.UserName, ok = claims["userName"].(string); !ok {
				Error("Net:Sec:API", "Invalid user name in JWT")
			}
			if uinfo.UserType, ok = claims["userType"].(string); !ok {
				Error("Net:Sec:API", "Invalid user type in JWT")
				success = false
			}
			if access, ok = claims["access"].(float64); !ok {
				Error("Net:Sec:API", "Invalid access level in JWT")
				success = false
			} else {
				uinfo.Role = AuthLevel(access)
			}
			uinfo.Valid = token.Valid
		}
	}
	if !success {
		err = errors.New("Could not find relevent information in JWT token")
	}
	return uinfo, err
}

//IsAdmin - returns true if user associated with request is an admin
func IsAdmin(ctx echo.Context) (yes bool) {
	yes = false
	uinfo, err := RetrieveSessionInfo(ctx)
	if err == nil {
		yes = uinfo.Role <= Admin
	}
	return yes
}

//IsSuperUser - returns true if user is a super user
func IsSuperUser(ctx echo.Context) (yes bool) {
	yes = false
	uinfo, err := RetrieveSessionInfo(ctx)
	if err == nil {
		yes = uinfo.Role == Super
	}
	return yes
}

//IsNormalUser - returns true if user is a normal user
func IsNormalUser(ctx echo.Context) (yes bool) {
	yes = false
	uinfo, err := RetrieveSessionInfo(ctx)
	if err == nil {
		yes = uinfo.Role <= Normal
	}
	return yes
}

//BinderFunc - a function that binds data struct to response body
type BinderFunc func(ctx echo.Context) (interface{}, error)

//AddEndpoint - registers an REST endpoint
func AddEndpoint(ep *Endpoint) {
	endpoints = append(endpoints, ep)
}

//AddEndpoints - registers multiple REST categories
func AddEndpoints(eps ...*Endpoint) {
	for _, ep := range eps {
		AddEndpoint(ep)
	}
}

// ModifiedHTTPErrorHandler is the default HTTP error handler. It sends a
// JSON response with status code. [Modefied from echo.DefaultHTTPErrorHandler]
func ModifiedHTTPErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
		msg = fmt.Sprintf("%v, %v", err, he.Error())
	} else if e.Debug {
		msg = err.Error()
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg}
	}

	LogError("Net:HTTP", err)

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, msg)
		}
		if err != nil {
			LogError("Net:HTTP", err)
		}
	}
}

//InitServer - initializes all the registered endpoints
func InitServer(opts Options) {
	e.HideBanner = true
	e.HTTPErrorHandler = ModifiedHTTPErrorHandler
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[ACCSS] [Net:HTTP] ${status} : ${method} => ${path}\n",
	}))
	//Add middleware
	authenticator = opts.Authenticator
	authorizer = opts.Authorizer

	//rootPath is a package variable
	rootPath = opts.RootName + "/api/v" + opts.APIVersion + "/"
	accessPos = len(rootPath) + len("in/")
	root := e.Group(rootPath)
	in := root.Group("in/")

	//For checking token
	in.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: GetJWTKey(),
		ContextKey: "token",
	}))

	//For checking authorization level
	in.Use(authMiddleware)

	for _, ep := range endpoints {
		switch ep.Access {
		case Super:
			configure(in, "r0/", ep)
		case Admin:
			configure(in, "r1/", ep)
		case Normal:
			configure(in, "r2/", ep)
		case Monitor:
			configure(in, "r3/", ep)
		case Public:
			configure(root, "", ep)
		}
	}
}

//Serve - start the server
func Serve(port int) (err error) {
	printConfig()
	address := fmt.Sprintf(":%d", port)
	err = e.Start(address)
	return err
}

//GetRootPath - get base URL of the configured application's REST Endpoints
func GetRootPath() string {
	return rootPath
}

func configure(grp *echo.Group, urlPrefix string, ep *Endpoint) {
	var route *echo.Route
	switch ep.Method {
	case echo.CONNECT:
		route = grp.CONNECT(urlPrefix+ep.URL, ep.Func)
	case echo.DELETE:
		route = grp.DELETE(urlPrefix+ep.URL, ep.Func)
	case echo.GET:
		route = grp.GET(urlPrefix+ep.URL, ep.Func)
	case echo.HEAD:
		route = grp.HEAD(urlPrefix+ep.URL, ep.Func)
	case echo.OPTIONS:
		route = grp.OPTIONS(urlPrefix+ep.URL, ep.Func)
	case echo.PATCH:
		route = grp.PATCH(urlPrefix+ep.URL, ep.Func)
	case echo.POST:
		route = grp.POST(urlPrefix+ep.URL, ep.Func)
	case echo.PUT:
		route = grp.PUT(urlPrefix+ep.URL, ep.Func)
	case echo.TRACE:
		route = grp.TRACE(urlPrefix+ep.URL, ep.Func)
	}
	ep.Route = route
	if _, found := categories[ep.Category]; !found {
		categories[ep.Category] = make([]*Endpoint, 0, 20)
	}
	categories[ep.Category] = append(categories[ep.Category], ep)
}

func printConfig() {
	fmt.Println()
	fmt.Println("Endpoints: ")
	for category, eps := range categories {
		fmt.Printf("\t%10s\n", category)
		for _, ep := range eps {
			fmt.Printf("\t\t|-%10s - %10v - %-50s - %s\n",
				ep.Method,
				ep.Access,
				ep.Route.Path,
				ep.Comment)
		}
		fmt.Println()
	}
}
