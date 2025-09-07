# sbcntr-batch

書籍「AWSコンテナ設計・構築[本格]入門 第2版」のバッチ処理アプリケーションリポジトリです。

## 概要

echoフレームワークを利用した、Golang製のバッチ処理サービスです。
APIサーバーとDB(Postgres)の接続はO/Rマッパライブラリであるsqlx[^sqlx]を利用しています。

[^sqlx]: <https://jmoiron.github.io/sqlx/>

本バッチ処理は次のサービスを備えています。

1. 予約バッチ処理
    - 保留中の予約を処理
    - 重複予約のチェック
    - 予約ステータスの更新
2. 通知バッチ処理
    - 通知の生成

## 利用想定

本書の内容に沿って、ご利用ください。

## ローカル利用方法

### 事前準備

- Goのバージョンは1.23系を利用します。
- GOPATHの場所に応じて適切なディレクトリに、このリポジトリのコードをクローンしてください。
- 次のコマンドを利用してモジュールをダウンロードしてください。

```bash
go get golang.org/x/lint/golint
go install
go mod download
```

- 本バックエンドAPIではDB接続があります。DB接続のために次の環境変数を設定してください。
    - DB_HOST
    - DB_USERNAME
    - DB_PASSWORD
    - DB_NAME
    - DB_CONN

### DBの用意

事前にローカルでPostgresサーバを立ち上げてください。

### ビルド＆デプロイ

#### ローカルで動かす場合

```text
export DB_HOST=localhost
export DB_USERNAME=sbcntrapp
export DB_PASSWORD=password
export DB_NAME=sbcntrapp
export DB_CONN=1
```

```bash
make all
```

#### Dockerから動かす場合

```bash
$ docker build -t sbcntr-batch:latest .
$ docker images
REPOSITORY                      TAG                 IMAGE ID            CREATED             SIZE
sbcntr-batch      latest              cdb20b70f267        58 minutes ago      15.2MB
:
$ docker run -d -e DB_HOST=host.docker.internal \
              -e DB_USERNAME=sbcntrapp \
              -e DB_PASSWORD=password \
              -e DB_NAME=sbcntrapp \
              -e DB_CONN=1 \
              sbcntr-batch:latest
```

### デプロイ後の動作確認

バッチ処理のログを確認して、正常に動作していることを確認してください。

```bash
docker logs <container-id>
```

## 環境変数

| 変数名      | 説明                     | デフォルト値 |
| ----------- | ------------------------ | ------------ |
| DB_HOST     | データベースホスト       | localhost    |
| DB_PORT     | データベースポート       | 5432         |
| DB_USERNAME | データベースユーザー     | sbcntrapp    |
| DB_PASSWORD | データベースパスワード   | password     |
| DB_NAME     | データベース名           | sbcntrapp    |
| DB_CONN     | データベース接続プール数 | 1            |

## 開発コマンド

- `make all`: ビルドとテストの実行
- `make build`: アプリケーションのビルド
- `make test`: テストの実行
- `make test-coverage`: カバレッジ付きテストの実行
- `make validate`: コードの検証
- `make clean`: ビルド成果物の削除
- `make install-tools`: 開発ツールのインストール

## 注意事項

- Mac OS Sequoia 15.6でのみ動作確認しています。

## ライセンス

Apache License 2.0
