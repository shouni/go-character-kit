// Package assets は、既定のキャラクター定義を埋め込みリソースとして提供します。
package assets

import (
	_ "embed"

	"github.com/shouni/go-character-kit/character"
)

// characterData は埋め込まれたキャラクター定義です。
//
//go:embed characters/characters.json
var characterData []byte

// LoadCharacters は埋め込まれたキャラクター定義を読み込みます。
func LoadCharacters() (*character.Characters, error) {
	return character.ParseCharacters(characterData)
}
