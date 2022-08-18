package db

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Poll struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Options     []string `json:"options"`
	Votes       int      `json:"votes,omitempty"`
	Voters      []string `json:"voters,omitempty"`
}

func (poll *Poll) ToJson() ([]byte, error) {
	data, err := json.Marshal(poll)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (poll *Poll) FromJson(jsonstr string) error {
	err := json.Unmarshal([]byte(jsonstr), poll)
	if err != nil {
		return err
	}
	return nil
}

func CreatePoll(db *sql.DB, title string, description string, options []string) (*Poll, error) { // bunu düzelt
	row := db.QueryRowContext(context.TODO(), "insert into poll (title, description) values ($1, $2) returning id, title, descriptions", title, description, options)
	poll := &Poll{}
	err := row.Scan(&poll.ID, &poll.Title, &poll.Description, &poll.Options)
	return poll, err
}

func ChoiceAndVote(db *sql.DB, pollID int64, optionid int64, ip string) error {
	row := db.QueryRowContext(context.TODO(), "select count(*) from votes where option_id = $1 and ip = $2 returning id", optionid, ip)
	poll := &Poll{}
	err := row.Scan(&poll.Votes)

	return err
}

func GetPollVotes(db *sql.DB, id int64) (*Poll, error) {
	row := db.QueryRowContext(context.TODO(), `select options.poll_id, json_object_agg(options.id,options.title), votes.ip as voter from options
	inner join votes on votes.option_id = options.id where options.poll_id = $1`, id)

	poll := &Poll{}
	err := row.Scan(poll.ID, poll.Options, poll.Voters)
	if err != nil {
		return nil, err
	}

	return poll, nil
}

func GetPolls(db *sql.DB) ([]*Poll, error) {
	rows, err := db.QueryContext(context.TODO(), `select polls.title, description, json_object_agg(options.id,options.title) from polls 
	inner join options on options.poll_id = polls.id 
	group by polls.id;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Poll

	for rows.Next() {
		item := &Poll{}
		err = rows.Scan(&item.Title, &item.Description, &item.Options)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}

func GetPoll(db *sql.DB, id int64) (*Poll, error) {
	row := db.QueryRowContext(context.TODO(), "select id, title, description, options, votes from polls where id= $1", id)
	poll := &Poll{}
	err := row.Scan(poll.ID, poll.Title, poll.Description, poll.Options, poll.Votes)
	if err != nil {
		return nil, err
	}

	return poll, nil
}
