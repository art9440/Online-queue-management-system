package pending

type PendingRegistration struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"password_hash"`
	BusinessName string  `json:"business_name"`
	BusinessType string  `json:"business_type"`
	Code         string  `json:"code"`
	ClientLink   *string `json:"client_link,omitempty"`
}
