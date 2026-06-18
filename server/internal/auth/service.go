package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/azizbek12234/quizarena/server/internal/store"
)

var (
	ErrEmailTaken   = errors.New("email allaqachon ro'yxatdan o'tgan")
	ErrInvalidCreds = errors.New("email yoki parol noto'g'ri")
)

// UserStore — auth uchun kerakli DB amallari (sqlc *store.Queries buni qondiradi).
// Interfeys orqali — auth DB'siz testlanadi.
type UserStore interface {
	CreateGuest(ctx context.Context) (store.User, error)
	CreateAccount(ctx context.Context, arg store.CreateAccountParams) (store.User, error)
	GetUserByEmail(ctx context.Context, email *string) (store.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID *int64) (store.User, error)
	CreateTelegramUser(ctx context.Context, telegramID *int64) (store.User, error)
}

type Service struct {
	users    UserStore
	tokens   *TokenManager
	botToken string // Telegram initData tekshiruvi uchun ("" bo'lsa telegram o'chiq)
}

func NewService(users UserStore, tokens *TokenManager, botToken string) *Service {
	return &Service{users: users, tokens: tokens, botToken: botToken}
}

// Result — foydalanuvchi + sessiya tokeni.
type Result struct {
	User  store.User
	Token string
}

func (s *Service) Guest(ctx context.Context) (Result, error) {
	u, err := s.users.CreateGuest(ctx)
	if err != nil {
		return Result{}, err
	}
	return s.withToken(u)
}

func (s *Service) Register(ctx context.Context, username, email, password string) (Result, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Result{}, err
	}
	uname := strings.TrimSpace(username)
	em := normalizeEmail(email)
	hp := string(hash)
	u, err := s.users.CreateAccount(ctx, store.CreateAccountParams{
		Username: &uname, Email: &em, PasswordHash: &hp,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Result{}, ErrEmailTaken
		}
		return Result{}, err
	}
	return s.withToken(u)
}

func (s *Service) Login(ctx context.Context, email, password string) (Result, error) {
	em := normalizeEmail(email)
	u, err := s.users.GetUserByEmail(ctx, &em)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Result{}, ErrInvalidCreds
		}
		return Result{}, err
	}
	if u.PasswordHash == nil {
		return Result{}, ErrInvalidCreds
	}
	if bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)) != nil {
		return Result{}, ErrInvalidCreds
	}
	return s.withToken(u)
}

// Telegram — Mini App initData'sini tekshirib, telegram_id bo'yicha user topadi/yaratadi.
func (s *Service) Telegram(ctx context.Context, initData string) (Result, error) {
	tg, err := ValidateInitData(initData, s.botToken)
	if err != nil {
		return Result{}, err
	}
	u, err := s.users.GetUserByTelegramID(ctx, &tg.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		u, err = s.users.CreateTelegramUser(ctx, &tg.ID)
		if err != nil {
			return Result{}, err
		}
	} else if err != nil {
		return Result{}, err
	}
	return s.withToken(u)
}

// Verify — WS/REST middleware uchun token tekshirish.
func (s *Service) Verify(token string) (*Claims, error) { return s.tokens.Verify(token) }

func (s *Service) withToken(u store.User) (Result, error) {
	tok, err := s.tokens.Issue(u.ID.String(), u.Role, u.IsGuest)
	if err != nil {
		return Result{}, err
	}
	return Result{User: u, Token: tok}, nil
}

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
