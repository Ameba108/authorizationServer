package main

import (
	db "authoriz/database"
	server "authoriz/http/server"
	"fmt"
	"log"
)

func main() {
	var err error
	server.Database, err = db.GetConnection() // Подключение к бд
	if err != nil {
		log.Fatal(err)
	}
	defer server.Database.Close()

	err = server.Database.Ping() // Проверка работы бд
	if err != nil {
		log.Fatal("Can't connect to DB:", err)
	}
	fmt.Println("The connection was successful!")
	err = server.StartServer() // Запуск сервера
	fmt.Println("Server is working")
	if err != nil {
		log.Fatal(err)
	}
}
