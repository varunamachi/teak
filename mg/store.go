package mg

import (
	"context"
	"runtime"
	"strings"

	"github.com/jinzhu/now"
	"github.com/varunamachi/teak"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/urfave/cli.v1"
)

//dataStorage - MongoDB implementation for dataStorage interface
type dataStorage struct{}

//NewStorage - creates a new mongodb based data storage implementation
func NewStorage() teak.DataStorage {
	return &dataStorage{}
	// TODO - make dataStorage satisfy teak.DataStorage interface
}

//logMongoError - if error is not mog.ErrNotFound return null otherwise log the
//error and return the given error
func logMongoError(module string, err error) (out error) {
	if err != nil && err != mongo.ErrNoDocuments {
		_, file, line, _ := runtime.Caller(1)
		teak.Error(module, "%s -- %s @ %d",
			err.Error(),
			file,
			line)
		out = err
	} else {
		err = nil
	}
	return out
}

func (mds *dataStorage) Name() string {
	return "mongo"
}

//Create - creates an record in 'dtype' collection
func (mds *dataStorage) Create(
	gtx context.Context,
	dtype string, value interface{}) error {
	_, err := C(dtype).InsertOne(gtx, value)
	return logMongoError("t.mongo.data", err)
}

//Update - updates the records in 'dtype' collection which are matched by
//the matcher query
func (mds *dataStorage) Update(
	gtx context.Context,
	dtype string,
	keyField string,
	key interface{},
	value interface{}) error {

	_, err := C(dtype).UpdateOne(gtx, bson.M{
		keyField: key,
	}, value)
	return logMongoError("t.mongo.store", err)
}

//Delete - deletes record matched by the matcher from collection 'dtype'
func (mds *dataStorage) Delete(
	gtx context.Context,
	dtype string,
	keyField string,
	key interface{}) error {

	_, err := C(dtype).DeleteOne(gtx, bson.M{
		keyField: key,
	})
	return logMongoError("t.mongo.store", err)
}

//RetrieveOne - gets a record matched by given matcher from collection 'dtype'
func (mds *dataStorage) RetrieveOne(
	gtx context.Context,
	dtype string,
	keyField string,
	key interface{},
	out interface{}) error {

	res := C(dtype).FindOne(gtx, bson.M{
		keyField: key,
	})
	err := Decode(res, out)
	return logMongoError("t.mongo.store", err)
}

//Count - counts the number of items for data type
func (mds *dataStorage) Count(
	gtx context.Context,
	dtype string,
	filter *teak.Filter) (int64, error) {
	//@TODO handle filters
	selector := GenerateSelector(filter)
	count, err := C(dtype).CountDocuments(gtx, selector)
	return count, logMongoError("t.mongo.store", err)
}

//Retrieve - gets all the items from collection 'dtype' selected by filter &
//paged
func (mds *dataStorage) Retrieve(
	gtx context.Context,
	dtype string,
	sortField string,
	offset int64,
	limit int64,
	filter *teak.Filter,
	out interface{}) error {
	selector := GenerateSelector(filter)
	fopts := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(GetSort(sortField))
	cur, err := C(dtype).Find(gtx, selector, fopts)
	if err != nil {
		return logMongoError("t.mongo.store", err)
	}
	defer cur.Close(gtx)
	err = cur.All(gtx, out)
	return logMongoError("t.mongo.store", err)
}

//RetrieveWithCount - gets all the items from collection 'dtype' selected by
//filter & paged also gives the total count of items selected by filter
func (mds *dataStorage) RetrieveWithCount(
	gtx context.Context,
	dtype string,
	sortField string,
	offset int64,
	limit int64,
	filter *teak.Filter,
	out interface{}) (int64, error) {
	selector := GenerateSelector(filter)
	fopts := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(GetSort(sortField))
	cur, err := C(dtype).Find(gtx, selector, fopts)
	if err != nil {
		return 0, logMongoError("t.mongo.store", err)
	}
	defer cur.Close(gtx)
	err = cur.All(gtx, out)
	if err != nil {
		return 0, logMongoError("t.mongo.store", err)
	}

	count, err := C(dtype).CountDocuments(gtx, selector)
	return count, logMongoError("t.mongo.store", err)
}

