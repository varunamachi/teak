package teak

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
)

func getDataEndpoints() []*Endpoint {
	return []*Endpoint{
		&Endpoint{
			Method:   echo.POST,
			URL:      "gen/:dataType",
			Access:   Normal,
			Category: "generic",
			Func:     createObject,
			Comment:  "Create a resource of given type",
		},
		&Endpoint{
			Method:   echo.PUT,
			URL:      "gen/:dataType",
			Access:   Normal,
			Category: "generic",
			Func:     updateObject,
			Comment:  "Update a resource of given type",
		},
		&Endpoint{
			Method:   echo.DELETE,
			URL:      "gen/:dataType/:id",
			Access:   Normal,
			Category: "generic",
			Func:     deleteObject,
			Comment:  "Delete a resource of given type",
		},
		&Endpoint{
			Method:   echo.GET,
			URL:      "gen/:dataType/:id",
			Access:   Monitor,
			Category: "generic",
			Func:     retrieveOne,
			Comment:  "retrieve a resource of given type",
		},
		&Endpoint{
			Method:   echo.GET,
			URL:      "gen/:dataType/list",
			Access:   Monitor,
			Category: "generic",
			Func:     retrieve,
			Comment:  "Retrieve a resource sub-list of given type",
		},
		&Endpoint{
			Method:   echo.GET,
			URL:      "gen/:dataType/count",
			Access:   Monitor,
			Category: "generic",
			Func:     countObjects,
			Comment:  "Get count of items of data type",
		},
		&Endpoint{
			Method:   echo.GET,
			URL:      "gen/:dataType",
			Access:   Monitor,
			Category: "generic",
			Func:     retrieveWithCount,
			Comment:  "Retrieve a resource sub-list of a type with total count",
		},
		&Endpoint{
			Method:   echo.GET,
			URL:      "gen/:dataType/fspec",
			Access:   Monitor,
			Category: "generic",
			Func:     getFilterValues,
			Comment:  "Get possible values for filter",
		},
		&Endpoint{
			Method:   echo.GET,
			URL:      "gen/:dataType/fvals/:field",
			Access:   Monitor,
			Category: "generic",
			Func:     getFilterValuesX,
			Comment:  "Get possible values for filter",
		},
		&Endpoint{
			Method:   echo.GET,
			URL:      "gen/:dataType/fvals/",
			Access:   Monitor,
			Category: "generic",
			Func:     getFilterValuesX,
			Comment:  "Get possible values for filter without field",
		},
	}
}

//StoredItem - represents a value that is stored in database and is
//compatible with generic queries and handlers. Any struct with a need to
//support generic CRUD operations must implement and register a factory
//method to return it
type StoredItem interface {
	ID() interface{}
	SetCreationInfo(at time.Time, by string)
	SetModInfo(at time.Time, by string)
}

//FactoryFunc - Function for creating an instance of data type
type FactoryFunc func() StoredItem

var factories = make(map[string]FactoryFunc)

//defaultSM - default status and message
func defaultSM(opern, name string) (int, string) {
	return http.StatusOK, fmt.Sprintf("%s %s - successful", opern, name)
}

type DataStorage interface {
	Create(dataType string, data interface{}) error
	Update(dataType string, key interface{}, data interface{}) error
	Delete(dataType string, key interface{}) error
	Count(dtype string, filter *Filter) (count int, err error)
	RetrieveOne(dataType string, key interface{}, data interface{}) error
	Retrieve(dtype string,
		sortFiled string,
		offset int,
		limit int,
		filter *Filter,
		out interface{}) error
	RetrieveWithCount(dtype string,
		sortFiled string,
		offset int,
		limit int,
		filter *Filter,
		out interface{}) (count int, err error)
	GetFilterValues(dtype string,
		specs FilterSpecList) (values M, err error)
	GetFilterValuesX(
		dtype string,
		field string,
		specs FilterSpecList,
		filter *Filter) (values M, err error)
}

var dataStorage DataStorage

