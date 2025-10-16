# Auto-Deployment Chat System

**Professional Go implementation** of an intelligent auto-deployment system that analyzes code repositories, determines optimal deployment strategies using AI, and automatically provisions infrastructure using Terraform.

## Features

🤖 **AI-Powered Decision Making**
- Uses Ollama LLM with comprehensive deployment knowledge base
- Few-shot learning with real-world examples
- Heuristic fallback rules for fast decisions
- Context-aware recommendations

🔍 **Intelligent Code Analysis**
- Automatic framework detection (Flask, Django, Express, Next.js, Go, etc.)
- Dependency extraction and analysis
- Port and command detection
- Environment variable parsing

🏗️ **Multi-Strategy Deployment**
- **VM (EC2)**: Traditional web apps, simple deployments
- **Kubernetes (EKS)**: Microservices, containerized apps (TODO)
- **Serverless (Lambda)**: Stateless APIs (TODO)

📝 **Comprehensive Logging**
- Detailed deployment logs
- LLM decision reasoning
- Validation warnings
- Optimization suggestions

✅ **Production Ready**
- Built with Go 1.24
- Cobra CLI framework
- Viper configuration management
- HashiCorp terraform-exec integration

---

## Quick Start

### Prerequisites

```bash
# Install Go 1.24+
go version

# Install Dagger (for build and CI tasks)
cd /usr/local
curl -L https://dl.dagger.io/dagger/install.sh | sh

# Install Task (replaces Make)
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh
ollama pull qwen2.5-coder:7b

# Install OpenTofu/Terraform
brew install opentofu  # macOS
# or
curl --proto '=https' --tlsv1.2 -fsSL https://get.opentofu.org/install-opentofu.sh | bash

# Configure AWS
aws configure
```

### Installation

```bash
# Clone repository
git clone <your-repo>/scia
cd scia

# Install dependencies
go mod download

# Build
go build -o scia .

# Install globally (optional)
go install
```

### Usage

```bash
# Deploy application
./scia deploy \
  "Deploy this Flask application on AWS" \
  https://github.com/Arvo-AI/hello_world

# Check status
./scia status

# Destroy infrastructure
./scia destroy --force
```

---

## How It Works

### 1. **Repository Analysis**

```
Clone Repo → Detect Framework → Extract Dependencies → Find Start Command
```

The analyzer examines:
- File structure (`requirements.txt`, `package.json`, `go.mod`)
- Framework patterns (Flask, Django, Express, etc.)
- Configuration files (Dockerfile, docker-compose.yml)
- Port numbers and start commands
- Environment variables

### 2. **AI-Powered Decision Making**

The system uses a **3-tier decision architecture**:

#### **Tier 1: Rule-Based (Fast Path)**
```yaml
# configs/deployment_rules.yaml
- name: multi_service_compose
  conditions:
    has_docker_compose: true
  recommendation: kubernetes
```

For common patterns, instant rule-based decisions provide:
- ⚡ Zero latency
- 🎯 Predictable outcomes
- 📋 Documented logic

#### **Tier 2: LLM with Knowledge Base (Smart Path)**

If no rule matches, the system consults Ollama LLM with:

**Comprehensive Knowledge Base**:
```
- Framework characteristics (memory, startup time, concurrency)
- Deployment decision rules
- AWS best practices
- Common port mappings
- Anti-patterns to avoid
```

**Few-Shot Examples**:
```
Example 1: Flask Hello World → VM
Example 2: Express Microservices → Kubernetes
Example 3: FastAPI Simple API → Serverless
...
```

**Current Analysis**:
```
Framework: Flask
Dependencies: 3
Has Dockerfile: No
Has docker-compose: No
→ LLM Decision: VM
```

#### **Tier 3: Heuristic Fallback (Safety Net)**

If LLM response is unclear:
```go
if hasDockerCompose → kubernetes
else if stateless && deps < 5 → serverless
else if deps > 20 → kubernetes
else → vm
```

### 3. **Terraform Generation**

```go
// Dynamic Terraform generation
template := selectTemplate(strategy)
vars := {
    AppName: "hello-world",
    Port: 5000,
    InstanceType: "t3.micro",
    StartCommand: "python app.py",
    ...
}
renderTemplate(template, vars)
```

### 4. **Infrastructure Provisioning**

```
terraform init → terraform plan → terraform apply → health check
```

Uses `hashicorp/terraform-exec` for robust Terraform automation.

### 5. **Deployment & Verification**

```bash
# User-data script on EC2:
1. Install dependencies (Python, Node.js, etc.)
2. Clone repository
3. Install packages (pip/npm install)
4. Replace localhost → 0.0.0.0
5. Start application
6. Health check endpoint
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI (Cobra)                               │
│  autodeploy deploy [prompt] [repo-url]                      │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│              Repository Analyzer                             │
│  • Framework detection      • Port detection                │
│  • Dependency extraction    • Command detection             │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│           Decision Engine (3-Tier)                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Rule Engine │→ │ LLM + KB    │→ │ Fallback    │        │
│  │  (Fast)     │  │  (Smart)    │  │  (Safe)     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│           Terraform Generator                                │
│  • Select template    • Render configuration                │
│  • Generate user-data • Security groups                     │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│           Terraform Executor (tfexec)                       │
│  init → plan → apply → outputs → health check              │
└─────────────────────────────────────────────────────────────┘
                         │
                         ▼
                   ☁️ AWS Infrastructure
```

