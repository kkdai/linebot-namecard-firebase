package main

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestSearchIfExist(t *testing.T) {
	uid := "INPUT_YOUR_UID_HERE"
	gap := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	firebaseURL := os.Getenv("FIREBASE_URL")

	// If no environment variable, goskip the test
	if gap == "" || firebaseURL == "" {
		t.Skip("No environment variable")
	}

	ctx := context.Background()
	initFirebase(gap, firebaseURL, ctx)
	userPath := fmt.Sprintf("%s/%s", DBCardPath, uid)
	fireDB.Path = userPath

	// Test case 1
	email := "search@email.com"
	got := fireDB.SearchIfExist(email)
	if got != true {
		t.Errorf("SearchIfExist() = %t; want true", got)
	}

}
