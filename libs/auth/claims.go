package auth

type AccessClaims struct {
	UserID     int64  `json:"user_id"`
	Login      string `json:"login"`
	RoleID     int64  `json:"role_id"`
	RoleName   string `json:"role_name,omitempty"`
	BusinessID int64  `json:"business_id"`
	BranchID   *int64 `json:"branch_id,omitempty"`
}
