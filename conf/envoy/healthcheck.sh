#!/bin/sh
# 尝试连接到管理端口
if nc -z localhost 9901 2>/dev/null; then
    exit 0
else
    exit 1
fi