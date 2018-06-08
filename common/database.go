package common

import (
	"fmt"
	"os"

	"github.com/recoilme/slowpoke"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Database struct {
	*slowpoke.Db
}

var DB *slowpoke.Db

// Opening a database and save the reference to `Database` struct.
func Init() *slowpoke.Db {
	db, err := slowpoke.Open("./../gorm.db")
	if err != nil {
		fmt.Println("db err: ", err)
	}
	DB = db
	return DB
}

// This function will create a temporarily database for running testing cases
func TestDBInit() *slowpoke.Db {
	test_db, err := slowpoke.Open("./../gorm_test.db") //gorm.Open("sqlite3", "./../gorm_test.db")

	if err != nil {
		fmt.Println("db err: ", err)
	}

	DB = test_db
	return nil
}

// Delete the database after running testing cases.
func TestDBFree(test_db *slowpoke.Db) error {
	//test_db.Close()
	os.RemoveAll("./../gorm_test.db")
	errdir := os.RemoveAll("db")
	return errdir
}

// Using this function to get a connection, you can create your connection pool here.
func GetDB() *slowpoke.Db {
	return DB
}
