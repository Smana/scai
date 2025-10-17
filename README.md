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

### Deploy Your First Application

```bash
# SCIA automatically sets up Ollama in Docker on first run
./scia deploy "Deploy this Flask app" https://github.com/Arvo-AI/hello_world
```

**First run will:**
- Pull and start Ollama Docker container
- Download the qwen2.5-coder:7b model (~4GB)
- Analyze the repository
- Generate and apply Terraform configuration
- Return your deployment URL

**Example output:**
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

### Configuration File

Create `~/.scia.yaml` for persistent configuration:

```yaml
workdir: /tmp/scia
verbose: false

ollama:
  model: qwen2.5-coder:7b
  use_docker: true  # Use Docker for Ollama (default)

aws:
  region: us-east-1

terraform:
  bin: tofu  # or "terraform"
```

Environment variables override config file (use `SCIA_` prefix):
```bash
export SCIA_AWS_REGION=eu-west-1
export SCIA_OLLAMA_MODEL=qwen2.5-coder:7b
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
- `cmd/` - CLI commands and entry point
- `internal/analyzer/` - Repository analysis and framework detection
- `internal/llm/` - AI decision engine with knowledge base
- `internal/terraform/` - Infrastructure provisioning and templates
- `internal/parser/` - Natural language prompt parsing
- `internal/deployer/` - Orchestration and health checking

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

- [x] EC2 VM deployments
- [x] Natural language prompt parsing
- [x] Docker-based Ollama integration
- [x] LLM-powered deployment decisions
- [ ] EKS Kubernetes deployments (in progress)
- [ ] AWS Lambda serverless deployments
- [ ] Support for GCP and Azure
- [ ] Cost estimation before deployment
- [ ] Deployment rollback mechanism
- [ ] Private GitHub repository support

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
