# RegistryX on Kubernetes ☸️

This directory contains the manifests to deploy RegistryX on any Kubernetes cluster (EKS, GKE, AKS, or local Minikube).

## 🚀 Installation Guide

### 1. Prerequisites
- A running Kubernetes cluster.
- `kubectl` installed and configured.

### 2. Configure Secrets
Edit `01-config.yaml` to set your passwords and keys:
```yaml
stringData:
  POSTGRES_PASSWORD: "secure-password-here"
  JWT_SECRET: "random-secret-key"
```

### 3. Deploy
Run the following commands in order:

```bash
# 1. Create Namespace & Configs
kubectl apply -f 01-config.yaml

# 2. Deploy Infrastructure (DB, Redis, MinIO)
kubectl apply -f 02-postgres.yaml
kubectl apply -f 03-redis.yaml
kubectl apply -f 04-minio.yaml

# 3. Deploy Applications
kubectl apply -f 05-backend.yaml
kubectl apply -f 06-frontend.yaml
```

### 4. Verify
Check if all pods are running:
```bash
kubectl get pods -n registryx
```

### 5. Access
If you are using a cloud provider (AWS/GCP), the services `registryx-frontend` and `registryx-backend` are set to `LoadBalancer`, so they will get a public IP.

Get the IP:
```bash
kubectl get svc -n registryx
```

*For production, we recommend setting up an Ingress Controller instead of LoadBalancer for every service.*
