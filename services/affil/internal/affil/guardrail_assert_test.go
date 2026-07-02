package affil

import (
	"testing"
)

func TestAssert_SinglePath_OK(t *testing.T) {
	routes := []Route{{Path: "/v1/affiliate/link", CreatesAffiliateLink: true, IncludesDisclosure: true}}
	if err := AssertSingleAffiliatePath(routes); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAssert_BackDoor_Fails(t *testing.T) {
	routes := []Route{
		{Path: "/v1/affiliate/link", CreatesAffiliateLink: true, IncludesDisclosure: true},
		{Path: "/v1/affiliate/auto", CreatesAffiliateLink: true, IncludesDisclosure: true}, // cửa thứ hai
	}
	if err := AssertSingleAffiliatePath(routes); err == nil {
		t.Fatal("expected error for back-door route, got nil")
	}
}

func TestAssert_NoDisclosure_Fails(t *testing.T) {
	routes := []Route{{Path: "/v1/affiliate/link", CreatesAffiliateLink: true, IncludesDisclosure: false}}
	if err := AssertSingleAffiliatePath(routes); err == nil {
		t.Fatal("expected error for missing disclosure, got nil")
	}
}
