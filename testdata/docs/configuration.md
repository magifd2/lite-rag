# 設定リファレンス

設定は `config.toml` ファイルで管理します。`--config` フラグでパスを指定できます（デフォルト: `./config.toml`）。

## 設定項目

### [api] セクション

```toml
[api]
base_url = "http://localhost:1234/v1"
api_key  = "lm-studio"
```

| キー       | 説明                                           | デフォルト                    |
|-----------|----------------------------------------------|------------------------------|
| base_url  | OpenAI 互換 API のベース URL                   | `http://localhost:1234/v1`   |
| api_key   | Authorization ヘッダーに送るAPIキー。LM Studio では任意の文字列で可 | `lm-studio` |

### [models] セクション

```toml
[models]
embedding = "text-embedding-nomic-embed-text-v1.5"
chat      = "openai/gpt-oss-20b"
```

| キー       | 説明                                    |
|-----------|----------------------------------------|
| embedding | 埋め込み生成に使用するモデルの識別子     |
| chat      | 回答生成に使用するチャットモデルの識別子 |

**重要**: `embedding` モデルを変更すると、既存のインデックスとベクター空間が変わるため、
すべてのドキュメントを再インデックスする必要があります。
`embedding_model` は documents テーブルに記録されており、クエリ時に自動的にフィルタリングされます。

### [database] セクション

```toml
[database]
path = "./lite-rag.db"
```

| キー  | 説明                           | デフォルト       |
|------|-------------------------------|-----------------|
| path | DuckDB データベースファイルのパス | `./lite-rag.db` |

### [retrieval] セクション

```toml
[retrieval]
top_k          = 5
context_window = 1
chunk_size     = 512
chunk_overlap  = 64
```

| キー            | 説明                                               | デフォルト |
|----------------|--------------------------------------------------|----------|
| top_k          | ベクター検索で取得する上位チャンク数                | 5        |
| context_window | 各ヒットの前後に拡張するチャンク数（0 = 拡張なし）  | 1        |
| chunk_size     | チャンクあたりの目標トークン数                      | 512      |
| chunk_overlap  | 隣接チャンク間のオーバーラップトークン数             | 64       |

## 環境変数オーバーライド

`config.toml` の設定は環境変数で上書きできます。

| 環境変数              | 対応する設定項目   |
|---------------------|-----------------|
| `LITE_RAG_API_BASE_URL` | `api.base_url` |
| `LITE_RAG_API_KEY`      | `api.api_key`  |
| `LITE_RAG_DB_PATH`      | `database.path`|
