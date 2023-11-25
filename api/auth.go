package api

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rclancey/authenticator"
	"github.com/rclancey/httpserver/v2/auth"
	//_ "modernc.org/sqlite"
)

type UserSource struct {
	db *sqlx.DB
}

func NewUserSource(dbfn string) (*UserSource, error) {
	log.Println("opening user db:", dbfn)
	db, err := sqlx.Connect("sqlite3", dbfn)
	if err != nil {
		return nil, err
	}
	query := `SELECT COUNT(*) FROM users`;
	var n int
	err = db.QueryRow(query).Scan(&n)
	if err != nil {
		log.Println("error reading from users, initializing db", err)
		pwauth, err := authenticator.NewPasswordAuthenticator("super-sekrit-default-pword")
		if err != nil {
			log.Println("bad default password:", err)
			return nil, err
		}
		tx, err := db.Beginx()
		if err != nil {
			log.Println("can't start transaction:", err)
			return nil, err
		}
		query := `CREATE TABLE users (
			id INTEGER NOT NULL PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			email VARCHAR(255),
			phone VARCHAR(255),
			avatar VARCHAR(255),
			apple_id VARCHAR(255),
			github_id VARCHAR(255),
			google_id VARCHAR(255),
			amazon_id VARCHAR(255),
			facebook_id VARCHAR(255),
			twitter_id VARCHAR(255),
			linkedin_id VARCHAR(255),
			slack_id VARCHAR(255),
			bitbucket_id VARCHAR(255),
			date_added INTEGER NOT NULL,
			date_modified INTEGER NOT NULL,
			active BOOL NOT NULL DEFAULT 't',
			admin BOOL NOT NULL DEFAULT 'f',
			password_auth TEXT,
			twofactor_auth TEXT,
			tmp_twofactor_auth TEXT)`
		_, err = tx.Exec(query)
		if err != nil {
			log.Println("error creating users table:", err)
			tx.Rollback()
			return nil, err
		}
		now := Now()
		query = `INSERT INTO users (username, date_added, date_modified, password_auth, admin) VALUES(?, ?, ?, ?, ?)`
		_, err = tx.Exec(query, "admin", now, now, pwauth, true)
		if err != nil {
			log.Println("error inserting default user:", err)
			tx.Rollback()
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			log.Println("error committing transaction:", err)
			tx.Rollback()
			return nil, err
		}
	}
	return &UserSource{db: db}, nil
}

