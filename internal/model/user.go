// Package model contains the core domain types of the smartContact
// application. This file replaces the JPA entity com.smartContact.model.User.
//
// MIGRATION_NOTE: The Java entity relied on Hibernate ddl-auto=update to
// create the `USER` table at startup and on Lombok to generate
// getters/setters/builder. In Go the struct fields are exported and accessed
// directly (no Lombok equivalent is needed), and schema creation is done
// explicitly by the store layer via UserTableDDL / EnsureSchema (see
// internal/store) instead of an ORM auto-creation step.
package model

import (
	"errors"
	"fmt"
	"strings"
)

// TableName is the physical table name the User entity maps to
// (@Table(name = "USER")). It is quoted because USER is a reserved word.
const TableName = "`USER`"

// Column names for the USER table, as declared by the @Column annotations
// on the Java entity. Kept here so the model owns its physical mapping and
// the store layer can build queries from these constants.
const (
	ColID       = "User_id"
	ColName     = "User_name"
	ColEmail    = "User_Email"
	ColPassword = "User_Password"
	ColRole     = "User_Role"
	ColAbout    = "User_About"
)

// UserTableDDL creates the USER table if it does not exist, mirroring the
// schema Hibernate would have generated from the entity:
//
//	id          INT AUTO_INCREMENT PRIMARY KEY  (@Id @GeneratedValue(AUTO))
//	User_name   VARCHAR(255) NOT NULL           (@NotBlank)
//	User_Email  VARCHAR(255) UNIQUE             (unique = true)
//	User_Password VARCHAR(255)
//	User_Role   VARCHAR(255)
//	User_About  VARCHAR(500)                    (length = 500)
//
// The store layer must execute this at startup in place of ddl-auto=update.
const UserTableDDL = "CREATE TABLE IF NOT EXISTS `USER` (" +
	"`User_id` INT AUTO_INCREMENT PRIMARY KEY, " +
	"`User_name` VARCHAR(255) NOT NULL, " +
	"`User_Email` VARCHAR(255) UNIQUE, " +
	"`User_Password` VARCHAR(255), " +
	"`User_Role` VARCHAR(255), " +
	"`User_About` VARCHAR(500)" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"

// User is the domain model for a contact-book user, replacing the JPA
// entity. ID == 0 means "not yet persisted" (the database assigns it via
// AUTO_INCREMENT, matching @GeneratedValue(strategy = AUTO)).
type User struct {
	ID       int    `json:"id"       db:"User_id"`
	Name     string `json:"name"     db:"User_name"`
	Email    string `json:"email"    db:"User_Email"`
	Password string `json:"password" db:"User_Password"`
	Role     string `json:"role"     db:"User_Role"`
	About    string `json:"about"    db:"User_About"`
}

// ErrNameRequired corresponds to the @NotBlank(message = "please Add the
// department Name") bean-validation constraint on User.name.
var ErrNameRequired = errors.New("please Add the department Name")

// Validate enforces the bean-validation constraints declared on the entity.
// It returns ErrNameRequired when Name is blank (whitespace-only counts as
// blank, as @NotBlank does). Callers (e.g. HTTP handlers) should map this to
// a 400 Bad Request response.
func (u User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return fmt.Errorf("user: %w", ErrNameRequired)
	}
	return nil
}

// String renders the user without ever exposing the password, which is the
// safe equivalent of the Lombok-generated toString on an entity holding a
// credential.
func (u User) String() string {
	return fmt.Sprintf("User{ID:%d Name:%q Email:%q Role:%q}", u.ID, u.Name, u.Email, u.Role)
}
