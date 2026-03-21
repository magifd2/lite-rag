# 開発ガイド

## ビルド

```sh
# 通常ビルド（ネイティブ）
make build
# → bin/lite-rag

# クロスコンパイル（darwin 向け）
make cross-build-darwin
# → bin/lite-rag-darwin-arm64, bin/lite-rag-darwin-amd64

# クロスコンパイル（linux 向け、podman または docker が必要）
make cross-build-linux
# → bin/lite-rag-linux-amd64, bin/lite-rag-linux-arm64
```

## テスト

```sh
# 全テスト実行
make test

# 品質ゲート（vet + lint + test + build + vuln）
make check
```

テストは DuckDB のインメモリDB（パス `""`）を使用するため、外部依存なしで実行できます。
LLM API は `httptest.Server` でモック化しています。

## コードカバレッジ

```sh
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## lint

```sh
make lint
```

`golangci-lint` を使用します。初回は自動インストールされます（`make setup` でも可）。

## セキュリティスキャン

```sh
make vuln
```

`govulncheck` で Go 依存関係の既知脆弱性をスキャンします。`make check` の一部として自動実行されます。

## Git フック

```sh
make setup
```

`pre-commit` と `pre-push` フックをインストールします。どちらも `make check` を実行し、
品質ゲートを通過しないとコミット・プッシュをブロックします。

## プロジェクト構成

```
lite-rag/
├── cmd/lite-rag/       # エントリーポイント（main, index, ask, version コマンド）
├── internal/
│   ├── config/         # TOML 設定読み込み + 環境変数オーバーライド
│   ├── database/       # DuckDB 接続・マイグレーション・クエリ
│   ├── indexer/        # ファイル走査・ハッシュ・チャンク・埋め込み・アップサート
│   ├── llm/            # OpenAI 互換 HTTP クライアント（Embed・Chat）
│   ├── normalizer/     # NFKC 正規化・空白・制御文字・Markdown 除去
│   └── retriever/      # ベクター検索・コンテキスト拡張・重複除去
├── pkg/
│   └── chunker/        # 見出し対応チャンク分割（外部再利用可能）
├── docs/               # 英語ドキュメント
├── docs/ja/            # 日本語ドキュメント
└── testdata/           # 機能検証用テストデータセット
```

## 依存関係

| パッケージ                              | バージョン | 用途                      |
|---------------------------------------|---------|--------------------------|
| `github.com/marcboeker/go-duckdb/v2`  | v2.4.3  | DuckDB Go ドライバー（v2 で `[]float32` 対応）|
| `github.com/spf13/cobra`              | v1.9.1  | CLI フレームワーク         |
| `github.com/BurntSushi/toml`          | v1.5.0  | TOML 設定パーサー          |
| `golang.org/x/text`                   | v0.25.0 | Unicode NFKC 正規化        |
