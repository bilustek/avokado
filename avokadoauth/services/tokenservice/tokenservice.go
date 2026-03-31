package tokenservice

import (
	"time"

	"github.com/bilustek/avokado/avokadoauth"
	"github.com/bilustek/avokado/avokadoerror"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	defaultAccessTokenTTL       = 15 * time.Minute
	defaultRefreshTokenTTL      = 30 * 24 * time.Hour // 30 days
	defaultEmailVerificationTTL = 24 * time.Hour
	defaultPasswordResetTTL     = 1 * time.Hour
	refreshTokenBytes           = 32
)

// BaseClaims contains the minimal JWT claims for avokado access tokens.
// Developers can embed BaseClaims in their own struct to add custom claims.
//
// RegisteredClaims is embedded as a value (NOT pointer) to avoid nil dereference
// panics during token parsing with golang-jwt/v5.
type BaseClaims struct {
	jwt.RegisteredClaims

	Email       string `json:"email"`
	IsStaff     bool   `json:"is_staff"`
	IsSuperuser bool   `json:"is_superuser"`
}

// PurposeClaims is used for single-purpose tokens (email verification, password reset).
// The Purpose field distinguishes between token types: "email-verify" or "password-reset".
type PurposeClaims struct {
	jwt.RegisteredClaims

	Purpose string `json:"purpose"`
}

// Servicer defines token service behaviour.
type Servicer interface {
	CreateAccessToken(user avokadoauth.UserModelizer) (string, error)
	ParseAccessToken(tokenString string) (*BaseClaims, error)
	CreatePurposeToken(userUID uuid.UUID, purpose string) (string, error)
	ParsePurposeToken(tokenString string, expectedPurpose string) (*PurposeClaims, error)
	GenerateRefreshToken() (plain string, hashed string, err error)
	HashToken(plain string) string
	PurposeTTL(purpose string) time.Duration
}

// Service handles JWT signing/parsing and refresh token generation.
type Service struct {
	signingKey           []byte
	accessTokenTTL       time.Duration
	refreshTokenTTL      time.Duration
	emailVerificationTTL time.Duration
	passwordResetTTL     time.Duration
}

// CreateAccessToken creates a signed HS256 JWT access token for the given user.
// Claims include sub (user UID), email, is_staff, is_superuser, jti, exp, and iat.
func (s *Service) CreateAccessToken(user avokadoauth.UserModelizer) (string, error) {
	now := time.Now()

	claims := &BaseClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.GetPublicID().String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		Email:       user.GetEmail(),
		IsStaff:     user.GetIsStaff(),
		IsSuperuser: user.GetIsSuperuser(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.signingKey)
	if err != nil {
		return "", avokadoerror.New("[tokenservice.CreateAccessToken] err").WithErr(err)
	}

	return signed, nil
}

// Option is a functional option for configuring Service.
type Option func(*Service)

// WithAccessTokenTTL sets the access token time-to-live (default 15min).
func WithAccessTokenTTL(d time.Duration) Option {
	return func(s *Service) {
		s.accessTokenTTL = d
	}
}

// WithRefreshTokenTTL sets the refresh token time-to-live (default 30 days).
func WithRefreshTokenTTL(d time.Duration) Option {
	return func(s *Service) {
		s.refreshTokenTTL = d
	}
}

// WithEmailVerificationTTL sets the email verification token time-to-live (default 24h).
func WithEmailVerificationTTL(d time.Duration) Option {
	return func(s *Service) {
		s.emailVerificationTTL = d
	}
}

// WithPasswordResetTTL sets the password reset token time-to-live (default 1h).
func WithPasswordResetTTL(d time.Duration) Option {
	return func(s *Service) {
		s.passwordResetTTL = d
	}
}

// New creates a TokenService with the given signing key and options.
// Default TTLs: access=15min, refresh=30days, emailVerify=24h, passwordReset=1h.
func New(signingKey []byte, opts ...Option) *Service {
	ts := &Service{
		signingKey:           signingKey,
		accessTokenTTL:       defaultAccessTokenTTL,
		refreshTokenTTL:      defaultRefreshTokenTTL,
		emailVerificationTTL: defaultEmailVerificationTTL,
		passwordResetTTL:     defaultPasswordResetTTL,
	}

	for _, opt := range opts {
		opt(ts)
	}

	return ts
}
