package assets

import "testing"

// TestLoadCharacters ensures the bundled characters.json loads cleanly, validation included.
// The JSON is edited by hand, so this test is what catches a broken definition slipping in.
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
		t.Error("GetDefault() = nil, want a default character to be defined")
	}

	for _, char := range chars.All() {
		found := chars.GetCharacter(char.ID)
		if found == nil || found.ID != char.ID {
			t.Errorf("GetCharacter(%q) = %+v, want the character itself", char.ID, found)
		}
	}
}
