package storage

type Settings struct {
	CacheDestroyerKey string `json:"cache_destroyer_key"`
	EnableGifEmotes   bool   `json:"enable_gif_emotes"`
}

type SettingsRepository interface {
	Load(userID string) (*Settings, error)
	Save(userID string, settings *Settings) error
}
