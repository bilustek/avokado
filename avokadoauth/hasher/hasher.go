package hasher

import "github.com/bilustek/avokado/avokadoerror"

// MinPasswordLength is the minimum allowed password length for hashing.
const MinPasswordLength = 8

// ErrPasswordTooShort is returned when a password is shorter than MinPasswordLength.
var ErrPasswordTooShort = avokadoerror.New("[avokadoauth.passwordhasher]: password too short")

// PasswordHasher defines the interface for password hashing and comparison.
// Implementations can use different hashing algorithms (bcrypt, argon2, etc.).
type PasswordHasher interface {
	// Hash takes a plaintext password and returns its hash.
	Hash(password string) (string, error)
	// Compare compares a hashed password with a plaintext password.
	// Returns nil if they match, or an error otherwise.
	Compare(hashedPassword, password string) error
}
