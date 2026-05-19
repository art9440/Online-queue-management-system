package dto

type Response struct {
	Status         string `json:"status"`
	RegistrationID string `json:"registration_id,omitempty"`
	RecoveryID     string `json:"recovery_id,omitempty"`
}
