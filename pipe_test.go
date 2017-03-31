package main

import (
	"testing"
)

func assertEqual(t *testing.T, result interface{}, expected interface{}) {
	if result != expected {
		t.Errorf("result: %v \n no match expected: %v", result, expected)
	}
}

func TestBuildBPFFilter(t *testing.T) {
	// one host
	result := buildBPFFilter(false,
		[]string{"127.0.0.1"}, "80")
	assertEqual(t, result, "tcp dst port 80 and ( dst host 127.0.0.1)")
	// multi host
	result = buildBPFFilter(false,
		[]string{"127.0.0.1", "10.0.0.10"}, "5010")
	assertEqual(t, result, "tcp dst port 5010 and ( dst host 127.0.0.1 or dst host 10.0.0.10)")
	// track response
	result = buildBPFFilter(true, []string{"127.0.0.1"}, "80")
	assertEqual(t, result, "tcp port 80 and ( host 127.0.0.1)")
}
