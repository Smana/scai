package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Smana/scai/internal/types"
)

func TestLoadRules(t *testing.T) {
	// Create temporary YAML file for testing
	tmpDir := t.TempDir()
	rulesFile := filepath.Join(tmpDir, "test_rules.yaml")

	yamlContent := `version: "1.0"
rules:
  - name: test_rule
    priority: 100
    description: Test rule
    conditions:
      has_docker_compose: true
    recommendation: kubernetes
    instance_type: null
    reason: Test reason
`

	if err := os.WriteFile(rulesFile, []byte(yamlContent), 0o640); err != nil {
		t.Fatalf("Failed to write test rules file: %v", err)
	}

	rules, err := LoadRules(rulesFile)
	if err != nil {
		t.Fatalf("LoadRules failed: %v", err)
	}

	if rules.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", rules.Version)
	}

	if len(rules.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules.Rules))
	}

	if rules.Rules[0].Name != "test_rule" {
		t.Errorf("Expected rule name 'test_rule', got %s", rules.Rules[0].Name)
	}
}

func TestLoadRulesFileNotFound(t *testing.T) {
	_, err := LoadRules("/nonexistent/path/rules.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestEvaluateRulesDockerCompose(t *testing.T) {
	hasDockerCompose := true
	rules := &types.DeploymentRules{
		Rules: []types.DeploymentRule{
			{
				Name:     "multi_service_compose",
				Priority: 100,
				Conditions: types.RuleConditions{
					HasDockerCompose: &hasDockerCompose,
				},
				Recommendation: "kubernetes",
				Reason:         "Multi-service architecture detected",
			},
		},
	}

	analysis := &types.Analysis{
		HasDockerCompose: true,
		Framework:        "flask",
		Language:         "python",
		Dependencies:     []string{"flask", "requests"},
	}

	match, matched := EvaluateRules(rules, analysis)
	if !matched {
		t.Error("Expected rule to match docker-compose condition")
	}

	if match.Strategy != "kubernetes" {
		t.Errorf("Expected strategy 'kubernetes', got %s", match.Strategy)
	}

	if match.RuleName != "multi_service_compose" {
		t.Errorf("Expected rule name 'multi_service_compose', got %s", match.RuleName)
	}
}

func TestEvaluateRulesFrameworkMatch(t *testing.T) {
	rules := &types.DeploymentRules{
		Rules: []types.DeploymentRule{
			{
				Name:     "django_app",
				Priority: 70,
				Conditions: types.RuleConditions{
					Framework: []string{"django"},
				},
				Recommendation: "vm",
				InstanceType:   "t3.small",
				Reason:         "Django web application",
			},
		},
	}

	analysis := &types.Analysis{
		Framework:    "django",
		Language:     "python",
		Dependencies: []string{"django", "gunicorn"},
	}

	match, matched := EvaluateRules(rules, analysis)
	if !matched {
		t.Error("Expected rule to match framework condition")
	}

	if match.Strategy != "vm" {
		t.Errorf("Expected strategy 'vm', got %s", match.Strategy)
	}

	if match.InstanceType != "t3.small" {
		t.Errorf("Expected instance type 't3.small', got %s", match.InstanceType)
	}
}

func TestEvaluateRulesDependencyCount(t *testing.T) {
	rules := &types.DeploymentRules{
		Rules: []types.DeploymentRule{
			{
				Name:     "high_complexity",
				Priority: 40,
				Conditions: types.RuleConditions{
					MinDependencies: 30,
				},
				Recommendation: "kubernetes",
				Reason:         "High complexity",
			},
		},
	}

	// Test with enough dependencies
	deps := make([]string, 35)
	for i := range deps {
		deps[i] = "dep"
	}

	analysis := &types.Analysis{
		Framework:    "express",
		Language:     "javascript",
		Dependencies: deps,
	}

	match, matched := EvaluateRules(rules, analysis)
	if !matched {
		t.Error("Expected rule to match min dependencies condition")
	}

	if match.Strategy != "kubernetes" {
		t.Errorf("Expected strategy 'kubernetes', got %s", match.Strategy)
	}

	// Test with not enough dependencies
	analysis.Dependencies = []string{"express", "lodash"}
	_, matched = EvaluateRules(rules, analysis)
	if matched {
		t.Error("Expected rule NOT to match with only 2 dependencies")
	}
}

func TestEvaluateRulesMaxDependencies(t *testing.T) {
	hasDockerfile := false
	rules := &types.DeploymentRules{
		Rules: []types.DeploymentRule{
			{
				Name:     "simple_web_app",
				Priority: 60,
				Conditions: types.RuleConditions{
					Framework:       []string{"flask", "express"},
					MaxDependencies: 15,
					HasDockerfile:   &hasDockerfile,
				},
				Recommendation: "vm",
				InstanceType:   "t3.micro",
				Reason:         "Simple web application",
			},
		},
	}

	analysis := &types.Analysis{
		Framework:     "flask",
		Language:      "python",
		Dependencies:  []string{"flask", "requests", "jinja2"},
		HasDockerfile: false,
	}

	match, matched := EvaluateRules(rules, analysis)
	if !matched {
		t.Error("Expected rule to match simple web app conditions")
	}

	if match.Strategy != "vm" {
		t.Errorf("Expected strategy 'vm', got %s", match.Strategy)
	}

	// Test with too many dependencies
	deps := make([]string, 20)
	for i := range deps {
		deps[i] = "dep"
	}
	analysis.Dependencies = deps
	_, matched = EvaluateRules(rules, analysis)
	if matched {
		t.Error("Expected rule NOT to match with 20 dependencies")
	}
}

func TestEvaluateRulesPriority(t *testing.T) {
	hasDockerCompose := true
	hasDockerfile := true

	rules := &types.DeploymentRules{
		Rules: []types.DeploymentRule{
			{
				Name:     "low_priority",
				Priority: 50,
				Conditions: types.RuleConditions{
					HasDockerfile: &hasDockerfile,
				},
				Recommendation: "vm",
			},
			{
				Name:     "high_priority",
				Priority: 100,
				Conditions: types.RuleConditions{
					HasDockerCompose: &hasDockerCompose,
				},
				Recommendation: "kubernetes",
			},
		},
	}

	analysis := &types.Analysis{
		HasDockerCompose: true,
		HasDockerfile:    true,
		Framework:        "flask",
		Language:         "python",
	}

	match, matched := EvaluateRules(rules, analysis)
	if !matched {
		t.Error("Expected rule to match")
	}

	// Should match high priority rule first
	if match.RuleName != "high_priority" {
		t.Errorf("Expected high_priority rule to match first, got %s", match.RuleName)
	}

	if match.Strategy != "kubernetes" {
		t.Errorf("Expected strategy 'kubernetes', got %s", match.Strategy)
	}
}

func TestEvaluateRulesNoMatch(t *testing.T) {
	hasDockerCompose := true
	rules := &types.DeploymentRules{
		Rules: []types.DeploymentRule{
			{
				Name:     "specific_rule",
				Priority: 100,
				Conditions: types.RuleConditions{
					Framework:        []string{"django"},
					Language:         "python",
					HasDockerCompose: &hasDockerCompose,
				},
				Recommendation: "kubernetes",
			},
		},
	}

	analysis := &types.Analysis{
		Framework:        "flask", // Different framework
		Language:         "python",
		HasDockerCompose: true,
	}

	_, matched := EvaluateRules(rules, analysis)
	if matched {
		t.Error("Expected no rule to match")
	}
}

func TestEvaluateRulesNilRules(t *testing.T) {
	analysis := &types.Analysis{
		Framework: "flask",
		Language:  "python",
	}

	_, matched := EvaluateRules(nil, analysis)
	if matched {
		t.Error("Expected no match with nil rules")
	}
}

func TestMatchesConditionsEmptyConditions(t *testing.T) {
	// Empty conditions should match everything (default fallback rule)
	conditions := types.RuleConditions{}

	analysis := &types.Analysis{
		Framework: "flask",
		Language:  "python",
	}

	if !matchesConditions(conditions, analysis) {
		t.Error("Expected empty conditions to match")
	}
}

func TestMatchesConditionsLanguage(t *testing.T) {
	conditions := types.RuleConditions{
		Language: "go",
	}

	analysis := &types.Analysis{
		Framework: "unknown",
		Language:  "go",
	}

	if !matchesConditions(conditions, analysis) {
		t.Error("Expected language condition to match")
	}

	analysis.Language = "python"
	if matchesConditions(conditions, analysis) {
		t.Error("Expected language condition NOT to match")
	}
}
