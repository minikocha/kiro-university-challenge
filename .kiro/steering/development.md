# Development Guide

## セットアップ

```bash
cd object-remover
go mod download
```

## ビルド

```bash
# ビルド確認（全パッケージ）
go build ./...

# バイナリ生成
go build -o object-remover ./cmd/object-remover
```

## テスト

```bash
# 全テスト実行
go test ./...

# 詳細出力付き
go test -v ./...
```

## 依存パッケージの追加

```bash
go get <package>@<version>
go mod tidy
```

## AWSの認証情報

ローカル実行時は以下のいずれかの方法で認証情報を設定する：

- `~/.aws/credentials` および `~/.aws/config` を設定する
- 環境変数 `AWS_ACCESS_KEY_ID`、`AWS_SECRET_ACCESS_KEY`、`AWS_REGION` を設定する
- AWS SSOを使用する（`aws sso login`）

## 動作確認

```bash
# ヘルプ表示
./object-remover --help

# 対象オブジェクトの確認（削除は行わない）
./object-remover --bucket <bucket> --prefix <prefix> --region <region>

# 確認なしで削除
./object-remover --bucket <bucket> --prefix <prefix> --region <region> --yes
```
