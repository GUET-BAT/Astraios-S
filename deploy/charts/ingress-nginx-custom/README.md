# Ingress-Nginx Custom Setup

这是一个基于官方 `ingress-nginx` 的定制 Helm Chart，预配置了全局 SSL 证书和 HTTPS 终止逻辑。

## 核心功能

1.  **全局 SSL**: 自动为所有子域名应用 `ingress-nginx/global-tls` 证书。
2.  **SSL 终止**: 在 Ingress 层卸载 SSL，将请求作为明文 HTTP 转发给集群内服务。
3.  **强制 HTTPS**: 自动将所有 HTTP 请求重定向到 HTTPS。

## 部署步骤

### 1. 准备全局 SSL Secret

首先，确保在 `ingress-nginx` 命名空间中存在名为 `global-tls` 的 Secret。如果没有，可以使用以下命令创建（替换为你的证书文件）：

```bash
kubectl create namespace ingress-nginx --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret tls global-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  -n ingress-nginx
```

### 2. 下载依赖

在当前目录下执行：

```bash
helm dependency update
```

### 3. 安装 Chart

```bash
helm upgrade --install ingress-nginx-custom . \
  --namespace ingress-nginx \
  --create-namespace
```

## 注意事项

- **泛域名证书**: 为了让所有子域名都生效，建议在 `global-tls` 中使用泛域名证书（如 `*.yourdomain.com`）。
- **Ingress 资源**: 业务服务的 `Ingress` 资源不再需要定义 `tls` 部分，除非你想为特定域名覆盖全局证书。
