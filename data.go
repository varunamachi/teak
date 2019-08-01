package teak

import (
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
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

//String - version to string
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

//DateRange - represents date ranges
type DateRange struct {
	// Name string    `json:"name" bson:"name"`
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
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
	Key   string `json:"key"`
	Value string `json:"value"`
}

//Range - integer range
type Range struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

//Param - represents generic parameter
type Param struct {
	Name    string      `json:"name"`
	Type    ParamType   `json:"type"`
	Desc    string      `json:"desc"`
	Range   Range       `json:"range"`
	Choices []Pair      `json:"choices"`
	Default interface{} `json:"def"`
	// Value   interface{} `json:"value" bson:"value"`
}

//CountList - paginated list returned from mongoDB along with total number of
//items in the list counted without pagination
type CountList struct {
	TotalCount int         `json:"total"`
	Data       interface{} `json:"data"`
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
	Field string     `json:"field"`
	Name  string     `json:"name"`
	Type  FilterType `json:"type"`
}

//Matcher - matches the given fields. If multiple fileds are given the; the
//joining condition is decided by the MatchStrategy given
type Matcher struct {
	Strategy MatchStrategy `json:"strategy"`
	Fields   []interface{} `json:"fields"`
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
	Props    map[string]Matcher     `json:"props"`
	Bools    map[string]interface{} `json:"bools"`
	Dates    map[string]DateRange   `json:"dates"`
	Lists    map[string]Matcher     `json:"lists"`
	Searches map[string]Matcher     `json:"searches"`
}

//FilterSpecList - alias for array of filter specs
type FilterSpecList []*FilterSpec

//FilterVal - values for filter along with the count
type FilterVal struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

//DataStorage - defines a data storage
type DataStorage interface {
	Name() string
	Count(dtype string, filter *Filter) (count int, err error)
	Create(dataType string, data interface{}) error
	Update(
		dataType string,
		keyField string,
		key interface{},
		data interface{}) error
	Delete(
		dataType string,
		keyField string,
		key interface{}) error
	RetrieveOne(
		dataType string,
		keyField string,
		key interface{},
		data interface{}) error
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
	GetFilterValues(
		dtype string,
		specs FilterSpecList) (values M, err error)
	GetFilterValuesX(
		dtype string,
		field string,
		specs FilterSpecList,
		filter *Filter) (values M, err error)

	Init() error
	Setup(params M) error
	Reset() error
	Destroy() error
	Wrap(cmd *cli.Command) *cli.Command
	GetManageCommands() []cli.Command
}

//@TODO Data store ini shall do these
// vevt.SetEventAuditor(auditor)

//VisitorFunc - function that will be called on each value of reflected type.
//The return value decides whether to continue with depth search in current
//branch
type VisitorFunc func(
	path string,
	tag reflect.StructTag,
	parent *reflect.Value,
	value *reflect.Value) (cont bool)

type WalkConfig struct {
	MaxDepth             int
	Visitor              VisitorFunc
	VisitContainerParent bool
}

//Walk - walk a given instance of struct/slice/map/basic type
func Walk(
	obj interface{},
	config *WalkConfig) {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)

	walkRecursive(config, 0, "", "", nil, &original)
}

func walkRecursive(
	config *WalkConfig,
	depth int,
	tag reflect.StructTag,
	path string,
	parent *reflect.Value,
	original *reflect.Value) {
	if config.MaxDepth > 0 && depth == config.MaxDepth+1 {
		return
	}
	depth++
	switch original.Kind() {
	case reflect.Ptr:
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return
		}
		walkRecursive(
			config,
			depth,
			"",
			path,
			original,
			&originalValue)

	case reflect.Interface:
		originalValue := original.Elem()
		walkRecursive(
			config,
			depth,
			"",
			path,
			original,
			&originalValue)

	case reflect.Struct:
		if !config.Visitor(path, "", parent, original) {
			return
		}
		for i := 0; i < original.NumField(); i++ {
			field := original.Field(i)
			structField := original.Type().Field(i)
			fieldPath := structField.Name
			if path != "" {
				fieldPath = path + "." + structField.Name
			}
			walkRecursive(
				config,
				depth,
				structField.Tag,
				fieldPath,
				original,
				&field)
		}
	case reflect.Slice:
		if config.VisitContainerParent &&
			!config.Visitor(path, "", parent, original) {
			return
		}
		for i := 0; i < original.Len(); i++ {
			itemPath := path + "[" + strconv.Itoa(i) + "]"
			value := original.Index(i)
			walkRecursive(
				config,
				depth,
				"",
				itemPath,
				original,
				&value)
		}
	case reflect.Map:
		if config.VisitContainerParent &&
			!config.Visitor(path, "", parent, original) {
			return
		}
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			itemPath := path + "[" + key.String() + "]"
			walkRecursive(
				config,
				depth,
				"",
				itemPath,
				original,
				&originalValue)
		}
	// And everything else will simply be taken from the original
	default:
		if cont := config.Visitor(path, tag, parent, original); !cont {
			return
		}

	}

}
