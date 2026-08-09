---
name: cosmic-lib-generator-go
description: "生成金蝶 COSMIC（苍穹）平台自定义组件包的 .lib 文件（Go 原生）。支持 full/group/patch 三种类型。触发词：生成 lib、COSMIC lib、苍穹部署包、自定义组件包。配合 validate_lib 使用。"
description_zh: "生成 COSMIC 平台 .lib 文件（Go 原生二进制）"
version: 1.0.0
allowed-tools: Bash
---

# COSMIC Lib Generator (Go)

金蝶 COSMIC（苍穹）平台自定义组件包 `.lib` 文件的生成工具。支持三种 lib 类型。

> **源项目**: [cosmic-lib-generator](https://github.com/yimiaoxiehou/cosmic-lib-generator)，用 Go 重写以节省运行资源。

## 快速开始

```bash
# 编译
cd generate_lib && go build -o generate_lib .

# 全量部署 lib
./generate_lib --type full \
    --cloud dicj_bcjc \
    --appids dicj_basedata,dicj_innernotice \
    --refs cus/dicj_bcjc-dicj_bcjc_base,bcjc-dicj_innernotice \
    --output dicj_bcjc.lib

# 模块组 lib
./generate_lib --type group \
    --appids dicj_public,dicj_tyfw \
    --refs cus/dicj-public.xml,dicj-tyfw.xml \
    --output dicj_public.lib

# 增量补丁 lib
./generate_lib --type patch \
    --refs cus/dicj_bcjc-dicj_venuemonitor \
    --output cus/cus.lib
```

## 参数

| 参数 | 必需 | 说明 |
|------|------|------|
| `--type` | 是 | lib 类型：`full` / `group` / `patch` |
| `--cloud` | full | cloud 域标识 |
| `--appids` | full/group | 逗号分隔的 appId 列表 |
| `--refs` | 是 | 逗号分隔的引用列表 |
| `--output` | 是 | 输出文件路径 |

## lib 类型

| 类型 | 结构 | 适用场景 |
|------|------|---------|
| `full` | `<cloud>` + `<appIds>` + `<libs>` | 全量部署 |
| `group` | `<appIds>` + `<libs>` | 模块分组 |
| `patch` | `<libs>` | 增量补丁 |

## 相关工具

- `validate_lib` — 验证生成的 .lib 文件合规性
