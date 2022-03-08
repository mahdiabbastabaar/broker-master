package broker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"therealbroker/pkg/broker"
	"time"
)

var mu sync.Mutex

type Module struct {
	subjects []*Subject
	isClosed bool
}

type Subject struct {
	name        string
	messages    map[broker.Message]time.Time
	idCounter   int
	subscribers []chan broker.Message
}

func NewModule() broker.Broker {
	return &Module{
		subjects: make([]*Subject, 0),
		isClosed: false,
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

	mu.Lock()
	defer mu.Unlock()

	s.messages[msg] = time.Now()

	for _, r := range s.subscribers {
		r <- msg
		fmt.Println(len(s.subscribers))
	}

	m.subjects = append(m.subjects, s)

	s.idCounter++
	msg.Id = s.idCounter

	return msg.Id, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	if m.isClosed {
		return nil, broker.ErrUnavailable
	}
	s := m.getSubjectByName(subject)
	ch := make(chan broker.Message, 100)
	s.subscribers = append(s.subscribers, ch)
	m.subjects = append(m.subjects, s)
	return ch, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	if m.isClosed {
		return broker.Message{}, broker.ErrUnavailable
	}
	s := m.getSubjectByName(subject)
	//TODO optimize searching
	for msg, t := range s.messages {
		duration := time.Since(t)
		if msg.Id == id {
			if duration <= msg.Expiration {
				return msg, nil
			} else {
				return broker.Message{}, broker.ErrExpiredID
			}
		}
	}
	return broker.Message{}, broker.ErrInvalidID
}

func (m *Module) getSubjectByName(name string) *Subject {
	for _, s := range m.subjects {
		if strings.Compare(s.name, name) == 0 {
			return s
		}
	}
	mes := make(map[broker.Message]time.Time)
	channels := make([]chan broker.Message, 0)
	s := Subject{
		name:        name,
		messages:    mes,
		idCounter:   0,
		subscribers: channels,
	}
	return &s
}
