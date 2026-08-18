package character

import (
	"strings"
	"testing"
)

func validList() []Character {
	seed := int64(10001)
	return []Character{
		{
			ID:           "zundamon",
			Name:         "Zundamon",
			ReferenceURL: "gs://bucket/zundamon.png",
			ReferenceURLs: map[string]string{
				"16:9": "gs://bucket/zundamon-16x9.png",
			},
			VisualCues: []string{"green hair"},
			Seed:       &seed,
			IsDefault:  true,
		},
		{
			ID:           "metan",
			Name:         "Metan",
			ReferenceURL: "gs://bucket/metan.png",
			VisualCues:   []string{"purple hair"},
		},
	}
}

func mustNewCharacters(t *testing.T, list []Character) *Characters {
	t.Helper()
	chars, err := NewCharacters(list)
	if err != nil {
		t.Fatalf("NewCharacters() error = %v", err)
	}
	return chars
}

func TestNewCharactersBuildsListAndLookup(t *testing.T) {
	t.Parallel()

	chars := mustNewCharacters(t, validList())

	if got := chars.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}
	if got := len(chars.All()); got != 2 {
		t.Fatalf("len(All()) = %d, want 2", got)
	}
	if got := chars.GetCharacter("ZUNDAMON"); got == nil || got.ID != "zundamon" {
		t.Fatalf("GetCharacter(\"ZUNDAMON\") = %+v, want zundamon", got)
	}
	def := chars.GetDefault()
	if def == nil || def.ID != "zundamon" {
		t.Fatalf("GetDefault() = %+v, want zundamon", def)
	}
	if def.Seed == nil || *def.Seed != 10001 {
		t.Fatalf("GetDefault().Seed = %v, want 10001", def.Seed)
	}
	if got := chars.GetCharacterWithDefault("metan"); got == nil || got.ID != "metan" {
		t.Fatalf("GetCharacterWithDefault(\"metan\") = %+v, want metan", got)
	}
	if got := chars.GetCharacterWithDefault(""); got == nil || got.ID != "zundamon" {
		t.Fatalf("GetCharacterWithDefault(\"\") = %+v, want default zundamon", got)
	}
	if got := chars.GetCharacter("unknown"); got != nil {
		t.Fatalf("GetCharacter(\"unknown\") = %+v, want nil", got)
	}
}

func TestParseCharactersBuildsListAndLookup(t *testing.T) {
	t.Parallel()

	chars, err := ParseCharacters([]byte(`[
		{
			"id": "zundamon",
			"name": "Zundamon",
			"reference_url": "gs://bucket/zundamon.png",
			"reference_urls": {"16:9": "gs://bucket/zundamon-16x9.png"},
			"visual_cues": ["green hair"],
			"seed": 10001,
			"is_default": true
		}
	]`))
	if err != nil {
		t.Fatalf("ParseCharacters() error = %v", err)
	}

	if got := chars.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}
	if got := chars.GetCharacter("ZUNDAMON"); got == nil || got.ID != "zundamon" {
		t.Fatalf("GetCharacter(\"ZUNDAMON\") = %+v, want zundamon", got)
	}
	def := chars.GetDefault()
	if def == nil || def.ID != "zundamon" || def.Seed == nil || *def.Seed != 10001 {
		t.Fatalf("GetDefault() = %+v, want zundamon with seed 10001", def)
	}
	if got := chars.GetCharacterWithDefault(""); got == nil || got.ID != "zundamon" {
		t.Fatalf("GetCharacterWithDefault(\"\") = %+v, want default zundamon", got)
	}
}

func TestParseCharactersRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	if _, err := ParseCharacters([]byte(`{not json`)); err == nil {
		t.Fatal("ParseCharacters() error = nil, want JSON parse error")
	}
}

