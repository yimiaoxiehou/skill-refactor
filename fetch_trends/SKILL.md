---
name: github-ai-trends-go
description: "Generate GitHub AI trending project reports as formatted text leaderboards. Go native binary — no Python/Node.js needed. Fetches top-starred AI/ML/LLM repos by daily, weekly, or monthly period. 触发词：AI趋势、AI日报、AI热点、GitHub AI、大模型项目。"
description_zh: "生成 GitHub AI 热门项目趋势排行榜（Go 原生二进制）"
version: 1.0.0
allowed-tools: Bash
---

# GitHub AI Trends (Go)

获取 GitHub AI/ML/LLM 项目趋势排行榜，输出格式化的 Markdown 排行榜。原生 Go 二进制，零运行时依赖。

> **源项目**: 基于 WorkBuddy Skill Marketplace 的 `github-ai-trends`，用 Go 重写以节省运行资源。

## 快速开始

```bash
# 编译
cd fetch_trends && go build -o fetch_trends .

# 本周 AI 趋势 Top 20
./fetch_trends --period weekly --limit 20

# 今日 AI 趋势，JSON 输出
./fetch_trends --period daily --json

# 本月 AI 趋势
./fetch_trends --period monthly --limit 30
```

## 参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--period` | `weekly` | 时间范围：`daily` / `weekly` / `monthly` |
| `--limit` | `20` | 返回项目数量 |
| `--token` | `$GITHUB_TOKEN` | GitHub PAT（提升 API 限额） |
| `--json` | `false` | 输出 JSON 而非格式化文本 |

## 原理

1. 通过 GitHub Search API 搜索 AI 相关关键词（ai, llm, gpt, agent, transformer, diffusion, rag, ml）
2. 按 topic 搜索（artificial-intelligence, llm, generative-ai, ai-agent）
3. 去重并按 star 数排序
4. 输出格式化 Markdown 排行榜

## 注意事项

- 无 token 时 API 限制 10次/分钟，带 token 可到 30次/分钟
- 零外部依赖，仅使用 Go 标准库
- 输出可直接用于聊天展示
