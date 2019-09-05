package pg

import (
	"fmt"
	"os/user"
	"strconv"
	"strings"

	"github.com/varunamachi/teak"
	"gopkg.in/urfave/cli.v1"
)

//dataStorage - Postgres implementation for dataStorage interface
type dataStorage struct{}

//NewStorage - creates a new mongodb based data storage implementation
func NewStorage() teak.DataStorage {
	return &dataStorage{}
}

func (pg *dataStorage) Name() string {
	return "postgres"
}

//Create - creates an record in 'dtype' collection
func (pg *dataStorage) Create(
	dtype string, value interface{}) (err error) {
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
	_, err = defDB.Exec(buf.String(), vals...)
	return err
}

//Update - updates the records in 'dtype' collection which are matched by
//the matcher query
func (pg *dataStorage) Update(
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
	_, err = defDB.Exec(buf.String(), vals...)
	return err
}

//Delete - deletes record matched by the matcher from collection 'dtype'
func (pg *dataStorage) Delete(
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
	_, err = defDB.Exec(buf.String(), key)
	return err
}

//RetrieveOne - gets a record matched by given matcher from collection 'dtype'
func (pg *dataStorage) RetrieveOne(
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
	err = defDB.Select(out, buf.String(), key)
	return err
}

//Count - counts the number of items for data type
func (pg *dataStorage) Count(
	dtype string, filter *teak.Filter) (count int, err error) {
	defer func() {
		teak.LogErrorX("t.pg.store", "Failed to delete item", err)
	}()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s %s",
		dtype, generateSelector(filter))
	err = defDB.Select(&count, query)
	return count, err
}

//Retrieve - gets all the items from collection 'dtype' selected by filter &
//paged
func (pg *dataStorage) Retrieve(
	dtype string,
	sortFiled string,
	offset int,
	limit int,
	filter *teak.Filter,
	out interface{}) (err error) {
	selector := generateSelector(filter)
	var buf strings.Builder
	buf.Grow(100)
	buf.WriteString("SELECT * FROM ")
	buf.WriteString(dtype)
	buf.WriteString(selector)
	buf.WriteString(" OFFSET ")
	buf.WriteString(strconv.Itoa(offset))
	buf.WriteString(" LIMIT ")
	buf.WriteString(strconv.Itoa(limit))
	buf.WriteString(" ORDER BY ")
	buf.WriteString(sortFiled) //Check for minus sign?? like in mongo??
	err = defDB.Select(out, buf.String())
	return teak.LogError("t.pg.store", err)
}

//RetrieveWithCount - gets all the items from collection 'dtype' selected by
//filter & paged also gives the total count of items selected by filter
func (pg *dataStorage) RetrieveWithCount(
	dtype string,
	sortField string,
	offset int,
	limit int,
	filter *teak.Filter,
	out interface{}) (count int, err error) {
	//For now this is going to be bit unoptimized - we generate selector twice
	err = pg.Retrieve(dtype, sortField, offset, limit, filter, out)
	if err != nil {
		return count, err
	}
	count, err = pg.Count(dtype, filter)
	return count, err
}

//GetFilterValues - provides values associated the fields defined in filter spec
func (pg *dataStorage) GetFilterValues(
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
func (pg *dataStorage) Init(admin *teak.User, adminPass string, param teak.M) (
	err error) {
	val, err := pg.IsInitialized()
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
	err = pg.Setup(teak.M{})
	if err != nil {
		err = teak.LogErrorX("t.pg.store", "Failed to setup app", err)
		return err
	}
	uStore := NewUserStorage()
	err = uStore.CreateUser(admin)
	if err != nil {
		err = teak.LogErrorX("t.pg.store",
			"Failed to create initial super admin", err)
		return err
	}
	err = uStore.SetPassword(admin.ID, adminPass)
	if err != nil {
		err = teak.LogErrorX("t.pg.store",
			"Failed to set initial super user password", err)
		return err
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
			id				CHAR(128)		NOT NULL,
			email			VARCHAR(100)	NOT NULL,
			auth			INTEGER			NOT NULL,
			firstName		VARCHAR(64)		NOT NULL,
			lastName		VARCHAR(64)		,
			title			CHAR(10)		NOT NULL,
			fullName		VARCHAR(128)	NOT NULL,
			state			CHAR(10)		NOT NULL DEFAULT 'disabled',
			verID			CHAR(38),
			pwdExpiry		TIMESTAMPTZ,
			createdAt		TIMESTAMPTZ,
			createdBy		CHAR(128),
			modifiedAt		TIMESTAMPTZ,
			modifiedBy		CHAR(128),
			verifiedAt		TIMESTAMPTZ,
			props			JSONB,
			CONSTRAINT pk_id PRIMARY KEY(id)
		);`,
	},
	{
		name: "user_secret",
		query: `CREATE TABLE user_secret(
			userID  	CHAR(128)		PRIMARY KEY,
			phash		VARCHAR(256),
			FOREIGN KEY (userID) REFERENCES teak_user(id) ON DELETE CASCADE
		)`,
	},
	{
		name: "teak_event",
		query: `CREATE TABLE teak_event(
			id			CHAR(60)		PRIMARY KEY,
			op			CHAR(60),
			userID		CHAR(60),
			userName	CHAR(60),
			success		CHAR(60),
			error		CHAR(60),
			time		CHAR(60),
			data		JSONB
		)`,
	},
	{
		name: "teak_internal",
		query: `CREATE TABLE teak_internal(
			key 	VARCHAR(100)	PRIMARY KEY,
			value 	JSONB
		)`,
	},
}

//Setup - setup has to be run when data storage structure changes, such as
//adding index, altering tables etc
func (pg *dataStorage) Setup(params teak.M) (err error) {
	// for name, query := range tables {
	for _, tab := range tables {
		_, err = defDB.Exec(tab.query)
		if err != nil {
			err = teak.LogErrorX("t.pg.store", "Failed to create table '%s'",
				err, tab.name)
			break
		}
	}
	return err
}

//Reset - reset clears the data without affecting the structure/schema
func (pg *dataStorage) Reset() (err error) {
	for _, tab := range tables {
		query := fmt.Sprintf("DELETE FROM %s;", tab.name)
		_, err = defDB.Exec(query)
		if err != nil {
			teak.Error(
				"t.pg.store", "Failed clear data from %s: %v", tab.name, err)
			//break??
		}
	}
	return err
}

//Destroy - deletes data and also structure
func (pg *dataStorage) Destroy() (err error) {
	// for _, tab := range tables {
	for i := len(tables) - 1; i >= 0; i-- {
		tab := tables[i]
		query := fmt.Sprintf("DROP TABLE %s;", tab.name)
		_, err = defDB.Exec(query)
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
	user, err := user.Current()
	if err == nil {
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

//GetManageCommands - commands that can be used to manage this data storage
func (pg *dataStorage) GetManageCommands() (commands []cli.Command) {
	return commands
}

//IsInitialized - tells if data source is initialized
func (pg *dataStorage) IsInitialized() (yes bool, err error) {
	yes, err = pg.hasTable("teak_internal")
	return yes, err
}

func (pg *dataStorage) hasTable(tableName string) (yes bool, err error) {
	err = defDB.Get(&yes,
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
