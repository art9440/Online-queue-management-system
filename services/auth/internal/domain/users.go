package domain

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	RoleID       int64
	BusinessID   int64
	BranchID     *int64
}