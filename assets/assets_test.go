package assets

import "testing"

// TestLoadCharacters は、同梱の characters.json が検証込みで正しく読み込めることを保証します。
// JSONは手編集されるため、壊れた定義の混入をこのテストで検出します。
func TestLoadCharacters(t *testing.T) {
	t.Parallel()

	chars, err := LoadCharacters()
	if err != nil {
		t.Fatalf("LoadCharacters() error = %v", err)
	}

	if chars.Len() == 0 {
		t.Fatal("Len() = 0, want at least one character")
	}
	if chars.GetDefault() == nil {
		t.Error("GetDefault() = nil, デフォルトキャラクターが定義されていること")
	}

	for _, char := range chars.All() {
		found := chars.GetCharacter(char.ID)
		if found == nil || found.ID != char.ID {
			t.Errorf("GetCharacter(%q) = %+v, want the character itself", char.ID, found)
		}
	}
}
