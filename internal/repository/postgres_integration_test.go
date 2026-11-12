package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// postgresTestDB connects to the database from TEST_POSTGRES_DSN. The test is
// skipped when the env var is unset, so `go test ./...` stays dependency-free.
func postgresTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not set; skipping PostgreSQL integration test")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func TestPostgresUserRepositoryAuthMethods(t *testing.T) {
	db := postgresTestDB(t)
	users := NewUserRepository(db)
	ctx := context.Background()
	suffix := uuid.NewString()[:8]

	user := &entity.User{
		ID:             "usr_it_" + suffix,
		Email:          "it-" + suffix + "@example.com",
		Name:           "Integration Test",
		AuthProvider:   "password",
		PasswordHash:   "hash",
		Tier:           "FREE",
		DailyCredits:   10,
		MonthlyCredits: 100,
	}
	if err := users.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _ = db.Exec("DELETE FROM users WHERE id = ?", user.ID).Error })

	duplicateEmail := &entity.User{
		ID:    "usr_it_dup_" + suffix,
		Email: user.Email,
		Name:  "Duplicate",
	}
	if err := users.Create(ctx, duplicateEmail); !errors.Is(err, entity.ErrDuplicateEmail) {
		t.Fatalf("duplicate email error = %v, want entity.ErrDuplicateEmail", err)
	}

	user.GoogleID = "google-" + suffix
	user.AuthProvider = "both"
	if err := users.Update(ctx, user); err != nil {
		t.Fatalf("update user: %v", err)
	}

	linked, err := users.FindByGoogleID(ctx, user.GoogleID)
	if err != nil {
		t.Fatalf("find by google id: %v", err)
	}
	if linked == nil || linked.ID != user.ID {
		t.Fatalf("find by google id = %+v, want user %q", linked, user.ID)
	}

	duplicateGoogle := &entity.User{
		ID:       "usr_it_dup_google_" + suffix,
		Email:    "it-dup-google-" + suffix + "@example.com",
		Name:     "Duplicate Google",
		GoogleID: user.GoogleID,
	}
	if err := users.Create(ctx, duplicateGoogle); !errors.Is(err, entity.ErrDuplicateGoogleID) {
		t.Fatalf("duplicate google id error = %v, want entity.ErrDuplicateGoogleID", err)
	}
}

func TestPostgresSessionRepository(t *testing.T) {
	db := postgresTestDB(t)
	sessions := NewSessionRepository(db)
	ctx := context.Background()
	now := time.Now()
	suffix := uuid.NewString()[:8]

	session := &entity.AuthSession{
		ID:               "ses_it_" + suffix,
		UserID:           "usr_it_" + suffix,
		RefreshTokenHash: "hash-" + suffix,
		ExpiresAt:        now.Add(time.Hour),
	}
	if err := sessions.Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}
	t.Cleanup(func() { _ = db.Exec("DELETE FROM auth_sessions WHERE id = ?", session.ID).Error })

	found, err := sessions.FindByRefreshTokenHash(ctx, session.RefreshTokenHash)
	if err != nil {
		t.Fatalf("find by refresh hash: %v", err)
	}
	if found == nil || found.ID != session.ID {
		t.Fatalf("find by refresh hash = %+v, want session %q", found, session.ID)
	}

	duplicate := &entity.AuthSession{
		ID:               "ses_it_dup_" + suffix,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		ExpiresAt:        now.Add(time.Hour),
	}
	if err := sessions.Create(ctx, duplicate); !errors.Is(err, entity.ErrDuplicateRefresh) {
		t.Fatalf("duplicate refresh error = %v, want entity.ErrDuplicateRefresh", err)
	}

	if err := sessions.Revoke(ctx, session.ID, now); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	revoked, err := sessions.FindByID(ctx, session.ID)
	if err != nil {
		t.Fatalf("find revoked session: %v", err)
	}
	if revoked == nil || revoked.IsActive(now) {
		t.Fatal("session should be revoked")
	}
}
