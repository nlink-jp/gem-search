# アーキテクチャ: gem-search

> 最終更新: 2026-04-12

## なぜこのツールが必要か

本ツールは[agentic-web-search](https://github.com/nlink-jp/agentic-web-search)（アーカイブ済み）の後継。前プロジェクトはサードパーティの検索API（DuckDuckGo、次にBrave）でエージェンティックWeb検索を試みたが、両方とも不採用となった：

- **DuckDuckGo** — Web Search APIが存在しない。HTMLエンドポイントのrobots.txtは`Disallow: /`。利用はスクレイピングに該当。
- **Brave Search API** — 技術的には動作したが、利用規約のデータ再利用制限（保存・再配布・AI学習禁止）が攻撃的であり、個人ツールに有償登録が必要という心理的障壁もあった。

結論：どうせ有償なら、Vertex AI Web Groundingの方が合理的。公式のGoogle Cloud API機能であり、課金基盤は既に整っており、検索結果のLLM消費に関するToSの曖昧さがない（それが設計された用途だから）。

## なぜ固定3フェーズ（ラウンド数可変ではなく）

元の`agentic-web-search`は`--max-rounds`でLLMに検索終了の判断を委ねていた。Vertex AI Groundingではこれが機能しなかった：Groundingは1回のAPI呼び出しで完全な回答+ソースを返すため、LLMは常にラウンド1で「情報は十分」と判断した。

修正策は、同じことをもう1回やらせることではない。各フェーズが根本的に異なる質問をする：

| フェーズ | 目的 | スキップできない理由 |
|---------|------|-------------------|
| **Survey** | 全体像の把握 — 主要トピック、用語、主要ソースの特定 | 広さがなければ、トピックの側面を丸ごと見落とす |
| **Deep-dive** | 不足の深掘り — 具体的データ、一次ソース、専門家分析 | 深さがなければ、レポートは表面的なまま |
| **Verify** | 矛盾の検証 — 情報の矛盾、陳腐化、最新データの確認 | 検証がなければ、フェーズ1-2のエラーがそのまま伝播する |

フェーズ数は設定可能にしない。フェーズをスキップすると品質が特定の予測可能な方法で劣化する。3フェーズは徹底的な調査の最小構成。

## なぜVertex AI Web Grounding（別の検索APIではなく）

Groundingは検索+コンテンツ抽出+回答生成を1回のAPI呼び出しで実行する。これにより以下が不要になった：

- 検索エンジンパッケージ（`internal/search/`）
- robots.txt/SSRF防止付きWebフェッチャー（`internal/fetch/`）
- スニペットの保存・再配布に関するコンプライアンス問題
- 検索エンジンへのレートリミット

検索インフラ全体がGoogleの責任。我々のコードは、我々にしかできないことに集中する：3フェーズの調査戦略とレポート統合。

## なぜGo

- **シングルバイナリ配布** — パイプライン利用に不可欠（`echo "query" | gem-search --format json | jq`）
- **nlk統合** — guard（プロンプトインジェクション防御）、strip（thinking tags除去）
- **gem-cliの実績** — GroundingのGo実装パターンが既に検証済み
- **Vertex AIのGo化推進** — product-researchとnews-collectorはPython。gem-searchをGoで構築することでGo SDKのGroundingワークロード対応を検証

## セキュリティ

| 懸念 | 対策 |
|------|------|
| プロンプトインジェクション | 全4回のLLM呼び出しでユーザークエリをnlk/guardノンスタグXMLでラップ |
| SSRFリスクなし | URL取得なし — GroundingがWeb内部アクセスを処理 |
| robots.txt懸念なし | APIアクセスであり、スクレイピングではない |
| ToS曖昧さなし | Groundingは検索結果のLLM消費のために設計された機能 |
| 認証情報 | ADCのみ、コードや設定ファイルに秘密情報なし |
