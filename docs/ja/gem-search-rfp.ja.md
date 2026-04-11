# RFP: gem-search

> Generated: 2026-04-12
> Status: Draft

## 1. Problem Statement

Web上の情報を調査する際、検索→結果読解→追加検索のループを手動で回す非効率がある。既存の`product-research`や`news-collector`はVertex AI Web Groundingを使っているが特定用途に最適化されており、汎用的な調査には使えない。

gem-searchはVertex AI Web Groundingを使ったエージェンティックWeb検索CLIである。自然言語クエリからWeb検索・情報収集・レポート生成を自律的に実行し、Markdown/JSON出力する。単体での調査ツールとしてだけでなく、stdin/stdoutによるパイプライン連携で情報収集パーツとしても機能する。

凍結した`agentic-web-search`のUI設計を踏襲し、バックエンドを公式APIであるVertex AI Web Groundingに置き換える。Vertex AIは既に利用・課金基盤が整っており、利用規約上の懸念がない。

主な利用者はnlink-jp開発者。

## 2. Functional Specification

### Commands / API Surface

```bash
# 基本（引数、stdout にMarkdown出力）
gem-search "質問文"

# stdin入力
echo "質問文" | gem-search

# 出力形式指定
gem-search --format json "..."
gem-search --format markdown "..."
gem-search --format both -o ./result "..."
# → ./result.md + ./result.json

# エージェント制御
gem-search --max-rounds 5 "..."

# 出力言語指定
gem-search --lang ja "..."
```

主要フラグ:

| フラグ | 説明 | デフォルト |
|--------|------|-----------|
| `--format` | 出力形式 (`json`, `markdown`, `both`) | `markdown` |
| `-o` | 出力ファイルプレフィックス（`.md`, `.json`を付与） | なし（stdout） |
| `--max-rounds` | 自律検索ループの上限回数 | TBD |
| `--lang` | 最終出力の言語 | なし（検索結果の言語のまま） |

### Input / Output

- **入力**: 自然言語クエリ（コマンド引数またはstdin）
- **出力**:
  - Markdown: 調査結果のレポート（ソースURL付き）
  - JSON: 構造化データ（検索クエリ、ソースURL/タイトル、最終回答）
  - `-o`指定時はファイル出力、未指定時はstdout

### Configuration

環境変数方式（`GEMSEARCH_`プレフィックス）:

| 環境変数 | 必須 | デフォルト | 説明 |
|----------|------|-----------|------|
| `GEMSEARCH_PROJECT` | Yes | — | GCPプロジェクトID |
| `GEMSEARCH_LOCATION` | No | `us-central1` | Vertex AIリージョン |
| `GEMSEARCH_MODEL` | No | `gemini-2.5-flash` | モデル名 |
| `GEMSEARCH_LANG` | No | — | 出力言語（`--lang`と同等） |

認証: ADC（`gcloud auth application-default login`）

### External Dependencies

- **Vertex AI API**: Gemini + Google Search Grounding
- **nlk**: LLM前後処理ライブラリ（guard, strip, jsonfix, backoff, validate）
- **cobra**: CLIフレームワーク
- **google-genai Go SDK**: `google.golang.org/genai`

## 3. Design Decisions

- **言語: Go** — シングルバイナリ配布（パイプラインパーツとして重要）、nlk統合、Vertex AI Go SDK（`google.golang.org/genai`）のGrounding実績がgem-cliにある。Vertex AI専用機能のGo化を推進する方針。
- **Vertex AI Web Grounding** — 公式API機能。利用規約上の懸念なし（Brave APIで経験した利用規約リスクを回避）。既にGCP課金基盤が整っている。
- **`agentic-web-search`のUI設計踏襲** — フラグ体系（`--format`, `-o`, `--max-rounds`, `--lang`）、出力構造（Markdown+JSON）、stdin/stdout対応を流用。
- **gem-cliのGroundingパターン流用** — `internal/client/client.go`のGoogle Search Grounding有効化、grounding_metadataからのソース抽出パターン。
- **nlk統合**: guard（プロンプトインジェクション防御）、strip（thinking tags除去）、jsonfix（JSON修復）、backoff（リトライ）、validate（出力検証）
- **product-research/news-collectorとは完全独立** — 用途特化ツールとの統合は行わない。
- **スコープ外**:
  - GUI/TUI
  - MCP Tool化（将来の拡張候補）
  - ローカルLLM対応（Vertex AI専用）

