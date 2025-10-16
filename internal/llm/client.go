package llm

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	// TODO: Replace with actual Ollama Go client
	// Common options:
	//   - github.com/ollama/ollama/api (official)
	//   - github.com/jmorganca/ollama/api
	"github.com/Smana/scia/internal/types"
)

// Stub types for Ollama client - replace with actual implementation
type ollamaClient struct {
	baseURL string
}

type GenerateRequest struct {
	Model   string
	Prompt  string
	Options *Options
}

type Options struct {
	Temperature float64
	NumPredict  int
}

type GenerateResponse struct {
	Response string
}

type Client struct {
	client *ollamaClient
	model  string
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		client: &ollamaClient{baseURL: baseURL},
		model:  model,
	}
}

// generate is a stub method - replace with actual Ollama API call
func (c *Client) generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	// TODO: Implement actual Ollama API call
	// For now, return empty response to allow compilation
	return &GenerateResponse{Response: ""}, fmt.Errorf("Ollama client not implemented - add actual Ollama Go SDK")
}

// DetermineStrategy uses LLM with comprehensive context to determine deployment strategy
func (c *Client) DetermineStrategy(userPrompt string, analysis *types.Analysis) (string, error) {
	// Build comprehensive prompt with knowledge base and examples
	prompt := c.buildStrategyPrompt(userPrompt, analysis)

	// Generate response
	req := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Options: &Options{
			Temperature: 0.7,
			NumPredict:  200,
		},
	}

	resp, err := c.generate(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse response
	strategy, reason := c.parseStrategyResponse(resp.Response)

	if strategy == "" {
		// Fallback to simple heuristics if LLM response is unclear
		strategy = c.fallbackStrategy(analysis)
		reason = "Fallback heuristic (LLM response unclear)"
	}

	// Log the decision (optional, for debugging)
	if analysis.Verbose {
		fmt.Printf("LLM Decision: %s\nReason: %s\n", strategy, reason)
	}

	return strategy, nil
}

// buildStrategyPrompt constructs the full prompt with context
func (c *Client) buildStrategyPrompt(userPrompt string, analysis *types.Analysis) string {
	var sb strings.Builder

	// Add knowledge base
	sb.WriteString(DeploymentKnowledgeBase)
	sb.WriteString("\n\n")

	// Add few-shot examples
	sb.WriteString(FewShotExamples)
	sb.WriteString("\n\n")

	// Add the specific question with analysis
	prompt := fmt.Sprintf(DecisionPromptTemplate,
		userPrompt,
		analysis.Framework,
		analysis.Language,
		len(analysis.Dependencies),
		analysis.HasDockerfile,
		analysis.HasDockerCompose,
		analysis.Port,
		analysis.StartCommand,
		c.estimateMemory(analysis),
	)

	sb.WriteString(prompt)

	return sb.String()
}

// parseStrategyResponse extracts strategy and reason from LLM response
func (c *Client) parseStrategyResponse(response string) (strategy string, reason string) {
	response = strings.TrimSpace(response)
	responseLower := strings.ToLower(response)

	// Try to parse structured response
	// Format: STRATEGY: <strategy>\nREASON: <reason>
	strategyRe := regexp.MustCompile(`(?i)STRATEGY:\s*(vm|kubernetes|serverless)`)
	reasonRe := regexp.MustCompile(`(?i)REASON:\s*(.+)`)

	if matches := strategyRe.FindStringSubmatch(response); len(matches) > 1 {
		strategy = strings.ToLower(matches[1])
	}

	if matches := reasonRe.FindStringSubmatch(response); len(matches) > 1 {
		reason = strings.TrimSpace(matches[1])
	}

	// Fallback: check for keywords in response
	if strategy == "" {
		if strings.Contains(responseLower, "kubernetes") || strings.Contains(responseLower, "k8s") {
			strategy = "kubernetes"
		} else if strings.Contains(responseLower, "serverless") || strings.Contains(responseLower, "lambda") {
			strategy = "serverless"
		} else if strings.Contains(responseLower, "vm") || strings.Contains(responseLower, "ec2") {
			strategy = "vm"
		}
	}

	return strategy, reason
}

// fallbackStrategy provides heuristic-based fallback when LLM is unclear
func (c *Client) fallbackStrategy(analysis *types.Analysis) string {
	// Rule 1: Has docker-compose → Kubernetes
	if analysis.HasDockerCompose {
		return "kubernetes"
	}

	// Rule 2: Stateless + minimal deps → Serverless
	if c.isStateless(analysis) && len(analysis.Dependencies) < 5 {
		return "serverless"
	}

	// Rule 3: High dependency count → Kubernetes
	if len(analysis.Dependencies) > 20 {
		return "kubernetes"
	}

	// Rule 4: Has Dockerfile but simple → VM
	if analysis.HasDockerfile && len(analysis.Dependencies) < 15 {
		return "vm"
	}

	// Default: VM (safest choice)
	return "vm"
}

// isStateless checks if application is likely stateless
func (c *Client) isStateless(analysis *types.Analysis) bool {
	// Check framework patterns
	statelessFrameworks := []string{"fastapi", "express"}
	for _, fw := range statelessFrameworks {
		if strings.Contains(strings.ToLower(analysis.Framework), fw) {
			return true
		}
	}

	// Check for stateful dependencies
	statefulDeps := []string{"django", "rails", "postgres", "mysql", "mongodb", "redis", "sqlite"}
	for _, dep := range analysis.Dependencies {
		depLower := strings.ToLower(dep)
		for _, stateful := range statefulDeps {
			if strings.Contains(depLower, stateful) {
				return false
			}
		}
	}

	return false
}

