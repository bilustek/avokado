package avokadoauth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/bilustek/avokado/avokadoauth/hasher"
	"github.com/bilustek/avokado/avokadodb"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	_ avokadodb.DBModelizer = (*User)(nil)
	_ avokadodb.DBModelizer = (*BaseUser)(nil)
)

const csrfTokenBytes = 32

// Service provides core authentication operations: login, register,
// logout, token refresh, and Google OAuth.
type Service struct {
	db          *gorm.DB                    //nolint:unused // will implement soon
	hasher      hasher.PasswordHasher       //nolint:unused // will implement soon
	emailSender avokadonotifier.EmailSender //nolint:unused // will implement soon
	// tokenService    *TokenService
	// googleVerifier  GoogleVerifier
	frontendBaseURL string //nolint:unused // will implement soon
	cookieDomain    string //nolint:unused // will implement soon
}

// Option is a functional option for configuring AuthService.
type Option func(*Service)

// TokenPair holds the access token, refresh token, and CSRF token returned
// by authentication operations (login, register, refresh).
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	CSRFToken    string `json:"csrf_token"`
}

// BaseUser provides common authentication fields for user models.
type BaseUser struct {
	avokadodb.BaseModel

	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	IsStaff       bool       `json:"is_staff"`
	IsSuperuser   bool       `json:"is_superuser"`
	EmailVerified bool       `json:"email_verified"`
	LastLogin     *time.Time `json:"last_login"`
	Provider      string     `json:"provider"`
	ProviderID    *string    `json:"provider_id"`
}

// GetPublicID returns the public UUID for the user, used for API responses.
func (b BaseUser) GetPublicID() uuid.UUID { return b.UID }

// TableName returns the default table name for BaseUser.
// Custom types embedding BaseUser should override this method.
func (BaseUser) TableName() string { return "users" }

// SetPasswordHash sets the user's hashed password.
func (b *BaseUser) SetPasswordHash(hash string) { b.PasswordHash = hash }

// SetEmailVerified sets the user's email verification status.
func (b *BaseUser) SetEmailVerified(verified bool) { b.EmailVerified = verified }

// SetLastLogin sets the user's last login time.
func (b *BaseUser) SetLastLogin(t time.Time) { b.LastLogin = &t }

// UserModelizer extends avokadodb.DBModelizer with authentication-specific methods.
// Both the default User and custom user types (embedding BaseUser) satisfy this interface.
type UserModelizer interface {
	avokadodb.DBModelizer

	GetEmail() string
	GetPasswordHash() string
	SetPasswordHash(hash string)
	GetFirstName() string
	GetLastName() string
	GetIsStaff() bool
	GetIsSuperuser() bool
	GetIsActive() bool
	GetEmailVerified() bool
	SetEmailVerified(verified bool)
	SetLastLogin(t time.Time)
	GetProvider() string
	GetProviderID() *string
}

// User represents the user model for authentication purposes.
type User struct {
	BaseUser
}

// TableName returns the database table name for the default User model.
func (User) TableName() string { return "users" }

// GetPublicID returns the public UUID for the user, used for API responses.
func (u User) GetPublicID() uuid.UUID { return u.UID }

// AuthServicer ...
type AuthServicer interface {
	Register(ctx context.Context, email, password, firstName, lastName string) (*TokenPair, *User, error)
}

// GenerateCSRFToken generates a cryptographically secure CSRF token.
func GenerateCSRFToken() (string, error) {
	b := make([]byte, csrfTokenBytes)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// FindActiveUserByEmail looks up an active user by normalized email address.
func FindActiveUserByEmail(db *gorm.DB, email string) (*User, error) {
	var user User

	err := db.Where("email = @email AND is_active = @active", sql.Named("email", email), sql.Named("active", true)).
		First(&user).
		Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
