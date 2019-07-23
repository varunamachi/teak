package pg

import (
	"fmt"

	"github.com/varunamachi/teak"
)

type ConnOpts struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
}

func (c *ConnOpts) String() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName)
}

func Connect(optStr string) (err error) {
	defer teak.LogErrorX("t.pg", "Failed to connect to postgres", err)
	return err
}

func ConnectWithOpts(opts ConnOpts) (err error) {
	return Connect(opts.String())
}
