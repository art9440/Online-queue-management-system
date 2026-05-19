package recovery

type PasswordRecovery struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Code  string `json:"code"`
}
