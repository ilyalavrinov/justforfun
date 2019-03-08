package hlcup

type FilterType string

const (
	FilterEq       FilterType = "eq"
	FilterNeq      FilterType = "neq"
	FilterLt       FilterType = "lt"
	FilterGt       FilterType = "gt"
	FilterAny      FilterType = "any"
	FilterNull     FilterType = "null"
	FilterContains FilterType = "contains"
	FilterDomain   FilterType = "domain"
	FilterStarts   FilterType = "starts"
	FilterCode     FilterType = "code"
	FilterYear     FilterType = "year"
	FilterNow      FilterType = "now"
)

type Filter struct {
	Type   FilterType
	Values []string
}

type FilterSet struct {
	Sex       *Filter
	Email     *Filter
	Status    *Filter
	Firstname *Filter
	Surname   *Filter
	Phone     *Filter
	Country   *Filter
	City      *Filter
	Birth     *Filter
	Interests *Filter
	Likes     *Filter
	Premium   *Filter
}

type AccountFilter interface {
	Filter(f FilterSet) []int32
}
