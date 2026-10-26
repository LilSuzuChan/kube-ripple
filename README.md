# kube-ripple

根拠付き・読み取り専用の Kubernetes 診断 CLI。初期プロトタイプです。

## ローカル実行

Go 1.24 以上を使用します。サンプル実行にはクラスタ、GPU、Docker は不要です。

```sh
go mod download
go run . --before examples/before.json --after examples/after.json
go run . --before examples/before.json --after examples/after.json --output json
make test
make build
```

## 結果と制限

- `confirmed`: 入力に存在する観測事実。
- `predicted`: 入力が対象を網羅するという前提での静的判定。
- `unknown`: 対象・データ不足、未対応条件。
- 終了コード 0 は診断処理完了を意味し、正常性の保証ではありません。入力・引数エラーは 1。
- `--output text|json|markdown`。現段階の Markdown はプレーンテキストと同じ内容です。
- Secret の値や任意の Event/Condition メッセージは出力しません。出力のリソース名やラベルに機密を含めないでください。
- スナップショットは複数 YAML ドキュメント、単一オブジェクト、Kubernetes List 形式に対応。入力は最大 32 MiB を想定。
- 各診断の具体的な未評価項目は `limitations` に出力します。
- 実クラスタ、実 GPU、Linux での実行検証は別途必要です。

詳細は [開発・権限メモ](docs/design.md) を参照してください。
