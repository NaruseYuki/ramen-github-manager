# rgm — GitHub Multi-Repo Manager

複数のGitHubリポジトリのIssue/PRをターミナルから一元管理するCLIツール。
任意のプロジェクトで使えます。

## インストール

```bash
# ビルド
make build

# GOPATH/bin にインストール
make install
```

## 初期設定

```bash
# 対話的に設定（owner + リポジトリを入力）
rgm config init

# または手動でリポジトリを追加
rgm config set-owner your-org
rgm config add-repo your-app --alias app
rgm config add-repo your-api --alias api

# 設定を確認
rgm config show
```

## 使い方

### Issue 管理

```bash
# 全リポジトリのIssue一覧
rgm issue list

# 特定リポジトリのみ（エイリアス使用可）
rgm issue list -r app
rgm issue list -r api --state closed

# ラベル・アサイニーでフィルタ
rgm issue list -l bug -l urgent
rgm issue list -a username

# Issue詳細
rgm issue view app 42

# Issue作成
rgm issue create app -t "新しい機能" -b "詳細説明"

# Issue操作
rgm issue close app 42
rgm issue reopen app 42
rgm issue label app 42 --add bug --add urgent
rgm issue assign app 42 username
rgm issue comment app 42 -b "コメント内容"
```

### PR 管理

```bash
# 全リポジトリのPR一覧
rgm pr list

# フィルタ
rgm pr list -r api --state open
rgm pr list -a username

# PR詳細・操作
rgm pr view app 100
rgm pr approve app 100
rgm pr merge app 100 -m squash
rgm pr comment app 100 -b "LGTM!"
```

### ダッシュボード

```bash
# プロジェクト全体の概要
rgm dashboard

# 週次 / 月次レポート
rgm dashboard --weekly
rgm dashboard --monthly
```

### 設定管理

```bash
rgm config show                          # 現在の設定を表示
rgm config set-owner new-org             # owner変更
rgm config add-repo new-repo --alias nr  # リポジトリ追加
rgm config remove-repo old-repo          # リポジトリ削除
rgm config path                          # 設定ファイルのパス
```

## 認証

`gh` CLI の認証情報を再利用します。事前に `gh auth login` でログインしてください。

環境変数 `GITHUB_TOKEN` または `GH_TOKEN` でも設定可能です。

## 設定ファイル

`~/.config/rgm/config.yaml`

```yaml
owner: your-org
repositories:
  - name: your-app
    alias: app
  - name: your-api
    alias: api
  - name: your-infra
    alias: infra
defaults:
  sort: updated
  limit: 30
  state: open
```
