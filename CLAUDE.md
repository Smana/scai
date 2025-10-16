# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SCIA (Smart Cloud Infrastructure Automation) is a Go-based CLI tool that analyzes code repositories, uses AI (Ollama LLM) to determine optimal deployment strategies, and automatically provisions infrastructure using Terraform. It supports VM (EC2), Kubernetes (EKS), and Serverless (Lambda) deployments.

**Core Architecture**: The system uses a 3-tier decision architecture:
1. **Rule-Based Fast Path**: YAML-defined heuristics for common patterns (configs/deployment_rules.yaml)
2. **LLM Smart Path**: Ollama with comprehensive knowledge base (3000+ lines) and few-shot examples
3. **Heuristic Fallback**: Safety net for edge cases

## Build & Development Commands

**IMPORTANT**: This project uses **Taskfile** (not Makefile) and **Dagger** for all build and CI tasks. All build operations use Dagger modules from the Daggerverse for consistency and reproducibility.

### Go Version
This project uses **Go 1.25** which includes:
- **DWARF5 debug information** for smaller binaries and faster linking
- **Improved stack allocation** for slices, reducing heap allocations
- **Container-aware GOMAXPROCS** respecting cgroup CPU limits
- **Enhanced error handling** - always check errors immediately before dereferencing values
- **New testing/synctest package** for robust concurrent code testing

### Prerequisites
```bash
# Install Go 1.25
# Download from https://go.dev/dl/

# Install Dagger
curl -L https://dl.dagger.io/dagger/install.sh | sh

# Install Task
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
```

### Building
```bash
# Build binary (uses Dagger with caching)
task build

# Install globally
task install
```

### Testing
```bash
# Run tests (via Dagger)
task test

# Run tests with verbose output
task test-verbose

# Lint code (uses golangci-lint via Dagger)
task lint

# Format code (uses gofumpt via Dagger)
task lint-format

# Check for vulnerabilities (uses govulncheck via Dagger)
task vulncheck

# Run benchmarks
task bench
```

### CI Tasks
```bash
# Run all checks (test + lint + vulncheck)
task check

# Complete CI pipeline (build + test + lint + vulncheck)
task ci
```

### Running
```bash
# Deploy from GitHub URL
./scia deploy "Deploy this Flask app on AWS" https://github.com/Arvo-AI/hello_world

# Deploy from ZIP file
./scia deploy "Deploy this application" /path/to/app.zip

# Example deployment (builds first)
task run-example
```

### Dagger Modules Used
The project leverages popular Daggerverse modules:
- **`github.com/sagikazarmark/daggerverse/go@v0.9.0`**: Build with module/build cache support
- **`github.com/purpleclay/daggerverse/golang@v0.5.0`**: Testing, linting, vulnerability scanning

All Dagger operations are defined in [Taskfile.yml](Taskfile.yml).

### Configuration
The application uses Viper for configuration with precedence: **Flags > Environment Variables > Config File > Defaults**

Config file location: `~/.scia.yaml`
Environment variables: Use `SCIA_` prefix (e.g., `SCIA_OLLAMA_MODEL`)

## Code Architecture

### Package Structure

```
internal/
├── analyzer/           # Repository analysis & framework detection
│   └── zip.go         # ZIP file extraction & analysis
├── llm/               # AI decision engine
│   ├── client.go      # Ollama client wrapper
│   └── knowledge.go   # 3000+ line knowledge base + few-shot examples
├── terraform/         # Infrastructure provisioning
│   ├── generator.go   # Dynamic template generation
│   ├── executor.go    # terraform-exec wrapper
│   └── templates/     # Terraform templates (.tmpl files)
│       ├── ec2_generic.tf.tmpl
│       ├── eks_deployment.tf.tmpl
│       └── lambda_deployment.tf.tmpl
├── deployer/          # Orchestration layer
│   └── health.go      # Health checking
└── types/
    └── types.go       # Shared data structures
```

### Key Data Structures

