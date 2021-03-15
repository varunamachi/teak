package pg

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/varunamachi/teak"
	"gopkg.in/urfave/cli.v1"
)

//DBAttr - used to store generic data in a JSONB column
type DBAttr map[string]interface{}

//Value - convert attribute to JSON while storing
func (a DBAttr) Value() (driver.Value, error) {
	return json.Marshal(a)
}

//Scan - read JSON into attributes
func (a *DBAttr) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("Invalid scan, expected bytes")
	}
	return json.Unmarshal(b, &a)
}

//dataStorage - Postgres implementation for dataStorage interface
type dataStorage struct{}

//NewStorage - creates a new mongodb based data storage implementation
func NewStorage() teak.DataStorage {
	// TODO - update the data store to satisfy the DataStorage interface
	return &dataStorage{}
	// return nil
}

func (pg *dataStorage) Name() string {
	return "postgres"
}

//Create - creates an record in 'dtype' collection
func (pg *dataStorage) Create(
	gtx context.Context, dtype string, value interface{}) (err error) {
	defer func() {
		teak.LogErrorX("t.pg.store", "Failed to create item", err)
	}()
	hdl := teak.GetItemHandler(dtype)
	if hdl == nil {
		err = fmt.Errorf("Failed to get handler for data type %s", dtype)
		return err
	}
	mp := teak.ToFlatMap(value, "json")
	buf := strings.Builder{}
	buf.WriteString("INSERT INTO ")
	buf.WriteString(dtype)
	buf.WriteString("(")
	for i, propName := range hdl.PropNames() {
		if _, has := mp[propName]; !has {
			continue
		}
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(propName)
	}
	buf.WriteString(") VALUES (")
	vals := make([]interface{}, 0, len(mp))
	for i, propName := range hdl.PropNames() {
		val, has := mp[propName]
		if !has {
			continue
		}
		vals = append(vals, val)
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("$") //fix flat map to give strings
		buf.WriteString(strconv.Itoa(i + 1))
	}
	buf.WriteString(");")
	// fmt.Println(buf.String())
	_, err = defDB.ExecContext(gtx, buf.String(), vals...)
	return err
}

//Update - updates the records in 'dtype' collection which are matched by
//the matcher query
func (pg *dataStorage) Update(
	gtx context.Context,
	dtype string,
	keyField string,
	key interface{},
	value interface{}) (err error) {
	defer func() {
		teak.LogErrorX("t.pg.store", "Failed to update item", err)
	}()
	hdl := teak.GetItemHandler(dtype)
	if hdl == nil {
		err = fmt.Errorf("Failed to get handler for data type %s", dtype)
		return err
	}
	mp := teak.ToFlatMap(value, "json")
	buf := strings.Builder{}
	buf.WriteString("UPDATE ")
	buf.WriteString(dtype)
	buf.WriteString(" SET ")
	vals := make([]interface{}, 0, len(mp))
	for i, propName := range hdl.PropNames() {
		if val, has := mp[propName]; has && val != keyField {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(propName)
			buf.WriteString(" = $")
			buf.WriteString(strconv.Itoa(i + 1))
			vals = append(vals, val)
		}
	}
	buf.WriteString(" WHERE ")
	buf.WriteString(keyField)
	buf.WriteString(" = $")
	buf.WriteString(strconv.Itoa(len(hdl.PropNames()) + 1))
	vals = append(vals, key)
	// fmt.Println(buf.String())
	_, err = defDB.ExecContext(gtx, buf.String(), vals...)
	return err
}

//Delete - deletes record matched by the matcher from collection 'dtype'
func (pg *dataStorage) Delete(
	gtx context.Context,
	dtype string,
	keyField string,
	key interface{}) (err error) {
	defer func() {
		teak.LogErrorX("t.pg.store", "Failed to delete item", err)
	}()
	var buf strings.Builder
	buf.WriteString("DELETE FROM ")
	buf.WriteString(dtype)
	buf.WriteString(" WHERE ")
	buf.WriteString(keyField)
	buf.WriteString(" = $1")
	_, err = defDB.ExecContext(gtx, buf.String(), key)
	return err
}

//RetrieveOne - gets a record matched by given matcher from collection 'dtype'
func (pg *dataStorage) RetrieveOne(
	gtx context.Context,
	dtype string,
	keyField string,
	key interface{},
	out interface{}) (err error) {
	defer func() {
		teak.LogErrorX("t.pg.store", "Failed to delete item", err)
	}()
	var buf strings.Builder
	buf.WriteString("SELECT * FROM ")
	buf.WriteString(dtype)
	buf.WriteString(" WHERE ")
	buf.WriteString(keyField)
	buf.WriteString(" = $1")
	err = defDB.SelectContext(gtx, out, buf.String(), key)
	return err
}

