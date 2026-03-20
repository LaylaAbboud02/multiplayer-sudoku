package main

import (
	"html/template"
	"log"
	"multiplayer-sudoku/internal/game"
	"net/http"
)

type PageData struct {
	Board game.Board
}

func main () {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Board: game.NewSampleBoard(),
		}

		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("template execute error:", err)
			return
		}
	})

	log.Println("Server running at http://localhost:8080")
	
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}