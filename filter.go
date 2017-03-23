package main

import (
	"regexp"
)

// url: /tet & method: get & body: balabal & Content-Type: application/json
type HttpFilter struct {
	filterStr string
	filters   map[string]*regexp.Regexp
}

func (f HttpFilter) String() string {
	return f.filterStr
}

func (f *HttpFilter) IsEmpty() bool {
	return len(f.filters) == 0
}

func (f *HttpFilter) compile() {
	if len(f.filterStr) == 0 {
		return
	}
	pattern := regexp.MustCompile("\\s+&\\s+")
	filters := pattern.Split(f.filterStr, -1)
	subPattern := regexp.MustCompile("\\s*:\\s*")
	for _, filter := range filters {
		result := subPattern.Split(filter, 2)
		f.filters[result[0]] = regexp.MustCompile("(?im:" + result[1] + ")")
	}
}

func NewHttpFilter(filterStr string) *HttpFilter {
	f := &HttpFilter{filterStr, map[string]*regexp.Regexp{}}
	f.compile()
	return f
}