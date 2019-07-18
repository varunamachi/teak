package teak

import echo "github.com/labstack/echo/v4"

//auth
var authenticator Authenticator
var authorizer Authorizer

//Network
var categories = make(map[string][]*Endpoint)
var endpoints = make([]*Endpoint, 0, 200)
var e = echo.New()
var accessPos = 0
var rootPath = ""
var jwtKey []byte
