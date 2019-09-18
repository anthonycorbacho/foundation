// +build integration

package database

import (
	"context"
	"testing"
)

func TestOpenAndStatusCheck(t *testing.T) {
	// Given
	connection := "foundation:secret1234@tcp(127.0.0.1:3306)/db_test"

	db, close, err := Open("mysql", connection)
	if err != nil {
		t.Errorf("fail opening connection to db: %v", err)
		return
	}
	defer close()

	err = StatusCheck(context.TODO(), db)
	if err != nil {
		t.Errorf("fail statusCkeck: %v", err)
		return
	}
}
