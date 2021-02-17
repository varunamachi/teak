package mg

import (
	"bytes"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnType - type of cluster to connect to
type ConnType string

// Single - No cluster, single instance
const Single ConnType = "Single"

// ReplicaSet - A replica set cluster
const ReplicaSet ConnType = "ReplicaSet"

// Sharded - Sharded database
const Sharded ConnType = "Sharded"

//ConnOpts - options for connecting to a mongodb instance
type ConnOpts struct {
	URLs     []string `json:"uri"`
	Type     ConnType `json:"type"`
	User     string   `json:"user"`
	Password string   `json:"password"`
}

func (co *ConnOpts) String() string {
	var buf bytes.Buffer
	buf.WriteString("mongodb://")
	for i, url := range co.URLs {
		buf.WriteString(url)
		if len(co.URLs) > 1 && i < len(co.URLs)-1 {
			buf.WriteString(",")
		}
	}
	if co.Type == ReplicaSet {
		buf.WriteString("/?replicaSet=replset")
	}
	return co.String()
}

//store - holds mongodb connection handle and information
type store struct {
	client *mongo.Client
}

var mongoStore *store
var defaultDB = "teak"

// Connect - connects to a mongoDB instance or a cluster based on the
// the options provided
func Connect(gtx context.Context, opts *ConnOpts) error {
	clientOpts := options.Client().ApplyURI(opts.String())
	if opts.User != "" {
		creds := options.Credential{
			AuthMechanism: "PLAIN",
			Username:      opts.User,
			Password:      opts.Password,
		}
		clientOpts.SetAuth(creds)
	}
	client, err := mongo.Connect(gtx, clientOpts)
	if err != nil {
		return err
	}
	mongoStore = &store{
		client: client,
	}
	return nil
}

// C - get a handle to collection in the default database, single letter name
// to have nice way to transition from mgo
func C(collectionName string) *mongo.Collection {
	return CcollectionWithDB(defaultDB, collectionName)
}

// CcollectionWithDB - gives a reference to a collection in given database
func CcollectionWithDB(db, coll string) *mongo.Collection {
	return mongoStore.client.Database(db).Collection(coll)
}

// //Conn - represents a mongdb connection
// type Conn struct {
// 	*mongo.Collection
// }

// //SetDefaultDB - sets the default DB
// func SetDefaultDB(defDB string) {
// 	defaultDB = defDB
// }

// //Close - closes mongodb connection
// func (conn *Conn) Close() {
// 	conn.Database().Close()
// }

// //toOptStr - converts mongodb options to a string that can be used as URL to
// //connect to a mongodb instance
// func toOptStr(options []*ConnOpts) string {
// 	var buf bytes.Buffer
// 	for i, opt := range options {
// 		//userName:password@host:port[,userName:password@host:port...]
// 		//buf.WriteString("mongo://")
// 		if len(opt.User) != 0 {
// 			buf.WriteString(opt.User)
// 			buf.WriteString(":")
// 			buf.WriteString(opt.Password)
// 			buf.WriteString("@")
// 		}
// 		if len(opt.Host) != 0 {
// 			buf.WriteString(opt.Host)
// 		} else {
// 			buf.WriteString("localhost")
// 		}

// 		if opt.Port != 0 {
// 			buf.WriteString(":")
// 			buf.WriteString(strconv.Itoa(opt.Port))
// 		}
// 		if len(options) > 1 && i < len(options)-1 {
// 			//In case of multiple addresses
// 			buf.WriteString(",")
// 		}
// 	}
// 	co := buf.String()
// 	return co
// }

// //ConnectSingle - connects to single instance of mongodb server
// func ConnectSingle(opts *ConnOpts) (err error) {
// 	err = Connect([]*ConnOpts{opts})
// 	if err == nil {
// 		teak.Info("DB:Mongo", "Connected to mongo://%s:%d",
// 			opts.Host, opts.Port)
// 	} else {
// 		teak.Error("DB:Mongo", "Failed to connected to mongo://%s:%d",
// 			opts.Host, opts.Port)
// 	}
// 	return err
// }

// //Connect - connects to one or more mirrors of mongodb server
// func Connect(opts []*ConnOpts) (err error) {
// 	var sess *mgo.Session
// 	optString := toOptStr(opts)
// 	sess, err = mgo.Dial(optString)
// 	if err == nil {
// 		sess.SetMode(mgo.Monotonic, true)
// 		mongoStore = &store{
// 			session: sess,
// 			opts:    opts,
// 		}
// 		teak.Info("DB:Mongo", "Connected to mongoDB")
// 	}
// 	return teak.LogError("DB:Mongo", err)
// }

// //NewConn - creates a new connection to mogodb
// func NewConn(dbName string) (conn *Conn) {
// 	conn = &Conn{
// 		Database: mongoStore.session.Copy().DB(dbName),
// 	}
// 	return conn
// }

// //DefaultConn - creates a connection to default DB
// func DefaultConn() *Conn {
// 	return NewConn(defaultDB)
// }

// //CloseConn - closes the mongodb connection
// func CloseConn() {
// 	mongoStore.session.Close()
// }
