package broker

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"therealbroker/pkg/broker"
	"time"
)

var mu sync.Mutex
var psDb *sql.DB

type Module struct {
	subjects []*Subject
	isClosed bool
}

type Subject struct {
	name string
	//messages    map[broker.Message]time.Time
	//idCounter   int
	subscribers []chan broker.Message
}

func NewModule(db *sql.DB) broker.Broker {
	psDb = db
	createTable()
	return &Module{
		subjects: make([]*Subject, 0),
		isClosed: false,
	}
}

func createTable() {
	_, err := psDb.Exec(`CREATE TABLE IF NOT EXISTS messages (subject TEXT, id INT, message TEXT, times TIMESTAMP)`)
	if err != nil {
		log.Fatal("2\n", err)
	}
}

func (m *Module) Close() error {
	if m.isClosed {
		return broker.ErrUnavailable
	}
	m.isClosed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	if m.isClosed {
		return 0, broker.ErrUnavailable
	}
	s := m.getSubjectByName(subject)

	sqlStmt1 := `INSERT INTO users (subject, id, message, time) VALUES ($1, $2, $3, $4) RETURNING id`
	sqlStmt2 := `SELECT COUNT id FROM messages WHERE subject='subject';`

	var id int
	err := psDb.QueryRow(sqlStmt2).Scan(&id)

	if err != nil {
		log.Fatal("35\n", err)
	}

	if err != nil {
		log.Fatal("23\n", err)
	}

	mu.Lock()

	var finalId int
	err = psDb.QueryRow(sqlStmt1, subject, int(id+1), msg, time.Now()).Scan(&finalId)
	if err != nil {
		panic(err)
	}

	for _, r := range s.subscribers {
		r <- msg
		//fmt.Println(len(s.subscribers))
	}
	m.subjects = append(m.subjects, s)
	mu.Unlock()
	return finalId, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	if m.isClosed {
		return nil, broker.ErrUnavailable
	}
	s := m.getSubjectByName(subject)

	mu.Lock()
	defer mu.Unlock()

	ch := make(chan broker.Message, 100)
	s.subscribers = append(s.subscribers, ch)
	m.subjects = append(m.subjects, s)
	return ch, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	if m.isClosed {
		return broker.Message{}, broker.ErrUnavailable
	}
	//s := m.getSubjectByName(subject)

	mu.Lock()
	defer mu.Unlock()

	sqlStmt := `SELECT message FROM messages WHERE subject='subject' AND id='id'`
	var result string
	err := psDb.QueryRow(sqlStmt).Scan(&result)
	if err != nil {
		log.Fatal("43\n", err)
	}
	//for msg, t := range s.messages {
	//	duration := time.Since(t)
	//	if msg.Id == id {
	//		if duration <= msg.Expiration {
	//			return msg, nil
	//		} else {
	//			return broker.Message{}, broker.ErrExpiredID
	//		}
	//	}
	//}
	msg := broker.Message{Id: id, Body: result}
	return msg, broker.ErrInvalidID
}

func (m *Module) getSubjectByName(name string) *Subject {
	for _, s := range m.subjects {
		if strings.Compare(s.name, name) == 0 {
			return s
		}
	}
	channels := make([]chan broker.Message, 0)
	s := Subject{
		name:        name,
		subscribers: channels,
	}
	return &s
}
