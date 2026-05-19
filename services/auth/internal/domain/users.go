package domain

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	RoleID       int64
	RoleName     string
	BusinessID   int64
	BranchID     *int64
}
