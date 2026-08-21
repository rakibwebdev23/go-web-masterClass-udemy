package main

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(name, email, hashedPassword, avatar string) (int64, error)
	GetUserByEmail(email string) (*User, error)
	GetUsers() ([]User, error)
}

type SQLUserRepository struct {
	db *sql.DB // Assuming you have a database connection here
}

func NewSQLUserRepository(db *sql.DB) UserRepository {
	return &SQLUserRepository{
		db: db,
	}
}

func (r *SQLUserRepository) CreateUser(name, email, hashedPassword, avatar string) (int64, error) {
	ctx := context.Background()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO users (name, email, hashed_password) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	hp, err := bcrypt.GenerateFromPassword([]byte(hashedPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(name, email, string(hp))
	if err != nil {
		return 0, err
	}

	userID, err := res.LastInsertId();
	if err != nil {
		return 0, err
	}

	profileStmt, err := tx.PrepareContext(ctx, "INSERT INTO profiles (user_id, avatar_url) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer profileStmt.Close();

	_, err = profileStmt.Exec(userID, avatar)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *SQLUserRepository) GetUserByEmail(email string) (*User, error) {
	stmt := "SELECT u.id, u.name, u.email, u.hashed_password, u.created_at, p.id, p.avatar_url, p.created_at FROM users u LEFT JOIN profiles p ON u.id = p.user_id WHERE u.email = ?"
	row := r.db.QueryRow(stmt, email)

	var user User

	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword, &user.CreatedAt, &user.Profile.ID, &user.Profile.AvatarURL, &user.Profile.CreatedAt);
	if err != nil {
		return nil, err
	}

	user.Profile.UserId = user.ID
	return &user, nil
}

func (r *SQLUserRepository) GetUsers() ([]User, error) {
	stmt := "SELECT u.id, u.name, u.email, u.hashed_password, u.created_at, p.id, p.avatar_url, p.created_at FROM users u LEFT JOIN profiles p ON u.id = p.user_id"
	rows, err := r.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	
	for rows.Next() {
		var user User
		var profile Profile

		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword, &user.CreatedAt, &profile.ID, &profile.AvatarURL, &profile.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}	