#!/bin/bash

# 启动 minikube（如果还没启动）
minikube status || minikube start

# 设置 Docker 环境以使用 minikube 的 Docker daemon
eval $(minikube docker-env)

# 构建应用镜像
echo "Building application image..."
docker build -t agricultural-vision:latest .

echo "Load image to minikube"
minikube image load agricultural-vision:latest

# 创建命名空间
kubectl create namespace agricultural-vision --dry-run=client -o yaml | kubectl apply -f -

# 设置默认命名空间
kubectl config set-context --current --namespace=agricultural-vision

# 部署基础配置
echo "Deploying base configurations..."
kubectl apply -f k8s/base/

# 部署 MySQL
echo "Deploying MySQL..."
kubectl apply -f k8s/mysql/

# 部署 Redis
echo "Deploying Redis..."
kubectl apply -f k8s/redis/

# 等待 MySQL 和 Redis 就绪
echo "Waiting for MySQL and Redis to be ready..."
kubectl wait --for=condition=ready pod -l app=mysql --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis --timeout=60s

#kubectl apply -f k8s/app/
kubectl apply -f k8s/app/deployment.yaml

# 等待部署完成
echo "Waiting for deployment to be ready..."
kubectl rollout status deployment/agricultural-vision

# 查看 Pod 状态
echo "Pod status:"
kubectl get pods -l app=agricultural-vision -o wide

# 部署 Envoy
kubectl apply -f k8s/envoy/


# 5. 查看 Envoy 状态
kubectl get pods -n agricultural-vision -l app=envoy
kubectl get svc envoy-service -n agricultural-vision

# 等待所有部署完成
echo "Waiting for all deployments to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment --all

# 显示部署状态
echo "Deployment status:"
kubectl get all

# 获取 Envoy 服务的 URL
echo ""
echo "Getting service URL..."
minikube service envoy-service --url -n agricultural-vision