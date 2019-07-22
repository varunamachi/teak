package pg

import (
	"github.com/varunamachi/teak"
)

//DataStorage - Postgres implementation for DataStorage interface
type DataStorage struct{}

//Create - creates an record in 'dtype' collection
func (mds *DataStorage) Create(
	dtype string, value interface{}) (err error) {
	return teak.LogError("t.crud.pg", err)
}

//Update - updates the records in 'dtype' collection which are matched by
//the matcher query
func (mds *DataStorage) Update(
	dtype string,
	keyField string,
	key interface{},
	value interface{}) (err error) {
	return teak.LogError("t.crud.pg", err)
}

//Delete - deletes record matched by the matcher from collection 'dtype'
func (mds *DataStorage) Delete(
	dtype string,
	keyField string,
	key interface{}) (err error) {
	return teak.LogError("t.crud.pg", err)
}

//RetrieveOne - gets a record matched by given matcher from collection 'dtype'
func (mds *DataStorage) RetrieveOne(
	dtype string,
	keyField string,
	key interface{},
	out interface{}) (err error) {
	return teak.LogError("t.crud.pg", err)
}

//Count - counts the number of items for data type
func (mds *DataStorage) Count(
	dtype string, filter *teak.Filter) (count int, err error) {
	return count, teak.LogError("t.crud.pg", err)
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
	return teak.LogError("t.crud.pg", err)
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
	return count, teak.LogError("t.crud.pg", err)
}

//GetFilterValues - provides values associated the fields defined in filter spec
func (mds *DataStorage) GetFilterValues(
	dtype string,
	specs teak.FilterSpecList) (values teak.M, err error) {
	return values, teak.LogError("t.crud.pg", err)
}

//GetFilterValuesX - get values for filter based on given filter
func (mds *DataStorage) GetFilterValuesX(
	dtype string,
	field string,
	specs teak.FilterSpecList,
	filter *teak.Filter) (values teak.M, err error) {
	return values, teak.LogError("t.crud.pg", err)
}
