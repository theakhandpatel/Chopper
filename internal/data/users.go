package data

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"
	"url_shortner/internal/validator"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)
var AnonymousUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Type      int       `json:"type"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (u *User) IsPremium() bool {
	return u.Type == 2
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(v.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}

}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {

	query := `
	INSERT INTO users (name,email,password_hash,type,created_at) 
	VALUES (?,?,?,?,?)
	`
	curTime := time.Now()
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Type, curTime}

	result, err := m.DB.Exec(query, args...)
	if err != nil {
		sqliteErr, ok := err.(sqlite3.Error)
		if ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	user.ID, _ = result.LastInsertId()
	user.CreatedAt = curTime

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, type
		FROM users
		WHERE email = ?`

	var user User

	err := m.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Type,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) SetUserType(userID int64, userType int) error {
	query := `
			UPDATE users
			SET type = ?
			WHERE id = ?
	`
	args := []interface{}{userType, userID}

	_, err := m.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	query := `
		SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.type
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.hash = ?
		AND tokens.scope = ?
		AND tokens.expiry > ?`

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}
	var user User

	err := m.DB.QueryRow(query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Type,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Return the matching user.
	return &user, nil
}
