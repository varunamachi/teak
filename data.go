package teak

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/urfave/cli.v1"
)

var dataStorage DataStorage

//GetStore - get the data store
func GetStore() DataStorage {
	return dataStorage
}

//Version - represents version of the application
type Version struct {
	Major int `json:"major" db:"major"`
	Minor int `json:"minor" db:"minor"`
	Patch int `json:"patch" db:"patch"`
}

// Ver - quick and dirty way to create a version object
func Ver(major, minor, patch int) Version {
	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

//String - version to string
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

//DateRange - represents date ranges
type DateRange struct {
	// Name string    `json:"name" bson:"name"`
	From time.Time `json:"from" db:"from"`
	To   time.Time `json:"to" db:"to"`
}

//IsValid - returns true if both From and To dates are non-zero
func (r *DateRange) IsValid() bool {
	return !(r.From.IsZero() || r.To.IsZero())
}

//ParamType - type of the parameter
type ParamType int

const (
	//Bool - bool parameter
	Bool ParamType = iota

	//NumberRange - number range parameter
	NumberRange

	//Choice - parameter with choices
	Choice

	//Text - arbitrary string
	Text
)

//Pair - association of key and value
type Pair struct {
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`
}

//Range - integer range
type Range struct {
	Min int `json:"min" db:"min"`
	Max int `json:"max" db:"max"`
}

//Param - represents generic parameter
type Param struct {
	Name    string      `json:"name" db:"name"`
	Type    ParamType   `json:"type" db:"type"`
	Desc    string      `json:"desc" db:"desc"`
	Range   Range       `json:"range" db:"range"`
	Choices []Pair      `json:"choices" db:"choices"`
	Default interface{} `json:"def" db:"def"`
	// Value   interface{} `json:"value" bson:"value"`
}

//CountList - paginated list returned from mongoDB along with total number of
//items in the list counted without pagination
type CountList struct {
	TotalCount int64       `json:"total" db:"total"`
	Data       interface{} `json:"data" db:"data"`
}

//FilterType - Type of filter item
type FilterType string

//Prop - filter for a value
const Prop FilterType = "prop"

//Array - filter for an array
const Array FilterType = "array'"

//Date - filter for data range
const Date FilterType = "dateRange"

//Boolean - filter for boolean field
const Boolean FilterType = "boolean"

//Search - filter for search text field
const Search FilterType = "search"

//Constant - constant filter value
const Constant FilterType = "constant"

//Static - constant filter value
const Static FilterType = "static"

//MatchStrategy - strategy to match multiple fields passed as part of the
//filters
type MatchStrategy string

//MatchAll - match all provided values while executing filter
const MatchAll MatchStrategy = "all"

//MatchOne - match atleast one of the  provided values while executing filter
const MatchOne MatchStrategy = "one"

//MatchNone - match values that are not part of the provided list while
//executing filter
const MatchNone MatchStrategy = "none"

//FilterSpec - filter specification
type FilterSpec struct {
	Field string     `json:"field" db:"field"`
	Name  string     `json:"name" db:"name"`
	Type  FilterType `json:"type" db:"type"`
}

//Matcher - matches the given fields. If multiple fileds are given the; the
//joining condition is decided by the MatchStrategy given
type Matcher struct {
	Strategy MatchStrategy `json:"strategy" db:"strategy"`
	Fields   []interface{} `json:"fields" db:"fields"`
}

//SearchField - contains search string and info for performing the search
// type SearchField struct {
// 	MatchAll  bool   `json:"matchAll" bson:"matchAll"`
// 	Regex     bool   `json:"regex" bson:"regex"`
// 	SearchStr string `json:"searchStr" bson:"searchStr"`
// }

//PropMatcher - matches props
type PropMatcher []interface{}

//Filter - generic filter used to filter data in any mongodb collection
type Filter struct {
	Props    map[string]Matcher     `json:"props" db:"props"`
	Bools    map[string]interface{} `json:"bools" db:"bools"`
	Dates    map[string]DateRange   `json:"dates" db:"dates"`
	Lists    map[string]Matcher     `json:"lists" db:"lists"`
	Searches map[string]Matcher     `json:"searches" db:"searches"`
}

//FilterSpecList - alias for array of filter specs
type FilterSpecList []*FilterSpec

//FilterVal - values for filter along with the count
type FilterVal struct {
	Name  string `json:"name" db:"name"`
	Count int    `json:"count" db:"count"`
}

//DataStorage - defines a data storage
type DataStorage interface {
	Name() string
	Count(
		gtx context.Context,
		dtype string,
		filter *Filter) (count int64, err error)
	Create(
		gtx context.Context,
		dataType string,
		data interface{}) error
	Update(
		gtx context.Context,
		dataType string,
		keyField string,
		key interface{},
		data interface{}) error
	Delete(
		gtx context.Context,
		dataType string,
		keyField string,
		key interface{}) error
	RetrieveOne(
		gtx context.Context,
		dataType string,
		keyField string,
		key interface{},
		data interface{}) error
	Retrieve(
		gtx context.Context,
		dtype string,
		sortFiled string,
		offset int64,
		limit int64,
		filter *Filter,
		out interface{}) error
	RetrieveWithCount(
		gtx context.Context,
		dtype string,
		sortFiled string,
		offset int64,
		limit int64,
		filter *Filter,
		out interface{}) (count int64, err error)
	GetFilterValues(
		gtx context.Context,
		dtype string,
		specs FilterSpecList) (values M, err error)
	GetFilterValuesX(
		gtx context.Context,
		dtype string,
		field string,
		specs FilterSpecList,
		filter *Filter) (values M, err error)

	IsInitialized(gtx context.Context) (yes bool, err error)
	Init(gtx context.Context, admin *User, adminPass string, param M) error
	Setup(gtx context.Context, params M) error
	Reset(gtx context.Context) error
	Destroy(gtx context.Context) error
	// WithFlags(flags ...cli.Flag) []cli.Flag
	// ConnectDefault(ag *ArgGetter) error
	Wrap(cmd *cli.Command) *cli.Command
	GetManageCommands() []cli.Command
}

//@TODO Data store ini shall do these
// vevt.SetEventAuditor(auditor)

//IsBasicType - tells if the kind of data type is basic or composite
func IsBasicType(rt reflect.Kind) bool {
	switch rt {
	case reflect.Bool:
		return true
	case reflect.Int:
		return true
	case reflect.Int8:
		return true
	case reflect.Int16:
		return true
	case reflect.Int32:
		return true
	case reflect.Int64:
		return true
	case reflect.Uint:
		return true
	case reflect.Uint8:
		return true
	case reflect.Uint16:
		return true
	case reflect.Uint32:
		return true
	case reflect.Uint64:
		return true
	case reflect.Uintptr:
		return true
	case reflect.Float32:
		return true
	case reflect.Float64:
		return true
	case reflect.Complex64:
		return true
	case reflect.Complex128:
		return true
	case reflect.Array:
		return false
	case reflect.Chan:
		return false
	case reflect.Func:
		return false
	case reflect.Interface:
		return false
	case reflect.Map:
		return false
	case reflect.Ptr:
		return false
	case reflect.Slice:
		return false
	case reflect.String:
		return true
	case reflect.Struct:
		return false
	case reflect.UnsafePointer:
		return false
	}
	return false
}

//IsTime - tells if a reflected value is time
func IsTime(val *reflect.Value) bool {
	return val.IsValid() &&
		val.Kind() == reflect.Struct &&
		val.Type() == reflect.TypeOf(time.Time{})
}

//ToFlatMap - converts given composite data structure into a map of string to
//interfaces. The heirarchy of types are flattened into single level. The
//keys of the map indicate the original heirarchy
func ToFlatMap(obj interface{}, tagName string) (out map[string]interface{}) {
	out = make(map[string]interface{})
	Walk(obj, &WalkConfig{
		MaxDepth:         InfiniteDepth,
		IgnoreContainers: false,
		VisitPrivate:     false,
		VisitRootStruct:  false,
		FieldNameRetriever: func(field *reflect.StructField) string {
			jt := field.Tag.Get(tagName)
			if jt != "" {
				return jt
			}
			return field.Name
		},
		Visitor: func(state *WalkerState) bool {
			if IsBasicType(state.Current.Kind()) || IsTime(state.Current) {
				out[state.Path] = state.Current.Interface()
			}
			return true
		},
	})
	return out
}

//VisitorFunc - function that will be called on each value of reflected type.
//The return value decides whether to continue with depth search in current
//branch
type VisitorFunc func(state *WalkerState) (cont bool)

//FieldNameRetriever - retrieves name for the field from given
type FieldNameRetriever func(field *reflect.StructField) (name string)

//WalkConfig - determines how Walk is carried out
type WalkConfig struct {
	Visitor            VisitorFunc        //visitor function
	FieldNameRetriever FieldNameRetriever //func to get name from struct field
	MaxDepth           int                //Stop walk at this depth
	IgnoreContainers   bool               //Ignore slice and map parent objects
	VisitPrivate       bool               //Visit private fields
	VisitRootStruct    bool               //Visit the root struct thats passed
}

//WalkerState - current state of the walk
type WalkerState struct {
	Depth   int
	Field   *reflect.StructField
	Path    string
	Parent  *reflect.Value
	Current *reflect.Value
}

//InfiniteDepth - used to indicate that Walk should continue till all the nodes
//in the heirarchy are visited
const InfiniteDepth int = -1

//Walk - walk a given instance of struct/slice/map/basic type
func Walk(
	obj interface{},
	config *WalkConfig) {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)
	if config.Visitor == nil {
		return
	}
	if config.FieldNameRetriever == nil {
		config.FieldNameRetriever = func(field *reflect.StructField) string {
			return field.Name
		}
	}
	walkRecursive(
		config,
		WalkerState{
			Depth:   0,
			Field:   nil,
			Path:    "",
			Parent:  nil,
			Current: &original,
		})
}

func walkRecursive(config *WalkConfig, state WalkerState) {
	if config.MaxDepth > 0 && state.Depth == config.MaxDepth+1 {
		return
	}
	//We copy any field from state which is used inside the loops, so that
	//state is not cumulatevily modified in a loop
	cur := state.Current
	path := state.Path
	switch state.Current.Kind() {
	case reflect.Ptr:
		originalValue := state.Current.Elem()
		if !originalValue.IsValid() {
			return
		}
		state.Parent = state.Current
		state.Current = &originalValue
		walkRecursive(config, state)

	case reflect.Interface:
		originalValue := state.Current.Elem()
		state.Parent = state.Current
		state.Current = &originalValue
		walkRecursive(config, state)

	case reflect.Struct:
		state.Depth++
		if state.Depth == 1 {
			if config.VisitRootStruct && !config.Visitor(&state) {
				return
			}
		} else if !config.Visitor(&state) {
			return
		}
		for i := 0; i < cur.NumField(); i++ {
			field := cur.Field(i)
			//Dont want to walk unexported fields if VisitPrivate is false
			if !(config.VisitPrivate || field.CanSet()) {
				continue
			}
			structField := cur.Type().Field(i)
			state.Field = &structField
			if path != "" {
				state.Path = path + "." +
					config.FieldNameRetriever(&structField)
			} else {
				state.Path = config.FieldNameRetriever(&structField)
			}
			state.Parent = state.Current
			state.Current = &field
			walkRecursive(config, state)
		}

	case reflect.Slice:
		state.Depth++
		if config.IgnoreContainers {
			return
		}

		for i := 0; i < cur.Len(); i++ {
			state.Field = nil
			state.Path = path + "." + strconv.Itoa(i)
			value := cur.Index(i)
			// state.Parent = state.Current
			state.Current = &value
			walkRecursive(config, state)
		}
	case reflect.Map:
		state.Depth++
		if config.IgnoreContainers {
			return
		}
		for _, key := range cur.MapKeys() {
			originalValue := cur.MapIndex(key)
			state.Field = nil
			state.Path = path + "." + key.String()
			state.Parent = state.Current
			state.Current = &originalValue
			walkRecursive(config, state)
		}
	// And everything else will simply be taken from the original
	default:
		if cont := config.Visitor(&state); !cont {
			return
		}

	}

}
