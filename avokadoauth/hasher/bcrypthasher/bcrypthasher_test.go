package bcrypthasher_test

import (
	"errors"
	"testing"

	"github.com/bilustek/avokado/avokadoauth/hasher"
	"github.com/bilustek/avokado/avokadoauth/hasher/bcrypthasher"
)

func TestBcryptHasher_HashAndCompare(t *testing.T) {
	h := bcrypthasher.New(bcrypthasher.WithCost(4))

	password := "securepassword123"

	hashed, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	if hashed == "" {
		t.Fatal("expected non-empty hash")
	}

	if hashed == password {
		t.Error("hash should not equal plaintext password")
	}

	if err := h.Compare(hashed, password); err != nil {
		t.Errorf("Compare should return nil for correct password, got: %v", err)
	}
}

func TestBcryptHasher_WrongPassword(t *testing.T) {
	h := bcrypthasher.New(bcrypthasher.WithCost(4))

	hashed, err := h.Hash("correctpassword1")
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	if err := h.Compare(hashed, "wrongpassword11"); err == nil {
		t.Error("Compare should return error for wrong password")
	}
}

func TestBcryptHasher_CustomCost(t *testing.T) {
	h := bcrypthasher.New(bcrypthasher.WithCost(4))

	hashed, err := h.Hash("testpassword123")
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	if hashed == "" {
		t.Fatal("expected non-empty hash with custom cost")
	}

	if err := h.Compare(hashed, "testpassword123"); err != nil {
		t.Errorf("Compare should return nil, got: %v", err)
	}
}

func TestBcryptHasher_TooShortPassword(t *testing.T) {
	h := bcrypthasher.New(bcrypthasher.WithCost(4))

	_, err := h.Hash("short")
	if err == nil {
		t.Fatal("expected error for password shorter than MinPasswordLength")
	}
	if !errors.Is(err, hasher.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got: %v", err)
	}
}

func TestBcryptHasher_ExactMinLength(t *testing.T) {
	h := bcrypthasher.New(bcrypthasher.WithCost(4))

	// Exactly MinPasswordLength characters should succeed.
	password := "12345678" // 8 chars

	_, err := h.Hash(password)
	if err != nil {
		t.Fatalf("expected no error for password of exactly MinPasswordLength, got: %v", err)
	}
}

func TestBcryptHasher_EmptyPassword(t *testing.T) {
	h := bcrypthasher.New(bcrypthasher.WithCost(4))

	_, err := h.Hash("")
	if err == nil {
		t.Fatal("expected error for empty password")
	}

	if !errors.Is(err, hasher.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got: %v", err)
	}
}

func TestBcryptHasher_DefaultCost(t *testing.T) {
	h := bcrypthasher.New()

	// Default cost should work (bcrypt.DefaultCost = 10).
	hashed, err := h.Hash("defaultcosttest")
	if err != nil {
		t.Fatalf("Hash with default cost failed: %v", err)
	}

	if err := h.Compare(hashed, "defaultcosttest"); err != nil {
		t.Errorf("Compare with default cost failed: %v", err)
	}
}

func TestBcryptHasher_BcryptError(t *testing.T) {
	h := bcrypthasher.New(bcrypthasher.WithCost(100))
	if _, err := h.Hash("validpasswd"); err == nil {
		t.Errorf("expected avokadoerror, got: %v", err)
	}
}

func TestPasswordHasher_InterfaceCompliance(t *testing.T) {
	// Mock hasher satisfies PasswordHasher interface.
	type mockHasher struct{}

	mockHasherHash := func(_ string) (string, error) {
		return "mocked-hash", nil
	}
	mockHasherCompare := func(_, _ string) error {
		return nil
	}

	// Use a wrapper to satisfy the interface.
	_ = mockHasherHash
	_ = mockHasherCompare

	// Verify through a function accepting the interface.
	var ph hasher.PasswordHasher = bcrypthasher.New(bcrypthasher.WithCost(4))
	if ph == nil {
		t.Fatal("expected non-nil PasswordHasher")
	}
}
