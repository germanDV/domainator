package services

import "golang.org/x/crypto/bcrypt"

type pwd struct {
	plain *string
	hash  []byte
}

// hashPwd calculates the bcrypt hash of a plaintext password
func hashPwd(plain string, cost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(plain), cost)
}

// Matches checks whether the provided plaintext password matches the
// hashed password stored in the struct.
func (p *pwd) Matches(candidate string) bool {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(candidate))
	if err != nil {
		return false
	}
	return true
}
