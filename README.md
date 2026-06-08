# 🎨 Go Character Kit

[![Language](https://img.shields.io/badge/Language-Go-blue)](https://go.dev/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/go-character-kit)](https://go.dev/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/go-character-kit)](https://github.com/shouni/go-character-kit/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/shouni/go-character-kit)](https://goreportcard.com/report/github.com/shouni/go-character-kit)
[![Go Reference](https://pkg.go.dev/badge/github.com/shouni/go-character-kit.svg)](https://pkg.go.dev/github.com/shouni/go-character-kit)
[![Status](https://img.shields.io/badge/Status-Active-brightgreen)](#)

## 🚀 概要 (About) - キャラクターDNA管理キット

**Go Character Kit** は、画像生成・漫画生成ワークフローで利用する **キャラクターDNA** を、JSON 定義として安全に読み込み・検証・参照するための小さな Go ライブラリです。

キャラクターごとの **Seed値**、**参照アセットURL**、**VisualCues/外見指示**、**デフォルトキャラクター** を一元管理し、生成パイプライン側から安定して利用できる形に整えます。

---

## ✨ コア・コンセプト (Core Concepts)

* **🧬 Character DNA Definition**:
  * `id` / `name` / `seed` / `reference_url` / `visual_cues` をまとめて定義し、キャラクターの一貫性維持に必要な情報を扱います。
* **🔍 Safe Lookup Helpers**:
  * ID 検索、大小文字を吸収した検索、未指定時のデフォルトキャラクター fallback を提供します。
* **🛡 Validation First**:
  * 空ID、重複ID、参照画像URL不足、VisualCues不足、複数デフォルト指定など、設定ミスをパース時に検出します。
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
| `reference_url` | ✅ | 一貫性保持のための参照画像URL。`gs://...` などを指定可能。 |
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

### 2. 同梱キャラクター定義を読み込む

```go
package main

import (
	"fmt"
	"log"

	"github.com/shouni/go-character-kit/assets"
)

func main() {
	chars, err := assets.LoadCharacters()
	if err != nil {
		log.Fatal(err)
	}

	for _, char := range chars.List {
		fmt.Println(char.ID, char.ReferenceURL)
	}
}
```

---

## 🧩 パッケージ構成 (Packages)

| パッケージ | 内容 |
| --- | --- |
| `github.com/shouni/go-character-kit/character` | キャラクターのドメインモデル、JSONパース、検証、検索ヘルパー。 |
| `github.com/shouni/go-character-kit/assets` | `go:embed` された同梱キャラクター定義JSONの読み込み。 |

> `character` ディレクトリの package 名も `character` に揃えているため、利用側は alias なしで自然に import できます。

---

## 🛡 バリデーション (Validation)

`ParseCharacters` は、読み込み時に以下の設定ミスを検出します。

* キャラクター定義が空
* `id` / `name` / `reference_url` / `visual_cues` の不足
* `id` の前後空白
* 大小文字を無視した重複ID
* `is_default` が複数指定されている状態

---

## 📂 プロジェクト構造 (Project Structure)

```text
go-character-kit/
├── assets/                   # 【同梱定義】キャラクターJSONの埋め込みと読み込み。
│   ├── assets.go
│   └── characters/
│       └── characters.json
├── character/                # 【ドメイン】キャラクターの型定義・検証・検索。
│   ├── character.go
│   ├── character_helpers.go
│   └── character_test.go
├── go.mod
└── README.md
```

---

## 📜 ライセンス (License)

このプロジェクトは [MIT License](LICENSE) の下で公開されています。
