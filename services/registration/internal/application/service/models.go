package service

type RegisterInput struct {
	Email        string
	Password     string
	BusinessName string
	BusinessType string
}

type RegisterOutput struct {
	Status         string
	RegistrationID string
}

type VerifyInput struct {
	RegistrationID string
	Code           string
}

type ResendInput struct {
	RegistrationID string
}

type PasswordRecoveryInput struct {
	Email string
}

type PasswordRecoveryOutput struct {
	Status     string
	RecoveryID string
}

type PasswordRecoveryVerifyInput struct {
	RecoveryID string
	Code       string
}
