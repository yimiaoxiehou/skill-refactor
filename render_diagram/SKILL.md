---
name: diagram-renderer-go
description: "Render Mermaid/Graphviz/PlantUML/SVG diagrams to PNG/SVG/PDF (Go native). Auto-detects diagram type. 触发词：渲染图表、导出图表、mermaid 转 png、graphviz 渲染。"
description_zh: "图表渲染工具（Go 原生）：Mermaid/Graphviz/PlantUML/SVG → PNG/SVG/PDF"
version: 1.0.0
allowed-tools: Bash
---

# Diagram Renderer (Go)

将 Mermaid、Graphviz (DOT)、PlantUML、SVG 图表渲染为 PNG/SVG/PDF 文件。自动检测图表类型。

> **源项目**: [drawio-skill](https://github.com/Agents365-ai/drawio-skill) by Agents365-ai。部分功能用 Go 重写以节省运行资源（渲染部分），完整的 draw.io 图表生成请使用原项目。

## 快速开始

```bash
# 编译
cd render_diagram && go build -o render_diagram .

# 渲染 Mermaid 图表为 SVG
./render_diagram diagram.mmd --format svg

# 渲染 Graphviz DOT 为 PNG
./render_diagram architecture.dot --format png

# 渲染 PlantUML 为 SVG
./render_diagram sequence.puml --format svg --out output/seq.svg

# 指定图表类型
./render_diagram input.txt --kind mermaid --format png
```

## 参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `<input>` | 必需 | 输入文件路径 |
| `--format` | `svg` | 输出格式：`svg` / `png` / `pdf` |
| `--out` | 自动 | 输出路径（默认：输入文件名.格式） |
| `--kind` | `auto` | 图表类型：`auto` / `mermaid` / `graphviz` / `plantuml` / `svg` |

## 支持的图表类型

| 类型 | 扩展名 | 依赖 |
|------|--------|------|
| Mermaid | `.mmd` `.mermaid` | `mmdc` (npm: @mermaid-js/mermaid-cli) |
| Graphviz | `.dot` `.gv` | `dot` (graphviz) |
| PlantUML | `.puml` `.plantuml` | `plantuml` 或 Java + `PLANTUML_JAR` |
| SVG | `.svg` | `rsvg-convert` 或 `inkscape`（非 SVG 输出时） |

## 相关工具

- `create_sample_diagrams` — 生成示例图表文件用于测试
- [drawio-skill](https://github.com/Agents365-ai/drawio-skill) — 完整的 draw.io 图表生成
