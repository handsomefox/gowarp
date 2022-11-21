package storage

import (
	"errors"
	"sync"

	"github.com/handsomefox/gowarp/client/account"
)

var ErrEmptyStore = errors.New("internal storage is empty")

// Stack is a thread-safe 'stack' for account data with push/pop functionality.
type Stack struct {
	mu    sync.Mutex
	store []account.Data
}

func NewStack() *Stack {
	return &Stack{
		store: make([]account.Data, 0),
		mu:    sync.Mutex{},
	}
}

func (s *Stack) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.store)
}

// Push adds the item to the collection.
func (s *Stack) Push(acc account.Data) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store = append(s.store, acc)
}

// Pop either returns the last item in the stack and removes it from the internal collection,
// or returns an error.
func (s *Stack) Pop() (*account.Data, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.store) < 1 {
		return nil, ErrEmptyStore
	}

	item := s.store[len(s.store)-1]
	s.store = s.store[:len(s.store)-1]

	return &item, nil
}
