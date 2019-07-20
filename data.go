package teak

import (
	"fmt"
	"time"
)

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
