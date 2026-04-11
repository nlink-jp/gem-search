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

# 検索の深さを制御
gem-search --max-rounds 5 "複雑な調査トピック"

# 出力言語を指定
gem-search --lang ja "English topic, Japanese report"
```

### フラグ

| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `--format` | `markdown` | 出力形式: `json`, `markdown`, `both` |
| `-o, --output` | (stdout) | 出力ファイルプレフィックス（`.md`/`.json`を付与） |
| `--max-rounds` | `3` | 自律検索ラウンドの上限（最大: 10） |
| `--lang` | (なし) | 出力言語コード（例: `ja`, `en`） |

## 仕組み

```
クエリ → Gemini + Google Search Grounding
          → LLMがGrounding結果を分析
            → 判断: 追加検索 / 完了
              → 最終レポート生成
                → 出力（Markdown / JSON）
```

エージェントループは`--max-rounds`回まで繰り返す。GeminiのGoogle Search Groundingが1回のAPI呼び出しでWeb検索結果と事前抽出コンテンツを提供する — 別途の検索APIやWebスクレイピングは不要。

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

- [RFP](docs/ja/gem-search-rfp.ja.md) — 要件定義書

## ライセンス

[LICENSE](LICENSE)を参照。
