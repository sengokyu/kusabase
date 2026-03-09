package core

import "time"

// I18nText is a locale-to-string map. Common keys are "en" and "ja".
type I18nText map[string]string

// Avatar describes an avatar for tools and presets.
type Avatar struct {
	Type string `json:"type"`
	Icon string `json:"icon,omitempty"`
	Path string `json:"path,omitempty"`
}

// Conversation represents a single chat conversation.
type Conversation struct {
	UUID              string     `json:"uuid"`
	Title             string     `json:"title"`
	GenerationTTL     *int       `json:"generationTtl"`
	LastMessageAt     time.Time  `json:"lastMessageAt"`
	IsOpenAIAssistant bool       `json:"isOpenAIAssistant"`
	IsBookmarked      bool       `json:"isBookmarked"`
	BookmarkedAt      *time.Time `json:"bookmarkedAt"`
}

// Tool represents an available tool in the system.
type Tool struct {
	UUID                  string    `json:"uuid"`
	Name                  string    `json:"name"`
	DisplayName           I18nText  `json:"displayName"`
	Description           I18nText  `json:"description"`
	ChatUsageNotes        I18nText  `json:"chatUsageNotes"`
	Avatar                Avatar    `json:"avatar"`
	UpdatedAt             time.Time `json:"updatedAt"`
	CreatedAt             time.Time `json:"createdAt"`
	DisplayOrder          *int      `json:"displayOrder"`
	InitialDisplayOrder   *int      `json:"initialDisplayOrder"`
	ConfigurationRequired []string  `json:"configurationRequired"`
}

// ModelParameters holds model-specific configuration.
type ModelParameters struct {
	Model string `json:"model"`
}

// InitialPrompts holds initial prompt configurations keyed by locale.
type InitialPrompts map[string][]any

// Preset represents a chat preset (model + configuration bundle).
type Preset struct {
	ID                          int             `json:"id"`
	UUID                        string          `json:"uuid"`
	Name                        I18nText        `json:"name"`
	Subtitle                    I18nText        `json:"subtitle"`
	Avatar                      Avatar          `json:"avatar"`
	FeatureFlag                 string          `json:"featureFlag"`
	FeatureFlags                []string        `json:"featureFlags"`
	ShowTools                   bool            `json:"showTools"`
	Position                    int             `json:"position"`
	Notes                       []any           `json:"notes"`
	ModelParameters             ModelParameters `json:"modelParameters"`
	Initial                     InitialPrompts  `json:"initial"`
	IsOpenAIAssistant           bool            `json:"isOpenAIAssistant"`
	IsFreeUsage                 bool            `json:"isFreeUsage"`
	AcceptsImageAttachment      bool            `json:"acceptsImageAttachment"`
	AcceptsAudioAttachment      *bool           `json:"acceptsAudioAttachment"`
	AcceptsOpenAIFileAttachment *bool           `json:"acceptsOpenAIFileAttachment"`
	AcceptsEmbeddedAttachments  bool            `json:"acceptsEmbeddedAttachments"`
	MaxAudioDuration            *int            `json:"maxAudioDuration"`
	DisallowedToolUUIDs         []string        `json:"disallowedToolUuids"`
	Cautions                    I18nText        `json:"cautions"`
	ImageUploadSize             int             `json:"imageUploadSize"`
}

// PresetListResponse is the response body for the preset list endpoint.
type PresetListResponse struct {
	Presets       []Preset `json:"presets"`
	DefaultPreset string   `json:"defaultPreset"`
}

// AuthProbeRequest is the request body for probing authentication methods.
type AuthProbeRequest struct {
	Email       string `json:"email"`
	RedirectURL string `json:"redirectURL"`
}

// AuthProbeResponse is the response for the auth probe endpoint.
type AuthProbeResponse struct {
	AllowPassword     bool     `json:"allowPassword"`
	ExternalProviders []string `json:"externalProviders"`
}

// ChatRequest is the public request type for sending a chat message.
type ChatRequest struct {
	Content         string
	AttachmentUUIDs []string
	ConfiguredTools []ConfiguredTool
	IsRetry         bool
}

// ConfiguredTool holds the enabled/disabled state of a tool for a chat request.
type ConfiguredTool struct {
	UUID     string                 `json:"uuid"`
	Settings ConfiguredToolSettings `json:"settings"`
}

// ConfiguredToolSettings holds the settings for a configured tool.
type ConfiguredToolSettings struct {
	Enabled bool `json:"enabled"`
}