**Analysis** (types/types.go): Repository analysis results including framework, dependencies, ports, environment variables, and Docker detection.

**DeploymentRule** (types/types.go): YAML-based heuristic rules with priority, conditions, and recommendations.

**TerraformConfig** (types/types.go): Generated Terraform configuration metadata.

**DeploymentResult** (types/types.go): Final deployment outcome with URLs, logs, warnings, and optimization suggestions.

### Decision Flow

1. **Repository Analysis**: Clone/extract repo → detect framework → extract dependencies → identify ports/commands
2. **Strategy Selection**:
   - First check rule engine (configs/deployment_rules.yaml) by priority
   - If no match, query LLM with knowledge base + few-shot examples
   - If LLM unclear, apply heuristic fallback
3. **Terraform Generation**: Select template based on strategy → populate variables → render
4. **Infrastructure Provisioning**: terraform init → plan → apply → extract outputs
5. **Health Check & Validation**: Verify deployment, provide optimization suggestions

### LLM Integration

The LLM system (internal/llm/) provides rich context through:
- **DeploymentKnowledgeBase**: Framework characteristics, deployment patterns, AWS best practices, anti-patterns
- **FewShotExamples**: 7 real-world deployment scenarios with reasoning
- **DecisionPromptTemplate**: Structured prompt for strategy selection

The knowledge base is injected into every LLM query to provide expert-level deployment knowledge.

## Important Implementation Details

### Terraform Templates
- Templates use Go text/template syntax
- Variables are injected dynamically based on analysis results
- All templates include proper security groups, IAM roles, and tagging
- VM template includes user-data script for bootstrapping
- Lambda template includes automatic code packaging with framework adapters

### Framework Detection
The analyzer uses pattern matching across multiple files:
- Python: requirements.txt, setup.py, Pipfile
- JavaScript: package.json, package-lock.json
- Go: go.mod, go.sum
- Detects Flask, Django, FastAPI, Express, Next.js, Rails, etc.

### Port Detection
Scans code files for common patterns like:
- `app.run(port=5000)`
- `app.listen(3000)`
- `:8080` in Go
Falls back to framework defaults if not found.

### ZIP File Support
- Automatic detection via .zip extension (analyzer/zip.go)
- Secure extraction with zip slip protection
- Same analysis pipeline as GitHub repos

### Configuration Precedence
Viper configuration order (cmd/root.go):
1. Command-line flags (--verbose, --work-dir)
2. Environment variables (SCIA_OLLAMA_MODEL, SCIA_AWS_REGION)
3. Config file (~/.scia.yaml)
4. Hardcoded defaults (us-east-1, qwen2.5-coder:7b, etc.)

## Common Development Workflows

### Adding a New Framework
1. Update framework detection logic in `internal/analyzer/framework.go`
2. Add framework characteristics to `internal/llm/knowledge.go` (DeploymentKnowledgeBase)
3. Add example deployment to `internal/llm/knowledge.go` (FewShotExamples)
4. Update deployment rules in `configs/deployment_rules.yaml` if needed
5. Add optimizations to `configs/deployment_rules.yaml` (optimizations section)
6. Test with real repository

### Adding a New Deployment Strategy
1. Create Terraform template in `internal/terraform/templates/` (e.g., `gcp_vm.tf.tmpl`)
2. Update generator.go to handle new strategy
3. Add knowledge to `internal/llm/knowledge.go` about when to use it
4. Add rules to `configs/deployment_rules.yaml`
5. Update types/types.go if new config fields needed

### Modifying LLM Behavior
- Edit `internal/llm/knowledge.go` to change DeploymentKnowledgeBase (decision rules, best practices)
- Add/modify FewShotExamples to teach new patterns
- Update DecisionPromptTemplate to change output format
- The knowledge base is version-controlled and does not require model retraining

### Debugging Deployments
- Use `--verbose` flag for detailed logs
- Check Terraform state in work directory (default: /tmp/scia/terraform/<timestamp>/)
- For VM deployments, SSH to instance and check `/var/log/user-data.log` and `/var/log/app.log`
- Health checks target root path by default

