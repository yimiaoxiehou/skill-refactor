---
name: sample-diagrams-generator-go
description: "Generate sample diagram files for testing diagram renderer. Creates Mermaid, Graphviz, PlantUML, and SVG samples. 触发词：生成示例图表、测试图表。配合 render_diagram 使用。"
description_zh: "生成示例图表文件（Go 原生），用于测试图表渲染工具"
version: 1.0.0
allowed-tools: Bash
---

# Sample Diagrams Generator (Go)

生成 Mermaid、Graphviz (DOT)、PlantUML、SVG 格式的示例图表文件，用于测试 `render_diagram` 工具。

> **源项目**: [drawio-skill](https://github.com/Agents365-ai/drawio-skill) by Agents365-ai。辅助工具，用 Go 重写。

## 快速开始

```bash
# 编译
cd create_sample_diagrams && go build -o create_sample_diagrams .

# 生成示例文件到 samples/ 目录
./create_sample_diagrams
```

生成的文件：
- `samples/sample_flow.mmd` — Mermaid 流程图
- `samples/sample_graph.dot` — Graphviz 有向图
- `samples/sample_sequence.puml` — PlantUML 时序图
- `samples/sample.svg` — 简单 SVG

## 相关工具

- `render_diagram` — 图表渲染工具
