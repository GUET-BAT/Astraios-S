# Gateway Service 灰度发布指南

## 概述

Gateway Service 支持通过 Ingress-nginx 的 Canary 功能实现灰度发布，支持两种方式：
- **权重灰度**：按百分比分配流量
- **Header 灰度**：基于请求头匹配

## 架构说明

```
用户请求
    ↓
Ingress (主版本)
    ├─ 90% 流量 → gateway-service (v1)
    └─ 10% 流量 → gateway-service-canary (v2) ← Canary Ingress
```

## 使用场景

### 场景 1：API 版本升级（/api/v1 → /api/v2）

**步骤 1：部署 v2 版本服务**

```bash
# 部署 gateway-service-canary（v2 版本）
helm upgrade --install gateway-service-canary ./deploy/charts/gateway-service \
  --namespace astraios \
  --set image.tag=v2-commit-sha \
  --set fullnameOverride=gateway-service-canary \
  --set ingress.enabled=false  # Canary 服务不需要主 Ingress
```

**步骤 2：启用灰度**

修改 `values.yaml` 或使用 `--set`：

```yaml
ingress:
  canary:
    enabled: true
    type: weight
    weight: 10  # 10% 流量到 v2
    serviceName: gateway-service-canary
    servicePort: 8888
```

```bash
helm upgrade --install gateway-service ./deploy/charts/gateway-service \
  --namespace astraios \
  --set ingress.canary.enabled=true \
  --set ingress.canary.weight=10
```

**步骤 3：逐步扩大灰度比例**

```bash
# 10% → 30% → 50% → 100%
helm upgrade gateway-service ./deploy/charts/gateway-service \
  --namespace astraios \
  --set ingress.canary.weight=30
```

**步骤 4：全量切换**

```bash
# 1. 将 v2 镜像 tag 更新到主服务
helm upgrade gateway-service ./deploy/charts/gateway-service \
  --namespace astraios \
  --set image.tag=v2-commit-sha \
  --set ingress.canary.enabled=false

# 2. 删除 canary 服务
helm uninstall gateway-service-canary --namespace astraios
```

### 场景 2：基于 Header 的灰度（内测用户）

**配置**：

```yaml
ingress:
  canary:
    enabled: true
    type: header
    header:
      name: X-Canary
      value: "true"
    serviceName: gateway-service-canary
    servicePort: 8888
```

**使用**：

```bash
# 普通用户（走 v1）
curl https://astraios.g-oss.top/api/v1/users/register

# 内测用户（走 v2）
curl -H "X-Canary: true" https://astraios.g-oss.top/api/v1/users/register
```

### 场景 3：路径级灰度（/api/v2/* 全部走新版本）

**方案 A：在 gateway-service 内部处理**

在 `routes.go` 中同时注册 v1 和 v2 路由：

```go
// v1 路由（旧版本）
server.AddRoutes(
    []rest.Route{
        {Method: http.MethodPost, Path: "/api/v1/users/register", Handler: ...},
    },
)

// v2 路由（新版本）
server.AddRoutes(
    []rest.Route{
        {Method: http.MethodPost, Path: "/api/v2/users/register", Handler: ...},
    },
)
```

**方案 B：使用独立的 Ingress 规则**

在 `values.yaml` 中添加：

```yaml
ingress:
  hosts:
    - host: astraios.g-oss.top
      paths:
        - path: /api/v1
          pathType: Prefix
        - path: /api/v2
          pathType: Prefix
          # 可以指向不同的后端服务
```

## 最佳实践

### ✅ 推荐做法

1. **小流量开始**：从 5-10% 开始，观察错误率和延迟
2. **逐步扩大**：每 24 小时增加 20-30%，直到 100%
3. **监控指标**：
   - 错误率（4xx/5xx）
   - 响应时间（P50/P95/P99）
   - 业务指标（注册成功率、登录成功率等）
4. **快速回滚**：发现问题立即将 `canary.weight=0` 或 `canary.enabled=false`

### ❌ 避免的做法

1. **不要跳过灰度直接全量**：风险太大
2. **不要一次性切 50%+**：难以定位问题
3. **不要忽略监控**：灰度期间必须密切观察

## 监控建议

在灰度期间，建议监控：

```bash
# 查看两个服务的 Pod 状态
kubectl get pods -n astraios -l app=gateway-service

# 查看请求分布（需要应用层日志）
kubectl logs -n astraios -l app=gateway-service --tail=100 | grep "X-Request-ID"

# 查看错误率
kubectl top pods -n astraios -l app=gateway-service
```

## 回滚步骤

如果灰度版本有问题，立即回滚：

```bash
# 方式 1：禁用灰度
helm upgrade gateway-service ./deploy/charts/gateway-service \
  --namespace astraios \
  --set ingress.canary.enabled=false

# 方式 2：将权重设为 0
helm upgrade gateway-service ./deploy/charts/gateway-service \
  --namespace astraios \
  --set ingress.canary.weight=0
```

## 注意事项

1. **Canary Ingress 必须和主 Ingress 有相同的 host 和 path**，否则不会生效
2. **权重范围是 0-100**，超过 100 会被忽略
3. **Header 灰度区分大小写**，`X-Canary: true` 和 `X-Canary: True` 是不同的
4. **灰度服务必须存在**，否则灰度流量会 502