func (us *UserSource) GetUser(username string) (auth.AuthUser, error) {
	query := `SELECT * FROM users WHERE username = ?`
	user := &AuthUser{db: us.db}
	err := us.db.QueryRowx(query, username).StructScan(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (us *UserSource) GetUserByEmail(email string) (auth.AuthUser, error) {
	query := `SELECT * FROM users WHERE email = ?`
	user := &AuthUser{db: us.db}
	err := us.db.QueryRowx(query, email).StructScan(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

type Time struct {
	time.Time
}

func Now() *Time {
	return &Time{time.Now()}
}

func (t *Time) Value() (driver.Value, error) {
	return t.Unix(), nil
}

func (t *Time) Scan(value interface{}) error {
	var s int64
	switch x := value.(type) {
	case int:
		s = int64(x)
	case int64:
		s = x
	case int32:
		s = int64(x)
	case uint:
		s = int64(x)
	case uint64:
		s = int64(x)
	case uint32:
		s = int64(x)
	case float64:
		s = int64(x)
	case float32:
		s = int64(x)
	case string:
		var err error
		s, err = strconv.ParseInt(x, 10, 64)
		if err != nil {
			xt, err := time.ParseInLocation("2006-01-02 15:04:05", x, time.UTC)
			if err == nil {
				t.Time = xt
				return nil
			}
			return err
		}
	}
	t.Time = time.Unix(s, 0)
	return nil
}

type AuthUser struct {
	ID            int64   `json:"id,omitempty" db:"id,primary_key"`
	Username      string  `json:"username" db:"username"`
	FirstName     *string `json:"first_name" db:"first_name"`
	LastName      *string `json:"last_name" db:"last_name"`
	Email         *string `json:"email" db:"email"`
	Phone         *string `json:"phone" db:"phone"`
	Avatar        *string `json:"avatar" db:"avatar"`
    AppleID       *string `json:"apple_id,omitempty" db:"apple_id"`
    GitHubID      *string `json:"github_id,omitempty" db:"github_id"`
    GoogleID      *string `json:"google_id,omitempty" db:"google_id"`
    AmazonID      *string `json:"amazon_id,omitempty" db:"amazon_id"`
    FacebookID    *string `json:"facebook_id,omitempty" db:"facebook_id"`
    TwitterID     *string `json:"twitter_id,omitempty" db:"twitter_id"`
    LinkedInID    *string `json:"linkedin_id,omitempty" db:"linkedin_id"`
    SlackID       *string `json:"slack_id,omitempty" db:"slack_id"`
    BitBucketID   *string `json:"bitbucket_id,omitempty" db:"bitbucket_id"`
    DateAdded     *Time   `json:"date_added,omitempty" db:"date_added" dbignore:"update"`
    DateModified  *Time   `json:"date_modified,omitempty" db:"date_modified"`
    Active        bool    `json:"active,omitempty" db:"active"`
    IsAdmin       bool    `json:"admin,omitempty" db:"admin"`
    Password      *authenticator.PasswordAuthenticator  `json:"password_auth,omitempty" db:"password_auth"`
    TwoFactor     *authenticator.TwoFactorAuthenticator `json:"twofactor_auth,omitempty" db:"twofactor_auth"`
    TmpTwoFactor  *authenticator.TwoFactorAuthenticator `json:"tmp_twofactor_auth,omitempty" db:"tmp_twofactor_auth"`
	db            *sqlx.DB
}

func (au *AuthUser) GetUserID() int64 {
	return au.ID
}

func (au *AuthUser) GetUsername() string {
	return au.Username
}

func (au *AuthUser) GetAuth() (authenticator.Authenticator, error) {
	return au.Password, nil
}

func (au *AuthUser) SetAuth(authen authenticator.Authenticator) error {
	if au.db == nil {
		return errors.New("no database handle")
	}
	pwauth, isa := authen.(*authenticator.PasswordAuthenticator)
	if !isa {
		return errors.New("invalid authentication mechanism")
	}
	now := Now()
	query := `UPDATE users SET password_auth = ?, date_modified = ? WHERE username = ?`
	_, err := au.db.Exec(query, pwauth, now, au.Username)
	if err != nil {
		return err
	}
	au.Password = pwauth
	au.DateModified = now
	return nil
}

func (au *AuthUser) GetTwoFactorAuth() (authenticator.Authenticator, error) {
	if au.TwoFactor == nil {
		return nil, nil
	}
	return au.TwoFactor, nil
}

func (au *AuthUser) SetTwoFactorAuth(authen authenticator.Authenticator) error {
	if au.db == nil {
		return errors.New("no database handle")
	}
	tfa, isa := authen.(*authenticator.TwoFactorAuthenticator)
	if !isa {
		return errors.New("invalida authentication mechanism")
	}
	now := Now()
	query := `UPDATE users SET twofactor_auth = ?, date_modified = ? WHERE username = ?`
	_, err := au.db.Exec(query, tfa, now, au.Username)
	if err != nil {
		return err
	}
	au.TwoFactor = tfa
	au.DateModified = now
	return nil
}

func (au *AuthUser) InitTwoFactorAuth() (authenticator.Authenticator, error) {
	if au.db == nil {
		return nil, errors.New("no database handle")
	}
	tfa, err := authenticator.NewTwoFactorAuthenticator(au.Username, "")
	if err != nil {
		return nil, err
	}
	now := Now()
	query := `UPDATE users SET tmp_twofactor_auth = ?, date_modified = ? WHERE username = ?`
	_, err = au.db.Exec(query, tfa, now, au.Username)
	if err != nil {
		return nil, err
	}
	au.TmpTwoFactor = tfa
	au.DateModified = now
	return tfa, nil
}

func (au *AuthUser) CompleteTwoFactorAuth(code string) error {
	if au.db == nil {
		return errors.New("no database handle")
	}
	tfa := au.TmpTwoFactor
	if tfa == nil {
		return errors.New("no authenticator configured")
	}
	err := tfa.Authenticate(code)
	if err != nil {
		return err
	}
	now := Now()
	query := `UPDATE users SET twofactor_auth = ?, tmp_twofactor_auth = NULL, date_modified = ? WHERE username = ?`
	_, err = au.db.Exec(query, tfa, now, au.Username)
	if err != nil {
		return err
	}
	au.TwoFactor = tfa
	au.TmpTwoFactor = nil
	au.DateModified = now
	return nil
}

var socialColumns = map[string]string{
	"apple": "apple_id",
	"github": "github_id",
	"google": "google_id",
	"amazon": "amazon_id",
	"facebook": "facebook_id",
	"twitter": "twitter_id",
	"linkedin": "linkedin_id",
	"slack": "slack_id",
	"bitbucket": "bitbucket_id",
}

func (au *AuthUser) SetSocialID(driver, id string) error {
	if au.db == nil {
		return errors.New("no database handle")
	}
	col, ok := socialColumns[driver]
	if !ok {
		return errors.New("unknown social driver")
	}
	now := Now()
	query := fmt.Sprintf(`UPDATE users SET %s = ?, date_modified = ? WHERE username = ?`, col)
	_, err := au.db.Exec(query, id, now, au.Username)
	if err != nil {
		return err
	}
	au.Reload(au.db)
	return nil
}

func (au *AuthUser) Clean() *AuthUser {
	clone := *au
	clone.AppleID = nil
	clone.GitHubID = nil
	clone.AmazonID = nil
	clone.FacebookID = nil
	clone.TwitterID = nil
	clone.LinkedInID = nil
	clone.SlackID = nil
	clone.BitBucketID = nil
	clone.Password = nil
	clone.TwoFactor = nil
	clone.TmpTwoFactor = nil
	return &clone
}

func (au *AuthUser) Reload(db *sqlx.DB) error {
	au.db = db
	query := `SELECT * FROM users WHERE `
	if au.ID != 0 {
		query += `id = :id`
	} else if au.Username != "" {
		query += `username = :username`
	} else {
		return errors.New("unknown user")
	}
	stmt, err := db.PrepareNamed(query)
	if err != nil {
		return err
	}
	return stmt.QueryRow(au).StructScan(au)
}

func (au *AuthUser) GetFirstName() string {
	if au.FirstName == nil {
		return ""
	}
	return *au.FirstName
}

func (au *AuthUser) GetLastName() string {
	if au.LastName == nil {
		return ""
	}
	return *au.LastName
}

func (au *AuthUser) GetEmailAddress() string {
	if au.Email == nil {
		return ""
	}
	return *au.Email
}

func (au *AuthUser) GetPhoneNumber() string {
	if au.Phone == nil {
		return ""
	}
	return *au.Phone
}

func (au *AuthUser) GetAvatar() string {
	if au.Avatar == nil {
		return ""
	}
	return *au.Avatar
}

func (au *AuthUser) GetID() int64 {
	return au.ID
}

func (au *AuthUser) SetID(id int64) {
	au.ID = id
}

func (au *AuthUser) Table() string {
	return "users"
}

func (au *AuthUser) Clone() *AuthUser {
	clone := *au
	return &clone
}

func (au *AuthUser) Create() error {
	if au.db == nil {
		return errors.New("no database handle")
	}
	tx, err := au.db.Beginx()
	if err != nil {
		return err
	}
	clone := au.Clone()
	now := Now()
	clone.DateAdded = now
	clone.DateModified = now
	err = InsertStruct(tx, clone)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	*au = *clone
	return nil
}

func (au *AuthUser) Update() error {
	if au.db == nil {
		return errors.New("no database handle")
	}
	tx, err := au.db.Beginx()
	if err != nil {
		return err
	}
	clone := au.Clone()
	clone.DateModified = Now()
	err = UpdateStruct(tx, clone)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	*au = *clone
	return nil
}