## 4. Development Plan

全フェーズ共通原則: **セキュリティファースト**、**テスタビリティ設計**、CONVENTIONS.md準拠。

### Phase 1: Core

- プロジェクトスキャフォールド
- Vertex AI Geminiクライアント（Google Search Grounding有効、nlk統合）
- エージェントループ（クエリ→Grounding検索→結果分析→追加検索/完了判断）
- Grounding metadataからソース抽出（URL+タイトル）
- JSON/Markdown出力（`--format`, `-o`）
- stdin/引数の両入力対応
- ユニットテスト（Vertex AIクライアントのモック可能な設計）

### Phase 2: Features

- 複数ラウンド自律検索（`--max-rounds`）
- 出力言語指定（`--lang` / `GEMSEARCH_LANG`）
- エラーハンドリング強化（API quota、ネットワークエラー、レートリミット）
- プロンプトチューニング

### Phase 3: Release

- README.md / README.ja.md
- CHANGELOG.md
- E2Eテスト（実Vertex AI環境）
- util-seriesへのサブモジュール登録
- リリース手順実行（CONVENTIONS.md準拠）

Phase 1で独立レビュー可能。

## 5. Required API Scopes / Permissions

- **Vertex AI API**: GCPプロジェクトで有効化が必要
- **認証**: ADC（Application Default Credentials）
- **IAMロール**: `roles/aiplatform.user`（Vertex AIエンドポイント呼び出し）
- **課金**: Gemini API呼び出し料 + Google Search Grounding利用料

## 6. Series Placement

- **Series**: util-series
- **Reason**: パイプ向けデータ処理CLI。JSON/Markdown出力、stdin/stdout対応で他ツールとのパイプライン連携を前提とした汎用ツール。Vertex AIという確立された基盤を使い、gem-cliでのGrounding実績もあるため、lab-series（実験段階）ではなくutil-seriesが適切。

## 7. External Platform Constraints

### Vertex AI API

- **Grounding + JSON mode 非併用**: Grounding有効時は`ResponseSchema`（structured output）が使えない。gem-cliと同じくクライアント側でJSON構築が必要。
- **課金**: Google Search Grounding はAPI呼び出しとは別に課金される。
- **リージョン制限**: Grounding対応リージョンが限定される可能性（`us-central1`はほぼ確実に対応）。
- **レートリミット**: Vertex AI Gemini API のデフォルトQPM/TPM制限に従う。
- **Groundingソースの扱い**: `grounding_metadata.grounding_chunks[].web`からURL/タイトルを抽出。Google利用規約に準拠（Brave APIのような攻撃的な利用制限はない）。

---

## Discussion Log

1. **前身プロジェクトの経緯**: `agentic-web-search`をBrave Search APIで実装したが、利用規約の攻撃的な制限（スニペット保存・再配布・AI学習禁止）と有償登録の心理的障壁で凍結。DuckDuckGoはrobots.txt違反で不採用。有償なら既にGCP課金基盤のあるVertex AI Web Groundingの方が合理的との判断。

2. **ツール名**: `gem-search` — gem-cliとの対称性、Geminiベースであることが明確。

3. **言語選択**: Go。Vertex AI専用機能のGo化を推進する方針。product-research/news-collector（Python）のパターンではなく、gem-cli（Go）のGrounding実装パターンを流用。シングルバイナリ配布がパイプラインパーツとして重要。

4. **モデル選択**: Web GroundingはGemini 2.5 Flashでも対応していることを確認（Pro専用ではない）。デフォルトは`gemini-2.5-flash`でコスト効率優先、`GEMSEARCH_MODEL`で変更可能。

5. **パイプライン連携**: MCP Tool化やライブラリ化は初期スコープ外。CLI + stdin/stdout のUNIX的な組み合わせをパイプライン連携手段とする。

6. **シリーズ配置**: util-series。Vertex AIという確立された基盤を使い、gem-cliでの実績もあるため実験段階（lab-series）ではない。

7. **Grounding制約**: Grounding + JSON mode非併用はgem-cliで既知・解決済みのパターン。クライアント側でJSON構築する方式を踏襲。
