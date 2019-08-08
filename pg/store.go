package pg

import (
	"fmt"
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

func (mds *dataStorage) Name() string {
	return "postgres"
}

//Create - creates an record in 'dtype' collection
func (mds *dataStorage) Create(
	dtype string, value interface{}) (err error) {
	defer func() { teak.LogErrorX("t.crud.pg", "Failed to create item", err) }()
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
		buf.WriteString("?") //fix flat map to give strings
	}
	buf.WriteString(");")
	// fmt.Println(buf.String())
	_, err = db.Exec(buf.String(), vals...)
	return err
}

//Update - updates the records in 'dtype' collection which are matched by
//the matcher query
func (mds *dataStorage) Update(
	dtype string,
	keyField string,
	key interface{},
	value interface{}) (err error) {
	defer func() { teak.LogErrorX("t.crud.pg", "Failed to update item", err) }()
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
			buf.WriteString(" = ?")
			vals = append(vals, val)
		}
	}
	buf.WriteString(" WHERE ")
	buf.WriteString(keyField)
	buf.WriteString(" = ?;")
	vals = append(vals, key)
	// fmt.Println(buf.String())
	_, err = db.Exec(buf.String(), vals...)
	return err
}

//Delete - deletes record matched by the matcher from collection 'dtype'
func (mds *dataStorage) Delete(
	dtype string,
	keyField string,
	key interface{}) (err error) {
	defer func() { teak.LogErrorX("t.crud.pg", "Failed to delete item", err) }()
	var buf strings.Builder
	buf.WriteString("DELETE FROM ")
	buf.WriteString(dtype)
	buf.WriteString(" WHERE ")
	buf.WriteString(keyField)
	buf.WriteString(" = ?;")
	_, err = db.Exec(buf.String(), key)
	return err
}

//RetrieveOne - gets a record matched by given matcher from collection 'dtype'
func (mds *dataStorage) RetrieveOne(
	dtype string,
	keyField string,
	key interface{},
	out interface{}) (err error) {
	defer func() { teak.LogErrorX("t.crud.pg", "Failed to delete item", err) }()
	var buf strings.Builder
	buf.WriteString("SELECT * FROM ")
	buf.WriteString(dtype)
	buf.WriteString(" WHERE ")
	buf.WriteString(keyField)
	buf.WriteString(" = ?;")
	err = db.Select(out, buf.String(), key)
	return err
}

//Count - counts the number of items for data type
func (mds *dataStorage) Count(
	dtype string, filter *teak.Filter) (count int, err error) {
	defer func() { teak.LogErrorX("t.crud.pg", "Failed to delete item", err) }()
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s",
		dtype, generateSelector(filter))
	err = db.Select(&count, query)
	return count, err
}

//Retrieve - gets all the items from collection 'dtype' selected by filter &
//paged
func (mds *dataStorage) Retrieve(
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
	buf.WriteString(" WHERE ")
	buf.WriteString(selector)
	buf.WriteString(" OFFSET ")
	buf.WriteString(strconv.Itoa(offset))
	buf.WriteString(" LIMIT ")
	buf.WriteString(strconv.Itoa(limit))
	buf.WriteString(" ORDER BY ")
	buf.WriteString(sortFiled) //Check for minus sign?? like in mongo??
	err = db.Select(out, buf.String())
	return teak.LogError("t.crud.pg", err)
}

//RetrieveWithCount - gets all the items from collection 'dtype' selected by
//filter & paged also gives the total count of items selected by filter
func (mds *dataStorage) RetrieveWithCount(
	dtype string,
	sortField string,
	offset int,
	limit int,
	filter *teak.Filter,
	out interface{}) (count int, err error) {
	//For now this is going to be bit unoptimized - we generate selector twice
	err = mds.Retrieve(dtype, sortField, offset, limit, filter, out)
	if err != nil {
		return count, err
	}
	count, err = mds.Count(dtype, filter)
	return count, err
}

//GetFilterValues - provides values associated the fields defined in filter spec
func (mds *dataStorage) GetFilterValues(
	dtype string,
	specs teak.FilterSpecList) (values teak.M, err error) {
	//@TODO later
	defer func() {
		teak.LogErrorX("t.crud.pg",
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

//GetFilterValuesX - get values for filter based on given filter
func (mds *dataStorage) GetFilterValuesX(
	dtype string,
	field string,
	specs teak.FilterSpecList,
	filter *teak.Filter) (values teak.M, err error) {
	defer func() {
		teak.LogErrorX("t.crud.pg",
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

//Init - initialize the data storage - this needs to be run on each application
//start up
func (mds *dataStorage) Init() (err error) {
	return err
}

//Setup - setup has to be run when data storage structure changes, such as
//adding index, altering tables etc
func (mds *dataStorage) Setup(params teak.M) (err error) {
	utq := `
		CREATE TABLE teak_user(
			id				CHAR(128)		PRIMARY KEY
			email			VARCHAR(100)	NOT NULL
			auth			INTEGER			NOT NULL
			firstName		VARCHAR(64)		NOT NULL
			lastName		VARCHAR(64)		
			title			CHAR(10)		NOT NULL
			fullName		VARCHAR(128)	NOT NULL
			state			CHAR(10)		NOT NULL DEFAULT 'disabled'
			verID			CHAR(38)		
			pwdExpiry		TIMESTAMPZ
			createdAt		TIMESTAMPZ
			createdBy		CHAR(128)
			modifiedAt		TIMESTAMPZ
			modifiedBy		CHAR(128)
			verified		BOOLEAN
			props			HSTORE
		);`
	_, err = db.Exec(utq)
	if err != nil {
		return err
	}

	// etq := `
	// 	CREATE TABLE teak_event(
	// 		op			CHAR(60)
	// 		userID		CHAR(60)
	// 		userName	CHAR(60)
	// 		success		CHAR(60)
	// 		error		CHAR(60)
	// 		time		CHAR(60)
	// 		data		HSTORE
	// 	)
	// `

	//Create events table
	//Create housekeeping table
	return err
}

//Reset - reset clears the data without affecting the structure/schema
func (mds *dataStorage) Reset() (err error) {
	//Clear user table
	//Clear events table
	//Reset house keeping table

	return err
}

//Destroy - deletes data and also structure
func (mds *dataStorage) Destroy() (err error) {
	//Delete everything
	return err
}

//Wrap - wraps a command with flags required to connect to this data source
func (mds *dataStorage) Wrap(cmd *cli.Command) *cli.Command {
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
func (mds *dataStorage) GetManageCommands() (commands []cli.Command) {
	return commands
}

func generateSelector(filter *teak.Filter) (selector string) {
	return selector
}
