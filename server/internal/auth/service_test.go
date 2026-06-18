package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/azizbek12234/quizarena/server/internal/store"
)

// fakeUsers — DB'siz UserStore.
type fakeUsers struct{ byEmail map[string]store.User }

func newFake() *fakeUsers { return &fakeUsers{byEmail: map[string]store.User{}} }

func (f *fakeUsers) CreateGuest(context.Context) (store.User, error) {
	return store.User{ID: uuid.New(), IsGuest: true, Role: "user"}, nil
}
func (f *fakeUsers) CreateAccount(_ context.Context, a store.CreateAccountParams) (store.User, error) {
	if _, ok := f.byEmail[*a.Email]; ok {
		return store.User{}, &pgconn.PgError{Code: "23505"}
	}
	u := store.User{ID: uuid.New(), Username: a.Username, Email: a.Email, PasswordHash: a.PasswordHash, Role: "user"}
	f.byEmail[*a.Email] = u
	return u, nil
}
func (f *fakeUsers) GetUserByEmail(_ context.Context, email *string) (store.User, error) {
	u, ok := f.byEmail[*email]
	if !ok {
		return store.User{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeUsers) GetUserByTelegramID(context.Context, *int64) (store.User, error) {
	return store.User{}, pgx.ErrNoRows
}

func (f *fakeUsers) CreateTelegramUser(_ context.Context, id *int64) (store.User, error) {
	return store.User{ID: uuid.New(), TelegramID: id, Role: "user"}, nil
}

func newSvc() *Service {
	return NewService(newFake(), NewTokenManager("test-secret", time.Hour), "")
}

func TestRegisterLogin(t *testing.T) {
	s := newSvc()
	ctx := context.Background()

	res, err := s.Register(ctx, "Ali", "Ali@Example.com", "secret123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if res.Token == "" {
		t.Fatal("token bo'sh")
	}
	claims, err := s.Verify(res.Token)
	if err != nil || claims.Subject != res.User.ID.String() {
		t.Fatalf("verify: %v, sub=%v", err, claims)
	}

	// Email kichik harfga normallashishi kerak.
	if res.User.Email == nil || *res.User.Email != "ali@example.com" {
		t.Fatalf("email normalize xato: %v", res.User.Email)
	}

	// To'g'ri parol bilan kirish.
	if _, err := s.Login(ctx, "ali@example.com", "secret123"); err != nil {
		t.Fatalf("login: %v", err)
	}
	// Noto'g'ri parol.
	if _, err := s.Login(ctx, "ali@example.com", "wrong"); err != ErrInvalidCreds {
		t.Fatalf("ErrInvalidCreds kutilgan, keldi %v", err)
	}
	// Yo'q email.
	if _, err := s.Login(ctx, "yoq@example.com", "x"); err != ErrInvalidCreds {
		t.Fatalf("ErrInvalidCreds kutilgan, keldi %v", err)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	s := newSvc()
	ctx := context.Background()
	if _, err := s.Register(ctx, "A", "dup@example.com", "secret123"); err != nil {
		t.Fatalf("birinchi register: %v", err)
	}
	if _, err := s.Register(ctx, "B", "dup@example.com", "secret123"); err != ErrEmailTaken {
		t.Fatalf("ErrEmailTaken kutilgan, keldi %v", err)
	}
}

func TestGuest(t *testing.T) {
	s := newSvc()
	res, err := s.Guest(context.Background())
	if err != nil {
		t.Fatalf("guest: %v", err)
	}
	claims, err := s.Verify(res.Token)
	if err != nil || !claims.IsGuest {
		t.Fatalf("guest claim kutilgan: %v, %v", err, claims)
	}
}
