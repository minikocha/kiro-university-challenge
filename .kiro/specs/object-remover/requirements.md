# Requirements Document

## Introduction

`object-remover` は指定したS3バケット上の指定したプレフィックスから始まるオブジェクト（過去バージョン・削除マーカーを含む）を削除するCLIツール。Go言語で実装する。

### CLI Usage

```
object-remover --bucket <bucket> --prefix <prefix> [--region <region>] [--yes] [--help]
```

### Flags

| フラグ | 型 | 必須 | 説明 |
|---|---|---|---|
| `--bucket` | string | ✅（`--help`なし時） | 対象のS3バケット名 |
| `--prefix` | string | ✅（`--help`なし時） | 対象オブジェクトのプレフィックス |
| `--region` | string | ❌ | S3バケットのリージョン。未指定時は`AWS_REGION`環境変数または`~/.aws/config`を参照 |
| `--yes` | bool | ❌ | 削除前のYes/No確認をスキップ |
| `--help` | bool | ❌ | ヘルプを表示して正常終了。他フラグの有無に関わらず優先される |

### Exit Codes

| Code | 条件 |
|---|---|
| `0` | 正常終了（ヘルプ表示、オブジェクトなし、Noでキャンセル、全削除成功） |
| `1` | 異常終了（必須フラグ未指定、バケット不存在、削除失敗あり） |

### Project Structure

```
object-remover/
├── cmd/
│   └── object-remover/
│       └── main.go       # エントリーポイント・フラグ解析・全体フロー
├── remover.go            # Remover構造体・列挙・削除ロジック（package remover）
├── remover_test.go       # ユニットテスト（package remover）
├── go.mod
└── go.sum
```

- モジュールパス: `github.com/minikocha/kiro-university-challenge/object-remover`
- パッケージ構成:
  - `remover.go` → `package remover`
  - `cmd/object-remover/main.go` → `package main`、`remover` パッケージをインポート

### Key Types

```go
// ObjectVersion は列挙・削除対象の1オブジェクトを表す
type ObjectVersion struct {
    Key            string
    VersionId      string
    IsDeleteMarker bool
}

// Remover はS3操作のコアロジックを持つ構造体
type Remover struct {
    client *s3.Client
    bucket string
    prefix string
}
```

### Key Functions

| 関数 | シグネチャ | 説明 |
|---|---|---|
| `NewRemover` | `(ctx, bucket, prefix, region string) (*Remover, error)` | クライアント初期化。`region`が空の場合はデフォルト設定を使用 |
| `CheckBucketExists` | `(ctx) error` | バケット存在確認 |
| `ListObjects` | `(ctx) ([]ObjectVersion, error)` | 全バージョン・削除マーカーを含む全オブジェクトを列挙 |
| `DeleteObjects` | `(ctx, []ObjectVersion) ([]ObjectVersion, error)` | バッチ削除。失敗したオブジェクトを返す |
| `chunkObjectVersions` | `([]ObjectVersion, int) [][]ObjectVersion` | スライスをN件ずつのチャンクに分割（テスト対象） |

### Test Coverage

| テスト対象 | 内容 |
|---|---|
| `chunkObjectVersions` | 1001件 → 2チャンク（1000+1）に分割されること |
| `chunkObjectVersions` | 999件 → 1チャンク（999）に分割されること |
| `chunkObjectVersions` | 0件 → 空スライスになること |

### Dependencies

- `github.com/aws/aws-sdk-go-v2`
- `github.com/aws/aws-sdk-go-v2/config`
- `github.com/aws/aws-sdk-go-v2/service/s3`

---

## Glossary

| 用語 | 説明 |
|---|---|
| オブジェクト | S3バケット上に保存されたデータの単位 |
| バージョン | S3のバージョニング機能によって保存されたオブジェクトの特定の状態。バージョンIDで識別される |
| 過去バージョン | 最新バージョン以外のバージョン |
| 削除マーカー | バージョニングが有効なバケットでオブジェクトを削除した際に作成される特殊なバージョン。実データを持たない |
| プレフィックス | S3オブジェクトキーの先頭部分。ディレクトリのような階層構造を表現するために使われる |
| バッチ削除 | S3の`DeleteObjects` APIを使って最大1000件のオブジェクトを1リクエストでまとめて削除すること |

---

## Requirements

### REQ-001: フラグ解析

- `--help` フラグが指定されている場合、他のフラグの有無に関わらずヘルプを表示して正常終了（exit 0）する
- `--help` フラグが指定されておらず `--bucket` または `--prefix` が未指定の場合、ヘルプを表示して異常終了（exit 1）する

### REQ-002: バケット存在確認

- `--bucket` で指定したS3バケットが存在するか `HeadBucket` APIで確認する
- バケットが存在しない場合、エラーメッセージを`stderr`に出力して異常終了（exit 1）する

### REQ-003: オブジェクト列挙

- `ListObjectVersions` APIで指定プレフィックスのオブジェクトを全件列挙する
- 列挙対象には通常バージョン・過去バージョン・削除マーカーをすべて含む
- ページネーションに対応し、1000件を超える場合も全件取得する
- 対象オブジェクトが0件の場合、その旨を出力して正常終了（exit 0）する

### REQ-004: 削除前確認（`--yes` 未指定時）

- `--yes` フラグが指定されていない場合、削除対象オブジェクトを以下の形式で1行ずつ出力する:
  ```
  s3://bucket/key (versionId: xxxx)
  ```
- ユーザーに `Yes/No` の入力を促す
- `No` が入力された場合、キャンセルメッセージを出力して正常終了（exit 0）する
- `Yes` が入力された場合、削除を実施する
- 入力の判定は大文字・小文字を区別しない

### REQ-005: 削除実施

- S3の `DeleteObjects` APIを使ってバッチ削除する（最大1000件/リクエスト）
- 1000件を超える場合は複数リクエストに分割して実行する
- `--yes` フラグが指定されている場合、REQ-004の確認をスキップして直接削除を実施する
- 削除に失敗したオブジェクトがある場合、失敗したオブジェクトを以下の形式でリスト出力して異常終了（exit 1）する:
  ```
  s3://bucket/key (versionId: xxxx)
  ```
- 全件削除に成功した場合、成功メッセージを出力して正常終了（exit 0）する

### REQ-006: リージョン設定

- `--region` フラグが指定されている場合、そのリージョンを使用する
- `--region` フラグが未指定の場合、`AWS_REGION` 環境変数または `~/.aws/config` の設定を参照する（AWS SDK v2のデフォルト動作に委ねる）