---

## Configuration

### Config File: `~/.scia.yaml`

```yaml
workdir: /tmp/scia
verbose: true

ollama:
  url: http://localhost:11434
  model: qwen2.5-coder:7b

aws:
  region: us-east-1

terraform:
  bin: tofu  # or terraform
```

### Environment Variables

```bash
export SCIA_OLLAMA_MODEL=qwen2.5-coder:7b
export SCIA_AWS_REGION=us-west-2
export SCIA_VERBOSE=true
```

### Command-Line Flags

```bash
./scia deploy \
  --work-dir /tmp/custom \
  --verbose \
  --keep \
  "Deploy Flask app" \
  https://github.com/example/repo
```

**Precedence**: Flags > Environment Variables > Config File > Defaults

---

## Example Output

```
$ ./scia deploy "Deploy Flask app" https://github.com/Arvo-AI/hello_world

🤖 Auto-Deployment Chat System
Prompt: Deploy Flask app on AWS
Repository: https://github.com/Arvo-AI/hello_world

📥 Step 1: Analyzing repository...
Cloning into '/tmp/scia/repos/hello_world'...
✅ Detected: flask (python)
   Dependencies: 3 packages
   Port: 5000
   Start Command: python app.py

🧠 Step 2: Determining deployment strategy...
LLM Decision: vm
Reason: Simple Flask application with minimal dependencies - traditional VM deployment suitable
✅ Recommended: vm deployment

📝 Step 3: Generating Terraform configuration...
✅ Generated: /tmp/scia/terraform/1704234567/main.tf

🚀 Step 4: Deploying infrastructure...
  • Initializing Terraform...
  • Planning infrastructure...
  • Provisioning infrastructure (2-3 minutes)...
  • Instance provisioned: 54.123.45.67
  • Waiting for application to start...
  • Application is healthy

============================================================
✅ DEPLOYMENT SUCCESSFUL!
============================================================

🌐 Public URL: http://54.123.45.67:5000
📍 Public IP: 54.123.45.67
🎯 Framework: flask
🔧 Strategy: vm

💡 Optimization Suggestions:
  • Consider using a production server (Gunicorn) instead of development server
  • Application runs on port 5000 - consider using Nginx on port 80/443

⚠️  Use --keep to preserve infrastructure
    Run 'scia destroy' to clean up
```

---

## Supported Frameworks

| Framework | Language | VM | K8s | Serverless | Status |
|-----------|----------|----|----|------------|--------|
| Flask | Python | ✅ | 🚧 | 🚧 | Stable |
| Django | Python | ✅ | 🚧 | ❌ | Stable |
| FastAPI | Python | ✅ | 🚧 | 🚧 | Stable |
| Express | JavaScript | ✅ | 🚧 | 🚧 | Stable |
| Next.js | JavaScript | ✅ | 🚧 | ❌ | Beta |
| Go | Go | ✅ | 🚧 | 🚧 | Beta |
| Rails | Ruby | 🚧 | 🚧 | ❌ | Planned |

Legend: ✅ Implemented | 🚧 In Progress | ❌ Not Suitable

---

## Development

### Project Structure

```
scia/
├── cmd/                    # Cobra commands
│   ├── root.go            # Root + Viper config
│   ├── deploy.go          # Deploy command
│   ├── status.go          # Status command
│   └── destroy.go         # Destroy command
├── internal/
│   ├── analyzer/          # Code analysis
│   │   ├── analyzer.go    # Main analyzer
│   │   ├── framework.go   # Framework detection
│   │   └── dependencies.go
│   ├── llm/               # AI decision engine
│   │   ├── client.go      # Ollama client
│   │   └── knowledge.go   # Knowledge base + examples
│   ├── terraform/         # Terraform automation
│   │   ├── generator.go   # Template generator
│   │   ├── executor.go    # tfexec wrapper
│   │   └── templates/     # TF templates
│   ├── deployer/          # Orchestration
│   │   ├── deployer.go
│   │   └── health.go
│   └── types/
│       └── types.go       # Shared types
├── configs/
│   └── deployment_rules.yaml  # Rule engine
├── go.mod
├── go.sum
├── main.go
├── README.md
└── Taskfile.yml          # Task runner config with Dagger integration
```

### Build & Test

All build and CI tasks use **Dagger** modules from the Daggerverse, orchestrated via **Taskfile**.

```bash
# Build binary
task build

# Run tests
task test

# Lint code
task lint

# Check for vulnerabilities
task vulncheck

# Run all checks (test + lint + vulncheck)
task check

# Run complete CI pipeline
task ci

# Run example deployment
task run-example

# Show all available tasks
task --list
```

