// Package store provides the MySQL-backed persistence implementation of
// service.UserStore, replacing the Spring Data JPA repository UserDao.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"migrated-app/internal/api"
	"migrated-app/internal/model"
)

// UserStore is the MySQL-backed implementation of service.UserStore.
type UserStore struct {
	db *sql.DB
}

// New opens a connection to the MySQL database at dsn and ensures the USER
// table exists (replacing Hibernate ddl-auto=update).
//
// A DSN assembled from multiple candidate hosts (comma-separated) is tried
// host by host: the first host that answers a ping wins. Each attempt is
// retried for a short window, because in containerized deployments the
// database container may still be starting (or its DNS entry may not be
// registered yet) when the application boots — a single immediate ping then
// fails with "no such host" and would otherwise crash the whole server.
func New(dsn string) (*UserStore, error) {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for {
		for _, candidate := range splitDSNHosts(dsn) {
			db, err := sql.Open("mysql", candidate)
			if err != nil {
				lastErr = fmt.Errorf("store: opening database: %w", err)
				continue
			}
			if err := db.Ping(); err != nil {
				_ = db.Close()
				lastErr = fmt.Errorf("store: pinging database: %w", err)
				continue
			}
			if _, err := db.Exec(model.UserTableDDL); err != nil {
				_ = db.Close()
				return nil, fmt.Errorf("store: ensuring schema: %w", err)
			}
			return &UserStore{db: db}, nil
		}
		if time.Now().After(deadline) {
			return nil, lastErr
		}
		time.Sleep(2 * time.Second)
	}
}

// splitDSNHosts expands a DSN whose tcp() target is a comma-separated host
// list into one DSN per host, preserving order. A DSN without a comma in the
// tcp target (e.g. an explicit DATABASE_URL) yields itself unchanged.
func splitDSNHosts(dsn string) []string {
	start := strings.Index(dsn, "tcp(")
	if start < 0 {
		return []string{dsn}
	}
	start += len("tcp(")
	end := strings.Index(dsn[start:], ")")
	if end < 0 {
		return []string{dsn}
	}
	end += start
	hosts := strings.Split(dsn[start:end], ",")
	if len(hosts) < 2 {
		return []string{dsn}
	}
	out := make([]string, 0, len(hosts))
	for _, h := range hosts {
		out = append(out, dsn[:start]+strings.TrimSpace(h)+dsn[end:])
	}
	return out
}

// Close closes the underlying database connection.
func (s *UserStore) Close() error { return s.db.Close() }

const userColumns = "`User_id`, `User_name`, `User_Email`, `User_Password`, `User_Role`, `User_About`"

func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.About)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Save inserts a user when ID is zero and updates it otherwise.
func (s *UserStore) Save(ctx context.Context, user *model.User) (*model.User, error) {
	if user.ID == 0 {
		res, err := s.db.ExecContext(ctx,
			"INSERT INTO `USER` (`User_name`, `User_Email`, `User_Password`, `User_Role`, `User_About`) VALUES (?, ?, ?, ?, ?)",
			user.Name, user.Email, user.Password, user.Role, user.About)
		if err != nil {
			return nil, fmt.Errorf("store: inserting user: %w", err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("store: reading inserted id: %w", err)
		}
		user.ID = int(id)
		return user, nil
	}
	_, err := s.db.ExecContext(ctx,
		"UPDATE `USER` SET `User_name` = ?, `User_Email` = ?, `User_Password` = ?, `User_Role` = ?, `User_About` = ? WHERE `User_id` = ?",
		user.Name, user.Email, user.Password, user.Role, user.About, user.ID)
	if err != nil {
		return nil, fmt.Errorf("store: updating user %d: %w", user.ID, err)
	}
	return user, nil
}

// FindByID returns the user with the given ID or api.ErrUserNotFound.
func (s *UserStore) FindByID(ctx context.Context, id int) (*model.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		"SELECT "+userColumns+" FROM `USER` WHERE `User_id` = ?", id))
	if err == sql.ErrNoRows {
		return nil, api.NewUserNotFoundError(fmt.Sprintf("user %d not found", id))
	}
	if err != nil {
		return nil, fmt.Errorf("store: finding user %d: %w", id, err)
	}
	return u, nil
}

// FindAll returns every user in the database.
func (s *UserStore) FindAll(ctx context.Context) ([]model.User, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT " + userColumns + " FROM `USER`")
	if err != nil {
		return nil, fmt.Errorf("store: listing users: %w", err)
	}
	defer rows.Close()
	users := []model.User{}
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.About); err != nil {
			return nil, fmt.Errorf("store: scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// DeleteByID removes the user with the given ID.
func (s *UserStore) DeleteByID(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM `USER` WHERE `User_id` = ?", id)
	if err != nil {
		return fmt.Errorf("store: deleting user %d: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return api.NewUserNotFoundError(fmt.Sprintf("user %d not found", id))
	}
	return nil
}

// FindByName returns the user whose name matches exactly, or
// api.ErrUserNotFound when absent.
func (s *UserStore) FindByName(ctx context.Context, name string) (*model.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		"SELECT "+userColumns+" FROM `USER` WHERE `User_name` = ?", name))
	if err == sql.ErrNoRows {
		return nil, api.NewUserNotFoundError(fmt.Sprintf("user %q not found", name))
	}
	if err != nil {
		return nil, fmt.Errorf("store: finding user by name %q: %w", name, err)
	}
	return u, nil
}
