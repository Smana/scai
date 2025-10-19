package rules

import (
	"fmt"
	"os"
	"slices"

	"github.com/Smana/scai/internal/types"
	"gopkg.in/yaml.v3"
)

// RuleMatch represents a matched deployment rule with its recommendation
type RuleMatch struct {
	Strategy     string
	Reason       string
	InstanceType string
	RuleName     string
}

// LoadRules loads deployment rules from YAML configuration file
func LoadRules(configPath string) (*types.DeploymentRules, error) {
	// #nosec G304 - configPath is from trusted source (hardcoded in code)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	var rules types.DeploymentRules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse rules YAML: %w", err)
	}

	// Sort rules by priority (highest first)
	slices.SortFunc(rules.Rules, func(a, b types.DeploymentRule) int {
		return b.Priority - a.Priority
	})

	return &rules, nil
}

// EvaluateRules evaluates deployment rules against repository analysis
// Returns the first matching rule by priority order
func EvaluateRules(rules *types.DeploymentRules, analysis *types.Analysis) (*RuleMatch, bool) {
	if rules == nil {
		return nil, false
	}

	// Sort rules by priority (highest first) in case they weren't sorted
	slices.SortFunc(rules.Rules, func(a, b types.DeploymentRule) int {
		return b.Priority - a.Priority
	})

	// Iterate through rules in priority order (using index to avoid copying)
	for i := range rules.Rules {
		rule := &rules.Rules[i]
		if matchesConditions(rule.Conditions, analysis) {
			return &RuleMatch{
				Strategy:     rule.Recommendation,
				Reason:       rule.Reason,
				InstanceType: rule.InstanceType,
				RuleName:     rule.Name,
			}, true
		}
	}

	return nil, false
}

// matchesConditions checks if a rule's conditions match the analysis
func matchesConditions(conditions types.RuleConditions, analysis *types.Analysis) bool {
	return matchesFramework(conditions, analysis) &&
		matchesLanguage(conditions, analysis) &&
		matchesDependencies(conditions, analysis) &&
		matchesDockerfile(conditions, analysis) &&
		matchesDockerCompose(conditions, analysis)
}

// matchesFramework checks if framework condition matches
func matchesFramework(conditions types.RuleConditions, analysis *types.Analysis) bool {
	if len(conditions.Framework) == 0 {
		return true
	}
	for _, fw := range conditions.Framework {
		if fw == analysis.Framework {
			return true
		}
	}
	return false
}

// matchesLanguage checks if language condition matches
func matchesLanguage(conditions types.RuleConditions, analysis *types.Analysis) bool {
	if conditions.Language == "" {
		return true
	}
	return conditions.Language == analysis.Language
}

// matchesDependencies checks if dependency count conditions match
func matchesDependencies(conditions types.RuleConditions, analysis *types.Analysis) bool {
	depCount := len(analysis.Dependencies)

	if conditions.MinDependencies > 0 && depCount < conditions.MinDependencies {
		return false
	}

	if conditions.MaxDependencies > 0 && depCount > conditions.MaxDependencies {
		return false
	}

	return true
}

// matchesDockerfile checks if Dockerfile condition matches
func matchesDockerfile(conditions types.RuleConditions, analysis *types.Analysis) bool {
	if conditions.HasDockerfile == nil {
		return true
	}
	return *conditions.HasDockerfile == analysis.HasDockerfile
}

// matchesDockerCompose checks if docker-compose condition matches
func matchesDockerCompose(conditions types.RuleConditions, analysis *types.Analysis) bool {
	if conditions.HasDockerCompose == nil {
		return true
	}
	return *conditions.HasDockerCompose == analysis.HasDockerCompose
}