**Dagger Modules Used:**
- `github.com/sagikazarmark/daggerverse/go` - Comprehensive Go tooling with caching
- `github.com/purpleclay/daggerverse/golang` - Testing, linting, vulnerability scanning

### Adding New Deployment Types

1. **Create Template**: `internal/terraform/templates/eks_deployment.tf.tmpl`
2. **Update Generator**: Handle new strategy in `generator.go`
3. **Update Knowledge Base**: Add examples to `internal/llm/knowledge.go`
4. **Add Rules**: Update `configs/deployment_rules.yaml`
5. **Test**: `go test ./...`

---

## Context System

### How the AI Makes Decisions

The system provides rich context to the LLM through multiple layers:

#### **Layer 1: Deployment Knowledge Base**
```
- Framework characteristics (memory, startup, concurrency)
- Deployment patterns and best practices
- AWS configuration guidelines
- Common port mappings
- Anti-patterns to avoid
```

#### **Layer 2: Few-Shot Examples**
```
7 real-world examples:
1. Flask Hello World → VM
2. Express Microservices → Kubernetes
3. FastAPI Simple API → Serverless
4. Django E-commerce → VM
5. Next.js React App → VM/K8s
6. Go Microservice → K8s/Serverless
7. Python Batch Job → VM/Lambda
```

#### **Layer 3: Application Analysis**
```
- Framework: Flask
- Language: Python
- Dependencies: 3 packages
- Has Dockerfile: No
- Has docker-compose: No
- Port: 5000
- Start Command: python app.py
- Estimated Memory: 256MB-512MB
```

#### **Layer 4: Decision Validation**
```
- Check deployment feasibility
- Identify warnings (e.g., "serverless recommended for stateful app")
- Suggest optimizations
- Validate requirements
```

### Why This Works

1. **Comprehensive Context**: LLM has expert-level deployment knowledge
2. **Learning by Example**: Few-shot examples guide decision patterns
3. **Fast Fallback**: Rule engine provides instant decisions for common cases
4. **Safety Net**: Heuristics ensure valid decisions even if LLM fails
5. **Explainable**: Every decision includes reasoning

---

## Roadmap

### Phase 1: Core Features (Current)
- [x] Repository analysis
- [x] AI-powered decision making with knowledge base
- [x] VM (EC2) deployment
- [x] Terraform automation
- [x] Health checking

### Phase 2: Extended Deployments (In Progress)
- [ ] Kubernetes (EKS) deployment
- [ ] Serverless (Lambda) deployment
- [ ] Zip file support (in addition to GitHub URLs)
- [ ] Private repository support

### Phase 3: Production Features
- [ ] Multi-region deployment
- [ ] Auto-scaling configuration
- [ ] Cost estimation
- [ ] Rollback support
- [ ] GitOps integration

### Phase 4: Advanced Features
- [ ] RAG for learning from past deployments
- [ ] Custom deployment templates
- [ ] Multi-cloud support (GCP, Azure)
- [ ] Web UI dashboard

---

## Troubleshooting

### Common Issues

**Ollama not responding**:
```bash
# Check Ollama status
curl http://localhost:11434/api/tags

# Restart Ollama
systemctl restart ollama
```

**AWS credentials error**:
```bash
# Verify credentials
aws sts get-caller-identity

# Reconfigure
aws configure
```

**Terraform errors**:
```bash
# Check Terraform version
tofu version

# Clean state
cd /tmp/scia/terraform/<timestamp>
tofu destroy
```

**Application not starting**:
```bash
# Check user-data logs on EC2
ssh ubuntu@<public-ip>
sudo tail -f /var/log/user-data.log
sudo tail -f /var/log/app.log
```

---

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests
4. Submit pull request

---

## License

MIT License - see [LICENSE](LICENSE)

---

## Sources & Dependencies

### AI/ML
- **Ollama**: https://ollama.com (MIT)
- **Qwen2.5-Coder**: https://github.com/QwenLM/Qwen2.5-Coder (Apache 2.0)

### Infrastructure
- **OpenTofu**: https://opentofu.org (MPL 2.0)
- **terraform-exec**: https://github.com/hashicorp/terraform-exec (MPL 2.0)
- **AWS Provider**: https://registry.terraform.io/providers/hashicorp/aws

### Go Libraries
- **Cobra**: https://github.com/spf13/cobra (Apache 2.0)
- **Viper**: https://github.com/spf13/viper (MIT)
- **go-git**: https://github.com/go-git/go-git (Apache 2.0)
- **go-ollama**: https://github.com/JexSrs/go-ollama (MIT)

### References
- AWS Best Practices: https://aws.amazon.com/architecture/well-architected/
- Go 1.24 Documentation: https://tip.golang.org/doc/go1.24
- Terraform Best Practices: https://www.terraform-best-practices.com/

---

## Acknowledgments

Built for the Arvo AI auto-deployment challenge. Special thanks to the open-source community for the excellent libraries and tools.

**Contact**: damian.loch@arvoai.ca
