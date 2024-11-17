package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var users []User

type User struct {
	Id          int    `db:"Id"`
	Name        string `db:"Name"`
	Second_Name string `db:"Second_Name"`
	Email       string `db:"Email"`
	Password    string `db:"Password"`
}

const (
	host   = "localhost"
	port   = 5432
	user   = "postgres"
	dbname = "users"
)

var mu sync.Mutex

func GetConnection() (*sqlx.DB, error) { // Подключение к бд
	mu.Lock()
	defer mu.Unlock()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	password := os.Getenv("PASSWORD")
	sqlxConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sqlx.Connect("postgres", sqlxConn)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	fmt.Println("Everything went good!")
	return db, nil
}

func AlreadyInDB(db *sqlx.DB, email string) bool { // Проверка есть ли пользователь в бд
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Checking user by email=%s", email)
	err := db.Select(&users, `SELECT "Id", "Name", "Second_Name", "Email", "Password" FROM public."autorizated_users_list" WHERE "Email"=$1 LIMIT 1;`, email)
	if err != nil {
		log.Fatal(err)
		return false
	}
	if len(users) > 0 {
		return true
	}
	fmt.Println("User is not in db")
	return false
}

func ShowProfile(db *sqlx.DB, email string) []User { // Показать всю информацию о пользователе
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("Searching for user: email=%s\n", email)
	err := db.Select(&users, `select * from public."autorizated_users_list" where "Email"=$1`, email)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d users\n", len(users))
	return users
}

func GetUsersList(db *sqlx.DB) { // Получить список пользователей
	mu.Lock()
	defer mu.Unlock()

	err := db.Select(&users, `select * from public."autorizated_users_list";`)
	if err != nil {
		log.Fatal(err)
	}
}

func PostData(db *sqlx.DB, name, secondName, email, password string) { // Отправить данные в бд
	mu.Lock()
	defer mu.Unlock()
	_, err := db.Exec(`insert into public."autorizated_users_list" ("Name", "Second_Name", "Email", "Password") values ($1, $2, $3, $4)`,
		name, secondName, email, password)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Data is posted!")
}
