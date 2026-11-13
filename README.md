# hls-live-window

读取游戏直播的 HLS 播放列表，形成可用于追赶直播边缘的片段时间线，只用 Go 标准库。

## 口径

- media sequence、片段时长、discontinuity 和 program date time 共同决定片段身份与时间线。
- discontinuity 只切换编码时代，不重置整个直播窗口的绝对位置。
- 时长按毫秒保留，连续片段的起点等于前一片段终点。
- 返回 `scanned`，解析量随片段数增长。

## 不变量

- 序号单调递增且不因 discontinuity 重复。
- 相邻片段不重叠，后片段起点等于前片段终点。
- program date time 只绑定到它修饰的片段。

## 输出契约

CLI 输出 JSON：`segments` 逐项包含 `sequence`、`uri`、`duration_ms`、`start_ms`、`discontinuity`，可选 `program_time`；另有 `scanned`。
