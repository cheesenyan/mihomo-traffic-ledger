# Traffic Monitor v2.3.0

## Highlights

- 新增“细节 / 汇总”前端展示模式切换。
- 汇总模式按天生成紧凑聚合，不展示连接明细，适合大流量和长时间数据保留场景。
- 汇总模式下，一级列表和二级列表都使用条形图展示。
- 一级、二级列表统一限制为前 50 条，避免页面被海量主机或域名撑满。
- 升级后首次清理会自动执行一次 WAL checkpoint 和 VACUUM，回收旧版本遗留空间。

## Storage Changes

- 新增 `traffic_summary` 表，按天聚合设备、主机、代理及必要的二级维度。
- `traffic_summary` 保留天数为当前“日志保留天数” × 4。
- 现有分钟级 `traffic_aggregated` 采集逻辑保持不变，详情模式继续使用原有数据。
- 后台每 10 分钟增量生成汇总数据，启动时会自动回填已有聚合数据。

## UI Changes

- 右上角新增“细节 / 汇总”切换。
- 汇总模式隐藏“当前看板上下文”和连接明细模块。
- 汇总模式趋势图默认范围更大，并按汇总保留期展示。
- 一级列表和二级列表在汇总模式中使用横向条形图。

## Upgrade Notes

- 直接替换二进制或升级容器镜像即可。
- 首次启动会自动创建并回填 `traffic_summary`。
- 首次采集清理会自动执行 `PRAGMA wal_checkpoint(TRUNCATE)` 和 `VACUUM`。
- 如果需要立即手动压缩现有数据库，可以在停服后执行：

```bash
sqlite3 /data/traffic_monitor.db "PRAGMA wal_checkpoint(TRUNCATE); VACUUM;"
```
