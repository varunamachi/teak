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
	return CollectionWithDB(defaultDB, collectionName)
}

// CollectionWithDB - gives a reference to a collection in given database
func CollectionWithDB(db, coll string) *mongo.Collection {
	return mongoStore.client.Database(db).Collection(coll)
}

//SetDefaultDB - sets the default DB
func SetDefaultDB(defDB string) {
	defaultDB = defDB
}

func GetDefaultDB()