func createObject(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Create", dtype)
	var data StoredItem
	if len(dtype) != 0 {
		data, err = bind(ctx, dtype)
		if err == nil {
			data.SetCreationInfo(time.Now(), GetString(ctx, "userID"))
			err = dataStorage.Create(dtype, data)
			if err != nil {
				msg = fmt.Sprintf("Failed to create %s in database", dtype)
				status = http.StatusInternalServerError
			}
		} else {
			msg = fmt.Sprintf(
				"Failed to retrieve %s information from the request", dtype)
			status = http.StatusBadRequest
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	AuditedSendX(ctx, &data, &Result{
		Status: status,
		Op:     dtype + "_create",
		Msg:    msg,
		OK:     err == nil,
		Data:   nil,
		Err:    ErrString(err),
	})
	return LogError("S:Entity", err)
}

func updateObject(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Update", dtype)
	var data StoredItem
	if len(dtype) != 0 {
		data, err = bind(ctx, dtype)
		if err == nil {
			data.SetModInfo(time.Now(), GetString(ctx, "userID"))
			err = dataStorage.Update(dtype, data.ID(), data)
			if err != nil {
				msg = fmt.Sprintf("Failed to update %s in database", dtype)
				status = http.StatusInternalServerError
			}
		} else {
			msg = fmt.Sprintf(
				"Failed to retrieve %s information from the request", dtype)
			status = http.StatusBadRequest
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	AuditedSendX(ctx, &data, &Result{
		Status: status,
		Op:     dtype + "_update",
		Msg:    msg,
		OK:     err == nil,
		Data:   nil,
		Err:    ErrString(err),
	})
	return LogError("S:Entity", err)
}

func deleteObject(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Delete", dtype)
	id := ctx.Param("id")
	if len(dtype) != 0 {
		err = dataStorage.Delete(dtype, id)
		if err != nil {
			msg = fmt.Sprintf("Failed to delete %s from database", dtype)
			status = http.StatusInternalServerError
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = AuditedSend(ctx, &Result{
		Status: status,
		Op:     dtype + "_delete",
		Msg:    msg,
		OK:     err == nil,
		Data:   id,
		Err:    ErrString(err),
	})
	return LogError("S:Entity", err)
}

func retrieveOne(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Get", dtype)
	data := M{}
	id := ctx.Param("id")
	if len(dtype) != 0 {
		err = dataStorage.RetrieveOne(dtype, id, &data)
		if err != nil {
			msg = fmt.Sprintf(
				"Failed to retrieve %s from database, entity with ID %s",
				dtype,
				id)
			status = http.StatusInternalServerError
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     dtype + "_fetch",
		Msg:    msg,
		OK:     err == nil,
		Data:   data,
		Err:    ErrString(err),
	})
	return LogError("S:Entity", err)
}

func retrieve(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Get All", dtype)
	var data []*M
	if len(dtype) != 0 {
		offset, limit, has := GetOffsetLimit(ctx)
		var filter Filter
		err = LoadJSONFromArgs(ctx, "filter", &filter)
		sortField := GetQueryParam(ctx, "sortField", "-createdAt")
		if has && err == nil {
			data = make([]*M, 0, limit)
			err = dataStorage.Retrieve(
				dtype,
				sortField,
				offset,
				limit,
				&filter,
				&data)
			if err != nil {
				msg = fmt.Sprintf("Failed to retrieve %s from database", dtype)
				status = http.StatusInternalServerError
			}
		} else {
			msg = "Invalid offset, limit or filter given"
			status = http.StatusBadRequest
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     dtype + "_fetch",
		Msg:    msg,
		OK:     err == nil,
		Data:   data,
		Err:    ErrString(err),
	})
	return LogError("S:Entity", err)
}

func retrieveWithCount(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Get all with count", dtype)
	var data []*M
	cnt := 0
	if len(dtype) != 0 {
		offset, limit, has := GetOffsetLimit(ctx)
		sortField := GetQueryParam(ctx, "sortField", "-createdAt")
		var filter Filter
		err = LoadJSONFromArgs(ctx, "filter", &filter)
		if has && err == nil {
			data = make([]*M, 0, limit)
			cnt, err = dataStorage.RetrieveWithCount(
				dtype,
				sortField,
				offset,
				limit,
				&filter,
				&data)
			if err != nil {
				msg = fmt.Sprintf("Failed to retrieve %s from database", dtype)
				status = http.StatusInternalServerError
			}
		} else {
			msg = "Invalid offset, limit or filter given"
			status = http.StatusBadRequest
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     dtype + "_fetch_n_count",
		Msg:    msg,
		OK:     err == nil,
		Data: CountList{
			Data:       data,
			TotalCount: cnt,
		},
		Err: ErrString(err),
	})
	return LogError("S:Entity", err)
}

func countObjects(ctx echo.Context) (err error) {
	//@TODO - handle filters
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Get All", dtype)
	count := 0
	if len(dtype) != 0 {
		if err == nil {
			var filter Filter
			err = LoadJSONFromArgs(ctx, "filter", &filter)
			count, err = dataStorage.Count(dtype, &filter)
			if err != nil {
				msg = fmt.Sprintf("Failed to retrieve %s from database", dtype)
				status = http.StatusInternalServerError
			}
		} else {
			msg = fmt.Sprintf("Failed to decode filter for '%s'", dtype)
			status = http.StatusInternalServerError
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     dtype + "_count",
		Msg:    msg,
		OK:     err == nil,
		Data:   count,
		Err:    ErrString(err),
	})
	return LogError("S:Entity", err)
}

func getFilterValues(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Filter Values of", dtype)
	var fspec []*FilterSpec
	var values M
	if len(dtype) != 0 {
		err = LoadJSONFromArgs(ctx, "fspec", &fspec)
		if err == nil {
			values, err = dataStorage.GetFilterValues(dtype, fspec)
		} else {
			msg = "Failed to load filter description from URL"
			status = http.StatusBadRequest
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     dtype + "_filter_fetch",
		Msg:    msg,
		OK:     err == nil,
		Data:   values,
		Err:    ErrString(err),
	})
	return LogError("S:Entity", err)
}

func getFilterValuesX(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	field := ctx.Param("field")
	status, msg := defaultSM("Filter Values of", dtype)
	var fspec []*FilterSpec
	var filter Filter
	var values M
	if len(dtype) != 0 {
		err1 := LoadJSONFromArgs(ctx, "fspec", &fspec)
		err2 := LoadJSONFromArgs(ctx, "filter", &filter)
		if !HasError("V:Generic", err1, err2) {
			values, err = dataStorage.GetFilterValuesX(
				dtype, field, fspec, &filter)
		} else {
			msg = "Failed to load filter description from URL"
			err = errors.New(msg)
			status = http.StatusBadRequest
		}
	} else {
		msg = "Invalid empty data type given"
		status = http.StatusBadRequest
		err = errors.New(msg)
	}
	err = SendAndAuditOnErr(ctx, &Result{
		Status: status,
		Op:     dtype + "_filter_fetch",
		Msg:    msg,
		OK:     err == nil,
		Data:   values,
		Err:    ErrString(err),
	})
	return LogError("S:Entity", err)
}

func bind(ctx echo.Context, dataType string) (
	data StoredItem, err error) {
	if creator, found := factories[dataType]; found {
		data = creator()
	}
	if data != nil {
		err = ctx.Bind(data)
	} else {
		err = fmt.Errorf("Could not find factory function for data type %s",
			dataType)
	}
	return data, err
}
