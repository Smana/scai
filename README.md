# Scia: AI-Powered Infrastructure Automation

> **"Just tell it what to deploy, it figures out how."**

**Scia** (from Latin "scio" - *I know* + IA) is an intelligent deployment assistant that analyzes your code, determines the best deployment strategy using AI, and automatically provisions cloud infrastructure - all from a single command.

**No YAML. No configuration files. Just natural language.**

```bash
scia deploy "Deploy this Flask app on AWS" https://github.com/your-org/app
```

That's it. Scia handles the rest - analyzing your code, choosing the right infrastructure (VM, Kubernetes, or Serverless), generating Terraform, and deploying to AWS.


## ⚡ Quick Start

### Prerequisites

You need three things:
1. **AWS credentials** configured (`aws configure`)
2. **Ollama** running locally with qwen2.5-coder model
3. **OpenTofu/Terraform** installed

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh
ollama pull qwen2.5-coder:7b

# Install OpenTofu
brew install opentofu  # macOS
# or: curl --proto '=https' --tlsv1.2 -fsSL https://get.opentofu.org/install-opentofu.sh | bash

# Configure AWS
aws configure
```

### Installation

```bash
# Download and build
git clone https://github.com/Smana/scia
cd scia
task build

# Or install directly with Go
go install github.com/Smana/scia@latest
```

### Deploy in 3 Steps

```bash
# 1. Tell Scia what to deploy
./scia deploy "Deploy this Flask app on AWS" https://github.com/your-org/your-app

# 2. Watch it analyze, decide, and provision
# (takes 2-3 minutes)

# 3. Get your URL
# 🌐 Public URL: http://54.123.45.67:5000
```

That's it! Scia analyzes your code, picks the right infrastructure (VM/K8s/Lambda), generates Terraform, and deploys.

## 🧠 How It Works

Scia uses a **3-tier decision system** to pick the right infrastructure:

1. **Analyze**: Detects your framework (Flask, Express, Go...), dependencies, and configuration
2. **Decide**:
   - **Rules** for common patterns (has docker-compose? → Kubernetes)
   - **AI** (Ollama LLM) for complex decisions with deployment knowledge
   - **Heuristics** as fallback if AI is unclear
3. **Deploy**: Generates Terraform, provisions infrastructure, health checks

```
Your Code → Analyzer → AI Decision → Terraform → AWS Infrastructure
```

**Supports**: Flask, Django, FastAPI, Express, Next.js, Go apps
**Deployment targets**: EC2 VMs (stable), EKS (coming soon), Lambda (coming soon)

---

## 🎯 Example Session

```
$ ./scia deploy "Deploy Flask app" https://github.com/Arvo-AI/hello_world

📥 Analyzing repository...
✅ Detected: flask (python), 3 dependencies, Port 5000

🧠 AI Decision: vm
Reason: Simple Flask app - VM deployment suitable

🚀 Deploying infrastructure... (2-3 min)
✅ DEPLOYMENT SUCCESSFUL!

🌐 http://54.123.45.67:5000
💡 Tip: Use Gunicorn for production
```

## ⚙️ Configuration (Optional)

```bash
# Environment variables
export SCIA_OLLAMA_MODEL=qwen2.5-coder:7b
export SCIA_AWS_REGION=us-west-2

# Or config file: ~/.scia.yaml
workdir: /tmp/scia
verbose: true
ollama:
  model: qwen2.5-coder:7b
aws:
  region: us-east-1
```

---

## 🛠️ Development

```bash
# Build & test
task build    # Build binary
task test     # Run tests
task lint     # Lint code
task ci       # Full CI pipeline

# Uses Dagger for reproducible builds
```

**Project structure**: `cmd/` (CLI) • `internal/analyzer/` (code analysis) • `internal/llm/` (AI decisions) • `internal/terraform/` (provisioning)

See [CLAUDE.md](CLAUDE.md) for detailed architecture and contribution guide

## ❓ Troubleshooting

| Problem | Solution |
|---------|----------|
| Ollama not responding | `curl http://localhost:11434/api/tags` then restart |
| AWS credentials error | Run `aws configure` |
| App not starting | SSH to EC2, check `/var/log/user-data.log` |

---

## 📄 License & Credits

MIT License • Built with Go 1.25, Ollama (qwen2.5-coder), OpenTofu, Terraform

**Contact**: [GitHub Issues](https://github.com/Smana/scia/issues)
