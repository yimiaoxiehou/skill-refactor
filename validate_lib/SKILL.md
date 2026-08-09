---
name: cosmic-lib-validator-go
description: "验证金蝶 COSMIC（苍穹）平台 .lib 文件的合规性（Go 原生）。检查 XML 格式、引用存在性、行尾格式等。触发词：验证 lib、检查 lib、lib 合规。配合 generate_lib 使用。"
description_zh: "验证 COSMIC 平台 .lib 文件合规性（Go 原生二进制）"
version: 1.0.0
allowed-tools: Bash
---

# COSMIC Lib Validator (Go)

金蝶 COSMIC（苍穹）平台 `.lib` 文件验证工具。检查 XML 合法性、引用文件存在性、格式规范等。

> **源项目**: [cosmic-lib-generator](https://github.com/yimiaoxiehou/cosmic-lib-generator)，用 Go 重写以节省运行资源。

## 快速开始

```bash
# 编译
cd validate_lib && go build -o validate_lib .

# 验证 lib 文件
./validate_lib dicj_bcjc.lib --cus-dir ./cus

# 验证补丁 lib
./validate_lib cus/cus.lib --cus-dir ./cus
```

## 验证检查项

| # | 检查项 | 说明 |
|---|--------|------|
| 1 | XML 合法性 | 文件是否为合法 XML |
| 2 | 必需元素 | 是否存在 `<libs>` / `<lib>` 元素 |
| 3 | 行尾格式 | full/group: CRLF，patch: LF |
| 4 | Cloud 匹配 | `<cloud>` 值是否与文件名一致 |
| 5 | 引用存在 | 引用的 xml/zip 文件是否存在于 cus 目录 |
| 6 | 逗号规范 | 逗号分隔条目是否有多余空格 |

## 参数

| 参数 | 说明 |
|------|------|
| `<lib_file>` | 要验证的 .lib 文件路径 |
| `--cus-dir <dir>` | cus 自定义组件目录（默认为 lib 文件同级的 `cus/`） |

## 相关工具

- `generate_lib` — 生成 .lib 文件
