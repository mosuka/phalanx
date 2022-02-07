package wildcard

import (
	"testing"
)

func TestMatch(t *testing.T) {
	actual := Match("*", "hoge")
	expected := true
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}

func TestMatchSuffix(t *testing.T) {
	actual := Match("*oge", "hoge")
	expected := true
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}

func TestMatchSuffixUnmatch(t *testing.T) {
	actual := Match("*age", "hoge")
	expected := false
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}

func TestMatchPrefix(t *testing.T) {
	actual := Match("ho*", "hoge")
	expected := true
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}

func TestMatchPrefixUnmatch(t *testing.T) {
	actual := Match("hu*", "hoge")
	expected := false
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}

func TestMatchExact(t *testing.T) {
	actual := Match("hoge", "hoge")
	expected := true
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}

func TestMatchExactUnmatch(t *testing.T) {
	actual := Match("huge", "hoge")
	expected := false
	if actual != expected {
		t.Fatalf("expected %v, but %v\n", expected, actual)
	}
}
