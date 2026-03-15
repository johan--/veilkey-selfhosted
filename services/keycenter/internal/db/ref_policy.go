package db

import (
	"fmt"
	"sort"
	"strings"
)

type RefPolicy struct {
	Family        string
	DefaultScope  string
	AllowedScopes map[string]string
}

var refPolicies = map[string]RefPolicy{
	"VK": {
		Family:       "VK",
		DefaultScope: "TEMP",
		AllowedScopes: map[string]string{
			"TEMP":     "temp",
			"LOCAL":    "active",
			"EXTERNAL": "active",
		},
	},
	"VE": {
		Family:       "VE",
		DefaultScope: "TEMP",
		AllowedScopes: map[string]string{
			"TEMP":     "temp",
			"LOCAL":    "active",
			"EXTERNAL": "active",
		},
	},
}

func GetRefPolicy(family string) (RefPolicy, bool) {
	policy, ok := refPolicies[strings.ToUpper(strings.TrimSpace(family))]
	return policy, ok
}

func ListRefPolicies() []RefPolicy {
	policies := make([]RefPolicy, 0, len(refPolicies))
	for _, policy := range refPolicies {
		policies = append(policies, policy)
	}
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Family < policies[j].Family
	})
	return policies
}

func NormalizeRefState(family, scope, status, fallbackScope string) (string, string, error) {
	policy, ok := GetRefPolicy(family)
	if !ok {
		return "", "", fmt.Errorf("unsupported ref family: %s", family)
	}
	scope = strings.ToUpper(strings.TrimSpace(scope))
	if scope == "" {
		scope = strings.ToUpper(strings.TrimSpace(fallbackScope))
	}
	if scope == "" {
		scope = policy.DefaultScope
	}
	defaultStatus, ok := policy.AllowedScopes[scope]
	if !ok {
		return "", "", fmt.Errorf("unsupported %s scope: %s", policy.Family, scope)
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = defaultStatus
	}
	return scope, status, nil
}

func DefaultRefStatusForFamily(family, scope string) string {
	_, status, err := NormalizeRefState(family, scope, "", "")
	if err != nil {
		return "temp"
	}
	return status
}
