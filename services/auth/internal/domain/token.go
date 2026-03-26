package domain

type AccessClaims struct {
	UserID     int64  `json:"user_id"`
	Login      string `json:"login"`
	RoleID     int64  `json:"role_id"`
	BusinessID int64  `json:"business_id"`
	BranchID   *int64 `json:"branch_id,omitempty"`
}

type RefreshClaims struct {
	UserID int64  `json:"user_id"`
	JTI    string `json:"jti"`
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}