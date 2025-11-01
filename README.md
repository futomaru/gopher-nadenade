# gopher-nadenade

ブラウザで Go 製 Gopher をなでるだけのミニマムアプリです。UI ロジックは Go で記述し、WASM にコンパイルして配信します。生成物は `web/` ディレクトリにまとまっており、静的ファイルをそのままホスティングできます。

## 前提

- Go 1.21 以上（WASM ターゲットを含む標準ツールチェーン）

## ビルド & 実行

```sh
# WASM をビルド（web/main.wasm が生成されます）
make wasm

# ローカルサーバを起動
make serve
```

`http://localhost:8080` にアクセスすると、Gopher をなでて表情を変えられます。

- `make wasm` は `web/wasm_exec.js` も併せてコピーします。Go のインストール先で `wasm_exec.js` が見つからない場合は、適宜ダウンロードして `web/` に配置してください。

## ディレクトリ構成

- `cmd/server` — 単純な静的ファイルサーバ
- `wasm` — ブラウザ側のロジック（Go -> WASM）
- `web` — 配信する静的ファイル（`make wasm` の成果物を含む）
