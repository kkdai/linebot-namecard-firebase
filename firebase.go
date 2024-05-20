package main

import (
	"context"
	"log"

	"firebase.google.com/go/db"
)

// Person 定義了 JSON 資料的結構體
type Person struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Company string `json:"company"`
}

// DBCardPath is the path to the namecard data in the database
const DBCardPath = "namecard"

// Define the context
var fireDB FireDB

// define firebase db
type FireDB struct {
	Path string
	CTX  context.Context
	*db.Client
}

// GetRef returns a reference to the location at the specified path.
func (f *FireDB) GetFromDB(data interface{}) error {
	if err := f.NewRef(f.Path).Get(f.CTX, data); err != nil {
		return err
	}
	return nil
}

// Insert data to firebase
func (f *FireDB) InsertDB(data interface{}) error {
	_, err := f.NewRef(f.Path).Push(f.CTX, data)
	if err != nil {
		return err
	}
	return nil
}

// SearchIfExist search if the email exist in the database
func (f *FireDB) SearchIfExist(email string) bool {
	var people map[string]Person
	err := f.NewRef(f.Path).OrderByChild("email").EqualTo(email).Get(f.CTX, &people)
	if err != nil {
		log.Println("Error getting data from DB:", err)
		return false
	}
	if len(people) > 0 {
		log.Println("Found data from DB:", people)
		return true
	}
	return false
}
