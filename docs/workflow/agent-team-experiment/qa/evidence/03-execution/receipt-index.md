# QA v4 脱敏 receipt 索引

正式候选 F：`1771d3037afa4c06187378171238bd1aa00d058219a231dcce0fc402ba59fa0d`；run ID：`20261101T063742Z`。原件位于 Git ignored 的 `test-results/qa-contract/<F>/<run-id>/<matrix-id>/1/`，每项含 `command.txt`、UTC `start.txt`/`end.txt`、`stdout.txt`、`stderr.txt`、`exit.txt` 和 `hashes.txt`。

| ID | exit | stdout bytes / SHA-256 | stderr bytes / SHA-256 |
|---|---:|---|---|
| A01 | 0 | 39 / `32a8e411…` | 0 / `e3b0c442…` |
| A02 | 1 | 1991 / `4604cb33…` | 0 / `e3b0c442…` |
| A03 | 1 | 1991 / `352b9c61…` | 0 / `e3b0c442…` |
| A04 | 0 | 39 / `ad77236c…` | 0 / `e3b0c442…` |
| A05 | 0 | 40 / `b1ba9180…` | 0 / `e3b0c442…` |
| C01 | 0 | 39 / `3a95e49f…` | 0 / `e3b0c442…` |
| C02 | 0 | 39 / `13979f21…` | 0 / `e3b0c442…` |
| C03 | 0 | 39 / `44675976…` | 0 / `e3b0c442…` |
| C04 | 0 | 39 / `0b78ea73…` | 0 / `e3b0c442…` |
| C05 | 0 | 39 / `c7ea7d3e…` | 0 / `e3b0c442…` |
| C06 | 0 | 39 / `76df946c…` | 0 / `e3b0c442…` |
| C07 | 0 | 39 / `76cb5207…` | 0 / `e3b0c442…` |
| B01 | 0 | 39 / `51c156ec…` | 0 / `e3b0c442…` |
| B02 | 0 | 39 / `f0640bd3…` | 0 / `e3b0c442…` |
| B03 | 0 (coverage invalid) | 39 / `a79f1961…` | 0 / `e3b0c442…` |
| B04 | 1 (coverage invalid) | 378 / `a4b0d9a8…` | 0 / `e3b0c442…` |
| B05 | 1 | 2748 / `7719fb3b…` | 0 / `e3b0c442…` |

完整 64 位哈希、命令和闭合时间保存在对应 ignored receipt；本索引有意不复制完整原始日志。X01–X03 因同版门禁阻断，无 receipt。
