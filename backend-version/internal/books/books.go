package books

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidTitle  = errors.New("Please enter a title.")
	ErrInvalidAuthor = errors.New("Please enter an author.")
	ErrInvalidStatus = errors.New("Status must be To Read, Reading, or Finished.")
	ErrNotFound      = errors.New("Book not found.")
)

var validStatuses = map[string]bool{
	"To Read":  true,
	"Reading":  true,
	"Finished": true,
}

type Book struct {
	ID        string `json:"id"`
	OwnerID   string `json:"ownerId"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
}

type Store struct {
	mu    sync.Mutex
	books []Book
	path  string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.books = []Book{}
		return nil
	}
	if err != nil {
		return err
	}

	var books []Book
	if err := json.Unmarshal(data, &books); err != nil {
		return fmt.Errorf("decode %s: %w", s.path, err)
	}
	s.books = make([]Book, 0, len(books))
	for _, book := range books {
		if book.ID == "" || book.Title == "" || book.Author == "" || book.CreatedAt <= 0 || !validStatuses[book.Status] {
			continue
		}
		s.books = append(s.books, book)
	}
	sort.SliceStable(s.books, func(i, j int) bool {
		return s.books[i].CreatedAt > s.books[j].CreatedAt
	})
	return nil
}

func (s *Store) List(ownerID string) []Book {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Book, 0)
	for _, book := range s.books {
		if book.OwnerID == ownerID {
			result = append(result, book)
		}
	}
	return result
}

func (s *Store) Create(ownerID, title, author, status string) (Book, error) {
	title = strings.TrimSpace(title)
	author = strings.TrimSpace(author)
	if title == "" {
		return Book{}, ErrInvalidTitle
	}
	if author == "" {
		return Book{}, ErrInvalidAuthor
	}
	if !validStatuses[status] {
		return Book{}, ErrInvalidStatus
	}

	id, err := newID()
	if err != nil {
		return Book{}, err
	}
	book := Book{ID: id, OwnerID: ownerID, Title: title, Author: author, Status: status, CreatedAt: time.Now().UnixMilli()}

	s.mu.Lock()
	defer s.mu.Unlock()
	oldBooks := append([]Book(nil), s.books...)
	s.books = append([]Book{book}, s.books...)
	if err := s.saveLocked(); err != nil {
		s.books = oldBooks
		return Book{}, fmt.Errorf("save book: %w", err)
	}
	return book, nil
}

func (s *Store) UpdateStatus(ownerID, id, status string) (Book, error) {
	if !validStatuses[status] {
		return Book{}, ErrInvalidStatus
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.books {
		if s.books[index].ID != id || s.books[index].OwnerID != ownerID {
			continue
		}
		oldStatus := s.books[index].Status
		s.books[index].Status = status
		if err := s.saveLocked(); err != nil {
			s.books[index].Status = oldStatus
			return Book{}, fmt.Errorf("save book: %w", err)
		}
		return s.books[index], nil
	}
	return Book{}, ErrNotFound
}

func (s *Store) Delete(ownerID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.books {
		if s.books[index].ID != id || s.books[index].OwnerID != ownerID {
			continue
		}
		oldBooks := append([]Book(nil), s.books...)
		s.books = append(s.books[:index], s.books[index+1:]...)
		if err := s.saveLocked(); err != nil {
			s.books = oldBooks
			return fmt.Errorf("save book: %w", err)
		}
		return nil
	}
	return ErrNotFound
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.books, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0644)
}

func newID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}
