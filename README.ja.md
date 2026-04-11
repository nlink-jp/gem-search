# gem-search

Vertex AI GeminiのGoogle Search Groundingを使ったエージェンティックWeb検索CLI。

自然言語で質問を入力すると、GeminiのGoogle Search Groundingで自律的にWeb検索・分析を行い、MarkdownまたはJSONレポートを出力する。単体の調査ツールとしてだけでなく、stdin/stdoutによるパイプライン連携でも使用可能。

## 前提条件

- **Google Cloudプロジェクト** — Vertex AI APIが有効であること
- **Application Default Credentials** — `gcloud auth application-default login` を実行

## インストール

```bash
git clone https://github.com/nlink-jp/gem-search.git
cd gem-search
make build
# バイナリ: dist/gem-search
```

## 設定

| 変数 | 必須 | デフォルト | 説明 |
|------|------|-----------|------|
| `GEMSEARCH_PROJECT` | はい | — | GCPプロジェクトID |
| `GEMSEARCH_LOCATION` | いいえ | `us-central1` | Vertex AIリージョン |
| `GEMSEARCH_MODEL` | いいえ | `gemini-2.5-flash` | Geminiモデル名 |
| `GEMSEARCH_LANG` | いいえ | — | 出力言語 |

## 使い方

```bash
# 基本（Markdownをstdoutに出力）
gem-search "Web Groundingとは何か"

# 標準入力（パイプライン）
echo "Go context.WithTimeoutのベストプラクティス" | gem-search

# JSON出力
gem-search --format json "Vertex AIの料金体系"

# MarkdownとJSON両方をファイルに出力
gem-search --format both -o ./report "調査したいトピック"
# → ./report.md + ./report.json

# 出力言語を指定
gem-search --lang ja "English topic, Japanese report"
```

### フラグ

| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `--format` | `markdown` | 出力形式: `json`, `markdown`, `both` |
| `-o, --output` | (stdout) | 出力ファイルプレフィックス（`.md`/`.json`を付与） |
| `--lang` | (なし) | 出力言語コード（例: `ja`, `en`） |

## 仕組み

```
クエリ → Phase 1: Survey（広く概要を把握）
          → Phase 2: Deep-dive（不足を深掘り）
            → Phase 3: Verify（矛盾・最新情報を検証）
              → 3フェーズの成果から最終レポート生成
                → 出力（Markdown / JSON）
```

すべての検索は固定3フェーズの調査パイプラインで実行される。各フェーズはGeminiのGoogle Search Groundingを異なる目的で使用する — Surveyで全体像を把握し、Deep-diveで不足を埋め、Verifyで矛盾と鮮度を検証する。LLMに「もう十分」と判断させるのではなく、常に3フェーズを実行して網羅性を確保する。

## 経緯

本ツールは[agentic-web-search](https://github.com/nlink-jp/agentic-web-search)（アーカイブ済み）の後継。前プロジェクトはBrave Search APIを使用していたが、利用規約の攻撃的な制限と有償API登録の要件で凍結。DuckDuckGoもHTMLエンドポイントのrobots.txtがボットアクセスを禁止（`Disallow: /`）しているため不採用。

Vertex AI Web Groundingはこれらの問題を回避する — 公式のGoogle Cloud API機能であり、明確な利用規約と既存の課金基盤がある。

## ビルド

```bash
make build       # ビルド → dist/gem-search
make build-all   # 5プラットフォーム向けクロスコンパイル
make test        # 全テスト実行
make clean       # dist/を削除
```

## ドキュメント

- [アーキテクチャ](docs/ja/architecture.ja.md) — 設計判断とその根拠
- [RFP](docs/ja/gem-search-rfp.ja.md) — 要件定義書

## ライセンス

[LICENSE](LICENSE)を参照。
