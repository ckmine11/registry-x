package policy

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/rego"
)

type Service struct {
	mu            sync.RWMutex
	DB            *sql.DB
	CurrentPolicy string
	Config        SecurityPolicyConfig
}

type SecurityPolicyConfig struct {
	CriticalThreshold int  `json:"critical_threshold"`
	HighThreshold     int  `json:"high_threshold"`
	MediumThreshold   int  `json:"medium_threshold"`
	LowThreshold      int  `json:"low_threshold"`
	BlockUnscanned    bool `json:"block_unscanned"`
}

func NewService(db *sql.DB) *Service {
	s := &Service{
		DB: db,
		Config: SecurityPolicyConfig{
			CriticalThreshold: 0,
			HighThreshold:     10,
			MediumThreshold:   100,
			LowThreshold:      1000,
			BlockUnscanned:    true,
		},
	}

	// Dynamic Policy generation is handled in Load
	if err := s.Load(context.Background()); err != nil {
		log.Printf("[Policy] Failed to load policy from DB: %v\n", err)
		// Set a safe default if load fails entirely
		s.CurrentPolicy = s.generateRego(s.Config)
	}

	return s
}

func (s *Service) Load(ctx context.Context) error {
	if s.DB == nil {
		return nil
	}

	var conf SecurityPolicyConfig
	err := s.DB.QueryRowContext(ctx, `
		SELECT critical_threshold, high_threshold, medium_threshold, low_threshold, block_unscanned 
		FROM security_policies 
		ORDER BY updated_at DESC LIMIT 1`).Scan(
		&conf.CriticalThreshold, &conf.HighThreshold, &conf.MediumThreshold, &conf.LowThreshold, &conf.BlockUnscanned)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// Save default policy config if not exists
			_, _ = s.DB.ExecContext(ctx, `
				INSERT INTO security_policies (critical_threshold, high_threshold, medium_threshold, low_threshold, block_unscanned) 
				VALUES ($1, $2, $3, $4, $5)`, 
				s.Config.CriticalThreshold, s.Config.HighThreshold, s.Config.MediumThreshold, s.Config.LowThreshold, s.Config.BlockUnscanned)
			s.mu.Lock()
			s.CurrentPolicy = s.generateRego(s.Config)
			s.mu.Unlock()
			return nil
		}
		return err
	}

	s.mu.Lock()
	s.Config = conf
	s.CurrentPolicy = s.generateRego(conf)
	s.mu.Unlock()
	return nil
}

// generateRego generates OPA policy string from a config.
func (s *Service) generateRego(conf SecurityPolicyConfig) string {
	return fmt.Sprintf(`
		package registryx.policy

		default allow = true
		
		violations[msg] {
			input.vulnerabilities.critical > %d
			msg := sprintf("Image has %%d critical vulnerabilities (threshold is %d).", [input.vulnerabilities.critical])
		}
		
		violations[msg] {
			input.vulnerabilities.high > %d
			msg := sprintf("Image has %%d high vulnerabilities (threshold is %d).", [input.vulnerabilities.high])
		}

		violations[msg] {
			input.vulnerabilities.medium > %d
			msg := sprintf("Image has %%d medium vulnerabilities (threshold is %d).", [input.vulnerabilities.medium])
		}

		violations[msg] {
			input.vulnerabilities.low > %d
			msg := sprintf("Image has %%d low vulnerabilities (threshold is %d).", [input.vulnerabilities.low])
		}

		violations[msg] {
			%t == true
			input.is_scanned == false
			msg := "Image has not been scanned yet. Scans are required before pull."
		}
		
		violations[msg] {
			input.is_signed == false
			input.is_scanned == false
			input.environment == "prod"
			msg := "Image is neither signed nor scanned. Access denied in Production."
		}
		
		allow = false {
			count(violations) > 0
		}
	`, conf.CriticalThreshold, conf.CriticalThreshold, 
	   conf.HighThreshold, conf.HighThreshold,
	   conf.MediumThreshold, conf.MediumThreshold,
	   conf.LowThreshold, conf.LowThreshold,
	   conf.BlockUnscanned)
}

