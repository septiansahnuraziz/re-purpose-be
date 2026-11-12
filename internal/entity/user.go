package entity

import "time"

type User struct {
	ID              string     `gorm:"column:id;primaryKey;type:varchar(64)" json:"id"`
	Email           string     `gorm:"column:email;type:varchar(255);uniqueIndex" json:"email"`
	PasswordHash    string     `gorm:"column:password_hash;type:varchar(255)" json:"-"`
	GoogleID        string     `gorm:"column:google_id;type:varchar(255);uniqueIndex" json:"-"`
	AuthProvider    string     `gorm:"column:auth_provider;type:varchar(32)" json:"authProvider"`
	PrimaryRole     string     `gorm:"column:primary_role;type:varchar(32)" json:"primaryRole,omitempty"`
	Name            string     `gorm:"column:name;type:varchar(255)" json:"name"`
	AvatarURL       string     `gorm:"column:avatar_url;type:varchar(512)" json:"avatarUrl"`
	Tier            string     `gorm:"column:tier;type:varchar(32)" json:"tier"`
	DailyCredits    int        `gorm:"column:daily_credits" json:"dailyCredits"`
	MonthlyCredits  int        `gorm:"column:monthly_credits" json:"monthlyCredits"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at" json:"-"`
	LastLoginAt     *time.Time `gorm:"column:last_login_at" json:"-"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

// HasPassword reports whether the user can sign in with email + password.
func (u *User) HasPassword() bool {
	return u != nil && u.PasswordHash != ""
}

// IsEmailVerified reports whether the user's email has been verified.
func (u *User) IsEmailVerified() bool {
	return u != nil && u.EmailVerifiedAt != nil
}

// ToAuthUser builds the payload returned by the auth endpoints.
func (u *User) ToAuthUser() *AuthUser {
	return &AuthUser{
		ID:             u.ID,
		Name:           u.Name,
		Email:          u.Email,
		AvatarURL:      u.AvatarURL,
		Tier:           u.Tier,
		Plan:           u.Tier,
		PrimaryRole:    u.PrimaryRole,
		DailyCredits:   u.DailyCredits,
		MonthlyCredits: u.MonthlyCredits,
	}
}

func (User) TableName() string { return "users" }

// UserProfile is the response for TITIK INTEGRASI API #5.
type UserProfile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	AvatarURL      string `json:"avatarUrl"`
	Tier           string `json:"tier"`
	DailyCredits   int    `json:"dailyCredits"`
	MonthlyCredits int    `json:"monthlyCredits"`
}

func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:             u.ID,
		Name:           u.Name,
		Email:          u.Email,
		AvatarURL:      u.AvatarURL,
		Tier:           u.Tier,
		DailyCredits:   u.DailyCredits,
		MonthlyCredits: u.MonthlyCredits,
	}
}
