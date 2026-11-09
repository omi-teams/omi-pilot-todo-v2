package platform

import (
	"context"
	"strings"
	"testing"
)

func TestDatabaseURLRequired(t *testing.T) {
	pool, err := OpenDatabase(context.Background(), "")
	if pool != nil || err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("got %v, %v", pool, err)
	}
}
func TestInvalidDatabaseURL(t *testing.T) {
	pool, err := OpenDatabase(context.Background(), "://invalid")
	if pool != nil || err == nil {
		t.Fatalf("got %v, %v", pool, err)
	}
}