//GetFilterValues - provides values associated the fields defined in filter spec
func (mds *dataStorage) GetFilterValues(
	gtx context.Context,
	dtype string,
	specs teak.FilterSpecList) (values teak.M, err error) {
	values = teak.M{}
	for _, spec := range specs {
		switch spec.Type {
		case teak.Prop:
			fallthrough
		case teak.Array:
			// props := make([]string, 0, 100)
			var props []interface{}
			props, err = C(dtype).Distinct(gtx, spec.Field, bson.D{})
			if err != nil {
				break
			}
			values[spec.Field] = props
		case teak.Date:
			var drange teak.DateRange
			var cur *mongo.Cursor
			cur, err = C(dtype).Aggregate(gtx, []bson.M{
				{"$group": bson.M{
					"_id": nil,
					"from": bson.M{
						"$max": spec.Field,
					},
					"to": bson.M{
						"$min": spec.Field,
					},
				},
				},
			})
			if err == nil && cur.Next(gtx) {
				err = cur.Decode(&drange)
			}
			if err != nil {
				break
			}
			values[spec.Field] = drange
		case teak.Boolean:
		case teak.Search:
		case teak.Static:
		}
	}
	return values, logMongoError("t.mongo.store", err)
}

//GetFilterValuesX - get values for filter based on given filter
func (mds *dataStorage) GetFilterValuesX(
	gtx context.Context,
	dtype string,
	field string,
	specs teak.FilterSpecList,
	filter *teak.Filter) (values teak.M, err error) {
	facet := teak.M{}
	for _, spec := range specs {
		if spec.Field != field {
			switch spec.Type {
			case teak.Prop:
				facet[spec.Field] = []bson.M{
					{
						"$sortByCount": "$" + spec.Field,
					},
				}
			case teak.Array:
				fd := "$" + spec.Field
				facet[spec.Field] = []bson.M{
					{
						"$unwind": fd,
					},
					{
						"$sortByCount": fd,
					},
				}
			case teak.Date:
			case teak.Boolean:
			case teak.Search:
			case teak.Static:
			}
		}
	}
	var selector bson.M
	if filter != nil {
		selector = GenerateSelector(filter)
	}
	values = teak.M{}
	cur, err := C(dtype).Aggregate(gtx, []bson.M{
		{
			"$match": selector,
		},
		{
			"$facet": facet,
		},
	})
	if err != nil {
		return nil, logMongoError("t.mongo.store", err)
	}
	err = cur.All(gtx, &values)
	return values, logMongoError("t.mongo.store", err)
}

//GenerateSelector - creates mongodb query for a generic filter
func GenerateSelector(
	filter *teak.Filter) (selector bson.M) {
	queries := make([]bson.M, 0, 100)
	// for key, values := range filter.Props {
	// 	if len(values) == 1 {
	// 		queries = append(queries, bson.M{key: values[0]})
	// 	} else if len(values) > 1 {
	// 		orProps := make([]bson.M, 0, len(values))
	// 		for _, val := range values {
	// 			orProps = append(orProps, bson.M{key: val})
	// 		}
	// 		queries = append(queries, bson.M{"$or": orProps})
	// 	}
	// }
	for field, matcher := range filter.Props {
		if len(matcher.Fields) != 0 {
			mode := "$in"
			if matcher.Strategy == teak.MatchAll {
				mode = "$all"
			} else if matcher.Strategy == teak.MatchNone {
				mode = "$nin"
			}
			queries = append(queries, bson.M{
				field: bson.M{
					mode: matcher.Fields,
				},
			})
		}
	}
	for field, val := range filter.Bools {
		if val != nil {
			queries = append(queries, bson.M{field: val})
		}
	}
	for field, dateRange := range filter.Dates {
		if dateRange.IsValid() {
			queries = append(queries,
				bson.M{
					field: bson.M{
						"$gte": now.New(dateRange.From).BeginningOfDay(),
						"$lte": now.New(dateRange.To).EndOfDay(),
					},
				},
			)
		}
	}
	for field, matcher := range filter.Lists {
		if len(matcher.Fields) != 0 {
			mode := "$in"
			if matcher.Strategy == teak.MatchAll {
				mode = "$all"
			} else if matcher.Strategy == teak.MatchNone {
				mode = "$nin"
			}
			queries = append(queries, bson.M{
				field: bson.M{
					mode: matcher.Fields,
				},
			})
		}
	}
	if len(queries) != 0 {
		selector = bson.M{
			"$and": queries,
		}
	}
	// teak.DumpJSON(filter)
	// teak.DumpJSON(selector)
	return selector
}

