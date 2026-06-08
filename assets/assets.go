package assets

import (
	"embed"
	"fmt"

	domain "github.com/shouni/go-character-kit/character"
)

var (
	characterPath = "characters/characters.json"
	// characterFiles はキャラクター定義です。
	//go:embed characters/characters.json
	characterFiles embed.FS
)

// LoadCharacters は埋め込まれたキャラクター定義を読み込みます。
func LoadCharacters() (*domain.Characters, error) {
	data, err := characterFiles.ReadFile(characterPath)
	if err != nil {
		return nil, fmt.Errorf("キャラクター定義の読み込み失敗: %w", err)
	}
	return domain.ParseCharacters(data)
}
