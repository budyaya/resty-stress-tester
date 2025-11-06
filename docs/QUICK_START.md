# 快速开始指南

Resty-Stress-Tester 是一个基于 go-resty 的高性能 HTTP 压测工具。

## 安装

### 从源码构建

```bash
# 克隆项目
git clone https://github.com/budyaya/resty-stress-tester
cd resty-stress-tester

# 构建
make build

# 安装到系统
sudo make install
```

### 使用 Docker

```bash
# 构建 Docker 镜像
docker build -t rst .

# 运行测试
docker run -v $(pwd):/data rst -url https://api.example.com/users -n 1000 -c 10
```

## 基本用法

### 简单压测

```bash
# 1000 个 GET 请求，10 个并发
rst -url https://api.example.com/users -n 1000 -c 10

# 基于时间的测试（运行 1 分钟）
rst -url https://api.example.com/users -c 50 -d 1m
```

### POST 请求测试

```bash
rst -url https://api.example.com/users \
  -method POST \
  -body '{"name":"test","email":"test@example.com"}' \
  -headers '{"Content-Type":"application/json"}' \
  -n 5000 -c 50
```

### 使用 CSV 参数化

```bash
rst -url "https://api.example.com/users/{{user_id}}" \
  -method GET \
  -csv users.csv \
  -headers '{"Authorization":"Bearer {{token}}"}' \
  -n 10000 -c 100
```

## 输出示例

```
=== Stress Test Report ===
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
```

## 下一步

- 查看 [高级用法](ADVANCED_USAGE.md) 了解更复杂的测试场景
- 学习 [CSV 格式说明](CSV_FORMAT.md) 了解参数化测试
- 查看 [API 参考](API_REFERENCE.md) 了解所有可用选项
