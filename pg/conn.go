package pg

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/varunamachi/teak"
)

var db *sqlx.DB

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

//Connect - connect to postgres db with given connection string
func Connect(optStr string) (err error) {
	defer teak.LogErrorX("t.pg", "Failed to connect to postgres", err)
	db, err = sqlx.Open("postgres", optStr)
	db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)
	return err
}

//ConnectWithOpts - connect to postgresdb based on given options
func ConnectWithOpts(opts *ConnOpts) (err error) {
	return Connect(opts.String())
}

//Conn - gives connection to database
func Conn() *sqlx.DB {
	return db
}
