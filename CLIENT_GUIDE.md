# 🚀 RegistryX: Client Deployment & User Guide

This guide provides comprehensive instructions for deploying and configuring RegistryX in a production or client environment.

---

## 📋 Prerequisites
Ensure the following are installed on the target server:
- **Docker** (v24.0 or higher)
- **Docker Compose** (v2.0 or higher)
- Minimum 4GB RAM and 50GB Disk Space.

---

## 🛠️ Phase 1: Environment Configuration
Before starting the services, navigate to the `deploy/` directory and configure the `.env` file.

### 1. Security Credentials & JWT Secret
Update these variables immediately to secure the installation:
- `POSTGRES_PASSWORD`: A strong password for the database.
- `MINIO_ROOT_PASSWORD`: A strong password for the storage backend.

**How to generate your `JWT_SECRET`:**
RegistryX uses this secret to sign authentication tokens. To generate a secure secret, run one of the following commands in your terminal and paste the output into the `JWT_SECRET` field in `.env`:

*   **Option A (Using OpenSSL):**
    ```bash
    openssl rand -base64 32
    ```
*   **Option B (Using PowerShell - Windows):**
    ```powershell
    [Convert]::ToBase64String((1..32 | % { [byte](Get-Random -Minimum 0 -Maximum 255) }))
    ```
*   **Option C (Manual):** Type a minimum of 32 characters consisting of random letters, numbers, and symbols.

### 2. URL Settings
Replace `localhost` with the client's actual domain or server IP:
- `BACKEND_URL`: `http://<your-server-ip>:5000`
- `FRONTEND_URL`: `http://<your-server-ip>:5173`

### 3. Enterprise Plan Activation
To unlock "Policy Enforcement" and "Cost Intelligence", add a valid license key:
- `REGISTRYX_LICENSE_KEY`: Paste your unique enterprise key here.

---

## 🚀 Phase 2: Deployment
From the project root, run the following command to build and start all services:

```bash
docker-compose -f deploy/docker-compose.yml up -d --build
```

---

## 🔐 Phase 3: Initial Login
Access the dashboard at `http://<your-server-ip>:5173`. Use the following default administrator credentials for the first login:

- **Username:** `admin`
- **Password:** `admin@123`

**CRITICAL:** Change the administrator password immediately after logging in via the **Settings > Users** menu.

---

## 📦 Phase 4: Basic Operations Guide

### 1. Creating a Repository
1. Navigate to the **Repositories** tab.
2. Click **+ Initiate Repository**.
3. Use the format `library/my-app-name`. This helps in organizing images logically.

### 2. Pushing your First Image
Use the **Quick Start** terminal guide for the following commands:

```bash
# Authenticate
docker login <your-server-ip>:5000

# Tag your local image
docker tag my-app:latest <your-server-ip>:5000/library/my-app:latest

# Push to RegistryX
docker push <your-server-ip>:5000/library/my-app:latest
```

### 🛡️ 3. Security Scanning
Once pushed, RegistryX automatically triggers a **Trivy Vulnerability Scan**.
- Go to the **Dashboard** to see the security health report.
- Check the **Lineage** tab for a detailed breakdown of vulnerabilities.

---

## 📧 Phase 5: Email Alerts (Optional)
To receive scan alerts, configure the SMTP section in the `.env` file:
- `SMTP_HOST`: e.g., `smtp.gmail.com`
- `SMTP_USER`: Client's email address.
- `SMTP_PASS`: Application-specific password.

---

## 🆘 Support & Maintenance
- **Check Logs:** `docker logs -f registryx-backend`
- **Database Backup:** Ensure periodic backups of the `postgres_data` volume.
- **Updates:** Pull the latest code and rerun the build command.

---
© 2026 RegistryX. All rights reserved.
