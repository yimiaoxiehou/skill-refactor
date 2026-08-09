---
name: longbridge-skills
description: "Longbridge 金融数据平台 Skills 合集。包含行情、基本面、技术分析、量化、期权、投资研究等 13 个专业 Skill。所有 Skill 均为 prompt 文档 + Go 重构的数据收集器。"
description_zh: "Longbridge 金融平台 Skills 合集（prompt 文档 + Go 数据收集器）"
version: 1.0.0
---

# Longbridge Skills 合集

基于 [Longbridge 开放平台](https://open.longbridge.com) 的金融数据分析 Skill 集合。

## 包含的 Skills

| Skill | 类型 | 说明 |
|-------|------|------|
| [longbridge](longbridge/) | Hub | 主入口 — 投资分析工作流 |
| [longbridge-content](longbridge-content/) | 数据 | 新闻、公告、SEC 申报文件 |
| [longbridge-derivatives](longbridge-derivatives/) | 数据 | 期权链、Greeks、窝轮/牛熊证 |
| [longbridge-earnings](longbridge-earnings/) | 分析 | 财报前/后分析 |
| [longbridge-fundamentals](longbridge-fundamentals/) | 分析 | 财务报表、估值、主营业务 |
| [longbridge-intel](longbridge-intel/) | 情报 | 策略筛选、市场异动、晨报 |
| [longbridge-market-data](longbridge-market-data/) | 数据 | 实时行情、K线、盘口 |
| [longbridge-portfolio](longbridge-portfolio/) | 交易 | 持仓、订单、组合诊断 |
| [longbridge-quant](longbridge-quant/) | 量化 | 因子模型、配对交易、ML |
| [longbridge-research](longbridge-research/) | 研究 | 机构评级、内部人交易、空头 |
| [longbridge-technical](longbridge-technical/) | 技术 | K线形态、缠论、海龟交易 |
| [longbridge-value-investing](longbridge-value-investing/) | 价值 | 格雷厄姆/巴菲特方法论 |
| [longbridge-watchlist](longbridge-watchlist/) | 工具 | 自选股、价格提醒 |

## Go 重构

仅 `longbridge-earnings` 包含可执行代码（`collect.py`），已用 Go 重写为 [longbridge-earnings-collect](../longbridge-earnings-collect/)。其余 12 个 Skill 均为纯 prompt 文档，无需重构。

## 源项目

所有 Skill 来自 [Longbridge Open Platform](https://open.longbridge.com)，基于 [skill-refactor](https://github.com/yimiaoxiehou/skill-refactor) 项目统一管理。
