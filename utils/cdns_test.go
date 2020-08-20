package utils

import (
	"fmt"
	"testing"
)

func TestGetCloudflareIPRanges(t *testing.T) {
	data, err := GetCloudflareIPRanges()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%v", data)
	if len(data) == 0 {
		t.Fatalf("expected some data")
	}
}
