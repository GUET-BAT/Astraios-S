# Kubernetes 部署配置

## ⚠️ 重要安全提醒

**当前 CI/CD 配置存在安全风险！**

如果您的 `KUBE_CONFIG` Secret 包含集群管理员权限，请立即按照以下步骤迁移到更安全的配置：

1. **阅读安全指南**：[k8s-deployment-security.md](../../docs/k8s-deployment-security.md)
2. **创建专用 ServiceAccount**：`kubectl apply -f ci-cd-serviceaccount.yaml`
3. **生成受限 kubeconfig**：`../../scripts/generate-deployer-kubeconfig.sh`
4. **更新 GitHub Secrets**：替换 `KUBE_CONFIG` 为新生成的内容
5. **验证权限**：确保只有部署所需的最小权限

## 文件说明

- `ci-cd-serviceaccount.yaml` - CI/CD 专用 ServiceAccount 配置
  - ServiceAccount: `github-actions-deployer`
  - Role: `deployer-role`（限制在单个 namespace）
  - 权限：仅包含部署应用所需的最小权限

## 快速开始

### 首次设置

```bash
# 1. 创建 namespace 和 ServiceAccount
kubectl apply -f ci-cd-serviceaccount.yaml

# 2. 验证创建成功
kubectl get serviceaccount github-actions-deployer -n astraios-prod
kubectl get role deployer-role -n astraios-prod

# 3. 生成 kubeconfig
cd ../..
./scripts/generate-deployer-kubeconfig.sh astraios-prod

# 4. 测试权限
export KUBECONFIG=deployer-kubeconfig.yaml
kubectl get pods -n astraios-prod

# 5. 添加到 GitHub Secrets
cat deployer-kubeconfig.yaml | base64 -w 0  # Linux
cat deployer-kubeconfig.yaml | base64        # macOS

# 6. 清理本地文件
rm deployer-kubeconfig.yaml
unset KUBECONFIG
```

### 定期维护

建议每 90 天轮换一次 token：

```bash
kubectl delete secret github-actions-deployer-token -n astraios-prod
kubectl apply -f ci-cd-serviceaccount.yaml
./scripts/generate-deployer-kubeconfig.sh astraios-prod
# 更新 GitHub Secrets
rm deployer-kubeconfig.yaml
```

## 权限范围

当前配置的权限范围：

✅ **允许的操作**
- 在 `astraios-prod` namespace 中创建/更新/删除 Deployments
- 管理 Services、ConfigMaps、Secrets（Helm 需要）
- 查看 Pods 和日志（用于调试）
- 管理 Ingresses 和 HPA

❌ **禁止的操作**
- 跨 namespace 操作
- 创建/删除 namespace
- 修改 RBAC 规则
- 管理集群级别资源
- 访问其他 namespace

## 故障排查

### 权限不足错误

```bash
# 查看当前权限
kubectl auth can-i --list --as=system:serviceaccount:astraios-prod:github-actions-deployer -n astraios-prod

# 测试特定操作
kubectl auth can-i create deployments -n astraios-prod \
  --as=system:serviceaccount:astraios-prod:github-actions-deployer
```

### Token 认证失败

重新生成 token 并更新 GitHub Secrets。

## 安全检查清单

- [ ] 使用专用 ServiceAccount（非管理员）
- [ ] 权限限制在单个 namespace
- [ ] 已验证权限范围
- [ ] 本地 kubeconfig 文件已删除
- [ ] 设置了 token 轮换计划

详细信息请参考：[k8s-deployment-security.md](../../docs/k8s-deployment-security.md)
