#!/bin/bash

# 保活机制：使用supervisor简化进程守护（Alpine内置无supervisor，用简单loop实现）
function start_proxy() {
  while true; do
    /usr/bin/proxy-bin  # 调用Go binary生成配置并启动Xray
    echo "Proxy crashed. Restarting in 5s..."
    sleep 5
  done
}

# 资源监控：简单检查内存，如果高则重启（可选，Koyeb会自动杀）
function monitor_resources() {
  while true; do
    MEM=$(free -m | awk '/Mem/{print $3}')
    if [ "$MEM" -gt 200 ]; then  # 阈值低于Koyeb限
      echo "High memory. Restarting..."
      killall xray
    fi
    sleep 60
  done
}

# 主逻辑
monitor_resources &  # 后台监控
start_proxy