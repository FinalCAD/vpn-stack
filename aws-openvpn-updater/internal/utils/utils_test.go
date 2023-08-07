package utils

import (
	"testing"
)

func TestExtractEmail(t *testing.T) {
	got, _ := ExtractEmail("test.test", "testdomain.com")
	want := "test.test@testdomain.com"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestExtractEmailUnderscore(t *testing.T) {
	got, _ := ExtractEmail("test-test.test_zrerez", "testdomain.com")
	want := "test-test.test@testdomain.com"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
