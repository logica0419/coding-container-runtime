# 低レベルコンテナランタイム自作講座 ハンズオン用テンプレート

ハンズオンパートで使う作業用テンプレートです。

```plain
.
├── main.go         メイン (特にNamespace関連) 処理
├── cgroup.go       cgroup関連処理
├── rootfs.go       rootfs関連処理
├── go.mod
├── go.sum
├── .devcontainer   Dev Container定義ファイル
├── Makefile
│
├── rootfs/         (自動生成) 今回作るコンテナでルートになるディレクトリ
└── config.json     (自動生成) 今回のプログラムの設定ファイル
```

## `make`コマンド一覧

- リポジトリの初期化
  - `config.json` (今回のプログラムの設定ファイル) の生成

```bash
make init
```

- `rootfs`ディレクトリの生成
  - `ubuntu:latest`イメージの中身がコピーされます

```bash
make rootfs
```

- 環境セットアップ完了チェック

```bash
make check
```

- 作成したプログラムの実行

```bash
make run
```
