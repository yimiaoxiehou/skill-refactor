---
name: imap-smtp-email-go
description: "IMAP/SMTP email tool (Go native). Read, search, and manage email via IMAP. Send email via SMTP. Supports Gmail, Outlook, 163/126/188/QQ and more. Multi-account support. 触发词：查邮件、发邮件、收件箱、未读邮件、搜索邮件。"
description_zh: "IMAP/SMTP 邮件工具（Go 原生）：收件、发件、搜索、多账户"
version: 1.0.0
allowed-tools: Bash
metadata:
  requires:
    env:
      - PROVIDER
      - USERNAME
      - PASSWORD
    primaryEnv: PROVIDER
---

# IMAP/SMTP Email Tool (Go)

通过 IMAP 协议收发邮件。支持 Gmail、Outlook、163、126、188、QQ 邮箱等。原生 Go 实现，使用少量的 Go 依赖（go-imap, go-message, godotenv）。

> **源项目**: [imap-smtp-email](https://clawhub.ai) by gzlicanyi。用 Go 重写以统一工具链并减少 Node.js 依赖。

## 配置

配置文件位于 `~/.config/mail-skills/.env`（共享格式，推荐）或 `~/.config/imap-smtp-email/.env`（传统格式）。

### 共享格式（推荐）

```bash
# 默认账户
PROVIDER=163
USERNAME=your@163.com
PASSWORD=your_password

# 文件访问白名单
ALLOWED_READ_DIRS=~/Downloads,~/Documents
ALLOWED_WRITE_DIRS=~/Downloads
```

### 传统格式

```bash
IMAP_HOST=imap.163.com
IMAP_PORT=993
IMAP_USER=your@163.com
IMAP_PASS=your_password
IMAP_TLS=true

SMTP_HOST=smtp.163.com
SMTP_PORT=465
SMTP_USER=your@163.com
SMTP_PASS=your_password
SMTP_SECURE=true
```

## 编译

```bash
cd imap_smtp && go build -o imap-smtp-email .
```

## IMAP 命令（收件）

```bash
# 检查收件箱
./imap-smtp-email check --limit 10

# 只查未读
./imap-smtp-email check --unseen --limit 20

# 搜索邮件
./imap-smtp-email search --from sender@example.com --subject "关键词"

# 获取邮件详情
./imap-smtp-email fetch <uid>

# 标记已读/未读
./imap-smtp-email mark-read <uid> [uid2 ...]
./imap-smtp-email mark-unread <uid> [uid2 ...]

# 列出邮箱文件夹
./imap-smtp-email list-mailboxes
```

## SMTP 命令（发件）

```bash
# 发送纯文本邮件
./imap-smtp-email send --to user@example.com --subject "Hello" --body "World"

# 发送 HTML 邮件
./imap-smtp-email send --to user@example.com --subject "News" --html --body "<h1>Title</h1>"

# 测试 SMTP 连接
./imap-smtp-email test
```

## 多账户

在配置文件中添加前缀账户：

```bash
# 工作账户
WORK_PROVIDER=gmail
WORK_USERNAME=me@company.com
WORK_PASSWORD=app_password
```

使用时加 `--account` 参数：

```bash
./imap-smtp-email --account work check
```

## 支持的邮箱

| 服务商 | IMAP 地址 | SMTP 地址 |
|--------|----------|----------|
| 163.com | imap.163.com:993 | smtp.163.com:465 |
| 126.com | imap.126.com:993 | smtp.126.com:465 |
| 188.com | imap.188.com:993 | smtp.188.com:465 |
| QQ Mail | imap.qq.com:993 | smtp.qq.com:587 |
| Gmail | imap.gmail.com:993 | smtp.gmail.com:587 |
| Outlook | outlook.office365.com:993 | smtp.office365.com:587 |

## 配置迁移

如果有旧的配置文件，使用迁移工具：

```bash
cd imap_smtp && go build -o migrate-legacy-config ./migrate/
./migrate-legacy-config ~/.config/imap-smtp-email/.env
```
