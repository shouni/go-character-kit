# 🎨 Go Character Kit

[![Language](https://img.shields.io/badge/Language-Go-blue)](https://go.dev/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-character-kit)](https://go.dev/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-character-kit)](https://github.com/shouni/go-character-kit/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Status](https://img.shields.io/badge/Status-Active-brightgreen)](#)

## 🚀 概要 (About) - キャラクターDNA管理キット

**Go Character Kit** は、画像生成・漫画生成ワークフローで利用する **キャラクターDNA** を、JSON 定義として安全に読み込み・検証・参照するための小さな Go ライブラリです。

キャラクターごとの **Seed値**、**参照アセットURL**、**VisualCues/外見指示**、**デフォルトキャラクター** を一元管理し、生成パイプライン側から安定して利用できる形に整えます。

---

## ✨ コア・コンセプト (Core Concepts)

* **🧬 Character DNA Definition**:
  * `id` / `name` / `seed` / `reference_url` / `reference_urls` / `visual_cues` をまとめて定義し、キャラクターの一貫性維持に必要な情報を扱います。
* **🔒 Immutable Collection**:
  * `Characters` は構築時に検証・複製され、以降は読み取り専用です。参照系APIはコピーを返すため、利用側から内部状態を壊せません。
* **🔍 Safe Lookup Helpers**:
  * ID 検索（大小文字を吸収）、未指定時のデフォルトキャラクター fallback、アスペクト比別参照URLの解決（`ReferenceURLFor`）、Seed差し替え済み集合の派生（`WithSeedOverride`）を提供します。
* **🛡 Validation First**:
  * 設定ミスをパース時に検出します（詳細は[バリデーション](#-バリデーション-validation)）。
* **📦 Embedded Character Assets**:
  * `assets` パッケージから、リポジトリ同梱のキャラクター定義JSONを `go:embed` されたデータとして読み込めます。

---

## 🎨 キャラクター定義 (Character Definition)

キャラクターは JSON 配列として定義します。

```json
[
  {
    "id": "zundamon",
    "name": "Zundamon",
    "seed": 10001,
    "reference_url": "gs://bucket/zundamon.png",
    "reference_urls": {
      "16:9": "gs://bucket/zundamon-16x9.png",
      "9:16": "gs://bucket/zundamon-9x16.png"
    },
    "visual_cues": [
      "vibrant emerald green hair",
      "soybean earmuffs",
      "strictly following the design from reference"
    ],
    "is_default": true
  }
]
```

| フィールド | 必須 | 内容 |
| --- | --- | --- |
| `id` | ✅ | キャラクターを識別する安定ID。前後の空白は許可されません。 |
| `name` | ✅ | 表示名・管理名。 |
| `visual_cues` | ✅ | 生成プロンプトへ注入する外見上の特徴。 |
| `reference_url` | ✅ | 一貫性保持のための参照画像URL（既定のフォールバック）。`gs://...` などを指定可能。 |
| `reference_urls` | - | `"16:9"` のようなアスペクト比文字列をキーにした参照画像URL。`ReferenceURLFor` が生成対象と同じアスペクト比のエントリを優先して解決します。 |
| `seed` | - | キャラクターに紐づく任意の生成シード。 |
| `is_default` | - | fallback 用のデフォルトキャラクター。指定できるのは1人まで。 |

---

## ⚙️ 使い方 (Usage)

### 1. JSON から読み込む

```go
package main

import (
	"fmt"
	"log"

	"github.com/shouni/go-character-kit/character"
)

func main() {
	chars, err := character.ParseCharacters([]byte(`[
		{
			"id": "zundamon",
			"name": "Zundamon",
			"reference_url": "gs://bucket/zundamon.png",
			"visual_cues": ["green hair"],
			"seed": 10001,
			"is_default": true
		}
	]`))
	if err != nil {
		log.Fatal(err)
	}

	char := chars.GetCharacterWithDefault("ZUNDAMON")
	fmt.Println(char.ID, char.Name)
}
```

### 2. Go の構造体から初期化する

```go
chars, err := character.NewCharacters([]character.Character{
	{
		ID:           "zundamon",
		Name:         "Zundamon",
		ReferenceURL: "gs://bucket/zundamon.png",
		VisualCues:   []string{"green hair"},
		IsDefault:    true,
	},
})
```

### 3. 同梱キャラクター定義を読み込む

```go
chars, err := assets.LoadCharacters() // import "github.com/shouni/go-character-kit/assets"
```

### 4. 検索・派生ヘルパー

```go
chars.Len()                            // キャラクター数
chars.All()                            // 定義順の一覧（コピー）
chars.GetCharacter("ZUNDAMON")         // 大小文字を無視したID検索。無ければ nil
chars.GetDefault()                     // is_default のキャラクター。無ければ nil
chars.GetCharacterWithDefault("xxx")   // 見つからなければデフォルトに fallback
char.ReferenceURLFor("9:16")           // アスペクト比に合う参照URL。無ければ reference_url

// zundamon の Seed だけ差し替えた新しい集合を派生（元の集合は不変）
overridden := chars.WithSeedOverride("zundamon", 999)
```

---

## 🧩 パッケージ構成 (Packages)

| パッケージ | 内容 |
| --- | --- |
| `github.com/shouni/go-character-kit/character` | キャラクターのドメインモデル、初期化、JSONパース、検証、検索ヘルパー。 |
| `github.com/shouni/go-character-kit/assets` | `go:embed` された同梱キャラクター定義JSONの読み込み。 |

---

## 🛡 バリデーション (Validation)

`NewCharacters` / `ParseCharacters` は、初期化時に以下の設定ミスを検出します。

* キャラクター定義が空（空配列・nil）
* `id` / `name` / `reference_url` / `visual_cues` の不足
* `id` の前後空白
* 大小文字を無視した重複ID
* `reference_urls` のアスペクト比キーの形式不正（`"16:9"` 形式のみ許可。`"16x9"` などのタイポを検出）
* `reference_urls` の空URL
* `is_default` が複数指定されている状態（該当IDをエラーメッセージに列挙）

---

## 📜 ライセンス (License)

このプロジェクトは [MIT License](LICENSE) の下で公開されています。
