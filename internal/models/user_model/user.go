package user_model

import "github.com/jackc/pgx/v5/pgtype"

type User struct {
	Id       pgtype.UUID
	Email    string
	Password []byte
	Role     UserRole
}

type UserRole string

const (
	Employee  UserRole = "employee"
	Moderator UserRole = "moderator"
)
