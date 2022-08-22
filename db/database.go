package db

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
)

type Polls struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Options     []*Options `json:"options"`
}

type Options struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Vote  int64  `json:"vote,omitempty"`
}

func (poll *Polls) ToJson() ([]byte, error) {
	data, err := json.Marshal(poll)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (poll *Polls) FromJson(jsonstr string) error {
	err := json.Unmarshal([]byte(jsonstr), poll)
	if err != nil {
		return err
	}
	return nil
}

func CreatePoll(db *sql.DB, title string, description string) (int64, error) {
	row := db.QueryRowContext(context.TODO(), "insert into polls (title, description) values ($1, $2) returning id", title, description)
	var poll_id int64
	err := row.Scan(&poll_id)
	return poll_id, err
}

func CreateOptions(db *sql.DB, title string, poll_id int64) (*Options, error) {
	row := db.QueryRowContext(context.TODO(), "insert into options (title, poll_id) values ($1, $2) returning id, title", title, poll_id)
	options := &Options{}
	err := row.Scan(&options.ID, &options.Title)
	return options, err
}

func ChoiceAndVote(db *sql.DB, option_id int64, ip string) (string, error) {
	row := db.QueryRowContext(context.TODO(), "insert into votes(ip, option_id) values ($1, $2) returning ip", ip, option_id)
	err := row.Scan(&ip)
	return ip, err
}

func GetPollVotes(db *sql.DB, options []int64) (map[int64]int64, error) {
	row, err := db.QueryContext(context.TODO(), "select option_id, count(ip) from votes where option_id = any($1) group by option_id", pq.Array(options))
	if err != nil {
		return nil, err
	}
	defer row.Close()
	result := map[int64]int64{}
	var id int64
	var vote int64
	for row.Next() {
		err = row.Scan(&id, &vote)
		if err != nil {
			return nil, err
		}
		result[id] = vote
	}

	return result, nil
}

func GetPolls(db *sql.DB) ([]*Polls, error) {
	rows, err := db.QueryContext(context.TODO(), `select polls.id, polls.title, polls.description, json_object_agg(options.id,options.title) from polls 
	inner join options on options.poll_id = polls.id 
	group by polls.id;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Polls

	for rows.Next() {
		item := &Polls{}
		options := map[int]string{}
		optionsBytes := json.RawMessage{}
		err = rows.Scan(&item.ID, &item.Title, &item.Description, &optionsBytes)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(optionsBytes, &options)
		if err != nil {
			return nil, err
		}
		for id, title := range options {
			item.Options = append(item.Options, &Options{ID: int64(id), Title: title})
		}
		result = append(result, item)
	}

	return result, nil
}
