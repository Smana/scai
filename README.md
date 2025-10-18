# SCIA: Smart Cloud Infrastructure Automation

> **AI-powered deployment that turns natural language into cloud infrastructure**

**SCIA** (from Latin *"scio"* - "I know" + IA) analyzes your code, determines the optimal deployment strategy using AI, and automatically provisions infrastructure on AWS.

```bash
scia deploy "Deploy this Flask app on AWS" https://github.com/your-org/app
```

SCIA will:
1. **Analyze** your application (framework, dependencies, ports)
2. **Decide** the best deployment strategy using AI (VM, Kubernetes, or Serverless)
3. **Deploy** by generating and applying Terraform configuration
4. **Return** a working deployment with access URL

## 🚀 Getting Started

### Prerequisites

You need:
1. **OpenTofu or Terraform** - Infrastructure provisioning tool
2. **Docker** - SCIA uses Docker to run Ollama LLM (automatic setup on first run)
3. **AWS credentials** - Configured via `aws configure`

### Installation

Download the latest binary from the [releases page](https://github.com/Smana/scia/releases):

```bash
# Download and install (replace VERSION with latest release)
curl -L https://github.com/Smana/scia/releases/download/VERSION/scia-linux-amd64 -o scia
chmod +x scia
sudo mv scia /usr/local/bin/
```

Or build from source:

```bash
git clone https://github.com/Smana/scia
cd scia
task build  # requires Task runner: https://taskfile.dev
sudo cp scia /usr/local/bin/
```

### Quick Start

**1. Initialize SCIA** (one-time setup)

```bash
scia init
```

This will:
- Configure your LLM provider (Ollama, Gemini, or OpenAI)
- Set up Terraform backend for state storage (optional)
- Validate AWS credentials and requirements

**2. Deploy your first application**

```bash
scia deploy "Deploy this Flask app" https://github.com/Arvo-AI/hello_world
```

SCIA will automatically:
- Set up Ollama in Docker (if needed)
- Download the AI model (~4GB on first run)
- Analyze and deploy your application

**3. Manage your deployments**

```bash
# List all deployments
scia list

# Show detailed deployment info
scia show <deployment-id>

# View deployment outputs (URLs, IPs)
scia outputs <deployment-id>

# Check deployment status
scia status <deployment-id>

# Destroy a deployment
scia destroy <deployment-id>
```

### Example Deployment Session
```
🐳 Setting up Ollama with Docker...
Creating Ollama container...
✓ Ollama container is ready
Pulling model qwen2.5-coder:7b...
✓ Model qwen2.5-coder:7b is ready

📊 Analyzing repository...
✅ Detected: flask (python), Port 5000

🤖 Determining deployment strategy...
   Recommended strategy: vm

📋 Preparing deployment plan...
┌─────────────────────────────────────────────┐
│         🚀 Deployment Plan                   │
├─────────────────────────────────────────────┤
│ Strategy: EC2 VM                             │
│ Region: us-east-1                            │
│ Application: hello_world                     │
│ Instance Type: t3.micro                      │
│ Port: 5000                                   │
└─────────────────────────────────────────────┘

Proceed with deployment? [y/N]: y

🚀 Deploying infrastructure...
✅ Deployment Complete!

📋 Deployment Summary:
   Strategy: vm
   Region: us-east-1

🔗 Access URLs:
   public_url: http://54.123.45.67:5000

💡 Optimization Suggestions:
   Consider using Gunicorn for production

🎉 Success! Your application is now deployed.
```

### Example: Managing Deployments

```bash
# List all your deployments
$ scia list
Found 3 deployment(s):

ID                                    APP NAME              STRATEGY    REGION        STATUS         CREATED
a1b2c3d4-e5f6-7890-abcd-ef1234567890  hello-world           vm          us-east-1     ✅ succeeded   2025-10-18 14:23
b2c3d4e5-f6a7-8901-bcde-f12345678901  api-service           vm          eu-west-1     ✅ succeeded   2025-10-18 13:45
c3d4e5f6-a7b8-9012-cdef-123456789012  microservices         kubernetes  us-west-2     🔄 running     2025-10-18 15:10

Use 'scia show <deployment-id>' to see detailed information

# Show detailed deployment information
$ scia show a1b2c3d4
═══════════════════════════════════════════════════════════════
  DEPLOYMENT: hello-world
═══════════════════════════════════════════════════════════════

📋 Basic Information:
   ID:           a1b2c3d4-e5f6-7890-abcd-ef1234567890
   App Name:     hello-world
   Status:       ✅ succeeded
   Strategy:     vm
   Region:       us-east-1

🔗 Outputs:
   public_url: http://54.123.45.67:5000

# Destroy when done
$ scia destroy a1b2c3d4
═══════════════════════════════════════════════════════════════
  DESTROY DEPLOYMENT: hello-world
═══════════════════════════════════════════════════════════════

   ID:           a1b2c3d4-e5f6-7890-abcd-ef1234567890
   Strategy:     vm
   Region:       us-east-1

⚠️  WARNING: This will destroy all infrastructure resources!

Do you want to proceed? (yes/no): yes

🔥 Destroying infrastructure...
✅ Deployment destroyed successfully!
```

## 🧠 How It Works

SCIA uses a **3-tier decision system**:

1. **Code Analysis**: Detects framework (Flask, Express, Go...), dependencies, and configuration
2. **AI Decision**:
   - **Rule-based** fast path for common patterns (docker-compose → Kubernetes)
   - **LLM-powered** smart path with deployment knowledge base
   - **Heuristic** fallback for edge cases
3. **Infrastructure Provisioning**: Generates Terraform, applies configuration, health checks

```
Repository → Analyzer → AI Decision Engine → Terraform → AWS Infrastructure
```

**Supported frameworks**: Flask, Django, FastAPI, Express, Next.js, Go apps, and more
**Deployment targets**: EC2 VMs (production-ready), EKS Kubernetes (in development), Lambda (planned)

## 🎯 Advanced Usage

### Natural Language Configuration

SCIA understands infrastructure specifications in your prompts:

```bash
# Specify instance type
./scia deploy "Deploy on a t3.large instance" https://github.com/your-org/app

# Specify region
./scia deploy "Deploy to us-west-2" https://github.com/your-org/app

# Combine multiple parameters
./scia deploy "Deploy to eu-west-1 on a t3.medium with 3 EKS nodes" https://github.com/your-org/app
```

### Command-Line Flags

```bash
# Force a specific strategy
./scia deploy --strategy kubernetes "Deploy this app" https://github.com/your-org/app

# Auto-approve deployment (no confirmation)
./scia deploy -y "Deploy this app" https://github.com/your-org/app

# Specify instance sizing
./scia deploy --ec2-instance-type t3.large --ec2-volume-size 50 "Deploy app" https://...

# EKS cluster sizing
./scia deploy --eks-node-type t3.medium --eks-desired-nodes 3 "Deploy app" https://...

# Verbose output for debugging
./scia --verbose deploy "Deploy app" https://github.com/your-org/app
```

### Configuration

**Using `scia init` (Recommended)**

The easiest way to configure SCIA:

```bash
scia init
```

This interactive command will:
- Configure your LLM provider (Ollama, Gemini, or OpenAI)
- Set up S3 backend for Terraform state (optional)
- Validate your AWS credentials
- Create `~/.scia.yaml` with your preferences

**Manual Configuration**

You can also create `~/.scia.yaml` manually:

```yaml
llm:
  provider: ollama  # or "gemini", "openai"
  ollama:
    model: qwen2.5-coder:7b
    use_docker: true
  # For Gemini:
  # gemini:
  #   api_key: your-api-key
  #   model: gemini-2.0-pro-exp
  # For OpenAI:
  # openai:
  #   api_key: your-api-key
  #   model: gpt-4o

cloud:
  provider: aws
  default_region: us-east-1

terraform:
  bin: tofu  # or "terraform"
  backend:
    type: s3
    s3_bucket: my-terraform-state-bucket
    s3_region: us-east-1
```

**Environment Variables**

Override any config with environment variables (use `SCIA_` prefix):
```bash
export SCIA_LLM_PROVIDER=ollama
export SCIA_CLOUD_DEFAULT_REGION=eu-west-1
export SCIA_VERBOSE=true
```

## 🛠️ Development

```bash
# Build binary
task build

# Run tests
task test

# Lint code
task lint

# Run all checks (test + lint + vulnerability scan)
task check

# Full CI pipeline
task ci
```

**Project structure:**
- `cmd/` - CLI commands (deploy, list, show, destroy, init, etc.)
- `internal/analyzer/` - Repository analysis and framework detection
- `internal/llm/` - AI decision engine with knowledge base
- `internal/terraform/` - Infrastructure provisioning (inline generation)
- `internal/parser/` - Natural language prompt parsing
- `internal/deployer/` - Orchestration and health checking
- `internal/store/` - SQLite database for deployment tracking
- `internal/backend/` - Terraform backend configuration (S3)
- `internal/config/` - Configuration management and validation
- `internal/cloud/` - Cloud provider abstractions (AWS, GCP)

See [CLAUDE.md](CLAUDE.md) for detailed architecture and contribution guidelines.

## ❓ Troubleshooting

### Ollama Issues

**Problem**: Docker container not starting
```bash
# Check Docker is running
docker ps

# Check logs
docker logs scia-ollama

# Restart container
docker restart scia-ollama
```

**Problem**: Model download is slow
- The qwen2.5-coder:7b model is ~4GB - first download takes time
- Use `--verbose` flag to see download progress
- Downloaded models are cached in Docker volume `ollama-data`

### AWS Issues

**Problem**: Deployment fails with credentials error
```bash
# Verify AWS credentials
aws sts get-caller-identity

# Reconfigure if needed
aws configure
```

**Problem**: EC2 instance not accessible
```bash
# Check security group allows inbound traffic on the application port
# SCIA creates security groups automatically but verify in AWS console

# SSH to instance to check logs (replace with your IP)
ssh -i ~/.ssh/your-key.pem ec2-user@<instance-ip>
sudo tail -f /var/log/user-data.log  # Bootstrap logs
sudo tail -f /var/log/app.log        # Application logs
```

### General Issues

**Problem**: Application not starting after deployment
```bash
# Use verbose mode to see detailed logs
./scia --verbose deploy "Deploy app" https://github.com/your-org/app

# Check Terraform state
cd /tmp/scia/terraform/<timestamp>
tofu show

# Check application logs on the deployed instance
```

**Problem**: Want to use local Ollama instead of Docker
```bash
# Set in config file (~/.scia.yaml)
ollama:
  use_docker: false
  url: http://localhost:11434

# Or via environment variable
export SCIA_OLLAMA_USE_DOCKER=false

# Make sure local Ollama is running
ollama serve
ollama pull qwen2.5-coder:7b
```

## 🗺️ Roadmap

- [x] EC2 VM deployments with Auto Scaling Groups
- [x] Natural language prompt parsing
- [x] Docker-based Ollama integration
- [x] LLM-powered deployment decisions with knowledge base
- [x] Multi-provider LLM support (Ollama, Gemini, OpenAI)
- [x] Deployment tracking with SQLite database
- [x] Deployment management (list, show, destroy, outputs, status)
- [x] Interactive configuration with `scia init`
- [x] Terraform state management with S3 backend
- [ ] EKS Kubernetes deployments (code ready, needs testing)
- [ ] AWS Lambda serverless deployments (code ready, needs testing)
- [ ] Support for GCP and Azure
- [ ] Cost estimation before deployment
- [ ] Deployment rollback mechanism
- [ ] Private GitHub repository support
- [ ] Web UI for deployment management

## 📄 License

MIT License

## 🙏 Credits

Built with:
- **Go 1.25** - Modern Go with DWARF5 and container-aware optimizations
- **Ollama** - Local LLM inference (qwen2.5-coder:7b model)
- **OpenTofu** - Open-source Terraform alternative
- **Dagger** - Container-based CI/CD engine
- **Cobra** - CLI framework
- **Viper** - Configuration management

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/Smana/scia/issues)
- **Discussions**: [GitHub Discussions](https://github.com/Smana/scia/discussions)
- **Documentation**: See [CLAUDE.md](CLAUDE.md) for architecture details
