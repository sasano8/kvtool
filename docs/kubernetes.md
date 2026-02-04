# Kubernetes 連携

kvtool を Kubernetes のシークレットプロバイダーとして使用する方法を説明します。

## 連携方法の比較

| 方法 | 難易度 | 特徴 |
|-----|-------|------|
| External Secrets Operator | 低 | Webhook 連携、宣言的管理 |
| Init Container | 低 | 追加コンポーネント不要 |
| Sidecar | 中 | シークレットの自動更新 |

## External Secrets Operator (ESO)

[External Secrets Operator](https://external-secrets.io/) は Kubernetes でシークレットを外部プロバイダーから同期するための標準的なツールです。

### ESO とは

ESO は以下の機能を提供します：

- 外部シークレットストア（Vault、AWS Secrets Manager 等）との連携
- Kubernetes Secret への自動同期
- **Webhook Provider** による任意の HTTP API との連携

kvtool は Webhook Provider を通じて ESO と連携します。

### ESO のインストール

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets --create-namespace
```

### kvtool のデプロイ

```yaml
# kvtool-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kvtool
  namespace: infra
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kvtool
  template:
    metadata:
      labels:
        app: kvtool
    spec:
      containers:
        - name: kvtool
          image: your-registry/kvtool:latest
          args: ["store", "serve", "--port", "8080"]
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: config
              mountPath: /app/.kvtool.yml
              subPath: .kvtool.yml
      volumes:
        - name: config
          configMap:
            name: kvtool-config
---
apiVersion: v1
kind: Service
metadata:
  name: kvtool
  namespace: infra
spec:
  selector:
    app: kvtool
  ports:
    - port: 8080
      targetPort: 8080
```

### SecretStore の設定

ESO の Webhook Provider は URL テンプレーティングをサポートしています。`remoteRef` オブジェクトの値を URL に埋め込めます。

```yaml
# secret-store.yaml
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: kvtool
spec:
  provider:
    webhook:
      url: "http://kvtool.infra.svc:8080/v1/namespaces/{{ .remoteRef.property }}/kv/{{ .remoteRef.key }}"
      result:
        jsonPath: "$"
      headers:
        Content-Type: application/json
```

**URL テンプレート変数:**

| 変数 | 説明 | 使用例 |
|-----|------|-------|
| `{{ .remoteRef.key }}` | ExternalSecret の `remoteRef.key` | `vault/db/password` |
| `{{ .remoteRef.property }}` | ExternalSecret の `remoteRef.property` | `production` |

### ExternalSecret の作成

```yaml
# external-secret.yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: my-app-secrets
  namespace: default
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: kvtool
    kind: ClusterSecretStore
  target:
    name: my-app-secrets  # 作成される K8s Secret の名前
    creationPolicy: Owner
  data:
    # 単一の値を取得
    - secretKey: db-password        # K8s Secret のキー
      remoteRef:
        key: vault/db/password      # kvtool の store/path
        property: production        # kvtool の namespace

    # JSON から特定のフィールドを抽出
    - secretKey: api-key
      remoteRef:
        key: vault/app/config
        property: production
      # JSONPath で値を抽出（オプション）
      # result:
      #   jsonPath: "$.api_key"
```

### 結果の確認

```bash
# ExternalSecret の状態確認
kubectl get externalsecret my-app-secrets

# 作成された Secret の確認
kubectl get secret my-app-secrets -o yaml

# 値の確認
kubectl get secret my-app-secrets -o jsonpath='{.data.db-password}' | base64 -d
```

## Init Container パターン

ESO を使わずに、Init Container で直接 kvtool を実行する方法です。

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-app
spec:
  initContainers:
    - name: load-secrets
      image: your-registry/kvtool:latest
      command:
        - sh
        - -c
        - |
          # 単一の値を取得
          kvtool store load -n production vault db/password > /secrets/db-password

          # .env 形式で複数の値を取得
          kvtool store load -n production local env/app.env > /secrets/.env
      volumeMounts:
        - name: secrets
          mountPath: /secrets
        - name: kvtool-config
          mountPath: /app/.kvtool.yml
          subPath: .kvtool.yml

  containers:
    - name: app
      image: my-app:latest
      volumeMounts:
        - name: secrets
          mountPath: /secrets
          readOnly: true
      env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-password-file
              key: password

  volumes:
    - name: secrets
      emptyDir:
        medium: Memory  # メモリ上に保持（セキュリティ向上）
    - name: kvtool-config
      configMap:
        name: kvtool-config
```

## Sidecar パターン

シークレットを定期的に更新する場合は、Sidecar パターンを使用します。

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-app
spec:
  containers:
    - name: app
      image: my-app:latest
      volumeMounts:
        - name: secrets
          mountPath: /secrets
          readOnly: true

    - name: secret-sync
      image: your-registry/kvtool:latest
      command:
        - sh
        - -c
        - |
          while true; do
            kvtool store load -n production vault db/password > /secrets/db-password
            sleep 300  # 5分ごとに更新
          done
      volumeMounts:
        - name: secrets
          mountPath: /secrets
        - name: kvtool-config
          mountPath: /app/.kvtool.yml
          subPath: .kvtool.yml

  volumes:
    - name: secrets
      emptyDir:
        medium: Memory
    - name: kvtool-config
      configMap:
        name: kvtool-config
```

## セキュリティ考慮事項

### ネットワークポリシー

kvtool サービスへのアクセスを制限します：

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kvtool-access
  namespace: infra
spec:
  podSelector:
    matchLabels:
      app: kvtool
  ingress:
    - from:
        # ESO からのアクセスのみ許可
        - namespaceSelector:
            matchLabels:
              name: external-secrets
        # または特定の namespace から
        - namespaceSelector:
            matchLabels:
              kvtool-access: "true"
      ports:
        - port: 8080
```

### RBAC

kvtool の ConfigMap へのアクセスを制限します：

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kvtool-config-reader
  namespace: infra
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    resourceNames: ["kvtool-config"]
    verbs: ["get"]
```

## トラブルシューティング

### ESO が Secret を作成しない

```bash
# ExternalSecret のイベント確認
kubectl describe externalsecret my-app-secrets

# ESO のログ確認
kubectl logs -n external-secrets -l app.kubernetes.io/name=external-secrets
```

### kvtool への接続エラー

```bash
# kvtool Pod の状態確認
kubectl get pods -n infra -l app=kvtool

# kvtool のログ確認
kubectl logs -n infra -l app=kvtool

# Service への疎通確認
kubectl run -it --rm debug --image=curlimages/curl -- \
  curl http://kvtool.infra.svc:8080/health
```