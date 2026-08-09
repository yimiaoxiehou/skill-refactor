---
name: github-trending-go
description: "GitHub Trending Monitor (Go native). Fetch GitHub trending repos by daily/weekly/monthly period using real GitHub Search API. Supports language filter. 触发词：GitHub、trending、开源、热门项目、今日热门、本周热门。"
description_zh: "获取 GitHub 热门项目（Go 原生，支持语言过滤）"
version: 1.0.0
allowed-tools: Bash
---

# GitHub Trending (Go)

获取 GitHub 真实热门项目数据，支持今日/本周/本月，可按编程语言过滤。原生 Go 二进制，零运行时依赖。

> **源项目**: 基于 WorkBuddy Skill Marketplace 的 `github-trending-cn`，用 Go 重写以节省运行资源。

## 快速开始

```bash
# 编译
cd github_trending && go build -o github_trending .

# 今日热门
./github_trending

# 本周热门 Top 25
./github_trending --period weekly --limit 25

# 本月 Python 热门项目
./github_trending --period monthly --language python

# JSON 输出
./github_trending --period daily --json
```

## 参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--period` | `daily` | 时间范围：`daily` / `weekly` / `monthly` |
| `--limit` | `20` | 返回项目数量 |
| `--language` | 全部 | 语言过滤：`python`、`go`、`rust`、`javascript` 等 |
| `--token` | `$GITHUB_TOKEN` | GitHub PAT |
| `--json` | `false` | 输出 JSON |

## 支持的编程语言

`python`, `javascript`, `typescript`, `go`, `rust`, `java`, `cpp`, `c`, `swift`, `kotlin`, `ruby`, `php`

## 注意事项

- 调用 GitHub Search API，按 `pushed + stars` 排序
- 多语言补充策略，确保结果多样性
- 无 token 时 10次/分钟，带 token 30次/分钟
- 零外部依赖，仅使用 Go 标准库
