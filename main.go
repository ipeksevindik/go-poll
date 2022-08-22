package main

import (
	"database/sql"
	"encoding/json"
	"go-poll/db"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"

	"github.com/go-chi/chi"
)

var pqsl *sql.DB

func main() {

	pqsl = connectDB()

	r := chi.NewRouter()

	r.Route("/poll", func(r chi.Router) {
		r.Post("/add", AddPoll)
		r.Get("/all", GetPolls)
		r.Patch("/", ChoicePoll)
	})

	http.ListenAndServe(":80", r)
}

func AddPoll(w http.ResponseWriter, r *http.Request) {
	addpoll := &db.Polls{}
	json.NewDecoder(r.Body).Decode(addpoll)

	poll_id, err := db.CreatePoll(pqsl, addpoll.Title, addpoll.Description)
	if err != nil {
		w.Write([]byte("creating poll error: " + err.Error()))
	}

	for i, option := range addpoll.Options {
		option, err := db.CreateOptions(pqsl, option.Title, poll_id)
		if err != nil {
			w.Write([]byte("creating poll error: " + err.Error()))
			return
		}
		addpoll.Options[i] = option
	}

	addpoll.ID = poll_id
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(addpoll)

}

func GetPolls(w http.ResponseWriter, r *http.Request) {
	polls, err := db.GetPolls(pqsl)
	if err != nil {
		w.Write([]byte("error: " + err.Error()))
	}

	for _, poll := range polls {
		option_ids := []int64{}
		for _, option := range poll.Options {
			option_ids = append(option_ids, option.ID)
		}
		votes, err := db.GetPollVotes(pqsl, option_ids)
		if err != nil {
			w.Write([]byte("error: " + err.Error()))
		}
		for _, option := range poll.Options {
			option.Vote = votes[option.ID]
		}

	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(polls)
}

func ChoicePoll(w http.ResponseWriter, r *http.Request) {
	voteStr := r.URL.Query().Get("v")
	vote, err := strconv.Atoi(voteStr)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	splitted := strings.Split(r.RemoteAddr, ":")
	choice, err := db.ChoiceAndVote(pqsl, int64(vote), splitted[0])
	if err != nil {
		w.Write([]byte("voted with the same ip"))
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(choice)

}

func connectDB() *sql.DB {
	conn, err := sql.Open("postgres", "postgresql://ipek:123456@localhost:5432/go-poll?sslmode=disable")
	if err != nil {
		log.Fatalf("error creating database: %v \n", err)
	}

	return conn
}
