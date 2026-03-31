package pending

type PendingRegistration struct {
	ID           string
	Email        string
	PasswordHash string
	BusinessName string
	BusinessType string
	Code         string
}
