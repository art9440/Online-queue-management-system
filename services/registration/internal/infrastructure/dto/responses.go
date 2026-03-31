package dto

type Response struct {
	Status         string `json:"status"`
	RegistrationID string `json:"registration_id,omitempty"`
}
