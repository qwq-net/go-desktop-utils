# Desktop Widget

Windows デスクトップに常駐するシステムモニター & マーケットデータウィジェットです。
透過レイヤードウィンドウで表示し、デスクトップ背景への埋め込みにも対応しています。

## 表示項目

### System

| 項目 | 表示内容                                                 | 更新間隔 |
| ---- | -------------------------------------------------------- | -------- |
| CPU  | 使用率 (%) + バーグラフ                                  | 5秒      |
| MEM  | 使用率 (%) + 使用量/総量 (GB) + バーグラフ               | 5秒      |
| GPU  | 使用率 (%) + バーグラフ                                  | 5秒      |
| VRAM | 使用率 (%) + 使用量/総量 (GB) + バーグラフ               | 5秒      |
| DISK | ドライブごとの使用率 (%) + 使用量/総量 (GB) + バーグラフ | 60秒     |
| NET  | ダウンロード / アップロード速度                          | 5秒      |

- GPU / VRAM は NVIDIA GPU のみ対応 (`nvidia-smi` を使用)。未検出時は自動で非表示になります。
- DISK は C: ~ F: ドライブを自動検出し、存在するドライブのみ表示します。
- バーグラフは使用率に応じて色が変化します (0-79%: 青, 80-94%: オレンジ, 95%+: 赤)。

### Exchange (為替レート)

基準通貨からの為替レートをグループ別に表示します。
データソース: [open.er-api.com](https://open.er-api.com/)

### Stocks (株価)

ティッカーシンボルごとの株価と変動率を表示します。
データソース: [Alpha Vantage](https://www.alphavantage.co/) (API キーが必要)

## 設定 (config.json)

実行ファイルと同じディレクトリに `config.json` を配置します。
存在しない場合は初回起動時にデフォルト設定で自動生成されます。
トレイアイコンの右クリックメニューから設定ファイルを開いたり、リロードできます。

### window

| キー        | 型     | デフォルト   | 説明                                                          |
| ----------- | ------ | ------------ | ------------------------------------------------------------- |
| `monitor`   | int    | `0`          | 表示するモニターのインデックス (0始まり)                      |
| `alignment` | string | `"topRight"` | 配置位置 (`topRight`, `topLeft`, `bottomRight`, `bottomLeft`) |
| `marginX`   | int    | `30`         | 画面端からの水平マージン (px)                                 |
| `marginY`   | int    | `30`         | 画面端からの垂直マージン (px)                                 |
| `width`     | int    | `420`        | ウィジェットの幅 (px)                                         |
| `height`    | int    | `1200`       | ウィジェットの高さ (px)                                       |
| `opacity`   | int    | `245`        | 不透明度 (0-255)                                              |

### font

| キー        | 型     | デフォルト | 説明                           |
| ----------- | ------ | ---------- | ------------------------------ |
| `family`    | string | `"Inter"`  | フォントファミリー             |
| `size`      | int    | `24`       | フォントサイズ (8-72)          |
| `boldTitle` | bool   | `true`     | セクションタイトルを太字にする |

### style

| キー             | 型     | デフォルト  | 説明                               |
| ---------------- | ------ | ----------- | ---------------------------------- |
| `textColor`      | string | `"#FFFFFF"` | テキスト色                         |
| `labelColor`     | string | `"#FFFFFF"` | セクションラベル色                 |
| `dimColor`       | string | `"#FFFFFF"` | 補助テキスト色                     |
| `errorColor`     | string | `"#FF8888"` | エラー表示色                       |
| `positiveColor`  | string | `"#88FF88"` | 正の変動色 (株価用)                |
| `negativeColor`  | string | `"#FF8888"` | 負の変動色 (株価用)                |
| `barBgColor`     | string | `"#444444"` | バーグラフ背景色                   |
| `barNormalColor` | string | `"#44AAFF"` | バーグラフ通常色 (0-79%)           |
| `barWarnColor`   | string | `"#FFAA44"` | バーグラフ警告色 (80-94%)          |
| `barCritColor`   | string | `"#FF4444"` | バーグラフ危険色 (95%+)            |
| `separatorColor` | string | `"#888888"` | セパレーター色                     |
| `sectionPadding` | int    | `16`        | セクション間の余白 (px)            |
| `linePadding`    | int    | `4`         | 行間の余白 (px)                    |
| `horizontalPad`  | int    | `16`        | 左右の余白 (px)                    |
| `barHeight`      | int    | `8`         | バーグラフの高さ (px)              |
| `textShadow`     | bool   | `false`     | テキストにドロップシャドウを付ける |
| `showSeparator`  | bool   | `false`     | セクション間にセパレーターを表示   |

### exchange

| キー             | 型     | デフォルト   | 説明                       |
| ---------------- | ------ | ------------ | -------------------------- |
| `enabled`        | bool   | `true`       | 為替レートセクションの表示 |
| `baseCurrency`   | string | `"JPY"`      | 基準通貨                   |
| `groups`         | array  | _(下記参照)_ | 通貨グループの配列         |
| `refreshMinutes` | int    | `60`         | 更新間隔 (分)              |

各グループは `name` (表示名) と `targets` (通貨コードの配列) を持ちます。

### stocks

| キー             | 型     | デフォルト                  | 説明                 |
| ---------------- | ------ | --------------------------- | -------------------- |
| `enabled`        | bool   | `false`                     | 株価セクションの表示 |
| `provider`       | string | `"alphavantage"`            | データプロバイダ     |
| `apiKey`         | string | `""`                        | API キー             |
| `symbols`        | array  | `["AAPL", "GOOGL", "MSFT"]` | ティッカーシンボル   |
| `refreshMinutes` | int    | `240`                       | 更新間隔 (分)        |
| `columns`        | int    | `2`                         | 表示カラム数 (1-4)   |

### system

| キー      | 型   | デフォルト | 説明                         |
| --------- | ---- | ---------- | ---------------------------- |
| `enabled` | bool | `true`     | システムセクション全体の表示 |
| `gpu`     | bool | `true`     | GPU / VRAM の表示            |
| `disk`    | bool | `true`     | ディスク使用量の表示         |
| `network` | bool | `true`     | ネットワーク速度の表示       |

## Task コマンド

[Task](https://taskfile.dev/) を使用してビルドや配布物の作成を行います。

| コマンド                   | 説明                                                    |
| -------------------------- | ------------------------------------------------------- |
| `task`                     | ビルド (デフォルト)                                     |
| `task build`               | `desktop-widget.exe` をビルド                           |
| `task dist`                | `dist/desktop-widget.zip` を作成                        |
| `task dist VERSION=v1.0.0` | バージョン付き zip (`desktop-widget-v1.0.0.zip`) を作成 |
| `task tidy`                | `go mod tidy` を実行                                    |
| `task clean`               | ビルド成果物を削除                                      |

## リリース手順

1. 変更をコミットしてプッシュ

   ```bash
   git add -A && git commit -m "リリース内容の説明" && git push
   ```

2. タグを作成してプッシュ

   ```bash
   git tag -a v1.x.x -m "v1.x.x"
   git push origin v1.x.x
   ```

3. 配布用 zip をビルド

   ```bash
   task dist VERSION=v1.x.x
   ```

4. GitHub Release を作成

   ```bash
   gh release create v1.x.x dist/desktop-widget-v1.x.x.zip \
     --title "v1.x.x" \
     --notes "リリースノート"
   ```