//Init - initialize the data storage for the first time, sets it upda and also
//creates the first admin user. Data store can be initialized only once
func (mds *dataStorage) Init(
	gtx context.Context,
	admin *teak.User,
	adminPass string,
	param teak.M) error {
	val, err := mds.IsInitialized(gtx)
	if err != nil {
		err = teak.LogErrorX("t.mongo.store",
			"Failed to check initialization status of PG store", err)
		return err
	}
	if val {
		teak.Info("t.mongo.store", "Store already initialized.")
		teak.Info("t.mongo.store",
			"If you want to update the structure of the store, use Setup")
		return err
	}
	err = mds.Setup(gtx, teak.M{})
	if err != nil {
		err = teak.LogErrorX("t.mongo.store", "Failed to setup app", err)
		return err
	}
	uStore := NewUserStorage()
	idHash, err := uStore.CreateUser(gtx, admin)
	if err != nil {
		err = teak.LogErrorX("t.mongo.store",
			"Failed to create initial super admin", err)
		return err
	}
	err = uStore.SetPassword(gtx, idHash, adminPass)
	if err != nil {
		err = teak.LogErrorX("t.mongo.store",
			"Failed to set initial super user password", err)
		return err
	}
	return err
}

//Setup - setup has to be run when data storage structure changes, such as
//adding index, altering tables etc
func (mds *dataStorage) Setup(gtx context.Context, params teak.M) (err error) {
	//Setup indices for user collection
	//Setup indices for event collection
	return err
}

//Reset - reset clears the data without affecting the structure/schema
func (mds *dataStorage) Reset(gtx context.Context) (err error) {
	return err
}

//Destroy - deletes data and also structure
func (mds *dataStorage) Destroy(gtx context.Context) (err error) {
	return err
}

// //Wrap - wraps a command with flags required to connect to this data source
// func (mds *dataStorage) Wrap(cmd *cli.Command) *cli.Command {
// 	cmd.Flags = append(cmd.Flags, mongoFlags...)
// 	if cmd.Before == nil {
// 		cmd.Before = requireMongo
// 	} else {
// 		otherBefore := cmd.Before
// 		cmd.Before = func(ctx *cli.Context) (err error) {
// 			err = requireMongo(ctx)
// 			if err == nil {
// 				err = otherBefore(ctx)
// 			}
// 			return err
// 		}
// 	}
// 	return cmd
// }

func (mds *dataStorage) WithFlags(flags ...cli.Flag) []cli.Flag {
	return append(flags, mongoFlags...)
}

//GetManageCommands - commands that can be used to manage this data storage
func (mds *dataStorage) GetManageCommands() (commands []cli.Command) {
	return commands
}

//IsInitialized - tells if data source is initialized
func (mds *dataStorage) IsInitialized(gtx context.Context) (bool, error) {
	return true, nil
}

// GetSort - get sort statement for the field
func GetSort(sortField string) bson.D {
	sortDir := 1
	if strings.HasPrefix(sortField, "-") {
		sortDir = -1
		sortField = sortField[1:]
	}
	return bson.D{{sortField, sortDir}}

}