//Count - counts the number of items for data type
func (pg *dataStorage) Count(
	gtx context.Context,
	dtype string,
	filter *teak.Filter) (count int64, err error) {
	defer func() {
		teak.LogErrorX("t.pg.store", "Failed to delete item", err)
	}()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s %s",
		dtype, generateSelector(filter))
	err = defDB.SelectContext(gtx, &count, query)
	return count, err
}

//Retrieve - gets all the items from collection 'dtype' selected by filter &
//paged
func (pg *dataStorage) Retrieve(
	gtx context.Context,
	dtype string,
	sortFiled string,
	offset int64,
	limit int64,
	filter *teak.Filter,
	out interface{}) (err error) {
	selector := generateSelector(filter)
	var buf strings.Builder
	buf.Grow(100)
	buf.WriteString("SELECT * FROM ")
	buf.WriteString(dtype)
	buf.WriteString(selector)
	buf.WriteString(" OFFSET ")
	buf.WriteString(strconv.FormatInt(offset, 10))
	buf.WriteString(" LIMIT ")
	buf.WriteString(strconv.FormatInt(limit, 10))
	buf.WriteString(" ORDER BY ")
	buf.WriteString(sortFiled) //Check for minus sign?? like in mongo??
	err = defDB.SelectContext(gtx, out, buf.String())
	return teak.LogError("t.pg.store", err)
}

//RetrieveWithCount - gets all the items from collection 'dtype' selected by
//filter & paged also gives the total count of items selected by filter
func (pg *dataStorage) RetrieveWithCount(
	gtx context.Context,
	dtype string,
	sortField string,
	offset int64,
	limit int64,
	filter *teak.Filter,
	out interface{}) (count int64, err error) {
	//For now this is going to be bit unoptimized - we generate selector twice
	err = pg.Retrieve(gtx, dtype, sortField, offset, limit, filter, out)
	if err != nil {
		return count, err
	}
	count, err = pg.Count(gtx, dtype, filter)
	return count, err
}

//GetFilterValues - provides values associated the fields defined in filter spec
func (pg *dataStorage) GetFilterValues(
	gtx context.Context,
	dtype string,
	specs teak.FilterSpecList) (values teak.M, err error) {
	//@TODO later
	defer func() {
		teak.LogErrorX("t.pg.store",
			"Failed to fetch filter values", err)
	}()
	///@TODO - implement
	for _, spec := range specs {
		switch spec.Type {
		case teak.Prop:
			//select distinct
		case teak.Array:
			//select distinct hstore?
		case teak.Date:
			//max-min
		case teak.Boolean: //nothing
		case teak.Search: //nothing
		case teak.Static: //nothing
		}
	}
	return values, err
}

//GetFilterValuesX - get values for filter based on given filter
func (pg *dataStorage) GetFilterValuesX(
	gtx context.Context,
	dtype string,
	field string,
	specs teak.FilterSpecList,
	filter *teak.Filter) (values teak.M, err error) {
	defer func() {
		teak.LogErrorX("t.pg.store",
			"Failed to fetch filter values", err)
	}()
	///@TODO - implement
	for _, spec := range specs {
		switch spec.Type {
		case teak.Prop:
		case teak.Array:
		case teak.Date:
		case teak.Boolean:
		case teak.Search:
		case teak.Static:
		}
	}
	return values, err
}

