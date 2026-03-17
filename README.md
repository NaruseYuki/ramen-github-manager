# 🍜 ramen-github-manager (rgm)

ラーメンレコメンドプロジェクトの複数リポジトリのGitHub Issue/PRを一元管理するCLIツール。

## インストール

```bash
# ビルド
make build

# GOPATH/bin にインストール
make install
```

## 初期設定

```bash
# デフォルト設定ファイルを作成
rgm config init

# 設定を確認
rgm config show
```

## 使い方

### Issue 管理

```bash
# 全リポジトリのIssue一覧
rgm issue list

# 特定リポジトリのみ（エイリアス使用可）
rgm issue list -r mobile
rgm issue list -r backend --state closed

# ラベル・アサイニーでフィルタ
rgm issue list -l bug -l urgent
rgm issue list -a NaruseYuki

# Issue詳細
rgm issue view mobile 215

# Issue作成
rgm issue create mobile -t "新しい機能" -b "詳細説明"

# Issue操作
rgm issue close mobile 215
rgm issue reopen mobile 215
rgm issue label mobile 215 --add bug --add urgent
rgm issue assign mobile 215 NaruseYuki
rgm issue comment mobile 215 -b "コメント内容"
```

### PR 管理

```bash
# 全リポジトリのPR一覧
rgm pr list

# 特定リポジトリ・フィルタ
rgm pr list -r backend --state open
rgm pr list -a NaruseYuki

# PR詳細
rgm pr view mobile 216

# PR操作
rgm pr approve mobile 216
rgm pr merge mobile 216 -m squash
rgm pr comment mobile 216 -b "LGTM!"
```

### ダッシュボード

```bash
# プロジェクト全体の概要
rgm dashboard

# 週次レポート
rgm dashboard --weekly

# 月次レポート
rgm dashboard --monthly
```

## リポジトリエイリアス

| エイリアス | リポジトリ |
|-----------|-----------|
| `mobile` | ramen_recommendation |
| `backend` | ramen_recommendation_backend |
| `infra` | ramen-infrastructure |
| `design` | ramen_recommendation_design |

## 認証

`gh` CLI の認証情報を再利用します。事前に `gh auth login` でログインしてください。

環境変数 `GITHUB_TOKEN` または `GH_TOKEN` でも設定可能です。

## 設定ファイル

`~/.config/ramen-github-manager/config.yaml`

```yaml
owner: NaruseYuki
repositories:
  - name: ramen_recommendation
    alias: mobile
  - name: ramen_recommendation_backend
    alias: backend
  - name: ramen-infrastructure
    alias: infra
  - name: ramen_recommendation_design
    alias: design
defaults:
  sort: updated
  limit: 30
  state: open
```
