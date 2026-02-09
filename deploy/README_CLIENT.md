# RegistryX Enterprise - Deployment Guide

Welcome to **RegistryX Enterprise**. This package contains everything you need to run your own private, secure, and intelligent container registry.

You can deploy RegistryX using either **Docker Compose** (Simplest) or **Kubernetes** (Scalable).

---

## 🏗️ Option 1: Docker Compose (Recommended for Single Server)

Best for: Single VPS, Testing, Small Teams.

### Prerequisites
- **Docker Engine** (v20.10+)
- **Docker Compose** (v2.0+)
- **Hardware**: Minimum 4GB RAM, 2 CPU Cores, 20GB Disk.

### 🚀 Installation Steps

1.  **Extract the Package**
    Unzip the folder `registryx-enterprise.zip` to your desired location (e.g., `/opt/registryx`).

2.  **Configure Environment**
    Edit the `.env` file to set your secure passwords and domain settings.
    ```bash
    nano .env
    ```
    *Important: Change `POSTGRES_PASSWORD`, `MINIO_ROOT_PASSWORD`, and `JWT_SECRET` immediately.*

3.  **Start Services**
    Run the following command to download and start the entire stack:
    ```bash
    docker-compose up -d
    ```

4.  **Verify Installation**
    Look at the logs to ensure everything started correctly:
    ```bash
    docker-compose logs -f
    ```
    Wait until you see "Server started on :5000" in the backend logs.

---

## ☸️ Option 2: Kubernetes (Recommended for Production/Scaling)

Best for: High Availability, Cloud Clusters (EKS/GKE/AKS), Auto-scaling.

### Prerequisites
- **Running Kubernetes Cluster** (v1.24+)
- **kubectl** configured to access your cluster.
- **Default StorageClass**: Ensure your cluster has a default storage class for dynamic provisioning (Run `kubectl get sc`).
    - *Cloud (AWS/GCP/Azure):* Configuring this is automatic.
    - *Bare Metal/Minikube:* You must enable a provisioner (e.g., `minikube addons enable storage-provisioner`).

### 🚀 Deployment Steps

1.  **Navigate to Kubernetes Manifests**
    Go to the `deploy/k8s` directory:
    ```bash
    cd deploy/k8s
    ```

2.  **Configure Secrets**
    Edit `01-config.yaml` to set your passwords and keys (Base64 encoding is NOT required here as we use `stringData`):
    ```bash
    nano 01-config.yaml
    ```
    *Note: Change `POSTGRES_PASSWORD`, `MINIO_ROOT_PASSWORD`, and `JWT_SECRET` here too!*

3.  **Deploy Resources**
    Apply the manifests in order:
    ```bash
    # 1. Config & Secrets
    kubectl apply -f 01-config.yaml

    # 2. Database, Redis, MinIO (Stateful Services)
    kubectl apply -f 02-postgres.yaml
    kubectl apply -f 03-redis.yaml
    kubectl apply -f 04-minio.yaml

    # 3. Application Services
    kubectl apply -f 05-backend.yaml
    kubectl apply -f 06-frontend.yaml
    ```

4.  **Verify Pods**
    Check if all pods are running:
    ```bash
    kubectl get pods -n registryx -w
    ```
    *Note: It may take a minute for the database to initialize. Any "CrashLoopBackOff" initially is normal while waiting for DB.*

5.  **Access the Application**
    Get the external IP address (LoadBalancer):
    ```bash
    kubectl get svc -n registryx
    ```
    - Access Dashboard: `http://<EXTERNAL-IP>:80`
    - Registry API: `http://<EXTERNAL-IP>:5000`

---

## 💰 Cost Intelligence Configuration

RegistryX supports two calculation modes:

1.  **Cloud Mode (Default):** Uses standard public cloud pricing (e.g., AWS S3).
2.  **On-Prem Mode:** Uses local hardware capacity and efficiency metrics.

### How to Configure

#### For Docker Compose Users
Edit `deploy/.env`:

**Option A: Cloud**
```ini
COST_MODE=CLOUD
STORAGE_COST_PER_GB_MONTH=0.023  # Per GB price (USD)
```

**Option B: On-Premise (Local Server)**
```ini
COST_MODE=ONPREM
STORAGE_CAPACITY_TB=50           # Total Server Storage in TB
STORAGE_COST_PER_GB_MONTH=0.5    # Amortized Hardware Cost Per GB
```
*After changes run: `docker-compose up -d`*

#### For Kubernetes Users
Edit `deploy/k8s/01-config.yaml`:

```yaml
data:
  COST_MODE: "ONPREM"
  STORAGE_CAPACITY_TB: "50"
  STORAGE_COST_PER_GB_MONTH: "0.5"
```
*After changes run: `kubectl apply -f 01-config.yaml && kubectl rollout restart deployment/registryx-backend -n registryx`*

---

## 🌐 Common Information

### Access Credentials
**Default Admin Login:**
- **Username**: `admin`
- **Password**: `password123` (Change this after first login!)

### Maintenance & Backups

**Docker Users:**
- **Restart**: `docker-compose restart`
- **Update**: `docker-compose pull && docker-compose up -d`
- **Backup**: Backup the `./data` directory or the named volumes.

**Kubernetes Users:**
- **Restart**: `kubectl rollout restart deployment/registryx-backend -n registryx`
- **Update**: Update the image tag in `05-backend.yaml` and run `kubectl apply -f 05-backend.yaml`.
- **Backup**: Use a tool like Velero to backup Persistent Volumes (PVCs).

## 📞 Support
For enterprise support or license activation issues, please contact:
**support@registryx.io**
