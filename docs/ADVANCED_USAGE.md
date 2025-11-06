# 高级用法

## 配置文件

您可以使用 YAML 或 JSON 配置文件：

```yaml
# config.yaml
url: "https://api.example.com/users/{{id}}"
method: "GET"
total_requests: 5000
concurrency: 50
timeout: 30s
csv_file: "users.csv"
headers:
  Content-Type: "application/json"
  Authorization: "Bearer {{token}}"
output_file: "results.json"
report_format: "json"
verbose: true
```

使用配置文件：
```bash
rst -config config.yaml
```

## 动态参数化

### CSV 文件格式

```csv
id,username,email,token,category
1,user1,user1@example.com,token1,premium
2,user2,user2@example.com,token2,standard
```

### 模板语法

在 URL、Headers 和 Body 中使用 `{{column_name}}` 引用 CSV 列：

```bash
rst -url "https://api.example.com/users/{{id}}/profile" \
  -method PUT \
  -csv users.csv \
  -body '{
    "user": {
      "id": {{id}},
      "username": "{{username}}",
      "email": "{{email}}"
    }
  }' \
  -headers '{
    "Content-Type": "application/json",
    "Authorization": "Bearer {{token}}",
    "X-User-Category": "{{category}}"
  }'
```

## 报告格式

### JSON 报告

```bash
rst -url https://api.example.com/users -n 1000 -c 10 -report json -output results.json
```

### HTML 报告

```bash
rst -url https://api.example.com/users -n 1000 -c 10 -report html -output report.html
```

## 性能调优

### 调整并发数

```bash
# 低并发，适合 API 有速率限制的情况
rst -url https://api.example.com/users -n 1000 -c 5

# 高并发，适合高性能后端
rst -url https://api.example.com/users -n 10000 -c 200
```

### 连接管理

```bash
# 禁用 Keep-Alive（模拟短连接）
rst -url https://api.example.com/users -n 1000 -c 10 -keep-alive=false

# 调整超时时间
rst -url https://api.example.com/users -n 1000 -c 10 -timeout 60s
```

## 监控和调试

### 详细日志

```bash
rst -url https://api.example.com/users -n 1000 -c 10 -verbose
```

### 实时进度

启用详细模式后，工具会显示实时进度：

```
Progress: 543/1000 (54.3%) - 156.7 req/sec
```

## 集成到 CI/CD

### 基本集成

```yaml
# GitHub Actions 示例
- name: Run API Stress Test
  run: |
    ./bin/rst \
      -url ${{ secrets.API_URL }} \
      -n 1000 \
      -c 10 \
      -timeout 30s
```

### 失败条件

工具会在错误率超过 10% 时返回非零退出码：

```bash
#!/bin/bash
if ./bin/rst -url https://api.example.com/users -n 1000 -c 10; then
  echo "Test passed"
else
  echo "Test failed - high error rate detected"
  exit 1
fi
```

## 最佳实践

1. **循序渐进**：从低并发开始，逐步增加
2. **监控资源**：注意客户端和服务器的资源使用情况
3. **参数化测试**：使用 CSV 文件模拟真实场景
4. **设置合理的超时**：避免测试卡住
5. **保存结果**：使用 JSON 输出进行后续分析
