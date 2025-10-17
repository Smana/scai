# Security Policy

## Known Vulnerabilities

### Ollama Client Library (github.com/ollama/ollama)

**Status:** Acknowledged - Awaiting upstream fixes

The project currently uses `github.com/ollama/ollama v0.12.6` which has 8 known vulnerabilities reported by govulncheck. All vulnerabilities show **"Fixed in: N/A"**, indicating no patched version is available yet.

| CVE ID | Description | Severity | Status |
|--------|-------------|----------|--------|
| [GO-2025-3824](https://pkg.go.dev/vuln/GO-2025-3824) | Cross-Domain Token Exposure | TBD | Awaiting fix |
| [GO-2025-3695](https://pkg.go.dev/vuln/GO-2025-3695) | Denial of Service (DoS) Attack | TBD | Awaiting fix |
| [GO-2025-3689](https://pkg.go.dev/vuln/GO-2025-3689) | Divide by Zero Vulnerability | TBD | Awaiting fix |
| [GO-2025-3582](https://pkg.go.dev/vuln/GO-2025-3582) | DoS via Null Pointer Dereference | TBD | Awaiting fix |
| [GO-2025-3559](https://pkg.go.dev/vuln/GO-2025-3559) | Divide By Zero vulnerability | TBD | Awaiting fix |
| [GO-2025-3558](https://pkg.go.dev/vuln/GO-2025-3558) | Out-of-Bounds Read | TBD | Awaiting fix |
| [GO-2025-3557](https://pkg.go.dev/vuln/GO-2025-3557) | Allocation Without Limits | TBD | Awaiting fix |
| [GO-2025-3548](https://pkg.go.dev/vuln/GO-2025-3548) | DoS via Crafted GZIP | TBD | Awaiting fix |

### Risk Assessment

**Impact Level:** Low to Medium

**Rationale:**
- **Client-Only Usage**: SCIA uses the Ollama **client library** to communicate with a local Ollama server. These vulnerabilities primarily affect the Ollama server, not the client.
- **Local Deployment**: SCIA is designed to communicate with a locally-running, user-controlled Ollama instance, not exposed to untrusted networks.
- **No Server Exposure**: SCIA does not run an Ollama server or expose Ollama endpoints to external networks.
- **DoS Scope**: Most vulnerabilities are DoS-related, which would affect the local Ollama service the user controls, not the SCIA tool itself.

### Mitigation Strategy

1. **Monitor for Patches**: Actively monitor Ollama releases for security patches
   - Check: https://github.com/ollama/ollama/releases
   - Check: https://pkg.go.dev/github.com/ollama/ollama

2. **Update Promptly**: Upgrade to patched versions immediately when available
   ```bash
   go get github.com/ollama/ollama@latest
   ```

3. **User Guidance**: Document in README that users should:
   - Run Ollama locally (not exposed to internet)
   - Keep Ollama server updated to latest version
   - Use firewall rules to restrict Ollama access

4. **Track Progress**: Created TODO to check for Ollama fixes monthly

### Reporting Security Issues

If you discover a security vulnerability in SCIA itself (not the Ollama dependency), please report it by creating a private security advisory on GitHub:

https://github.com/Smana/scia/security/advisories/new

**Do not create public issues for security vulnerabilities.**

## Security Best Practices

When using SCIA:

1. **Local Ollama**: Run Ollama locally, not exposed to the internet
2. **Latest Versions**: Keep both SCIA and Ollama updated
3. **Restricted Access**: Use firewall rules to limit Ollama access
4. **Trusted Repositories**: Only deploy code from trusted sources
5. **AWS Credentials**: Use IAM roles with least-privilege permissions

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

We currently support the latest released version.