// FetchRepositoryPolicy retrieves overrides for a specific repository.
func (s *Service) FetchRepositoryPolicy(ctx context.Context, repository string) (*SecurityPolicyConfig, error) {
	if s.DB == nil {
		return nil, nil
	}

	var conf SecurityPolicyConfig
	err := s.DB.QueryRowContext(ctx, `
		SELECT critical_threshold, high_threshold, medium_threshold, low_threshold, block_unscanned 
		FROM repository_policies 
		WHERE repository = $1`, repository).Scan(
		&conf.CriticalThreshold, &conf.HighThreshold, &conf.MediumThreshold, &conf.LowThreshold, &conf.BlockUnscanned)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &conf, nil
}

// RepositoryOverride represents a database record for a repo-specific policy.
type RepositoryOverride struct {
	ID                string               `json:"id"`
	RepositoryPath    string               `json:"repository_path"`
	Config            SecurityPolicyConfig `json:"config"`
	UpdatedAt         time.Time            `json:"updated_at"`
}

// ListRepositoryPolicies returns all custom policies.
func (s *Service) ListRepositoryPolicies(ctx context.Context) ([]RepositoryOverride, error) {
	if s.DB == nil {
		return nil, nil
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT repository, critical_threshold, high_threshold, medium_threshold, low_threshold, block_unscanned, updated_at 
		FROM repository_policies ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overrides []RepositoryOverride
	for rows.Next() {
		var o RepositoryOverride
		err := rows.Scan(&o.RepositoryPath, &o.Config.CriticalThreshold, &o.Config.HighThreshold, 
			&o.Config.MediumThreshold, &o.Config.LowThreshold, &o.Config.BlockUnscanned, &o.UpdatedAt)
		if err != nil {
			continue
		}
		overrides = append(overrides, o)
	}

	return overrides, nil
}

// UpdateRepositoryPolicy sets or updates an override.
func (s *Service) UpdateRepositoryPolicy(ctx context.Context, repository string, conf SecurityPolicyConfig) error {
	if s.DB == nil {
		return nil
	}

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO repository_policies (repository, critical_threshold, high_threshold, medium_threshold, low_threshold, block_unscanned, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (repository) 
		DO UPDATE SET critical_threshold = $2, high_threshold = $3, medium_threshold = $4, low_threshold = $5, block_unscanned = $6, updated_at = $7`,
		repository, conf.CriticalThreshold, conf.HighThreshold, conf.MediumThreshold, conf.LowThreshold, conf.BlockUnscanned, time.Now())
	
	return err
}

// DeleteRepositoryPolicy removes an override.
func (s *Service) DeleteRepositoryPolicy(ctx context.Context, repository string) error {
	if s.DB == nil {
		return nil
	}

	_, err := s.DB.ExecContext(ctx, "DELETE FROM repository_policies WHERE repository = $1", repository)
	return err
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

	// Update DB
	if s.DB != nil {
		_, err := s.DB.ExecContext(context.Background(), `
			INSERT INTO system_policies (name, policy_text, updated_at) 
			VALUES ('default', $1, $2)
			ON CONFLICT (name) DO UPDATE SET policy_text = $1, updated_at = $2`,
			policy, time.Now())
		if err != nil {
			log.Printf("[Policy] Failed to save policy to DB: %v\n", err)
		}
	}

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
	IsScanned       bool                   `json:"is_scanned"`
}

type VulnerabilitySummary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// Evaluate checks if the action is allowed.
// Returns allowed (bool) and a list of violation messages.
func (s *Service) Evaluate(ctx context.Context, input EvaluationInput) (bool, []string, error) {
	s.mu.RLock()
	policyStr := s.CurrentPolicy
	s.mu.RUnlock()

	// Check for repository-specific overrides
	effectivePolicyStr := policyStr
	repoConf, err := s.FetchRepositoryPolicy(ctx, input.Repository)
	if err == nil && repoConf != nil {
		effectivePolicyStr = s.generateRego(*repoConf)
	}

	query, err := rego.New(
		rego.Query("data.registryx.policy.allow"),
		rego.Module("policy.rego", effectivePolicyStr),
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
		// Use the same effective policy for violations
		vQuery, _ := rego.New(
			rego.Query("data.registryx.policy.violations"),
			rego.Module("policy.rego", effectivePolicyStr),
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
