package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	GetAccounts() ([]*Account, error)
	GetAccount(int) (*Account, error)
	GetAccountByFirstName(string) (*Account, error)
}

type PostgresStore struct {
	db    *sql.DB
	mutex sync.RWMutex
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=gobank host=go_bank_postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Configure the connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS account (
		id SERIAL PRIMARY KEY,
		first_name VARCHAR(50),
		last_name VARCHAR(50),
		number SERIAL,
		balance BIGINT,
		created_at TIMESTAMP
	)`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *Account) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	query := `INSERT INTO account (first_name, last_name, number, balance, created_at) 
              VALUES ($1, $2, $3, $4, $5)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, acc.FirstName, acc.LastName, acc.Number, acc.Balance, acc.CreatedAt)
	return err
}

func (s *PostgresStore) DeleteAccount(id int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	query := `DELETE FROM account WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account with ID %d not found", id)
	}

	return nil
}

func (s *PostgresStore) GetAccount(id int) (*Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	query := `SELECT id, first_name, last_name, number, balance, created_at 
              FROM account WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	account := &Account{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID, &account.FirstName, &account.LastName,
		&account.Number, &account.Balance, &account.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account with ID %d not found", id)
		}
		return nil, err
	}

	return account, nil
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	query := `SELECT id, first_name, last_name, number, balance, created_at FROM account`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}
	for rows.Next() {
		account := new(Account)
		err := rows.Scan(
			&account.ID, &account.FirstName, &account.LastName,
			&account.Number, &account.Balance, &account.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (s *PostgresStore) GetAccountByFirstName(firstName string) (*Account, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	query := `SELECT id, first_name, last_name, number, balance, created_at 
              FROM account WHERE first_name = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	account := &Account{}
	err := s.db.QueryRowContext(ctx, query, firstName).Scan(
		&account.ID, &account.FirstName, &account.LastName,
		&account.Number, &account.Balance, &account.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no account found with first name: %s", firstName)
		}
		return nil, err
	}

	return account, nil
}