---
name: longbridge-earnings-collect-go
description: "Longbridge 财报数据并行收集器（Go 原生）。替代 collect.py，并发调用 longbridge CLI 获取财报分析所需的全部数据。触发词：财报收集、earnings collect、并行抓取财报数据。配合 longbridge-earnings skill 使用。"
description_zh: "Longbridge 财报数据并行收集器（Go 原生）"
version: 1.0.0
allowed-tools: Bash
---

# Longbridge Earnings Data Collector (Go)

并行调用 Longbridge CLI 收集财报分析所需的全部数据源，输出精简 JSON 摘要。

> **源项目**: [longbridge-earnings](https://github.com/yimiaoxiehou/skill-refactor/tree/main/longbridge-skills/longbridge-earnings) skill 的 `collect.py`，用 Go 重写以节省运行资源。

## 快速开始

```bash
# 编译
cd longbridge-earnings-collect && go build -o longbridge-earnings-collect .

# 收集基础数据（lite 模式）
./longbridge-earnings-collect TSLA.US

# 收集完整数据（含资产负债表、现金流、行业对比）
./longbridge-earnings-collect --full 700.HK
```

## 参数

| 参数 | 默认 | 说明 |
|------|------|------|
| `<SYMBOL>` | 必需 | 股票代码（TSLA.US / 700.HK / 600519.SH） |
| `--full` | false | 包含资产负债表、现金流、申报文件、行业估值、同行对比 |

## 收集的数据（lite 模式）

| 数据源 | CLI 命令 | 说明 |
|--------|---------|------|
| snapshot | financial-report snapshot | 最新财务快照 |
| is_qf | financial-report IS qf | 利润表（最近8季度） |
| consensus | consensus | 一致预期 |
| forecast_eps | forecast-eps | EPS预测 |
| quote | quote | 实时行情 |
| calc_index | calc-index | PE/PB/市值 |
| rating | institution-rating | 机构评级 |
| segments | business-segments | 业务分部 |
| news | news | 最新10条新闻 |
| kline | kline day 250 | K线数据 |

## 前置条件

- 已安装并配置 `longbridge` CLI（https://open.longbridge.com）
- 编译需要 Go 1.21+

## 相关 Skills

- [longbridge-earnings](../longbridge-skills/longbridge-earnings/) — 财报分析 skill
- [longbridge](../longbridge-skills/longbridge/) — Longbridge 平台主 skill