func TestNewCharactersCopiesInput(t *testing.T) {
	t.Parallel()

	list := validList()
	chars := mustNewCharacters(t, list)

	// 入力スライスを書き換えても集合には影響しない。
	list[0].Name = "mutated"
	list[0].VisualCues[0] = "mutated"
	list[0].ReferenceURLs["16:9"] = "mutated"
	*list[0].Seed = -1

	got := chars.GetCharacter("zundamon")
	if got.Name != "Zundamon" {
		t.Errorf("Name = %q, want Zundamon", got.Name)
	}
	if got.VisualCues[0] != "green hair" {
		t.Errorf("VisualCues[0] = %q, want green hair", got.VisualCues[0])
	}
	if got.ReferenceURLs["16:9"] != "gs://bucket/zundamon-16x9.png" {
		t.Errorf("ReferenceURLs[16:9] = %q, want original URL", got.ReferenceURLs["16:9"])
	}
	if got.Seed == nil || *got.Seed != 10001 {
		t.Errorf("Seed = %v, want 10001", got.Seed)
	}
}

func TestAccessorsReturnCopies(t *testing.T) {
	t.Parallel()

	chars := mustNewCharacters(t, validList())

	// 返り値を書き換えても集合には影響しない。
	all := chars.All()
	all[0].Name = "mutated"
	all[0].VisualCues[0] = "mutated"

	got := chars.GetCharacter("zundamon")
	got.Name = "mutated"
	got.VisualCues[0] = "mutated"
	*got.Seed = -1

	def := chars.GetDefault()
	if def.Name != "Zundamon" || def.VisualCues[0] != "green hair" || *def.Seed != 10001 {
		t.Errorf("GetDefault() = %+v, want unmodified zundamon", def)
	}
	if name := chars.All()[0].Name; name != "Zundamon" {
		t.Errorf("All()[0].Name = %q, want Zundamon", name)
	}
}

func TestWithSeedOverride(t *testing.T) {
	t.Parallel()

	chars := mustNewCharacters(t, validList())

	overridden := chars.WithSeedOverride("ZUNDAMON", 999)

	if overridden == nil {
		t.Fatal("WithSeedOverride() = nil")
	}
	if seed := overridden.GetCharacter("zundamon").Seed; seed == nil || *seed != 999 {
		t.Errorf("overridden zundamon Seed = %v, want 999", seed)
	}
	// 他のキャラクターと元の集合は変わらない。
	if seed := overridden.GetCharacter("metan").Seed; seed != nil {
		t.Errorf("metan Seed = %v, want nil", seed)
	}
	if seed := chars.GetCharacter("zundamon").Seed; seed == nil || *seed != 10001 {
		t.Errorf("original zundamon Seed = %v, want 10001", seed)
	}
}

func TestWithSeedOverrideUnknownIDReturnsReceiver(t *testing.T) {
	t.Parallel()

	chars := mustNewCharacters(t, validList())

	if got := chars.WithSeedOverride("unknown", 999); got != chars {
		t.Fatalf("WithSeedOverride(unknown) = %p, want receiver %p", got, chars)
	}
}

func TestNilReceiverSafety(t *testing.T) {
	t.Parallel()

	var chars *Characters
	if got := chars.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
	if got := chars.All(); got != nil {
		t.Errorf("All() = %v, want nil", got)
	}
	if got := chars.GetCharacter("zundamon"); got != nil {
		t.Errorf("GetCharacter() = %v, want nil", got)
	}
	if got := chars.GetDefault(); got != nil {
		t.Errorf("GetDefault() = %v, want nil", got)
	}
	if got := chars.GetCharacterWithDefault("zundamon"); got != nil {
		t.Errorf("GetCharacterWithDefault() = %v, want nil", got)
	}
	if got := chars.WithSeedOverride("zundamon", 1); got != nil {
		t.Errorf("WithSeedOverride() = %v, want nil", got)
	}
}

