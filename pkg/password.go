package pkg

import "golang.org/x/crypto/bcrypt"

func HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}
func ComparePassword(hash, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}
