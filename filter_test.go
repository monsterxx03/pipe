package main

import (
	_ "fmt"
	"testing"
)

func TestHttpFilterPlainString(t *testing.T) {
	filter := NewHttpFilter("url: /home & method: get")
	pattern, ok := filter.filters["url"]
	assertEqual(t, ok, true)
	assertEqual(t, pattern.MatchString("/home/page"), true)
	assertEqual(t, pattern.MatchString("xxhome"), false)
	pattern, ok = filter.filters["method"]
	assertEqual(t, ok, true)
	assertEqual(t, pattern.MatchString("xxget"), true)
	assertEqual(t, pattern.MatchString("et"), false)
}

func TestHttpFilterRegexp(t *testing.T) {
	filter := NewHttpFilter("url : ^/home$ & method: PUT")
	pattern, ok := filter.filters["url"]
	assertEqual(t, ok, true)
	assertEqual(t, pattern.MatchString("/home/page"), false)
	assertEqual(t, pattern.MatchString("/home"), true)
	pattern, ok = filter.filters["method"]
	assertEqual(t, ok, true)
	assertEqual(t, pattern.MatchString("put"), true)
}
