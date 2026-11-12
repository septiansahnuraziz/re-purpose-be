package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/repurpose-ai/re-purpose-be/internal/entity"
)

func TestMemoryUserRepositoryAuthMethods(t *testing.T) {
	_, users, _ := NewMemoryRepositories()
	ctx := context.Background()

	user := &entity.User{
		ID:           "usr_test",
		Email:        "Test@Example.com",
		Name:         "Test",
		PasswordHash: "hash",
		AuthProvider: "password",
		Tier:         "FREE",
	}
	if err := users.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, err := users.FindByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}
	if found == nil || found.ID != user.ID {
		t.Fatalf("find by email = %+v, want user %q", found, user.ID)
	}

	if found, err := users.FindByGoogleID(ctx, ""); err != nil || found != nil {
		t.Fatalf("find by empty google id = (%+v, %v), want (nil, nil)", found, err)
	}

	duplicate := &entity.User{ID: "usr_dup", Email: "test@example.com"}
	if err := users.Create(ctx, duplicate); !errors.Is(err, entity.ErrDuplicateEmail) {
		t.Fatalf("duplicate email error = %v, want entity.ErrDuplicateEmail", err)
	}

	user.GoogleID = "google-1"
	user.AuthProvider = "both"
	if err := users.Update(ctx, user); err != nil {
		t.Fatalf("update user: %v", err)
	}

	linked, err := users.FindByGoogleID(ctx, "google-1")
	if err != nil {
		t.Fatalf("find by google id: %v", err)
	}
	if linked == nil || linked.ID != user.ID {
		t.Fatalf("find by google id = %+v, want user %q", linked, user.ID)
	}

	other := &entity.User{ID: "usr_other", Email: "other@example.com", GoogleID: "google-1"}
	if err := users.Create(ctx, other); !errors.Is(err, entity.ErrDuplicateGoogleID) {
		t.Fatalf("duplicate google id error = %v, want entity.ErrDuplicateGoogleID", err)
	}

	if err := users.Update(ctx, &entity.User{ID: "usr_missing"}); !errors.Is(err, entity.ErrNotFound) {
		t.Fatalf("update missing user error = %v, want entity.ErrNotFound", err)
	}
}

func TestMemorySessionRepository(t *testing.T) {
	sessions := NewMemorySessionRepository()
	ctx := context.Background()
	now := time.Now()

	session := &entity.AuthSession{
		ID:               "ses_1",
		UserID:           "usr_1",
		RefreshTokenHash: "hash-1",
		ExpiresAt:        now.Add(time.Hour),
	}
	if err := sessions.Create(ctx, session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	found, err := sessions.FindByRefreshTokenHash(ctx, "hash-1")
	if err != nil {
		t.Fatalf("find by refresh hash: %v", err)
	}
	if found == nil || found.ID != session.ID {
		t.Fatalf("find by refresh hash = %+v, want session %q", found, session.ID)
	}

	if !found.IsActive(now) {
		t.Fatal("new session should be active")
	}

	duplicate := &entity.AuthSession{
		ID:               "ses_dup",
		UserID:           "usr_1",
		RefreshTokenHash: "hash-1",
		ExpiresAt:        now.Add(time.Hour),
	}
	if err := sessions.Create(ctx, duplicate); !errors.Is(err, entity.ErrDuplicateRefresh) {
		t.Fatalf("duplicate refresh error = %v, want entity.ErrDuplicateRefresh", err)
	}

	if err := sessions.Revoke(ctx, session.ID, now); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	revoked, _ := sessions.FindByID(ctx, session.ID)
	if revoked == nil || revoked.IsActive(now) {
		t.Fatal("session should be revoked")
	}

	if err := sessions.Revoke(ctx, "ses_missing", now); err != nil {
		t.Fatalf("revoke missing session should be idempotent, got %v", err)
	}

	second := &entity.AuthSession{
		ID:               "ses_2",
		UserID:           "usr_1",
		RefreshTokenHash: "hash-2",
		ExpiresAt:        now.Add(time.Hour),
	}
	if err := sessions.Create(ctx, second); err != nil {
		t.Fatalf("create second session: %v", err)
	}
	if err := sessions.RevokeByUserID(ctx, "usr_1", now); err != nil {
		t.Fatalf("revoke by user: %v", err)
	}
	second, _ = sessions.FindByID(ctx, second.ID)
	if second == nil || second.IsActive(now) {
		t.Fatal("second session should be revoked by user id")
	}

	expired := &entity.AuthSession{
		ID:               "ses_expired",
		UserID:           "usr_2",
		RefreshTokenHash: "hash-expired",
		ExpiresAt:        now.Add(-time.Hour),
	}
	if err := sessions.Create(ctx, expired); err != nil {
		t.Fatalf("create expired session: %v", err)
	}
	if err := sessions.DeleteExpired(ctx, now); err != nil {
		t.Fatalf("delete expired sessions: %v", err)
	}
	if got, _ := sessions.FindByID(ctx, expired.ID); got != nil {
		t.Fatal("expired session should be deleted")
	}
}
