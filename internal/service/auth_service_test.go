package service

import (
	"context"
	"errors"
	"testing"

	"github.com/iluha481/shieldgate/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	repo := &stubUserRepo{}
	svc := NewAuthService(repo, &stubJWT{})

	user, err := svc.Register(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if user.Email != "user@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "user@example.com")
	}
	if user.PasswordHash == "password123" || user.PasswordHash == "" {
		t.Error("PasswordHash was not hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
		t.Errorf("PasswordHash does not match the original password: %v", err)
	}
	if repo.created != user {
		t.Error("Register() did not pass the created user to the repository")
	}
}

func TestAuthService_Register_RepoError(t *testing.T) {
	repo := &stubUserRepo{createErr: domain.ErrUserAlreadyExists}
	svc := NewAuthService(repo, &stubJWT{})

	_, err := svc.Register(context.Background(), "user@example.com", "password123")
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("Register() error = %v, want ErrUserAlreadyExists", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() error = %v", err)
	}

	repo := &stubUserRepo{user: &domain.User{ID: 7, Email: "user@example.com", PasswordHash: string(hash)}}
	jwtStub := &stubJWT{token: "signed-token"}
	svc := NewAuthService(repo, jwtStub)

	token, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token != "signed-token" {
		t.Errorf("token = %q, want %q", token, "signed-token")
	}
	if jwtStub.gotID != "7" {
		t.Errorf("GenerateToken() userID = %q, want %q", jwtStub.gotID, "7")
	}
}

func TestAuthService_Login_Failures(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() error = %v", err)
	}
	validUser := &domain.User{ID: 1, PasswordHash: string(hash)}

	tests := []struct {
		name string
		repo *stubUserRepo
		jwt  *stubJWT
		pass string
	}{
		{
			name: "user not found",
			repo: &stubUserRepo{getByEmailErr: errors.New("db down")},
			jwt:  &stubJWT{},
			pass: "password123",
		},
		{
			name: "wrong password",
			repo: &stubUserRepo{user: validUser},
			jwt:  &stubJWT{},
			pass: "wrong-password",
		},
		{
			name: "token generation failed",
			repo: &stubUserRepo{user: validUser},
			jwt:  &stubJWT{genErr: errors.New("sign failed")},
			pass: "password123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthService(tt.repo, tt.jwt)

			token, err := svc.Login(context.Background(), "user@example.com", tt.pass)
			if err == nil {
				t.Fatal("Login() expected error")
			}
			if token != "" {
				t.Errorf("Login() token = %q, want empty", token)
			}
			if tt.name != "token generation failed" && !errors.Is(err, domain.ErrUserNotFound) {
				t.Errorf("Login() error = %v, want ErrUserNotFound", err)
			}
		})
	}
}

func TestAuthService_GetUser(t *testing.T) {
	want := &domain.User{ID: 3, Email: "user@example.com"}

	got, err := NewAuthService(&stubUserRepo{user: want}, &stubJWT{}).GetUser(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if got != want {
		t.Errorf("GetUser() = %+v, want %+v", got, want)
	}
}

func TestAuthService_GetUser_Error(t *testing.T) {
	repoErr := errors.New("db down")

	_, err := NewAuthService(&stubUserRepo{getByIDErr: repoErr}, &stubJWT{}).GetUser(context.Background(), 3)
	if !errors.Is(err, repoErr) {
		t.Fatalf("GetUser() error = %v, want %v", err, repoErr)
	}
}
