---
name: burp-mcp-bridge-go
description: "Burp Suite MCP protocol bridge (Go native). JSON-RPC bridge between MCP clients and Burp Suite's HTTP API. 触发词：Burp MCP、Burp 桥接、Burp Suite 集成。配合 add_methods 使用。"
description_zh: "Burp Suite MCP 协议桥接（Go 原生）"
version: 1.0.0
allowed-tools: Bash
---

# Burp MCP Bridge (Go)

将 MCP (Model Context Protocol) JSON-RPC 请求桥接到 Burp Suite 的 HTTP API。使 AI 助手可以通过标准 MCP 协议与 Burp Suite 交互。

> **源项目**: MCP Community 的 Burp MCP Bridge。用 Go 重写以节省运行资源（原为 Python）。

## 前置条件

- Burp Suite 正在运行
- "MCP Full Control" 扩展已加载
- Burp MCP HTTP API 监听在 `127.0.0.1:9876`

## 快速开始

```bash
# 编译
cd mcp_bridge && go build -o mcp_bridge .

# 启动桥接（从 stdin 读取 JSON-RPC，输出到 stdout）
./mcp_bridge
```

## 工作原理

1. 启动时连接 Burp API (`http://127.0.0.1:9876/tools`) 获取可用工具列表
2. 监听 stdin 接收 MCP JSON-RPC 请求
3. 将请求转发到 Burp HTTP API
4. 返回 MCP 格式的响应到 stdout

## 支持的 MCP 方法

- `initialize` — MCP 握手
- `tools/list` — 列出 Burp 工具
- `tools/call` — 调用 Burp 工具（proxy_history, scan, send_request, etc.）

## 配置

环境变量（可选）：
- `BURP_MCP_HOST` — Burp 地址（默认 `127.0.0.1`）
- `BURP_MCP_PORT` — Burp 端口（默认 `9876`）

## 相关工具

- `add_methods` — Burp MCP 扩展方法注入工具
