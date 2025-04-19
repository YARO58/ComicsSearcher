package main

import (
	"log"
	"net/http"
	"os"

	"yadro.com/course/frontend/internal/handlers"
	"yadro.com/course/frontend/internal/templates"
)

func main() {
	templates.Init()

	handler := handlers.NewHandler()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handler.HomeHandler)
	http.HandleFunc("/search", handler.SearchHandler)
	http.HandleFunc("/login", handler.LoginHandler)
	http.HandleFunc("/admin", handler.AdminHandler)
	http.HandleFunc("/admin/update", handler.UpdateHandler)
	http.HandleFunc("/admin/drop", handler.DropHandler)
	http.HandleFunc("/logout", handler.LogoutHandler)

	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Frontend сервис запущен на порту %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
