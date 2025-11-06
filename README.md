# Resty-Stress-Tester 🚀

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/budyaya/resty-stress-tester/actions/workflows/ci.yml/badge.svg)](https://github.com/budyaya/resty-stress-tester/actions)
[![Release](https://img.shields.io/github/v/release/budyaya/resty-stress-tester)](https://github.com/budyaya/resty-stress-tester/releases)

一个基于 **go-resty** 的高性能 HTTP 压测工具，支持 CSV 参数化、实时统计和详细报告生成。

## ✨ 特性

- 🚀 **高性能**: 基于 go-resty，支持高并发压测
- 📊 **详细报告**: 实时统计、多种报告格式（控制台、JSON、HTML）
- 🔧 **参数化测试**: 支持 CSV 文件动态参数替换
- ⚡ **灵活配置**: 支持基于时长或请求数量的测试
- 🐳 **容器化**: 提供 Docker 镜像，开箱即用
- 📈 **实时监控**: 进度显示、性能指标实时更新
- 🔍 **错误分析**: 详细的错误分类和统计
- 🛠️ **易于扩展**: 模块化设计，易于定制和扩展

## 🚀 快速开始

### 安装

#### 从源码安装

```bash
git clone https://github.com/budyaya/resty-stress-tester
cd resty-stress-tester
go build -ldflags "-w -s" -o rst ./cmd/rst/
go build -ldflags "-w -s" -o rst.exe .\cmd\rst\
```

### 基本用法

```bash
# 简单 GET 请求测试
rst -url https://api.example.com/users -n 1000 -c 10 -t 1s -v

# 基于时长的测试
rst -url https://api.example.com/users -c 50 -d 1m

# POST 请求测试
rst -url https://api.example.com/users \
  -method POST \
  -body '{"name":"test","email":"test@example.com"}' \
  -n 5000 -c 50
```

### 高级用法：CSV 参数化

创建 CSV 文件 `users.csv`:
```csv
id,username,email,token
1,john_doe,john@example.com,token123
2,jane_smith,jane@example.com,token456
```

运行参数化测试:
```bash
rst -url "https://api.example.com/users/{{id}}" \
  -method GET \
  -csv users.csv \
  -headers '{"Authorization":"Bearer {{token}}","X-User-ID":"{{id}}"}' \
  -body '{"username":"{{username}}","email":"{{email}}"}' \
  -n 10000 -c 100 \
  -output results.json \
  -report json
```

## 📊 报告示例

```
=== HTTP STRESS TEST REPORT ===
Target URL:          https://api.example.com/users
HTTP Method:         GET
Concurrency:         100
Total Requests:      10000
Actual Duration:     15.23s
Successful:          9850
Failed:              150
Success Rate:        98.50%
Requests/sec:        656.86
Avg Response Time:   152ms
Min Response Time:   45ms
Max Response Time:   2.1s

Status Code Distribution:
  200: 9850 (98.50%)
  500: 150 (1.50%)
```

## 🔧 配置选项

| 参数 | 缩写 | 默认值 | 描述 |
|------|------|--------|------|
| `--url` | `-u` | - | 目标 URL (必需) |
| `--method` | `-X` | GET | HTTP 方法 |
| `--requests` | `-n` | 1000 | 总请求数 |
| `--concurrency` | `-c` | 10 | 并发数 |
| `--duration` | `-d` | - | 测试时长 (如 30s, 5m) |
| `--csv` | - | - | CSV 参数文件 |
| `--body` | `-b` | - | 请求体 |
| `--headers` | `-H` | - | 请求头 (JSON 格式) |
| `--timeout` | `-t` | 30s | 请求超时时间 |
| `--output` | `-o` | - | 输出文件 |
| `--report` | - | console | 报告格式 (console, json, html) |
| `--verbose` | `-v` | false | 详细输出 |
| `--version` | - | - | 显示版本信息 |

## 🏗️ 项目结构

```
resty-stress-tester/
├── cmd/rst/main.go          # 命令行入口
├── internal/                # 内部包
│   ├── config/             # 配置管理
│   ├── engine/             # 压测引擎
│   ├── parser/             # 数据解析器
│   ├── reporter/           # 报告生成器
│   └── util/               # 工具函数
├── pkg/                    # 公共包
│   ├── types/              # 类型定义
│   └── version/            # 版本信息
├── examples/               # 使用示例
├── test/                   # 测试文件
└── docs/                   # 文档
```

## 🧪 测试

```bash
# 运行单元测试
make test

# 运行集成测试
make test-integration

# 运行所有测试并生成覆盖率报告
make test-coverage
```

## 🤝 贡献

欢迎贡献！请阅读 [贡献指南](CONTRIBUTING.md) 了解如何参与项目开发。

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [go-resty](https://github.com/go-resty/resty) - 优秀的 Go HTTP 客户端库
- 所有贡献者和用户

---

**Resty-Stress-Tester** - 让 HTTP 压测变得简单高效！ 🚀
```
