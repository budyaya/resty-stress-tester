#!/bin/bash

# 高级压测示例脚本
echo "Running advanced stress test with CSV parameterization..."

./bin/rst \
  -url "https://httpbin.org/anything/user/{{user_id}}" \
  -method POST \
  -requests 5000 \
  -concurrency 50 \
  -csv ../advanced/users.csv \
  -body '{
    "user_id": "{{user_id}}",
    "username": "{{username}}",
    "email": "{{email}}",
    "category": "{{category}}",
    "timestamp": "2024-01-01T00:00:00Z"
  }' \
  -headers '{
    "Content-Type": "application/json",
    "Authorization": "Bearer {{user_id}}",
    "X-Request-ID": "req-{{user_id}}"
  }' \
  -output ../advanced/results.json \
  -report json \
  -verbose

echo "Advanced test completed!"