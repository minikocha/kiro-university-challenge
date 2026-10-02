# Coding Conventions

## 言語・パッケージ

- 言語はGo（`go.mod` に定義されたバージョンに従う）
- AWS操作には `github.com/aws/aws-sdk-go-v2` を使用する
- S3クライアントは `*s3.Client` を直接使用し、インターフェースによる抽象化は行わない
- 標準ライブラリで賄える処理に外部パッケージを追加しない

## パッケージ構成

- `go.mod` はリポジトリルート（`kiro-university-challenge/`）に置く
- モジュールパスは `github.com/minikocha/kiro-university-challenge`
- `remover.go` のパッケージ名は `remover`
- `cmd/object-remover/main.go` のパッケージ名は `main`
- `main.go` はフラグ解析・全体フロー制御のみを担い、S3操作のロジックは `remover` パッケージに置く

## エラーハンドリング

- エラーは `fmt.Errorf("...: %w", err)` でコンテキストを付けてラップして返す
- `os.Exit` は `main.go` 内でのみ使用する。`remover` パッケージ内では使用しない
- エラーメッセージは `stderr` に出力する（`fmt.Fprintln(os.Stderr, err)`）

## 命名規則

- 公開関数・型はGoの標準に従いキャメルケース（例: `NewRemover`, `ObjectVersion`）
- 非公開関数はローワーキャメルケース（例: `chunkObjectVersions`）
- エラー変数は `err` を基本とし、複数ある場合は `listErr`, `deleteErr` のように prefix をつける

## フォーマット

- コードは `gofmt` でフォーマットする
- インポートは標準ライブラリ・外部パッケージの順にグループ分けし、空行で区切る

## テスト

- テスト対象は純粋関数（AWS APIを呼ばないロジック）に限定する
- テストファイルは `remover_test.go`（`package remover`）に記述する
- テーブル駆動テストを基本とする
