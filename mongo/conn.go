package mongo

import (
	"bytes"
	"strconv"

	"github.com/varunamachi/teak"
	"gopkg.in/mgo.v2"
)

//ConnOpts - options for connecting to a mongodb instance
type ConnOpts struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

//store - holds mongodb connection handle and information
type store struct {
	session *mgo.Session
	opts    []*ConnOpts
}

var mongoStore *store
var defaultDB = "teak"

//Conn - represents a mongdb connection
type Conn struct {
	*mgo.Database
}

//SetDefaultDB - sets the default DB
func SetDefaultDB(defDB string) {
	defaultDB = defDB
}

//Close - closes mongodb connection
func (conn *Conn) Close() {
	conn.Session.Close()
}

//toOptStr - converts mongodb options to a string that can be used as URL to
//connect to a mongodb instance
func toOptStr(options []*ConnOpts) string {
	var buf bytes.Buffer
	for i, opt := range options {
		//userName:password@host:port[,userName:password@host:port...]
		//buf.WriteString("mongo://")
		if len(opt.User) != 0 {
			buf.WriteString(opt.User)
			buf.WriteString(":")
			buf.WriteString(opt.Password)
			buf.WriteString("@")
		}
		if len(opt.Host) != 0 {
			buf.WriteString(opt.Host)
		} else {
			buf.WriteString("localhost")
		}

		if opt.Port != 0 {
			buf.WriteString(":")
			buf.WriteString(strconv.Itoa(opt.Port))
		}
		if len(options) > 1 && i < len(options)-1 {
			//In case of multiple addresses
			buf.WriteString(",")
		}
	}
	co := buf.String()
	return co
}

//ConnectSingle - connects to single instance of mongodb server
func ConnectSingle(opts *ConnOpts) (err error) {
	err = Connect([]*ConnOpts{opts})
	if err == nil {
		teak.Info("DB:Mongo", "Connected to mongo://%s:%d",
			opts.Host, opts.Port)
	} else {
		teak.Error("DB:Mongo", "Failed to connected to mongo://%s:%d",
			opts.Host, opts.Port)
	}
	return err
}

//Connect - connects to one or more mirrors of mongodb server
func Connect(opts []*ConnOpts) (err error) {
	var sess *mgo.Session
	optString := toOptStr(opts)
	sess, err = mgo.Dial(optString)
	if err == nil {
		sess.SetMode(mgo.Monotonic, true)
		mongoStore = &store{
			session: sess,
			opts:    opts,
		}
		teak.Info("DB:Mongo", "Connected to mongoDB")
	}
	return teak.LogError("DB:Mongo", err)
}

//NewConn - creates a new connection to mogodb
func NewConn(dbName string) (conn *Conn) {
	conn = &Conn{
		Database: mongoStore.session.Copy().DB(dbName),
	}
	return conn
}

//DefaultConn - creates a connection to default DB
func DefaultConn() *Conn {
	return NewConn(defaultDB)
}

//CloseConn - closes the mongodb connection
func CloseConn() {
	mongoStore.session.Close()
}
