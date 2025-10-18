# SCIA: Smart Cloud Infrastructure Automation

> **AI-powered deployment that turns natural language into cloud infrastructure**

> ⚠️ **Experimental v1**: This is an initial experimentation and proof-of-concept. See [ROADMAP_V2.md](docs/V2_README.md) for production-ready v2 architecture and upcoming features.

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
```bash
$ scia deploy "Deploy this Flask app with 50GB disk and t3.medium instance" https://github.com/Arvo-AI/hello_world

Using config file: /home/user/.scia.yaml
✓ Database initialized: /home/user/.scia/deployments.db
🐳 Checking Docker Ollama...
✓ Ollama container is already running
✓ Model qwen2.5-coder:7b is already available

✓ Using LLM provider: ollama

🔍 Detected configuration from prompt:
   Strategy: vm
   Region: eu-west-3
   EC2 Instance: t3.medium

🚀 SCIA Deployment Starting...
   User Prompt: Deploy this Flask app with 50GB disk and t3.medium instance
   Repository: https://github.com/Arvo-AI/hello_world
   Work Directory: /tmp/scia
   AWS Region: eu-west-3
   Terraform Binary: tofu

📊 Analyzing repository...
Cloning repository: https://github.com/Arvo-AI/hello_world
   Framework: flask
   Language: python
   Port: 5000
   Dependencies: 1
   Docker: false

🤖 Determining deployment strategy...
   Strategy from prompt: vm

📋 Preparing deployment plan...

                               📋 DEPLOYMENT PLAN

  Strategy: vm
  Region: eu-west-3
  Application: hello-world

# Resources to be Created

┌────────────────────┬─────────────────────┬───────────────────────────┬──────────────────────────┐
│ Resource Type      │ Name                │ Configuration             │ Value                    │
├────────────────────┼─────────────────────┼───────────────────────────┼──────────────────────────┤
│ VPC                │ Default VPC         │   Type                    │ Default VPC              │
│                    │                     │   Region                  │ eu-west-3                │
│                    │                     │                           │                          │
│ Security Group *   │ hello-world-sg      │   Ingress Ports           │ 22 (SSH), 5000 (App)     │
│                    │                     │   Egress                  │ All traffic              │
│                    │                     │   CIDR                    │ 0.0.0.0/0                │
│                    │                     │                           │                          │
│ Auto Scaling Group*│ hello-world-asg     │   Min/Max/Desired         │ 1/1/1                    │
│                    │                     │   Health Check Type       │ EC2                      │
│                    │                     │   Health Check Grace      │ 300s                     │
│                    │                     │                           │                          │
│ EC2 Instance *     │ hello-world (ASG)   │   Instance Type           │ t3.medium                │
│                    │                     │   AMI                     │ Amazon Linux 2023        │
│                    │                     │   Volume Size             │ 50 GB                    │
│                    │                     │   Volume Type             │ GP3 (encrypted)          │
│                    │                     │   Monitoring              │ Enabled                  │
└────────────────────┴─────────────────────┴───────────────────────────┴──────────────────────────┘

 INFO  * = Important resources (will incur costs)

 SUCCESS  Auto-confirmed with --yes flag

   Created deployment record: b2c0091f-af3f-46a4-9b13-213f607b1e1b
   Creating Terraform configuration...
   Running Terraform...

✅ Deployment Complete!

💡 Optimization Suggestions:
   • Consider using a production server (Gunicorn/Uvicorn) instead of development server
   • Application runs on port 5000 - consider using a reverse proxy (Nginx) on port 80/443
   • No .env.example found - ensure environment variables are documented

🎉 Success! Your application is now deployed.
```

### Example: Managing Deployments

