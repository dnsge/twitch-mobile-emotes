package emotes

type Emote interface {
	// EmoteID is an identifier string for the emote within its provider's realm.
	EmoteID() string

	// TypedName is the name that people type into Twitch chat.
	TypedName() string

	// URL returns the URL to render the emote for the given size.
	URL(size ImageSize) string

	// LetterCode returns the provider prefix letter code.
	LetterCode() string

	// Type returns the emote type/mimetype of the image.
	Type() string
}
