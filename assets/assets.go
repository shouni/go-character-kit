// Package assets provides the bundled character definitions as an embedded resource.
package assets

import (
	_ "embed"

	"github.com/shouni/go-character-kit/character"
)

// characterData holds the embedded character definitions.
//
//go:embed characters/characters.json
var characterData []byte

// LoadCharacters loads the embedded character definitions.
func LoadCharacters() (*character.Characters, error) {
	return character.ParseCharacters(characterData)
}
