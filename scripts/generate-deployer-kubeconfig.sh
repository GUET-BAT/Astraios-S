#!/usr/bin/env bash
# 生成 CI/CD 专用的 kubeconfig
# 用法: ./scripts/generate-deployer-kubeconfig.sh [namespace] [cluster-name]
#
# 注意：首次使用前需要添加执行权限：
#   chmod +x scripts/generate-deployer-kubeconfig.sh

set -e

NAMESPACE="${1:-astraios-prod}"
CLUSTER_NAME="${2:-kubernetes}"
SERVICE_ACCOUNT="github-actions-deployer"
SECRET_NAME="github-actions-deployer-token"

echo "=== Generating kubeconfig for CI/CD deployment ==="
echo "Namespace: $NAMESPACE"
echo "Cluster: $CLUSTER_NAME"
echo "ServiceAccount: $SERVICE_ACCOUNT"
echo ""

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
  echo "Error: Namespace $NAMESPACE does not exist"
  echo "Please apply the ServiceAccount configuration first:"
  echo "  kubectl apply -f deploy/k8s/ci-cd-serviceaccount.yaml"
  exit 1
fi

# Check if ServiceAccount exists
if ! kubectl get serviceaccount "$SERVICE_ACCOUNT" -n "$NAMESPACE" &> /dev/null; then
  echo "Error: ServiceAccount $SERVICE_ACCOUNT does not exist in namespace $NAMESPACE"
  echo "Please apply the ServiceAccount configuration first:"
  echo "  kubectl apply -f deploy/k8s/ci-cd-serviceaccount.yaml"
  exit 1
fi

# Get the secret token
echo "Fetching token from secret..."
TOKEN=$(kubectl get secret "$SECRET_NAME" -n "$NAMESPACE" -o jsonpath='{.data.token}' | base64 -d)

if [ -z "$TOKEN" ]; then
  echo "Error: Failed to get token from secret $SECRET_NAME"
  exit 1
fi

# Get cluster info
echo "Fetching cluster information..."
CLUSTER_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
CLUSTER_CA=$(kubectl config view --minify --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')

if [ -z "$CLUSTER_SERVER" ]; then
  echo "Error: Failed to get cluster server URL"
  exit 1
fi

# Generate kubeconfig
OUTPUT_FILE="deployer-kubeconfig.yaml"
cat > "$OUTPUT_FILE" <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: $CLUSTER_CA
    server: $CLUSTER_SERVER
  name: $CLUSTER_NAME
contexts:
- context:
    cluster: $CLUSTER_NAME
    namespace: $NAMESPACE
    user: $SERVICE_ACCOUNT
  name: $SERVICE_ACCOUNT@$CLUSTER_NAME
current-context: $SERVICE_ACCOUNT@$CLUSTER_NAME
users:
- name: $SERVICE_ACCOUNT
  user:
    token: $TOKEN
EOF

echo ""
echo "✅ Kubeconfig generated successfully: $OUTPUT_FILE"
echo ""
echo "=== Next steps ==="
echo ""
echo "1. Test the kubeconfig:"
echo "   export KUBECONFIG=$OUTPUT_FILE"
echo "   kubectl get pods -n $NAMESPACE"
echo ""
echo "2. Add to GitHub Secrets (choose one method):"
echo ""
echo "   Method 1 (Recommended): Base64 encode"
if [[ "$OSTYPE" == "darwin"* ]]; then
  echo "   cat $OUTPUT_FILE | base64 | pbcopy  # Copies to clipboard"
  echo "   # Or manually copy the output:"
  echo "   cat $OUTPUT_FILE | base64"
else
  echo "   cat $OUTPUT_FILE | base64 -w 0  # Linux"
  echo "   # Or copy to clipboard (if xclip installed):"
  echo "   cat $OUTPUT_FILE | base64 -w 0 | xclip -selection clipboard"
fi
echo ""
echo "   Method 2 (Alternative): Direct YAML"
echo "   cat $OUTPUT_FILE"
echo ""
echo "   Then:"
echo "   - Go to GitHub repository Settings → Secrets → Actions"
echo "   - Update KUBE_CONFIG secret with the copied content"
echo "   - The workflow will automatically detect the format (base64 or YAML)"
echo ""
echo "3. Verify permissions:"
echo "   kubectl auth can-i list deployments -n $NAMESPACE --as=system:serviceaccount:$NAMESPACE:$SERVICE_ACCOUNT"
echo ""
echo "⚠️  IMPORTANT: Delete this file after adding to GitHub Secrets!"
echo "   rm $OUTPUT_FILE"
echo "   unset KUBECONFIG"
