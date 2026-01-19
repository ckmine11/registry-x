# RegistryX Deployment Guide

This guide explains how to deploy RegistryX to a production server (Linux/Ubuntu).

## 1. Prerequisites
- A Linux Server (Ubuntu 20.04+ recommended)
- Docker & Docker Compose installed
- The `registry-x-feature-complete.zip` backup file

## 2. Server Setup (Ubuntu)
```bash
# Update and Install Docker
sudo apt update
sudo apt install -y docker.io docker-compose
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
# (Log out and back in for group changes to take effect)
```

## 3. Deploying RegistryX
1. **Transfer the Zip**: Upload `registry-x-feature-complete.zip` to your server (use SCP or SFTP).
   ```bash
   scp registry-x-feature-complete.zip user@your-server-ip:~/
   ```

2. **Unzip and Setup**:
   ```bash
   unzip registry-x-feature-complete.zip -d registry-x
   cd registry-x
   ```

3. **Configure Environment**:
   - Edit `.env` to set strong passwords and secrets.
   ```bash
   nano .env
   # Change REGISTRY_PASSWORD, JWT_SECRET, etc.
   # New Production Flags:
   # S3_BUCKET=registryx-prod
   # MINIO_SECURE=true       # Enable for Production S3/Minio with SSL
   # POLICY_ENVIRONMENT=prod # Enforce strict policy (e.g., Image Signing)
   # SMTP_USER=apikey        # Secure SMTP credentials
   ```

4. **Start the Application**:
   ```bash
   docker-compose -f deploy/docker-compose.yml --env-file .env up -d --build
   ```

5. **Verify Deployment**:
   ```bash
   curl http://localhost:5000/health-check
   # Should return {"status": "ok", ...}
   ```

## 4. Setting up SSL (HTTPS) - Recommended
For production, you should use HTTPS. The easiest way is using Nginx as a reverse proxy with Certbot.

1. **Install Nginx**: `sudo apt install nginx certbot python3-certbot-nginx`
2. **Configure Nginx**: Proxy pass port 80/443 to port 5000.
   ```nginx
   server {
       server_name registry.yourdomain.com;
       client_max_body_size 20G; # Important for large layers

       location / {
           proxy_pass http://localhost:5000;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   ```
3. **Enable SSL**: `sudo certbot --nginx -d registry.yourdomain.com`

## 5. Maintenance
- **View Logs**: `docker-compose -f deploy/docker-compose.yml logs -f`
- **Stop**: `docker-compose -f deploy/docker-compose.yml down`
- **Backup**: Regularly backup the `registry-data` volume (MinIO data) and `postgres-data` volume.
