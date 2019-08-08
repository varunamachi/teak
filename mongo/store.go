package mongo

import (
	"runtime"

	"github.com/jinzhu/now"
	"github.com/varunamachi/teak"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/urfave/cli.v1"
)

//dataStorage - MongoDB implementation for dataStorage interface
type dataStorage struct{}

//NewStorage - creates a new mongodb based data storage implementation
func NewStorage() teak.DataStorage {
	return &dataStorage{}
}

//logMongoError - if error is not mog.ErrNotFound return null otherwise log the
//error and return the given error
func logMongoError(module string, err error) (out error) {
	if err != nil && err != mgo.ErrNotFound {
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
	dtype string, value interface{}) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	err = conn.C(dtype).Insert(value)
	return logMongoError("DB:Mongo", err)
}

//Update - updates the records in 'dtype' collection which are matched by
//the matcher query
func (mds *dataStorage) Update(
	dtype string,
	keyField string,
	key interface{},
	value interface{}) (err error) {

	conn := DefaultConn()
	defer conn.Close()
	err = conn.C(dtype).Update(bson.M{
		keyField: key,
	}, value)
	return logMongoError("DB:Mongo", err)
}

//Delete - deletes record matched by the matcher from collection 'dtype'
func (mds *dataStorage) Delete(
	dtype string,
	keyField string,
	key interface{}) (err error) {

	conn := DefaultConn()
	defer conn.Close()
	err = conn.C(dtype).Remove(bson.M{
		keyField: key,
	})
	return logMongoError("DB:Mongo", err)
}

//RetrieveOne - gets a record matched by given matcher from collection 'dtype'
func (mds *dataStorage) RetrieveOne(
	dtype string,
	keyField string,
	key interface{},
	out interface{}) (err error) {

	conn := DefaultConn()
	defer conn.Close()
	err = conn.C(dtype).Find(bson.M{
		keyField: key,
	}).One(out)
	return logMongoError("DB:Mongo", err)
}

//Count - counts the number of items for data type
func (mds *dataStorage) Count(
	dtype string, filter *teak.Filter) (count int, err error) {
	//@TODO handle filters
	conn := DefaultConn()
	defer conn.Close()
	selector := generateSelector(filter)
	count, err = conn.C(dtype).
		Find(selector).
		Count()
	return count, logMongoError("DB:Mongo", err)
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
	conn := DefaultConn()
	defer conn.Close()
	err = conn.C(dtype).
		Find(selector).
		Sort(sortFiled).
		Skip(offset).
		Limit(limit).
		All(out)
	return logMongoError("DB:Mongo", err)
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
	conn := DefaultConn()
	defer conn.Close()
	selector := generateSelector(filter)
	q := conn.C(dtype).Find(selector)
	count, err = q.Count()
	if err == nil {
		err = q.Sort(sortFiled).
			Skip(offset).
			Limit(limit).
			All(out)
	}
	return count, logMongoError("DB:Mongo", err)
}

//GetFilterValues - provides values associated the fields defined in filter spec
func (mds *dataStorage) GetFilterValues(
	dtype string,
	specs teak.FilterSpecList) (values teak.M, err error) {
	conn := DefaultConn()
	defer conn.Close()
	values = teak.M{}
	for _, spec := range specs {
		switch spec.Type {
		case teak.Prop:
			fallthrough
		case teak.Array:
			props := make([]string, 0, 100)
			err = conn.C(dtype).Find(nil).Distinct(spec.Field, &props)
			values[spec.Field] = props
		case teak.Date:
			var drange teak.DateRange
			err = conn.C(dtype).Pipe([]bson.M{
				bson.M{
					"$group": bson.M{
						"_id": nil,
						"from": bson.M{
							"$max": spec.Field,
						},
						"to": bson.M{
							"$min": spec.Field,
						},
					},
				},
			}).One(&drange)
			values[spec.Field] = drange
		case teak.Boolean:
		case teak.Search:
		case teak.Static:
		}
	}
	return values, logMongoError("DB:Mongo", err)
}

//GetFilterValuesX - get values for filter based on given filter
func (mds *dataStorage) GetFilterValuesX(
	dtype string,
	field string,
	specs teak.FilterSpecList,
	filter *teak.Filter) (values teak.M, err error) {
	conn := DefaultConn()
	defer conn.Close()
	facet := teak.M{}
	for _, spec := range specs {
		if spec.Field != field {
			switch spec.Type {
			case teak.Prop:
				facet[spec.Field] = []bson.M{
					bson.M{
						"$sortByCount": "$" + spec.Field,
					},
				}
			case teak.Array:
				fd := "$" + spec.Field
				facet[spec.Field] = []bson.M{
					bson.M{
						"$unwind": fd,
					},
					bson.M{
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
		selector = generateSelector(filter)
	}
	values = teak.M{}
	err = conn.C(dtype).Pipe([]bson.M{
		bson.M{
			"$match": selector,
		},
		bson.M{
			"$facet": facet,
		},
	}).One(&values)
	return values, logMongoError("DB:Mongo", err)
}

//GenerateSelector - creates mongodb query for a generic filter
func generateSelector(
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

//Init - initialize the data storage - this needs to be run on each application
//start up
func (mds *dataStorage) Init() (err error) {
	return err
}

//Setup - setup has to be run when data storage structure changes, such as
//adding index, altering tables etc
func (mds *dataStorage) Setup(params teak.M) (err error) {
	//Setup indices for user collection
	//Setup indices for event collection
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
	cmd.Flags = append(cmd.Flags, mongoFlags...)
	if cmd.Before == nil {
		cmd.Before = requireMongo
	} else {
		otherBefore := cmd.Before
		cmd.Before = func(ctx *cli.Context) (err error) {
			err = requireMongo(ctx)
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
