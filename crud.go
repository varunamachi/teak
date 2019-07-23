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

//StoredItemHandler - needs to be implemented for any data type that is expected
//to work with generic crud system
type StoredItemHandler interface {
	DataType() string
	UniqueKeyField() string
	GetKey(item interface{}) interface{}
	SetModInfo(item interface{}, at time.Time, by string)
	CreateInstance(by string) interface{}
}

var siHandlers = make(map[string]StoredItemHandler)

//FactoryFunc - Function for creating an instance of data type
// type FactoryFunc func() StoredItem
// var factories = make(map[string]FactoryFunc)

//defaultSM - default status and message
func defaultSM(opern, name string) (int, string) {
	return http.StatusOK, fmt.Sprintf("%s %s - successful", opern, name)
}

func createObject(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Create", dtype)
	var data interface{}
	//Log any potential error and send the response
	defer func() {
		AuditedSendX(ctx, &data, &Result{
			Status: status,
			Op:     dtype + "_create",
			Msg:    msg,
			OK:     err == nil,
			Data:   nil,
			Err:    ErrString(err),
		})
		LogError("t.crud.api", err)
	}()

	handler := siHandlers[dtype]
	if handler == nil {
		err = fmt.Errorf("Failed to find handler for data type '%s'", dtype)
		status = http.StatusBadRequest
		return err
	}

	data = handler.CreateInstance(GetString(ctx, "userID"))
	err = dataStorage.Create(dtype, data)
	if err != nil {
		msg = fmt.Sprintf("Failed to create item of type '%s' in data store",
			dtype)
		status = http.StatusInternalServerError
	}

	return err
}

func updateObject(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Update", dtype)
	var data interface{}
	//Log error if any and send response at the end
	defer func() {
		LogError("t.crud.api", err)
		AuditedSendX(ctx, &data, &Result{
			Status: status,
			Op:     dtype + "_update",
			Msg:    msg,
			OK:     err == nil,
			Data:   nil,
			Err:    ErrString(err),
		})
	}()
	//Get the updated object from request
	err = ctx.Bind(data)
	if err != nil {
		err = fmt.Errorf("Failed to retrive updated object for type '%s'",
			dtype)
		status = http.StatusBadRequest
		return err
	}

	//Get the data type handler for updating the modification info:
	handler := siHandlers[dtype]
	if handler == nil {
		err = fmt.Errorf("Failed to find handler for data type '%s'", dtype)
		status = http.StatusBadRequest
		return err
	}

	//Update the modification  info:
	handler.SetModInfo(data, time.Now(), GetString(ctx, "userID"))
	//Get the identifier for the item
	key := handler.GetKey(data)
	//And update...
	err = dataStorage.Update(dtype, handler.UniqueKeyField(), key, data)

	if err != nil {
		msg = fmt.Sprintf("Failed to create item of type '%s' in data store",
			dtype)
		status = http.StatusInternalServerError
	}
	return err
}

func deleteObject(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Delete", dtype)
	id := ctx.Param("id")

	defer func() {
		err = AuditedSend(ctx, &Result{
			Status: status,
			Op:     dtype + "_delete",
			Msg:    msg,
			OK:     err == nil,
			Data:   id,
			Err:    ErrString(err),
		})
		LogError("t.crud.api", err)
	}()

	//Get the data type handler for the given data type:
	handler := siHandlers[dtype]
	if handler == nil {
		err = fmt.Errorf("Failed to find handler for data type '%s'", dtype)
		status = http.StatusBadRequest
		return err
	}

	err = dataStorage.Delete(dtype, handler.UniqueKeyField(), id)
	if err != nil {
		msg = fmt.Sprintf("Failed to delete %s from database", dtype)
		status = http.StatusInternalServerError
	}
	return err
}

func retrieveOne(ctx echo.Context) (err error) {
	dtype := ctx.Param("dataType")
	status, msg := defaultSM("Get", dtype)
	id := ctx.Param("id")
	var data interface{}

	defer func() {
		err = SendAndAuditOnErr(ctx, &Result{
			Status: status,
			Op:     dtype + "_fetch",
			Msg:    msg,
			OK:     err == nil,
			Data:   data,
			Err:    ErrString(err),
		})
		LogError("t.crud.api", err)
	}()

	//Get the data type handler for the given data type:
	handler := siHandlers[dtype]
	if handler == nil {
		err = fmt.Errorf("Failed to find handler for data type '%s'", dtype)
		status = http.StatusBadRequest
		return err
	}
	data = handler.CreateInstance("")
	err = dataStorage.RetrieveOne(dtype, handler.UniqueKeyField(), id, &data)

	if err != nil {
		msg = fmt.Sprintf(
			"Failed to retrieve %s from database, entity with ID %s",
			dtype,
			id)
		status = http.StatusInternalServerError
	}
	return err
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
