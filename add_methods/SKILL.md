---
name: burp-mcp-methods-go
description: "Inject additional methods into Burp MCP Server Java source (Go native). Adds burp_version, add_issue, proxy_history_filtered, http_handler, and proxy_rule capabilities. 触发词：Burp 扩展、MCP 方法注入。配合 mcp_bridge 使用。"
description_zh: "Burp MCP 扩展方法注入工具（Go 原生）"
version: 1.0.0
allowed-tools: Bash
---

# Burp MCP Methods Injector (Go)

向 Burp MCP Server 的 Java 源代码注入额外方法，扩展 Burp Suite 的 AI 交互能力。

> **源项目**: MCP Community。辅助工具，用 Go 重写。

## 注入的方法

| 方法 | 功能 |
|------|------|
| `burp_version` | 获取 Burp Suite 版本 |
| `add_issue` | 手动添加漏洞问题 |
| `proxy_history_filtered` | 按条件过滤代理历史 |
| `register_http_handler` | 注册 HTTP 请求自动修改规则 |
| `remove_http_handler` | 清除 HTTP 处理规则 |
| `register_proxy_rule` | 注册代理拦截规则 |
| `remove_proxy_rule` | 清除代理规则 |

## 使用

```bash
# 编译
cd add_methods && go build -o add_methods .

# 注入方法到 McpHttpServer.java
./add_methods /path/to/McpHttpServer.java
```

完成后重新编译 Burp MCP 扩展。

## 相关工具

- `mcp_bridge` — Burp MCP 协议桥接
