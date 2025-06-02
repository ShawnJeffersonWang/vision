#!/bin/bash
set -e

# 等待 Kafka 服务就绪
until kafka-topics.sh --bootstrap-server kafka:9092 --list &> /dev/null; do
  echo "等待 Kafka 启动..."
  sleep 2
done

# 创建主题（替换为你的主题名称）
TOPIC="post-creation"
kafka-topics.sh --bootstrap-server kafka:9092 --create --topic ${TOPIC} --partitions 1 --replication-factor 1
echo "主题 ${TOPIC} 创建成功"