# Image Signing Guide for RegistryX

## Overview
RegistryX supports image signature verification using **Cosign**, which is the industry-standard tool for signing and verifying container images.

## Why Sign Images?

✅ **Security**: Verify images haven't been tampered with  
✅ **Trust**: Prove images come from a trusted source  
✅ **Compliance**: Meet security requirements for production deployments  
✅ **Supply Chain Security**: Part of SLSA framework

## Prerequisites

### Install Cosign

**Windows:**
```powershell
winget install sigstore.cosign
```

**macOS:**
```bash
brew install cosign
```

**Linux:**
```bash
# Download from GitHub releases
wget https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64
chmod +x cosign-linux-amd64
sudo mv cosign-linux-amd64 /usr/local/bin/cosign
```

## Quick Start

### 1. Generate Signing Keys (One-time setup)

```powershell
cosign generate-key-pair
```

This creates:
- `cosign.key` - Private key (keep secret!)
- `cosign.pub` - Public key (share for verification)

You'll be prompted to set a password for the private key.

### 2. Sign an Image

```powershell
# Sign a single image
cosign sign --key cosign.key localhost:5000/demo/alpine:latest

# Sign without password prompt (for automation)
$env:COSIGN_PASSWORD = "your-password"
cosign sign --key cosign.key --yes localhost:5000/demo/alpine:latest
```

### 3. Verify Signature

```powershell
cosign verify --key cosign.pub localhost:5000/demo/alpine:latest
```

### 4. Use the Automated Script

```powershell
# Sign all images in your registry
.\sign_images.ps1
```

## How It Works

### Cosign Signature Format

When you sign an image with digest `sha256:abc123...`, Cosign creates a signature tag:
```
sha256-abc123....sig
```

RegistryX automatically detects this tag pattern and marks the image as "Signed" in the UI.

### Example

```powershell
# Original image
localhost:5000/demo/alpine:latest
  └─ Digest: sha256:1882fa4569e0c591ea092d3766c4893e19b8901a8e649de7067188aba3cc0679

# After signing, Cosign pushes a signature
localhost:5000/demo/alpine:sha256-1882fa4569e0c591ea092d3766c4893e19b8901a8e649de7067188aba3cc0679.sig
```

## Production Best Practices

### 1. Secure Key Management

**DO:**
- Store `cosign.key` in a secrets manager (HashiCorp Vault, AWS Secrets Manager, etc.)
- Use strong passwords for private keys
- Rotate keys periodically
- Use different keys for different environments (dev, staging, prod)

**DON'T:**
- Commit `cosign.key` to Git
- Share private keys via email or chat
- Use empty passwords in production

### 2. CI/CD Integration

```yaml
# GitHub Actions example
- name: Sign Image
  env:
    COSIGN_PASSWORD: ${{ secrets.COSIGN_PASSWORD }}
  run: |
    echo "${{ secrets.COSIGN_PRIVATE_KEY }}" > cosign.key
    cosign sign --key cosign.key --yes $IMAGE_NAME
    rm cosign.key
```

### 3. Keyless Signing (Advanced)

For production, consider using Cosign's keyless signing with Fulcio:

```powershell
# Sign without managing keys
cosign sign localhost:5000/demo/alpine:latest

# Verify keyless signature
cosign verify --certificate-identity=user@example.com \
  --certificate-oidc-issuer=https://accounts.google.com \
  localhost:5000/demo/alpine:latest
```

## Troubleshooting

### "Image not signed" in RegistryX

**Check:**
1. Signature tag exists in registry:
   ```powershell
   curl http://localhost:5000/v2/demo/alpine/tags/list
   ```
   Look for tags like `sha256-....sig`

2. Signature is for the correct digest:
   ```powershell
   # Get image digest
   docker inspect localhost:5000/demo/alpine:latest | Select-String "RepoDigests"
   
   # Verify signature tag matches
   ```

3. Re-sign the image:
   ```powershell
   cosign sign --key cosign.key --yes localhost:5000/demo/alpine:latest
   ```

### "Failed to sign" error

**Common causes:**
- Image doesn't exist locally → Pull it first: `docker pull <image>`
- Wrong password → Check `COSIGN_PASSWORD` environment variable
- Network issues → Check registry connectivity

### Signature verification fails

```powershell
# Check if signature exists
cosign verify --key cosign.pub localhost:5000/demo/alpine:latest

# Common issues:
# - Wrong public key
# - Image was re-pushed without re-signing
# - Signature tag was deleted
```

## Advanced Features

### Sign Multiple Tags

```powershell
# Sign all tags of an image
$tags = @("latest", "v1.0", "stable")
foreach ($tag in $tags) {
    cosign sign --key cosign.key --yes localhost:5000/demo/alpine:$tag
}
```

### Attach SBOMs (Software Bill of Materials)

```powershell
# Generate SBOM
syft localhost:5000/demo/alpine:latest -o spdx-json > sbom.json

# Attach to image
cosign attach sbom --sbom sbom.json localhost:5000/demo/alpine:latest
```

### Policy Enforcement

Create a policy to only allow signed images:

```rego
# In RegistryX Policy Editor
package registry

deny[msg] {
    input.operation == "pull"
    not input.signed
    msg := "Only signed images are allowed"
}
```

## References

- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
- [Sigstore Project](https://www.sigstore.dev/)
- [Supply Chain Security](https://slsa.dev/)

## Support

For issues or questions:
1. Check RegistryX logs: `docker logs registryx-backend`
2. Verify Cosign version: `cosign version`
3. Test signature manually: `cosign verify --key cosign.pub <image>`
