package pg

import (
	"runtime"

	"github.com/varunamachi/teak"
)

//DataStorage - MongoDB implementation for DataStorage interface
type DataStorage struct{}

//logError - if error is not mog.ErrNotFound return null otherwise log the
//error and return the given error
func logError(module string, err error) (out error) {
	if err != nil {
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

//Create - creates an record in 'dtype' collection
func (mds *DataStorage) Create(
	dtype string, value interface{}) (err error) {
	return logError("DB:Mongo", err)
}

//Update - updates the records in 'dtype' collection which are matched by
//the matcher query
func (mds *DataStorage) Update(
	dtype string, key interface{}, value interface{}) (err error) {
	return logError("DB:Mongo", err)
}

//Delete - deletes record matched by the matcher from collection 'dtype'
func (mds *DataStorage) Delete(dtype string, key interface{}) (err error) {
	return logError("DB:Mongo", err)
}

//RetrieveOne - gets a record matched by given matcher from collection 'dtype'
func (mds *DataStorage) RetrieveOne(
	dtype string, key interface{}, out interface{}) (err error) {
	return logError("DB:Mongo", err)
}

//Count - counts the number of items for data type
func (mds *DataStorage) Count(
	dtype string, filter *teak.Filter) (count int, err error) {
	return count, logError("DB:Mongo", err)
}

//Retrieve - gets all the items from collection 'dtype' selected by filter &
//paged
func (mds *DataStorage) Retrieve(
	dtype string,
	sortFiled string,
	offset int,
	limit int,
	filter *teak.Filter,
	out interface{}) (err error) {
	return logError("DB:Mongo", err)
}

//RetrieveWithCount - gets all the items from collection 'dtype' selected by
//filter & paged also gives the total count of items selected by filter
func (mds *DataStorage) RetrieveWithCount(
	dtype string,
	sortFiled string,
	offset int,
	limit int,
	filter *teak.Filter,
	out interface{}) (count int, err error) {
	return count, logError("DB:Mongo", err)
}

//GetFilterValues - provides values associated the fields defined in filter spec
func (mds *DataStorage) GetFilterValues(
	dtype string,
	specs teak.FilterSpecList) (values teak.M, err error) {
	values = teak.M{}
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
	return values, logError("DB:Mongo", err)
}

//GetFilterValuesX - get values for filter based on given filter
func (mds *DataStorage) GetFilterValuesX(
	dtype string,
	field string,
	specs teak.FilterSpecList,
	filter *teak.Filter) (values teak.M, err error) {
	for _, spec := range specs {
		if spec.Field != field {
			switch spec.Type {
			case teak.Prop:
			case teak.Array:
			case teak.Date:
			case teak.Boolean:
			case teak.Search:
			case teak.Static:
			}
		}
	}
	return values, logError("DB:Mongo", err)
}
