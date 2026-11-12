package commons

// Tone identifiers accepted by the generation engine.
const (
	ToneViral     = "viral"
	ToneExecutive = "executive"
	ToneDeep      = "deep"
)

var ValidTones = map[string]bool{
	ToneViral:     true,
	ToneExecutive: true,
	ToneDeep:      true,
}

// Generation lifecycle.
const (
	GenerationStatusPending    = "PENDING"
	GenerationStatusProcessing = "PROCESSING"
	GenerationStatusCompleted  = "COMPLETED"
	GenerationStatusFailed     = "FAILED"
)

// Subscription tiers.
const (
	TierFree = "FREE"
	TierPro  = "PRO"
)

// Authentication providers stored in users.auth_provider.
const (
	AuthProviderPassword = "password"
	AuthProviderGoogle   = "google"
	AuthProviderBoth     = "both"
)

// Primary roles accepted during registration.
const (
	PrimaryRoleCreator  = "creator"
	PrimaryRoleFounder  = "founder"
	PrimaryRoleMarketer = "marketer"
	PrimaryRoleAgency   = "agency"
)

var ValidPrimaryRoles = map[string]bool{
	PrimaryRoleCreator:  true,
	PrimaryRoleFounder:  true,
	PrimaryRoleMarketer: true,
	PrimaryRoleAgency:   true,
}

// Sponsor analytics events.
const (
	SponsorEventClick      = "click"
	SponsorEventImpression = "impression"
	SponsorEventClose      = "close"
)

var ValidSponsorEvents = map[string]bool{
	SponsorEventClick:      true,
	SponsorEventImpression: true,
	SponsorEventClose:      true,
}

// Defaults used across services.
const (
	DefaultPage           = 1
	DefaultPageSize       = 10
	MaxPageSize           = 100
	DefaultChannelSize    = 15
	DefaultTimeoutSec     = 2
	DefaultDailyCredits   = 10
	DefaultMonthlyCredits = 100
)

// Job task types.
const (
	TaskTypeContentGenerate = "content:generate"
)