//Init - initialize the data storage for the first time, sets it upda and also
//creates the first admin user. Data store can be initialized only once
func (pg *dataStorage) Init(
	gtx context.Context,
	admin *teak.User,
	adminPass string,
	param teak.M) (err error) {
	val, err := pg.IsInitialized(gtx)
	if err != nil {
		err = teak.LogErrorX("t.pg.store",
			"Failed to check initialization status of PG store", err)
		return err
	}
	if val {
		teak.Info("t.pg.store", "Store already initialized.")
		teak.Info("t.pg.store",
			"If you want to update the structure of the store, use Setup")
		return err
	}
	err = pg.Setup(gtx, teak.M{})
	if err != nil {
		err = teak.LogErrorX("t.pg.store", "Failed to setup app", err)
		return err
	}
	uStore := NewUserStorage()
	idHash, err := uStore.CreateUser(gtx, admin)
	if err != nil {
		err = teak.LogErrorX("t.pg.store",
			"Failed to create initial super admin", err)
		return err
	}
	err = uStore.SetPassword(gtx, idHash, adminPass)
	if err != nil {
		err = teak.LogErrorX("t.pg.store",
			"Failed to set initial super user password", err)
		return err
	}
	_, err = defDB.ExecContext(gtx, fmt.Sprintf(`
			INSERT INTO teak_internal(name, val) VALUES 
				( 'initialized', '{ "value": true }'),
				( 'initializedAt', '{ "value": "%s" }')
		`, time.Now().Format(time.RFC3339)))
	if err != nil {
		err = teak.LogErrorX("t.pg.store",
			"Failed to mark store as initialized", err)
	}
	return err
}

var tables = []struct {
	name  string
	query string
}{
	{
		name: "teak_user",
		query: `CREATE TABLE teak_user(
			id				VARCHAR(128)	NOT NULL,
			email			VARCHAR(100)	NOT NULL,
			auth			INTEGER			NOT NULL,
			first_name		VARCHAR(64)		NOT NULL,
			last_name		VARCHAR(64)		,
			title			VARCHAR(10)		NOT NULL,
			full_name		VARCHAR(128)	NOT NULL,
			state			VARCHAR(10)		NOT NULL DEFAULT 'disabled',
			ver_id			VARCHAR(38),
			pwd_expiry		TIMESTAMPTZ,
			created_at		TIMESTAMPTZ,
			created_by		VARCHAR(128),
			modified_at		TIMESTAMPTZ,
			modified_by		VARCHAR(128),
			verified_at		TIMESTAMPTZ,
			props			JSONB,
			CONSTRAINT pk_id PRIMARY KEY(id)
		);`,
	},
	{
		name: "user_secret",
		query: `CREATE TABLE user_secret(
			user_id  	VARCHAR(128)		PRIMARY KEY,
			phash		VARCHAR(256),
			FOREIGN KEY (user_id) REFERENCES teak_user(id) ON DELETE CASCADE
		)`,
	},
	{
		name: "teak_event",
		query: `CREATE TABLE teak_event(
			id			VARCHAR(60)		PRIMARY KEY,
			op			VARCHAR(60),
			user_id		VARCHAR(60),
			user_name	VARCHAR(60),
			success		VARCHAR(60),
			error		VARCHAR(60),
			time		VARCHAR(60),
			data		JSONB
		)`,
	},
	{
		name: "teak_internal",
		query: `CREATE TABLE teak_internal(
			name	CHAR(128)	PRIMARY KEY,
			val		JSONB
		)`,
	},
}

//Setup - setup has to be run when data storage structure changes, such as
//adding index, altering tables etc
func (pg *dataStorage) Setup(gtx context.Context, params teak.M) (err error) {
	// for name, query := range tables {
	for _, tab := range tables {
		_, err = defDB.ExecContext(gtx, tab.query)
		if err != nil {
			err = teak.LogErrorX("t.pg.store", "Failed to create table '%s'",
				err, tab.name)
			break
		}
	}
	return err
}

//Reset - reset clears the data without affecting the structure/schema
func (pg *dataStorage) Reset(gtx context.Context) (err error) {
	for _, tab := range tables {
		query := fmt.Sprintf("DELETE FROM %s;", tab.name)
		_, err = defDB.ExecContext(gtx, query)
		if err != nil {
			teak.Error(
				"t.pg.store", "Failed clear data from %s: %v", tab.name, err)
			//break??
		}
	}
	return err
}

//Destroy - deletes data and also structure
func (pg *dataStorage) Destroy(gtx context.Context) (err error) {
	// for _, tab := range tables {
	for i := len(tables) - 1; i >= 0; i-- {
		tab := tables[i]
		query := fmt.Sprintf("DROP TABLE %s;", tab.name)
		_, err = defDB.ExecContext(gtx, query)
		if err != nil {
			teak.Warn(
				"t.pg.store", "Failed delete table '%s': %v", tab.name, err)
		}
	}
	return err
}

