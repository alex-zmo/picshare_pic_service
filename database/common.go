package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDatabase() {
	serverName := "localhost:3306"
	user := "root"
	password := ""
	databaseName := "pic_share"

	connectionString := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, serverName, databaseName)
	database, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	db = database
	createImageTable()
	createLinkTable()
	createLikesTable()
	createFavoritesTable()
}

func closeRows(rows *sql.Rows) {
	err := rows.Close()

	if err != nil {
		fmt.Println(err)
	}
}
