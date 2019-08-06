package pg

import (
	"fmt"
	"strings"

	"github.com/varunamachi/teak"
	"gopkg.in/urfave/cli.v1"
)

//dataStorage - Postgres implementation for dataStorage interface
type dataStorage struct{}

// func toMap(obj interface{}) (out map[string]string) {
// 	out = make(map[string]string)
// 	teak.Walk(obj, &teak.WalkConfig{
// 		MaxDepth:         teak.InfiniteDepth,
// 		IgnoreContainers: false,
// 		FieldNameRetriever: func(field *reflect.StructField) string {
// 			jt := field.Tag.Get("json")
// 			if jt != "" {
// 				return jt
// 			}
// 			return field.Name
// 		},
// 		Visitor: func(state *teak.WalkerState) bool {
// 			if teak.IsBasicType(state.Current.Kind()) {
// 				out[state.Path] = state.Current.String()
// 			} else if state.Current.Kind() == reflect.Struct &&
// 				state.Current.Type() == reflect.TypeOf(time.Time{}) {
// 				// out[state.Path] = state.Current.Interface()
// 				tm := state.Current.Interface().(time.Time)

// 				return false
// 			}
// 			return true
// 		},
// 	})
// 	return out
// }

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
	fmt.Println(buf.String())
	// db.Exec(buf.String(), vals...)
	return err
}

//Update - updates the records in 'dtype' collection which are matched by
//the matcher query
func (mds *dataStorage) Update(
	dtype string,
	keyField string,
	key interface{},
	value interface{}) (err error) {
	return teak.LogError("t.crud.pg", err)
}

//Delete - deletes record matched by the matcher from collection 'dtype'
func (mds *dataStorage) Delete(
	dtype string,
	keyField string,
	key interface{}) (err error) {
	return teak.LogError("t.crud.pg", err)
}

//RetrieveOne - gets a record matched by given matcher from collection 'dtype'
func (mds *dataStorage) RetrieveOne(
	dtype string,
	keyField string,
	key interface{},
	out interface{}) (err error) {
	return teak.LogError("t.crud.pg", err)
}

//Count - counts the number of items for data type
func (mds *dataStorage) Count(
	dtype string, filter *teak.Filter) (count int, err error) {
	return count, teak.LogError("t.crud.pg", err)
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
	return teak.LogError("t.crud.pg", err)
}

//RetrieveWithCount - gets all the items from collection 'dtype' selected by
//filter & paged also gives the total count of items selected by filter
func (mds *dataStorage) RetrieveWithCount(
	dtype string,
	sortFiled string,
	offset int,
	limit int,
	filter *teak.Filter,
	out interface{}) (count int, err error) {
	return count, teak.LogError("t.crud.pg", err)
}

//GetFilterValues - provides values associated the fields defined in filter spec
func (mds *dataStorage) GetFilterValues(
	dtype string,
	specs teak.FilterSpecList) (values teak.M, err error) {
	return values, teak.LogError("t.crud.pg", err)
}

//GetFilterValuesX - get values for filter based on given filter
func (mds *dataStorage) GetFilterValuesX(
	dtype string,
	field string,
	specs teak.FilterSpecList,
	filter *teak.Filter) (values teak.M, err error) {
	return values, teak.LogError("t.crud.pg", err)
}

//Init - initialize the data storage - this needs to be run on each application
//start up
func (mds *dataStorage) Init() (err error) {
	return err
}

//Setup - setup has to be run when data storage structure changes, such as
//adding index, altering tables etc
func (mds *dataStorage) Setup(params teak.M) (err error) {
	return err
}

//Reset - reset clears the data without affecting the structure/schema
func (mds *dataStorage) Reset() (err error) {
	return err
}

//Destroy - deletes data and also structure
func (mds *dataStorage) Destroy() (err error) {
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