//Wrap - wraps a command with flags required to connect to this data source
func (pg *dataStorage) Wrap(cmd *cli.Command) *cli.Command {
	var curUserName string
	if user, err := user.Current(); err == nil {
		curUserName = user.Username
	}
	pgFlags := []cli.Flag{
		cli.StringFlag{
			Name:   "pg-host",
			Value:  "localhost",
			Usage:  "Address of the host running postgres",
			EnvVar: "PG_HOST",
		},
		cli.IntFlag{
			Name:   "pg-port",
			Value:  5432,
			Usage:  "Port on which postgres is listening",
			EnvVar: "PG_PORT",
		},
		cli.StringFlag{
			Name:   "pg-db",
			Value:  "",
			Usage:  "Database name",
			EnvVar: "PG_DB",
		},
		cli.StringFlag{
			Name:   "pg-user",
			Value:  curUserName,
			Usage:  "Postgres user name",
			EnvVar: "PG_USER",
		},
		cli.StringFlag{
			Name:   "pg-pass",
			Value:  "",
			Usage:  "Postgres password for connection",
			EnvVar: "PG_PASS",
		},
	}

	cmd.Flags = append(cmd.Flags, pgFlags...)
	if cmd.Before == nil {
		cmd.Before = requirePostgres
	} else {
		otherBefore := cmd.Before
		cmd.Before = func(ctx *cli.Context) (err error) {
			err = requirePostgres(ctx)
			if err == nil {
				err = otherBefore(ctx)
			}
			return err
		}
	}
	return cmd
}

//Wrap - wraps a command with flags required to connect to this data source
// func (pg *dataStorage) WithFlags(flags ...cli.Flag) []cli.Flag {
// 	var curUserName string
// 	user, err := user.Current()
// 	if err == nil {
// 		curUserName = user.Username
// 	}
// 	return append(flags,
// 		cli.StringFlag{
// 			Name:   "pg-host",
// 			Value:  "localhost",
// 			Usage:  "Address of the host running postgres",
// 			EnvVar: "PG_HOST",
// 		},
// 		cli.IntFlag{
// 			Name:   "pg-port",
// 			Value:  5432,
// 			Usage:  "Port on which postgres is listening",
// 			EnvVar: "PG_PORT",
// 		},
// 		cli.StringFlag{
// 			Name:   "pg-db",
// 			Value:  "",
// 			Usage:  "Database name",
// 			EnvVar: "PG_DB",
// 		},
// 		cli.StringFlag{
// 			Name:   "pg-user",
// 			Value:  curUserName,
// 			Usage:  "Postgres user name",
// 			EnvVar: "PG_USER",
// 		},
// 		cli.StringFlag{
// 			Name:   "pg-pass",
// 			Value:  "",
// 			Usage:  "Postgres password for connection",
// 			EnvVar: "PG_PASS",
// 		},
// 	)
// }

//GetManageCommands - commands that can be used to manage this data storage
func (pg *dataStorage) GetManageCommands() (commands []cli.Command) {
	return commands
}

//IsInitialized - tells if data source is initialized
func (pg *dataStorage) IsInitialized(
	gtx context.Context) (yes bool, err error) {
	yes, err = pg.HasTable(gtx, "teak_internal")
	if err != nil {
		err = teak.LogErrorX("t.pg.store",
			"Failed check if store is initialized", err)
		return yes, err
	}
	err = defDB.GetContext(gtx, &yes,
		`SELECT val->'value' FROM teak_internal WHERE name = 'initialized'`)
	if err != nil {
		teak.Debug("t.pg.store",
			"Failed to check initialization status of storage: %v", err)
		yes, err = false, nil
	}
	return yes, err
}

func (pg *dataStorage) InitializedAt(
	gtx context.Context) (t time.Time, err error) {
	str := ""
	err = defDB.GetContext(gtx, &str,
		`SELECT val->'value' FROM teak_internal 
		WHERE name = 'initializedAt'`)
	if err != nil {
		err = teak.LogErrorX("t.pg.store", "Failed to get init date", err)
		return t, err
	}
	//Remove the double quotes with -> str[1:len(str)-1]
	t, err = time.Parse(time.RFC3339, str[1:len(str)-1])
	if err != nil {
		err = teak.LogErrorX("t.pg.store",
			"Failed to parse init date", err)
		return t, err
	}
	return t, err
}

func (pg *dataStorage) HasTable(
	gtx context.Context, tableName string) (yes bool, err error) {
	err = defDB.GetContext(gtx, &yes,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = $1)`,
		tableName)
	return yes, teak.LogErrorX("t.pg.store",
		"Failed to check if table %s exists", err, tableName)
}

func generateSelector(filter *teak.Filter) (selector string) {
	//Will have to generate WHERE keyword if the filter is not empty
	return selector
}
