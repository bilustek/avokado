package bcrypthasher

import (
	"fmt"

	"github.com/bilustek/avokado/avokadoauth/hasher"
	"github.com/bilustek/avokado/avokadoerror"
	"golang.org/x/crypto/bcrypt"
)

var _ hasher.PasswordHasher = (*Hasher)(nil)

// Hasher implements PasswordHasher using bcrypt.
type Hasher struct {
	cost int
}

// Hash generates a bcrypt hash of the given password.
// Returns ErrPasswordTooShort if the password is shorter than MinPasswordLength.
func (h *Hasher) Hash(password string) (string, error) {
	if len(password) < hasher.MinPasswordLength {
		return "", fmt.Errorf(
			"%w: minimum %d characters required",
			hasher.ErrPasswordTooShort,
			hasher.MinPasswordLength,
		)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", avokadoerror.New("[bcrypthasher.Hash] err").WithErr(err)
	}

	return string(hashed), nil
}

// Compare checks whether the given plaintext password matches the hashed password.
// Returns nil if they match, or bcrypt.ErrMismatchedHashAndPassword if they don't.
func (*Hasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Option is a functional option for configuring Hasher.
type Option func(*Hasher)

// WithCost sets the bcrypt cost parameter.
// The cost must be between bcrypt.MinCost (4) and bcrypt.MaxCost (31).
func WithCost(cost int) Option {
	return func(h *Hasher) {
		h.cost = cost
	}
}

// New creates a new Hasher (BcryptHasher) with the given options.
// Default cost is bcrypt.DefaultCost (10).
func New(opts ...Option) *Hasher {
	h := &Hasher{
		cost: bcrypt.DefaultCost,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}
