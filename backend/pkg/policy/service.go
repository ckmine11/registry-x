package policy

import (
	"context"
	"fmt"
	"sync"

	"github.com/open-policy-agent/opa/rego"
)

type Service struct {
	mu            sync.RWMutex
	CurrentPolicy string
}

func NewService() *Service {
	// Default Policy
	defaultPolicy := `
		package registryx.policy

		default allow = true
		
		violations[msg] {
			input.vulnerabilities.critical > 0
			input.environment == "prod"
			msg := sprintf("Image has %d critical vulnerabilities. Blocked in Prod.", [input.vulnerabilities.critical])
		}
		
		violations[msg] {
			input.environment == "prod"
			input.is_signed == false
			msg := "Image is not signed (cosign signature missing). Blocked in Prod."
		}
		
		allow = false {
			count(violations) > 0
		}
	`
	return &Service{
		CurrentPolicy: defaultPolicy,
	}
}

// GetPolicy returns the current Rego policy.
func (s *Service) GetPolicy() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentPolicy
}

// UpdatePolicy updates the current Rego policy.
func (s *Service) UpdatePolicy(policy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Validate syntax (Simple compile check)
	_, err := rego.New(
		rego.Query("data.registryx.policy.allow"),
		rego.Module("policy.rego", policy),
	).PrepareForEval(context.Background())
	if err != nil {
		return fmt.Errorf("invalid policy syntax: %w", err)
	}

	s.CurrentPolicy = policy
	return nil
}

// EvaluationInput represents the data sent to OPA.
type EvaluationInput struct {
	Repository      string                 `json:"repository"`
	Tag             string                 `json:"tag"`
	Vulnerabilities VulnerabilitySummary   `json:"vulnerabilities"`
	User            string                 `json:"user"`
	Environment     string                 `json:"environment"`
	IsSigned        bool                   `json:"is_signed"`
}

type VulnerabilitySummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
}

// Evaluate checks if the action is allowed.
// Returns allowed (bool) and a list of violation messages.
func (s *Service) Evaluate(ctx context.Context, input EvaluationInput) (bool, []string, error) {
	s.mu.RLock()
	policyStr := s.CurrentPolicy
	s.mu.RUnlock()

	query, err := rego.New(
		rego.Query("data.registryx.policy.allow"),
		rego.Module("policy.rego", policyStr),
	).PrepareForEval(ctx)

	if err != nil {
		return false, nil, fmt.Errorf("failed to prepare rego: %w", err)
	}

	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, nil, fmt.Errorf("failed to eval rego: %w", err)
	}

	if len(results) == 0 {
		return false, nil, fmt.Errorf("undefined result")
	}

	allowed, ok := results[0].Expressions[0].Value.(bool)
	if !ok {
		return false, nil, fmt.Errorf("unexpected result type")
	}

	// Retrieve violations if denied
	var violationMsgs []string
	if !allowed {
		// New query for violations
		vQuery, _ := rego.New(
			rego.Query("data.registryx.policy.violations"),
			rego.Module("policy.rego", policyStr),
		).PrepareForEval(ctx)
		
		vRes, _ := vQuery.Eval(ctx, rego.EvalInput(input))
		if len(vRes) > 0 {
			if msgs, ok := vRes[0].Expressions[0].Value.([]interface{}); ok {
				for _, m := range msgs {
					violationMsgs = append(violationMsgs, fmt.Sprint(m))
				}
			}
		}
	}

	return allowed, violationMsgs, nil
}
