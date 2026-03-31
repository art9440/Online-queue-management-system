package dto

type RegisterRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	BusinessName string `json:"business_name"`
	BusinessType string `json:"business_type"`
}

type VerifyRequest struct {
	RegistrationID string `json:"registration_id"`
	Code           string `json:"code"`
}
