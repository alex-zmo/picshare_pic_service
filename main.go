package main

import (
	"github.com/gmo-personal/picshare_pic_service/database"
	"github.com/gmo-personal/picshare_pic_service/server"
)

func main() {
	database.InitDatabase()
	server.InitServer()
}

