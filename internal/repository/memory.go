package repository

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/repurpose-ai/re-purpose-be/commons"
	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
)

// NewMemoryRepositories returns in-memory repositories so the API can run
// without PostgreSQL during local scaffolding/demo. Data is lost on restart.
func NewMemoryRepositories() (entity.IGenerationRepository, entity.IUserRepository, entity.ISponsorClickRepository) {
	now := time.Now()
	demoUserID := config.DemoUserID()

	users := map[string]*entity.User{
		demoUserID: {
			ID:             demoUserID,
			Email:          "demo@repurpose.ai",
			Name:           "Demo Creator",
			AvatarURL:      "",
			Tier:           commons.TierPro,
			DailyCredits:   25,
			MonthlyCredits: 250,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
	}

	return &memoryGenerationRepository{}, &memoryUserRepository{users: users}, &memorySponsorClickRepository{}
}

type memoryGenerationRepository struct {
	mu    sync.RWMutex
	items []*entity.Generation
}

func (r *memoryGenerationRepository) Create(_ context.Context, generation *entity.Generation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if generation.ID == "" {
		generation.ID = uuid.NewString()
	}
	if generation.CreatedAt.IsZero() {
		generation.CreatedAt = time.Now()
	}
	generation.UpdatedAt = generation.CreatedAt

	r.items = append(r.items, generation)
	return nil
}

func (r *memoryGenerationRepository) Update(_ context.Context, generation *entity.Generation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, item := range r.items {
		if item.ID == generation.ID {
			generation.UpdatedAt = time.Now()
			r.items[i] = generation
			return nil
		}
	}
	r.items = append(r.items, generation)
	return nil
}

func (r *memoryGenerationRepository) FindByID(_ context.Context, id string) (*entity.Generation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, item := range r.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

func (r *memoryGenerationRepository) FindByUser(_ context.Context, userID string, page, size int) ([]*entity.Generation, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]*entity.Generation, 0, len(r.items))
	for _, item := range r.items {
		if item.UserID == userID {
			filtered = append(filtered, item)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	total := int64(len(filtered))
	start := (page - 1) * size
	if start >= len(filtered) {
		return []*entity.Generation{}, total, nil
	}
	end := start + size
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total, nil
}

type memoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*entity.User
}

func (r *memoryUserRepository) FindByID(_ context.Context, id string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (r *memoryUserRepository) FindByEmail(_ context.Context, email string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if strings.EqualFold(user.Email, email) {
			return user, nil
		}
	}
	return nil, nil
}

func (r *memoryUserRepository) FindByGoogleID(_ context.Context, googleID string) (*entity.User, error) {
	if strings.TrimSpace(googleID) == "" {
		return nil, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.GoogleID == googleID {
			return user, nil
		}
	}
	return nil, nil
}

func (r *memoryUserRepository) Create(_ context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	for _, existing := range r.users {
		if strings.EqualFold(existing.Email, user.Email) {
			return entity.ErrDuplicateEmail
		}
		if user.GoogleID != "" && existing.GoogleID == user.GoogleID {
			return entity.ErrDuplicateGoogleID
		}
	}

	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now
	r.users[user.ID] = user
	return nil
}

func (r *memoryUserRepository) Update(_ context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.users[user.ID]
	if !ok {
		return entity.ErrNotFound
	}

	for id, other := range r.users {
		if id == user.ID {
			continue
		}
		if strings.EqualFold(other.Email, user.Email) {
			return entity.ErrDuplicateEmail
		}
		if user.GoogleID != "" && other.GoogleID == user.GoogleID {
			return entity.ErrDuplicateGoogleID
		}
	}

	user.CreatedAt = existing.CreatedAt
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

type memorySponsorClickRepository struct {
	mu    sync.RWMutex
	items []*entity.SponsorClick
}

func (r *memorySponsorClickRepository) Create(_ context.Context, click *entity.SponsorClick) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if click.ID == "" {
		click.ID = uuid.NewString()
	}
	if click.CreatedAt.IsZero() {
		click.CreatedAt = time.Now()
	}

	r.items = append(r.items, click)
	return nil
}

type memorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*entity.AuthSession
}

// NewMemorySessionRepository returns an in-memory session store so auth works
// without PostgreSQL during local scaffolding/demo. Data is lost on restart.
func NewMemorySessionRepository() entity.ISessionRepository {
	return &memorySessionRepository{sessions: make(map[string]*entity.AuthSession)}
}

func (r *memorySessionRepository) Create(_ context.Context, session *entity.AuthSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.ID == "" {
		session.ID = uuid.NewString()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	for _, existing := range r.sessions {
		if existing.RefreshTokenHash == session.RefreshTokenHash {
			return entity.ErrDuplicateRefresh
		}
	}

	r.sessions[session.ID] = session
	return nil
}

func (r *memorySessionRepository) FindByID(_ context.Context, id string) (*entity.AuthSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (r *memorySessionRepository) FindByRefreshTokenHash(_ context.Context, hash string) (*entity.AuthSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, session := range r.sessions {
		if session.RefreshTokenHash == hash {
			return session, nil
		}
	}
	return nil, nil
}

// Revoke is idempotent: revoking an unknown or already-revoked session succeeds.
func (r *memorySessionRepository) Revoke(_ context.Context, id string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil
	}
	if session.RevokedAt == nil {
		session.RevokedAt = &revokedAt
	}
	return nil
}

func (r *memorySessionRepository) RevokeByUserID(_ context.Context, userID string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, session := range r.sessions {
		if session.UserID == userID && session.RevokedAt == nil {
			session.RevokedAt = &revokedAt
		}
	}
	return nil
}

func (r *memorySessionRepository) DeleteExpired(_ context.Context, before time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, session := range r.sessions {
		if session.ExpiresAt.Before(before) {
			delete(r.sessions, id)
		}
	}
	return nil
}