```bash
# List all your deployments
$ scia list

                             Found 1 deployment(s)

ID                                   | APP NAME    | STRATEGY | REGION    | STATUS      | CREATED
b2c0091f-af3f-46a4-9b13-213f607b1e1b | hello_world | vm       | eu-west-3 | 🔄 running  | 2025-10-18 14:18

 INFO  Use 'scia show <deployment-id>' to see detailed information


# Show detailed deployment information
$ scia show b2c0091f-af3f-46a4-9b13-213f607b1e1b

                            DEPLOYMENT: hello_world

# 📋 Basic Information

   ID:           b2c0091f-af3f-46a4-9b13-213f607b1e1b
   App Name:     hello_world
   Status:       🔄 running
   Strategy:     vm
   Region:       eu-west-3


# 📦 Repository

   URL:          https://github.com/Arvo-AI/hello_world
   Commit:       21eaaab0957681f6527813b33f1c887e06c20bcf


# 💬 User Prompt

   Deploy this Flask app with 50GB disk and t3.medium instance


# 🔧 Terraform

   State Key:    deployments/b2c0091f-af3f-46a4-9b13-213f607b1e1b/terraform.tfstate
   Directory:    /tmp/scia/terraform


# ⚙️  Configuration

   Framework:    flask
   Language:     python
   Port:         5000
   Instance:     t3.medium
   Start Cmd:    python3 app.py


# 🔗 Outputs

   security_group_id: sg-0e2e442bfb7b6b05e
   application_url: App will be available on port 5000 after instance launches
   asg_name: hello_world-asg-20251018121916369300000007


# 💡 Optimization Suggestions

   • Consider using a production server (Gunicorn/Uvicorn) instead of development server
   • Application runs on port 5000 - consider using a reverse proxy (Nginx) on port 80/443
   • No .env.example found - ensure environment variables are documented


# 🕐 Timestamps

   Created:      2025-10-18 14:18:58 +0200
   Updated:      2025-10-18 14:19:32 +0200


# View outputs only
$ scia outputs b2c0091f-af3f-46a4-9b13-213f607b1e1b

                              Outputs: hello_world

  application_url   = App will be available on port 5000 after instance launches
  asg_name          = hello_world-asg-20251018121916369300000007
  security_group_id = sg-0e2e442bfb7b6b05e


# Destroy when done
$ scia destroy --yes b2c0091f-af3f-46a4-9b13-213f607b1e1b
═══════════════════════════════════════════════════════════════
  DESTROY DEPLOYMENT: hello_world
═══════════════════════════════════════════════════════════════

   ID:           b2c0091f-af3f-46a4-9b13-213f607b1e1b
   App Name:     hello_world
   Strategy:     vm
   Region:       eu-west-3
   Status:       running

 SUCCESS  Auto-confirmed with --yes flag
 INFO  Destroying infrastructure...

 SUCCESS  Deployment destroyed successfully!
 INFO  Deployment ID: b2c0091f-af3f-46a4-9b13-213f607b1e1b
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

SCIA understands infrastructure specifications in your prompts using Ollama LLM:

```bash
# Specify instance type and disk size
scia deploy "Deploy this Flask app with 50GB disk and t3.medium instance" https://github.com/your-org/app

# Specify instance type only
scia deploy "Deploy on a t3.large instance" https://github.com/your-org/app

# Specify region
scia deploy "Deploy to us-west-2" https://github.com/your-org/app

# Combine multiple parameters
scia deploy "Deploy to eu-west-1 on a t3.medium with 3 EKS nodes" https://github.com/your-org/app

# The LLM extracts:
# - ec2_instance_type: t3.medium, t3.large, etc.
# - volume_size: 50, 100, etc. (in GB)
# - region: eu-west-3, us-west-2, etc.
# - eks_min_nodes, eks_max_nodes, eks_desired_nodes
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

**Current Status:**
- [x] EC2 VM deployments with Auto Scaling Groups
- [x] Natural language prompt parsing
- [x] Docker-based Ollama integration
- [x] LLM-powered deployment decisions with knowledge base
- [x] Multi-provider LLM support (Ollama, Gemini, OpenAI)
- [x] Deployment tracking with SQLite database
- [x] Deployment management (list, show, destroy, outputs, status)
- [x] Interactive configuration with `scia init`
- [x] Terraform state management with S3 backend

**Coming Next:**
- [ ] EKS Kubernetes deployments (code ready, needs testing)
- [ ] AWS Lambda serverless deployments (code ready, needs testing)
- [ ] Health checks and application URL verification
- [ ] Support for GCP and Azure
- [ ] Cost estimation before deployment
- [ ] Deployment rollback mechanism
- [ ] Private GitHub repository support
- [ ] Web UI for deployment management

📋 **See [ROADMAP_V2.md](ROADMAP_V2.md) and [docs/](docs/) for detailed future plans and architecture discussions.**

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
