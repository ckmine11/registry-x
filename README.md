# RegistryX - Secure & Intelligent Container Registry

RegistryX is a modern, production-ready OCI-compliant container registry built for security, cost-efficiency, and ease of use. It goes beyond simple image storage by offering **Real-time Vulnerability Scanning**, **Exploit Prediction (EPSS)**, **Cost Intelligence**, and **Granular Access Control**.

![RegistryX Dashboard](https://via.placeholder.com/1200x600?text=RegistryX+Dashboard+Preview)

## 🚀 Key Features

### 🛡️ Advanced Security Pipeline
*   **Real-time Vulnerability Scanning**: Automatically scans images upon push using **Trivy**.
*   **Smart Threat Intelligence**: Integrates with **EPSS (Exploit Prediction Scoring System)** to prioritize vulnerabilities based on the likelihood of real-world exploitation, not just static CVSS scores.
*   **Image Signing**: Supports Docker Content Trust (DCT) to ensure image integrity and provenance.
*   **Granular Isolation**: Strict multi-tenant data isolation ensures users can access only their own namespaces.

### 💰 Cost Intelligence
*   **Real-time Cost Tracking**: Visualizes storage and bandwidth costs per repository and tag.
*   **Zombie Image Detection**: Identifies unused or "zombie" images (not pulled in >90 days) to help reclaim storage space.
*   **Storage Quotas**: Enforces limits to prevent resource exhaustion.

### ⚡ Performance & Usability
*   **Modern Web UI**: A sleak, responsive React-based dashboard for managing repositories, policies, and settings.
*   **S3-Compatible Storage**: Built on MinIO for scalable, cloud-native object storage.
*   **Production Ready**: Includes comprehensive logging, audit trails, and health monitoring.

---

## 🏗️ Architecture

RegistryX follows a microservices-ready architecture:

*   **Frontend**: React (Vite + TailwindCSS)
*   **Backend**: Go (Golang) REST API
*   **Database**: PostgreSQL 15
*   **Storage**: MinIO (S3 Compatible)
*   **Caching/Queues**: Redis
*   **Scanning Engine**: Trivy

---

## 🛠️ Quick Start

### Prerequisites
*   Docker & Docker Compose
*   Git

### Installation

1.  **Clone the repository**
    ```bash
    git clone https://github.com/ckmine11/registry-x.git
    cd registry-x
    ```

2.  **Start the Application** (Production Mode)
    Windows (PowerShell):
    ```powershell
    .\scripts\start_production.ps1
    ```
    Linux/Mac:
    ```bash
    docker-compose -f deploy/docker-compose.yml up --build -d
    ```

3.  **Access the Dashboard**
    *   **Frontend UI**: [http://localhost:5173](http://localhost:5173)
    *   **Registry API**: [http://localhost:5000](http://localhost:5000)

4.  **Default Credentials**
    *   **Username**: `admin`
    *   **Password**: `password123`

---

## 📖 Usage Guide

### 1. Pushing Images

Login with your secure RegistryX credentials:
```bash
docker login localhost:5000
```

Tag and push your image:
```bash
docker tag my-app:latest localhost:5000/my-user/my-app:v1
docker push localhost:5000/my-user/my-app:v1
```
*The scanner will automatically trigger upon upload.*

### 2. Checking Vulnerabilities

Navigate to the **Repositories** page in the UI to view scan results.
*   **Critical/High**: Immediate action required.
*   **EPSS Score**: Use the "Smart Resolution" tab to focus on bugs with active exploits.

### 3. Managing Costs

Visit the **Cost Intelligence** tab to:
*   View your monthly burn rate.
*   Identify expensive, large images.
*   Clean up "Zombie Images" with one click.

---

## 🔧 Deployment Configuration

Environment variables can be configured in `.env` or `deploy/docker-compose.yml`:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `DB_HOST` | PostgreSQL Hostname | `db` |
| `S3_ENDPOINT` | MinIO Address | `minio:9000` |
| `S3_BUCKET` | Storage Bucket Name | `registryx-data` |
| `MINIO_SECURE` | Use SSL for Storage | `false` |
| `JWT_SECRET` | Secret for Session Tokens | *(Change in Prod)* |

---

## 🤝 Contributing

Contributions are welcome!
1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
