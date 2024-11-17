package server

import (
	db "authoriz/database"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var Database *sqlx.DB
var user db.User

func handleEnter(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token") // Проверяем, есть ли куки
	if err == nil && cookie != nil {         // Если куки есть, проверяем пользователя
		if db.AlreadyInDB(Database, cookie.Value) {
			http.Redirect(w, r, "/success", http.StatusSeeOther)
			return
		}
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	fmt.Printf("Got data: email=%s, password=%s\n", email, password)
	if email != "" && password != "" {
		if db.AlreadyInDB(Database, email) { // Если пользователь уже есть в бд
			fmt.Println("User is already in DB, redirecting to /success")
			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   email,
				Expires: time.Now().Add(500 * time.Second),
			})
			http.Redirect(w, r, "/success", http.StatusSeeOther)
		} else {
			fmt.Println("User is not in DB, redirecting to /regist") // Если нет - на регистрацию
			http.Redirect(w, r, "/regist", http.StatusSeeOther)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func handleAuthorization(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // Проверяем метод, чтобы при обновлении страницы не отправлялись пустые поля в бд
		http.Redirect(w, r, "/regist", http.StatusSeeOther)
		return
	}
	name := r.FormValue("username")
	secondName := r.FormValue("second_name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if name != "" && email != "" && password != "" {
		if !db.AlreadyInDB(Database, email) { // Если пользователя нет в бд, то его данные отправляются в бд
			user = db.User{Email: email, Password: password}
			db.PostData(Database, name, secondName, email, password)
			http.SetCookie(w, &http.Cookie{ // Отправляем куки
				Name:    "session_token",
				Value:   email,
				Expires: time.Now().Add(500 * time.Second),
			})

			http.Redirect(w, r, "/enter", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/enter", http.StatusSeeOther) // Если пользователь уже есть в бд, то отправляем его на страницу входа
		}
	} else {
		http.Redirect(w, r, "/regist", http.StatusSeeOther)
	}
}

func handleSuccess(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token") // Проверяем наличие куки
	if err != nil || cookie == nil || !db.AlreadyInDB(Database, cookie.Value) {
		http.Redirect(w, r, "/enter", http.StatusSeeOther) // Перенаправляем на вход если куки нет или пользователь не найден
		return
	}

	users := db.ShowProfile(Database, cookie.Value)
	tmpl, err := template.ParseFiles("http/templates/profile.html")
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(w, users)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleRegist(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "http/templates/regist.html")
}

func handleAuthor(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "http/templates/enter.html")
}

func StartServer() error {
	mux := http.NewServeMux()
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("http/templates"))))
	mux.HandleFunc("/", handleAuthor)
	mux.HandleFunc("/enter", handleEnter)
	mux.HandleFunc("/regist", handleRegist)
	mux.HandleFunc("/postform", handleAuthorization)
	mux.HandleFunc("/success", handleSuccess)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return err
	} else {
		fmt.Println("Server is running")
	}
	return nil
}
