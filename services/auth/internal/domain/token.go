package domain

type RefreshClaims struct {
	UserID int64  `json:"user_id"`
	JTI    string `json:"jti"`
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}
