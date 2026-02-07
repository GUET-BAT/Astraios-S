# 生成 CI/CD 专用的 kubeconfig (PowerShell 版本)
# 用法: .\scripts\generate-deployer-kubeconfig.ps1 [namespace] [cluster-name]

param(
    [string]$Namespace = "astraios-prod",
    [string]$ClusterName = "kubernetes"
)

$ServiceAccount = "github-actions-deployer"
$SecretName = "github-actions-deployer-token"
$OutputFile = "deployer-kubeconfig.yaml"

Write-Host "=== Generating kubeconfig for CI/CD deployment ===" -ForegroundColor Cyan
Write-Host "Namespace: $Namespace"
Write-Host "Cluster: $ClusterName"
Write-Host "ServiceAccount: $ServiceAccount"
Write-Host ""

# Check if namespace exists
try {
    kubectl get namespace $Namespace | Out-Null
} catch {
    Write-Host "Error: Namespace $Namespace does not exist" -ForegroundColor Red
    Write-Host "Please apply the ServiceAccount configuration first:" -ForegroundColor Yellow
    Write-Host "  kubectl apply -f deploy/k8s/ci-cd-serviceaccount.yaml"
    exit 1
}

# Check if ServiceAccount exists
try {
    kubectl get serviceaccount $ServiceAccount -n $Namespace | Out-Null
} catch {
    Write-Host "Error: ServiceAccount $ServiceAccount does not exist in namespace $Namespace" -ForegroundColor Red
    Write-Host "Please apply the ServiceAccount configuration first:" -ForegroundColor Yellow
    Write-Host "  kubectl apply -f deploy/k8s/ci-cd-serviceaccount.yaml"
    exit 1
}

# Get the secret token
Write-Host "Fetching token from secret..."
$TokenBase64 = kubectl get secret $SecretName -n $Namespace -o jsonpath='{.data.token}'
if (-not $TokenBase64) {
    Write-Host "Error: Failed to get token from secret $SecretName" -ForegroundColor Red
    exit 1
}
$Token = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($TokenBase64))

# Get cluster info
Write-Host "Fetching cluster information..."
$ClusterServer = kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}'
$ClusterCA = kubectl config view --minify --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}'

if (-not $ClusterServer) {
    Write-Host "Error: Failed to get cluster server URL" -ForegroundColor Red
    exit 1
}

# Generate kubeconfig
$KubeconfigContent = @"
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: $ClusterCA
    server: $ClusterServer
  name: $ClusterName
contexts:
- context:
    cluster: $ClusterName
    namespace: $Namespace
    user: $ServiceAccount
  name: $ServiceAccount@$ClusterName
current-context: $ServiceAccount@$ClusterName
users:
- name: $ServiceAccount
  user:
    token: $Token
"@

Set-Content -Path $OutputFile -Value $KubeconfigContent -Encoding UTF8

Write-Host ""
Write-Host "✅ Kubeconfig generated successfully: $OutputFile" -ForegroundColor Green
Write-Host ""
Write-Host "=== Next steps ===" -ForegroundColor Cyan
Write-Host "1. Test the kubeconfig:"
Write-Host "   `$env:KUBECONFIG = `"$OutputFile`""
Write-Host "   kubectl get pods -n $Namespace"
Write-Host ""
Write-Host "2. Add to GitHub Secrets:"
Write-Host "   `$content = Get-Content $OutputFile -Raw"
Write-Host "   `$base64 = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes(`$content))"
Write-Host "   Write-Host `$base64"
Write-Host "   Then add the output to GitHub repository secrets as KUBE_CONFIG"
Write-Host ""
Write-Host "3. Verify permissions:"
Write-Host "   kubectl auth can-i list deployments -n $Namespace --as=system:serviceaccount:${Namespace}:${ServiceAccount}"
Write-Host ""
Write-Host "⚠️  IMPORTANT: Delete this file after adding to GitHub Secrets!" -ForegroundColor Yellow
Write-Host "   Remove-Item $OutputFile"