// estimateMemory provides a rough memory estimate based on framework
func (c *Client) estimateMemory(analysis *types.Analysis) string {
	framework := strings.ToLower(analysis.Framework)

	memoryMap := map[string]string{
		"flask":     "256MB-512MB",
		"django":    "512MB-1GB",
		"fastapi":   "128MB-256MB",
		"express":   "128MB-256MB",
		"nextjs":    "256MB-512MB",
		"go":        "50MB-200MB",
		"rails":     "512MB-1GB",
		"streamlit": "256MB-512MB",
	}

	if mem, ok := memoryMap[framework]; ok {
		return mem
	}

	// Default estimate based on language
	languageMem := map[string]string{
		"python":     "256MB-512MB",
		"javascript": "128MB-256MB",
		"typescript": "128MB-256MB",
		"go":         "50MB-200MB",
		"ruby":       "512MB-1GB",
		"java":       "512MB-2GB",
	}

	if mem, ok := languageMem[strings.ToLower(analysis.Language)]; ok {
		return mem
	}

	return "256MB-512MB" // Conservative default
}

// SuggestInstanceType recommends EC2 instance type based on analysis
func (c *Client) SuggestInstanceType(analysis *types.Analysis) string {
	framework := strings.ToLower(analysis.Framework)

	// Heavy frameworks
	heavyFrameworks := []string{"django", "rails", "nextjs"}
	for _, fw := range heavyFrameworks {
		if framework == fw || len(analysis.Dependencies) > 20 {
			return "t3.small" // 2 vCPU, 2GB
		}
	}

	// Very light frameworks
	if framework == "go" || (framework == "fastapi" && len(analysis.Dependencies) < 5) {
		return "t3.micro" // 1 vCPU, 1GB
	}

	// Default: balanced
	return "t3.micro" // 1 vCPU, 1GB (free tier eligible)
}

// SuggestOptimizations provides deployment optimization suggestions
func (c *Client) SuggestOptimizations(analysis *types.Analysis, strategy string) []string {
	var suggestions []string

	// Strategy-specific suggestions
	switch strategy {
	case "vm":
		if !strings.Contains(strings.ToLower(analysis.StartCommand), "gunicorn") &&
			!strings.Contains(strings.ToLower(analysis.StartCommand), "uvicorn") &&
			analysis.Language == "python" {
			suggestions = append(suggestions, "Consider using a production server (Gunicorn/Uvicorn) instead of development server")
		}

		if analysis.Port != 80 && analysis.Port != 443 {
			suggestions = append(suggestions, fmt.Sprintf("Application runs on port %d - consider using a reverse proxy (Nginx) on port 80/443", analysis.Port))
		}

	case "kubernetes":
		if !analysis.HasDockerfile {
			suggestions = append(suggestions, "Create a Dockerfile for containerization")
		}

		suggestions = append(suggestions, "Configure resource limits (CPU/memory) in Kubernetes manifests")
		suggestions = append(suggestions, "Set up horizontal pod autoscaling (HPA) for production")
		suggestions = append(suggestions, "Configure liveness and readiness probes")

	case "serverless":
		if len(analysis.Dependencies) > 10 {
			suggestions = append(suggestions, "High dependency count may increase cold start time - consider Lambda layers")
		}

		suggestions = append(suggestions, "Optimize for cold start by minimizing initialization code")
		suggestions = append(suggestions, "Consider provisioned concurrency for latency-sensitive workloads")
	}

	// General suggestions
	if len(analysis.EnvVars) == 0 {
		suggestions = append(suggestions, "No .env.example found - ensure environment variables are documented")
	}

	if !analysis.HasDockerfile && strategy != "vm" {
		suggestions = append(suggestions, "Creating Dockerfile recommended for consistency across environments")
	}

	return suggestions
}

// ValidateDeploymentRequirements checks if deployment is feasible
func (c *Client) ValidateDeploymentRequirements(analysis *types.Analysis, strategy string) []string {
	var warnings []string

	switch strategy {
	case "serverless":
		// Check for stateful patterns
		if analysis.HasDockerCompose {
			warnings = append(warnings, "⚠️ docker-compose detected but serverless recommended - this may not work")
		}

		statefulFrameworks := []string{"django", "rails"}
		for _, fw := range statefulFrameworks {
			if strings.ToLower(analysis.Framework) == fw {
				warnings = append(warnings, fmt.Sprintf("⚠️ %s is typically stateful - serverless may require significant modifications", analysis.Framework))
			}
		}

	case "kubernetes":
		if !analysis.HasDockerfile && !analysis.HasDockerCompose {
			warnings = append(warnings, "⚠️ Kubernetes recommended but no Dockerfile found - containerization needed")
		}

	case "vm":
		if len(analysis.Dependencies) > 30 {
			warnings = append(warnings, "⚠️ High dependency count - consider Kubernetes for better management")
		}
	}

	// Check for unknown frameworks
	if analysis.Framework == "unknown" {
		warnings = append(warnings, "⚠️ Unable to detect framework - deployment may require manual configuration")
	}

	if analysis.StartCommand == "unknown" {
		warnings = append(warnings, "⚠️ Unable to detect start command - manual configuration required")
	}

	return warnings
}