func TestCharacterReferenceURLForPrefersAspectRatioMatch(t *testing.T) {
	t.Parallel()

	char := Character{
		ID:           "tsumugi",
		ReferenceURL: "gs://bucket/tsumugi-16x9.png",
		ReferenceURLs: map[string]string{
			"9:16": "gs://bucket/tsumugi-9x16.png",
			"1:1":  "gs://bucket/tsumugi-1x1.png",
		},
	}

	if got := char.ReferenceURLFor("9:16"); got != "gs://bucket/tsumugi-9x16.png" {
		t.Errorf("ReferenceURLFor(9:16) = %q", got)
	}
	if got := char.ReferenceURLFor("1:1"); got != "gs://bucket/tsumugi-1x1.png" {
		t.Errorf("ReferenceURLFor(1:1) = %q", got)
	}
}

func TestCharacterReferenceURLForFallsBackWhenAspectRatioMissing(t *testing.T) {
	t.Parallel()

	char := Character{
		ID:           "tsumugi",
		ReferenceURL: "gs://bucket/tsumugi-16x9.png",
		ReferenceURLs: map[string]string{
			"9:16": "gs://bucket/tsumugi-9x16.png",
		},
	}

	// No "4:3" entry, and no aspect ratio requested at all: both fall back to ReferenceURL.
	if got := char.ReferenceURLFor("4:3"); got != "gs://bucket/tsumugi-16x9.png" {
		t.Errorf("ReferenceURLFor(4:3) = %q, want fallback", got)
	}
	if got := char.ReferenceURLFor(""); got != "gs://bucket/tsumugi-16x9.png" {
		t.Errorf("ReferenceURLFor(\"\") = %q, want fallback", got)
	}
}

func TestCharacterReferenceURLForNilCharacterReturnsEmpty(t *testing.T) {
	t.Parallel()

	var char *Character
	if got := char.ReferenceURLFor("9:16"); got != "" {
		t.Errorf("ReferenceURLFor() = %q, want empty", got)
	}
}

func TestNewCharactersValidation(t *testing.T) {
	t.Parallel()

	base := func(mutate func(list []Character)) []Character {
		list := validList()
		mutate(list)
		return list
	}

	tests := []struct {
		name    string
		list    []Character
		wantErr string
	}{
		{
			name:    "空リスト",
			list:    []Character{},
			wantErr: "キャラクター定義が空です",
		},
		{
			name:    "nilリスト",
			list:    nil,
			wantErr: "キャラクター定義が空です",
		},
		{
			name:    "空ID",
			list:    base(func(l []Character) { l[0].ID = "" }),
			wantErr: "キャラクターIDが空です",
		},
		{
			name:    "IDの前後空白",
			list:    base(func(l []Character) { l[0].ID = " zundamon " }),
			wantErr: "前後の空白があります",
		},
		{
			name:    "大小文字を無視した重複ID",
			list:    base(func(l []Character) { l[1].ID = "ZUNDAMON" }),
			wantErr: "キャラクターIDが重複しています",
		},
		{
			name:    "空の名前",
			list:    base(func(l []Character) { l[0].Name = " " }),
			wantErr: "キャラクター名が空です",
		},
		{
			name:    "空の参照URL",
			list:    base(func(l []Character) { l[0].ReferenceURL = "" }),
			wantErr: "参照画像URLが空です",
		},
		{
			name:    "空のvisual_cues",
			list:    base(func(l []Character) { l[0].VisualCues = nil }),
			wantErr: "visual_cuesが空です",
		},
		{
			name:    "不正なアスペクト比キー",
			list:    base(func(l []Character) { l[0].ReferenceURLs = map[string]string{"16x9": "gs://bucket/a.png"} }),
			wantErr: "アスペクト比キーが不正です",
		},
		{
			name:    "reference_urlsの空URL",
			list:    base(func(l []Character) { l[0].ReferenceURLs = map[string]string{"16:9": " "} }),
			wantErr: "reference_urlsのURLが空です",
		},
		{
			name:    "複数デフォルト",
			list:    base(func(l []Character) { l[1].IsDefault = true }),
			wantErr: "デフォルトキャラクターが複数あります: zundamon, metan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewCharacters(tt.list)
			if err == nil {
				t.Fatal("NewCharacters() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NewCharacters() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