## Dependencies

### Core Libraries
- **cobra**: CLI framework (commands in cmd/)
- **viper**: Configuration management with multi-source support
- **terraform-exec**: Official HashiCorp library for Terraform automation
- **go-ollama**: Ollama LLM client integration
- **go-git**: Git repository operations (if analyzer uses it)

### External Tools Required
- **Go 1.25+**: Required for building (leverages DWARF5, improved allocations, container-aware GOMAXPROCS)
- **Dagger**: Container-based CI/CD engine for build tasks
- **Task**: Modern task runner (replaces Make)
- **Ollama**: LLM inference (default: qwen2.5-coder:7b model)
- **OpenTofu/Terraform**: Infrastructure provisioning
- **AWS CLI**: AWS credentials configuration

## Testing Strategy

Current test coverage is minimal. When adding tests:
- Focus on analyzer package (framework detection, dependency extraction)
- Mock LLM responses for decision engine tests
- Test rule engine matching logic
- Integration tests should use small example repos
- Terraform tests can use `terraform plan` without apply

## Known Limitations

- Only supports AWS currently (no GCP/Azure)
- Kubernetes and Serverless templates are complete but may need real-world validation
- No rollback mechanism yet
- No cost estimation before deployment
- Health checks are basic (HTTP GET to root path)
- Private GitHub repositories not yet supported

## Deployment Rule Priority

Rules in configs/deployment_rules.yaml are evaluated by priority (highest first). Common patterns:
- 100: Multi-service architecture (docker-compose)
- 90: Complex containerized apps
- 80: Simple stateless APIs
- 70: Framework-specific defaults
- 0: Default fallback

When adding rules, use appropriate priority to ensure correct evaluation order.

## Working with the Knowledge Base

The knowledge base (internal/llm/knowledge.go) is the core of AI decision-making. It includes:
- Framework characteristics (memory, startup time, concurrency)
- Decision rules for each strategy (VM, Kubernetes, Serverless)
- Anti-patterns (what NOT to do)
- AWS best practices
- Common port mappings

When modifying, ensure consistency between knowledge base, few-shot examples, and deployment rules.

## Go 1.25 Best Practices

This project follows Go 1.25 best practices:

### Error Handling
**CRITICAL**: Go 1.25 fixed a nil pointer bug from Go 1.21. Always check errors **immediately** before accessing results:

```go
// ✅ CORRECT - Check error before using result
result, err := someFunc()
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}
// Now safe to use result

// ❌ WRONG - Will panic in Go 1.25+
result, err := someFunc()
doSomething(result.Field) // Panic if err != nil!
if err != nil {
    return err
}
```

### Build Flags
The project uses optimized build flags in `.goreleaser.yml`:
- `-trimpath`: Remove build paths for reproducible builds
- `-s -w`: Strip debug info and symbol table
- `mod_timestamp`: Deterministic builds
- DWARF5 is enabled by default in Go 1.25 (smaller binaries, faster linking)

### Performance
- Slice backing storage is auto-allocated on stack when possible (Go 1.25 optimization)
- Use container-aware GOMAXPROCS for proper CPU limit respect in Docker/K8s

### Testing
- Use `testing/synctest` package for testing concurrent code with virtual time
- Run tests with `-race` flag to catch data races (enabled in CI)

### Code Quality
All code must pass golangci-lint with:
- errcheck: No ignored errors
- gosec: Security scanning
- gocyclo: Complexity < 10 (exceptions in llm/ and analyzer/)
- goconst: Repeated strings as constants (exceptions in llm/ and analyzer/)

### Linting Exceptions
- `internal/llm/`: Excluded from gocyclo, goconst, gocritic (AI prompts have natural complexity)
- `internal/analyzer/`: Excluded from goconst (framework names are intentionally hardcoded)
- Test files: Excluded from most linters except critical ones
