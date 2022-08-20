package main

import (
	"database/sql"
	"encoding/json"
	"go-poll/db"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/go-chi/chi"
)

var pqsl *sql.DB

func main() {

	pqsl = connectDB()

	r := chi.NewRouter()

	r.Route("/poll", func(r chi.Router) {
		r.Post("/", AddPoll)
		r.Post("/", AddOptions)
		r.Get("/", GetPoll)
		r.Get("/all", GetPolls)
		r.Post("/vote", Vote)
		r.Get("/votes", GetPollVotes)
	})

	http.ListenAndServe(":80", r)
}

func AddPoll(w http.ResponseWriter, r *http.Request) {
	addpoll := &db.Poll{}
	json.NewDecoder(r.Body).Decode(addpoll)

	poll, err := db.CreatePoll(pqsl, addpoll.Title, addpoll.Description, addpoll.Options)
	if err != nil {
		w.Write([]byte("creating poll error: " + err.Error()))
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(poll)

}
func AddOptions(w http.ResponseWriter, r *http.Request) {
	addoptions := &db.Options{}
	json.NewDecoder(r.Body).Decode(addoptions)

	options, err := db.CreateOptions(pqsl, addoptions.Title, addoptions.PollID)
	if err != nil {
		w.Write([]byte("creating poll error: " + err.Error()))
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(options)

}

func GetPolls(w http.ResponseWriter, r *http.Request) {
	polls, err := db.GetPolls(pqsl)
	if err != nil {
		w.Write([]byte("error: " + err.Error()))
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(polls)
}

func GetPoll(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Write([]byte("id is not a number"))
		return
	}

	poll, err := db.GetPoll(pqsl, int64(id))
	if err != nil {
		w.Write([]byte("error: " + err.Error()))
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(poll)
}

func GetPollVotes(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Write([]byte("id is not a number"))
		return
	}

	poll, err := db.GetPollVotes(pqsl, int64(id))
	if err != nil {
		w.Write([]byte("error:"))
		return
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(poll)

}

func Vote(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	indexStr := r.URL.Query().Get("idx")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.Write([]byte("id is not a number"))
		return
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		w.Write([]byte("index is not a number"))
		return
	}
	poll := &db.Poll{}
	json.NewDecoder(r.Body).Decode(poll)

	err = db.ChoiceAndVote(pqsl, int64(id), int64(index), r.RemoteAddr)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

}

func connectDB() *sql.DB {
	conn, err := sql.Open("postgres", "postgresql://ipek:123456@localhost:5432/go-poll?sslmode=disable")
	if err != nil {
		log.Fatalf("error creating database: %v \n", err)
	}

	return conn
}
