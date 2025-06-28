#!/bin/bash

# 监控 Pod 状态
watch -n 2 kubectl get pods -o wide

# 在另一个终端查看日志
# kubectl logs -f deployment/agricultural-vision
# kubectl logs -f deployment/envoy