package pg

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/varunamachi/teak"
)

var defDB *sqlx.DB
var conns map[string]*sqlx.DB

//ConnOpts - postgres connection options
type ConnOpts struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
}

//String - get usable connection string
func (c *ConnOpts) String() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName)
}

//Connect - connects to DB based on connection string
func Connect(optStr string) (db *sqlx.DB, err error) {
	defer teak.LogErrorX("t.pg", "Failed to connect to postgres", err)
	db, err = sqlx.Open("postgres", optStr)
	if err == nil {
		defDB.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	}
	return db, err
}

//ConnectWithOpts - connect to postgresdb based on given options
func ConnectWithOpts(opts *ConnOpts) (db *sqlx.DB, err error) {
	return Connect(opts.String())
}

//NamedConn - gives connection to database associated with given name. If no
//connection exists with given name nil is returned. If name is empty default
//connection is returned
func NamedConn(name string) *sqlx.DB {
	if name == "" {
		return defDB
	}
	return conns[name]
}

//SetNamedConn - register a postgres connection against name
func SetNamedConn(name string, db *sqlx.DB) {
	conns[name] = db
}

//SetDefaultConn - sets the default postgres connection
func SetDefaultConn(db *sqlx.DB) {
	defDB = db
}

//Conn - Gives default connection
func Conn() *sqlx.DB {
	return defDB
}
