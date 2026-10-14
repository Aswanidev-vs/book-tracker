package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	ErrInvalidEmail       = errors.New("Enter a valid email address.")
	ErrWeakPassword       = errors.New("Password must be at least 8 characters and include a letter and a number.")
	ErrEmailAlreadyExists = errors.New("An account with that email already exists.")
	ErrInvalidCredentials = errors.New("Email or password is incorrect.")
)

const passwordIterations = 120000
const sessionLifetime = 24 * time.Hour

var emailPattern = regexp.MustCompile(`^[^@ \t\r\n]+@[^@ \t\r\n]+\.[^@ \t\r\n]+$`)

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	CreatedAt    int64  `json:"createdAt"`
}

type Store struct {
	mu       sync.Mutex
	users    []User
	path     string
	sessions map[string]session
}

type session struct {
	userID    string
	expiresAt time.Time
}

func NewStore(path string) *Store {
	return &Store{path: path, sessions: make(map[string]session)}
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.users = []User{}
		return nil
	}
	if err != nil {
		return err
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("decode %s: %w", s.path, err)
	}
	s.users = users
	return nil
}

func (s *Store) Signup(email, password string) (User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return User{}, err
	}
	if err := ValidatePassword(password); err != nil {
		return User{}, err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	id, err := newID()
	if err != nil {
		return User{}, err
	}
	user := User{ID: id, Email: email, PasswordHash: hash, CreatedAt: time.Now().UnixMilli()}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.users {
		if existing.Email == email {
			return User{}, ErrEmailAlreadyExists
		}
	}
	oldUsers := append([]User(nil), s.users...)
	s.users = append(s.users, user)
	if err := s.saveLocked(); err != nil {
		s.users = oldUsers
		return User{}, fmt.Errorf("save user: %w", err)
	}
	return user, nil
}

func (s *Store) Login(email, password string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.Email == email && verifyPassword(user.PasswordHash, password) {
			return user, nil
		}
	}
	return User{}, ErrInvalidCredentials
}

func (s *Store) CreateSession(userID string) (string, error) {
	token, err := randomHex(32)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.sessions[hashToken(token)] = session{userID: userID, expiresAt: time.Now().Add(sessionLifetime)}
	s.mu.Unlock()
	return token, nil
}

func (s *Store) UserIDForSession(token string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.sessions[hashToken(token)]
	if !ok || time.Now().After(current.expiresAt) {
		if ok {
			delete(s.sessions, hashToken(token))
		}
		return "", false
	}
	return current.userID, true
}

func (s *Store) DeleteSession(token string) {
	s.mu.Lock()
	delete(s.sessions, hashToken(token))
	s.mu.Unlock()
}

func (s *Store) FindUser(id string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.ID == id {
			return user, true
		}
	}
	return User{}, false
}

func ValidateEmail(email string) error {
	if !emailPattern.MatchString(strings.TrimSpace(email)) {
		return ErrInvalidEmail
	}
	return nil
}

func ValidatePassword(password string) error {
	if len([]rune(password)) < 8 {
		return ErrWeakPassword
	}
	hasLetter, hasNumber := false, false
	for _, character := range password {
		hasLetter = hasLetter || unicode.IsLetter(character)
		hasNumber = hasNumber || unicode.IsNumber(character)
	}
	if !hasLetter || !hasNumber {
		return ErrWeakPassword
	}
	return nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := ValidateEmail(email); err != nil {
		return "", err
	}
	return email, nil
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0600)
}

func hashPassword(password string) (string, error) {
	salt, err := randomBytes(16)
	if err != nil {
		return "", err
	}
	digest := digestPassword(salt, password)
	return hex.EncodeToString(salt) + "$" + hex.EncodeToString(digest[:]), nil
}

func verifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	actual := digestPassword(salt, password)
	return subtle.ConstantTimeCompare(actual[:], expected) == 1
}

func digestPassword(salt []byte, password string) [32]byte {
	input := append(append([]byte{}, salt...), []byte(password)...)
	digest := sha256.Sum256(input)
	for index := 1; index < passwordIterations; index++ {
		input = append(append([]byte{}, salt...), digest[:]...)
		digest = sha256.Sum256(input)
	}
	return digest
}

func newID() (string, error) {
	bytes, err := randomBytes(16)
	if err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}

func randomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	return bytes, err
}

func randomHex(size int) (string, error) {
	bytes, err := randomBytes(size)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}
