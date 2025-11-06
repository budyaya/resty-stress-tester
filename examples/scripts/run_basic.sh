#!/bin/bash

# 基础压测示例脚本
echo "Running basic stress test..."

./bin/rst \
  -url "https://httpbin.org/get" \
  -method GET \
  -requests 1000 \
  -concurrency 10 \
  -timeout 30s \
  -verbose

echo "Basic test completed!"