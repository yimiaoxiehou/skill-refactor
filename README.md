# skill-refactor

WorkBuddy Skills 的 Go 语言重构合集。将原有的 Python / Node.js / Shell 脚本重构为独立 Go 二进制文件，合并到一个统一的 monorepo 中。

## 重构目的

**本次重构的唯一目的：节省运行资源。**

原始 Skills 依赖 Python、Node.js 等外部运行时，每个 Skill 都需要独立的运行环境，导致资源占用较大。重构为 Go 原生二进制后：

- 零运行时依赖（除 imap-smtp 需少量 Go 库外）
- 单个静态二进制，极低内存占用
- 启动速度快，适合 WorkBuddy Skill 场景下的频繁调用
- 统一代码风格和维护方式

功能逻辑与原版保持等价，无新增特性。

## 包含的工具

| 二进制 | 原始 Skill | 原始实现 | 功能 |
|--------|-----------|---------|------|
| `fetch_trends` | GitHub AI趋势追踪 (`github-ai-trends`) | Go | 获取 GitHub AI/ML/LLM 项目趋势排行榜 |
| `github_trending` | GitHub热门项目 (`github-trending-cn`) | Go | 获取 GitHub 热门项目（支持语言过滤） |
| `generate_lib` | COSMIC Lib Generator (`cosmic-lib-generator`) | Go | 生成金蝶 COSMIC 平台 .lib 文件 |
| `validate_lib` | COSMIC Lib Generator (`cosmic-lib-generator`) | Go | 验证 COSMIC .lib 文件合规性 |
| `render_diagram` | Draw.io Diagrams (`drawio-skill`) | Python | 渲染 Mermaid/Graphviz/PlantUML/SVG 图表 |
| `create_sample_diagrams` | Draw.io Diagrams (`drawio-skill`) | Python | 生成示例图表文件 |
| `imap-smtp-email` | IMAP/SMTP Email (`imap-smtp-email`) | Go | IMAP 收件 / SMTP 发件完整邮件工具 |
| `mcp_bridge` | Burp MCP Bridge | Python | Burp Suite MCP 协议桥接 |
| `add_methods` | Burp MCP Helper | Python | Burp MCP 扩展方法注入 |

## 源项目

本项目中的工具分别改写自以下原始项目：

| 工具 | 原始项目 | 作者/来源 |
|------|---------|----------|
| `fetch_trends` | GitHub AI Trends Skill | Skill Marketplace |
| `github_trending` | GitHub Trending CN Skill | Skill Marketplace |
| `generate_lib` / `validate_lib` | [cosmic-lib-generator](https://github.com/yimiaoxiehou/cosmic-lib-generator) | yimiaoxiehou |
| `render_diagram` | [drawio-skill](https://github.com/Agents365-ai/drawio-skill) | Agents365-ai |
| `imap-smtp-email` | [imap-smtp-email](https://clawhub.ai) | gzlicanyi |
| `mcp_bridge` / `add_methods` | Burp MCP Bridge | MCP Community |

## 使用

```bash
# 编译所有工具
go build -o bin/ ./...

# GitHub AI 趋势
./bin/fetch_trends --period weekly --limit 20

# GitHub 热门项目
./bin/github_trending --period daily --language go

# 生成 COSMIC lib 文件
./bin/generate_lib --type full --cloud mycloud --appids app1,app2 --refs cus/module --output output.lib

# 验证 lib 文件
./bin/validate_lib output.lib --cus-dir ./cus

# 渲染图表
./bin/render_diagram input.mmd --format svg

# 邮件操作
./bin/imap-smtp-email check --limit 10
./bin/imap-smtp-email send --to user@example.com --subject "Hello" --body "Test"

# Burp MCP 桥接
./bin/mcp_bridge
```

## IMAP/SMTP 配置

`imap-smtp-email` 需要配置环境变量，支持两种配置模式：

**共享模式** (推荐，`~/.config/mail-skills/.env`)：
```env
# 默认账户（以 PROVIDER 预设自动推导服务器地址）
PROVIDER=163
USERNAME=user@163.com
PASSWORD=your_password

# 多账户支持
ANOTHER_PROVIDER=gmail
ANOTHER_USERNAME=user@gmail.com
ANOTHER_PASSWORD=your_password
```

**传统模式** (`~/.config/imap-smtp-email/.env`)：
```env
IMAP_HOST=imap.163.com
IMAP_PORT=993
IMAP_USER=user@163.com
IMAP_PASS=your_password
IMAP_TLS=true

SMTP_HOST=smtp.163.com
SMTP_PORT=465
SMTP_USER=user@163.com
SMTP_PASS=your_password
SMTP_SECURE=true
```

支持的邮箱服务商：163、126、188、QQ、企业邮箱、Gmail、Outlook、iCloud、Fastmail 等。

## 构建

```bash
# 构建所有工具（除 imap-smtp）
go build -o bin/ ./fetch_trends ./github_trending ./generate_lib ./validate_lib ./render_diagram ./create_sample_diagrams ./mcp_bridge ./add_methods

# 构建 imap-smtp（需要外部依赖）
cd imap_smtp && go build -o ../bin/imap-smtp-email .
```

## 许可证

本项目中的代码均为对原始项目功能等价改写，原始项目的许可证信息请参阅各源项目仓库。

---

*重构仅出于节省运行资源的目的，功能与原版保持一致。*
