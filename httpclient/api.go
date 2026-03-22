package httpclient

import core "github.com/sengokyu/kusabase/httpclient/internal"

// Re-exported types from the internal package.
// These aliases keep the public import path as "github.com/sengokyu/kusabase".

type (
	// CookieStore is the interface for persisting session data across client instances.
	CookieStore = core.CookieStore

	// APIError represents a non-2xx HTTP response from the API.
	APIError = core.Error

	// I18nText is a locale-to-string map (common keys: "en", "ja").
	I18nText = core.I18nText

	// Avatar describes an avatar for tools and presets.
	Avatar = core.Avatar

	// Conversation represents a single chat conversation.
	Conversation = core.Conversation

	// Tool represents an available tool in the system.
	Tool = core.Tool

	// ModelParameters holds model-specific configuration.
	ModelParameters = core.ModelParameters

	// InitialPrompts holds initial prompt configurations keyed by locale.
	InitialPrompts = core.InitialPrompts

	// Preset represents a chat preset (model + configuration bundle).
	Preset = core.Preset

	// PresetListResponse is the full response from the preset list endpoint.
	PresetListResponse = core.PresetListResponse

	// AuthProbeRequest is the request body for probing authentication methods.
	AuthProbeRequest = core.AuthProbeRequest

	// AuthProbeResponse is the response from the auth probe endpoint.
	AuthProbeResponse = core.AuthProbeResponse

	// ChatRequest is the request type for sending a chat message.
	ChatRequest = core.ChatRequest

	// ConfiguredTool holds the enabled/disabled state of a tool.
	ConfiguredTool = core.ConfiguredTool

	// ConfiguredToolSettings holds per-tool settings within a ChatRequest.
	ConfiguredToolSettings = core.ConfiguredToolSettings

	// Chat represents an ongoing conversation. Obtain via Client.Chat.New().
	Chat = core.Chat
)
