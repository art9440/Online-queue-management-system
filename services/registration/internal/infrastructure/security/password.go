package security

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost, // обычно 10–14
	)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}
