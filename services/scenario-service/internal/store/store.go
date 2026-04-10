package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"scenario-service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Migrate(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS scenario_scenarios (
    scenario_id TEXT PRIMARY KEY,
    edge_id TEXT NOT NULL,
    name TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    priority INTEGER NOT NULL DEFAULT 0,
    offline_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    triggers JSONB NOT NULL DEFAULT '[]'::jsonb,
    conditions JSONB NOT NULL DEFAULT '[]'::jsonb,
    actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS scenario_decisions (
    decision_id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL DEFAULT '',
    edge_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    matched_scenario_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scenario_scenarios_edge_id ON scenario_scenarios(edge_id);
CREATE INDEX IF NOT EXISTS idx_scenario_scenarios_enabled ON scenario_scenarios(enabled);
CREATE INDEX IF NOT EXISTS idx_scenario_scenarios_offline ON scenario_scenarios(offline_eligible);
CREATE INDEX IF NOT EXISTS idx_scenario_decisions_event_id ON scenario_decisions(event_id);
`

	_, err := s.pool.Exec(ctx, query)
	return err
}

func (s *Store) ListScenarios(ctx context.Context, edgeID string, offlineOnly bool) ([]model.Scenario, error) {
	query := `
SELECT scenario_id, edge_id, name, enabled, priority, offline_eligible, triggers, conditions, actions, created_at, updated_at
FROM scenario_scenarios
WHERE ($1 = '' OR edge_id = $1)
  AND (NOT $2 OR offline_eligible = TRUE)
ORDER BY priority DESC, updated_at DESC
`

	rows, err := s.pool.Query(ctx, query, edgeID, offlineOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scenarios []model.Scenario
	for rows.Next() {
		scenario, err := scanScenario(rows)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, scenario)
	}
	return scenarios, rows.Err()
}

func (s *Store) GetScenario(ctx context.Context, id string) (model.Scenario, error) {
	row := s.pool.QueryRow(ctx, `
SELECT scenario_id, edge_id, name, enabled, priority, offline_eligible, triggers, conditions, actions, created_at, updated_at
FROM scenario_scenarios
WHERE scenario_id = $1
`, id)

	scenario, err := scanScenario(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Scenario{}, ErrNotFound
		}
		return model.Scenario{}, err
	}
	return scenario, nil
}

func (s *Store) UpsertScenario(ctx context.Context, scenario model.Scenario) (model.Scenario, error) {
	triggers, err := json.Marshal(scenario.Triggers)
	if err != nil {
		return model.Scenario{}, err
	}
	conditions, err := json.Marshal(scenario.Conditions)
	if err != nil {
		return model.Scenario{}, err
	}
	actions, err := json.Marshal(scenario.Actions)
	if err != nil {
		return model.Scenario{}, err
	}

	row := s.pool.QueryRow(ctx, `
INSERT INTO scenario_scenarios (
    scenario_id, edge_id, name, enabled, priority, offline_eligible, triggers, conditions, actions, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, COALESCE($10, NOW()), NOW())
ON CONFLICT (scenario_id) DO UPDATE
SET edge_id = EXCLUDED.edge_id,
    name = EXCLUDED.name,
    enabled = EXCLUDED.enabled,
    priority = EXCLUDED.priority,
    offline_eligible = EXCLUDED.offline_eligible,
    triggers = EXCLUDED.triggers,
    conditions = EXCLUDED.conditions,
    actions = EXCLUDED.actions,
    updated_at = NOW()
RETURNING scenario_id, edge_id, name, enabled, priority, offline_eligible, triggers, conditions, actions, created_at, updated_at
`, scenario.ID, scenario.EdgeID, scenario.Name, scenario.Enabled, scenario.Priority, scenario.OfflineEligible, triggers, conditions, actions, zeroToNilTime(scenario.CreatedAt))

	saved, err := scanScenario(row)
	if err != nil {
		return model.Scenario{}, err
	}
	return saved, nil
}

func (s *Store) DeleteScenario(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM scenario_scenarios WHERE scenario_id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SaveDecision(ctx context.Context, eventID, edgeID string, decision model.Decision) error {
	actions, err := json.Marshal(decision.Actions)
	if err != nil {
		return err
	}
	matched, err := json.Marshal(decision.MatchedScenarioIDs)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
INSERT INTO scenario_decisions (decision_id, event_id, edge_id, status, actions, matched_scenario_ids, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, decision.ID, eventID, edgeID, decision.Status, actions, matched, decision.CreatedAt)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanScenario(row scanner) (model.Scenario, error) {
	var scenario model.Scenario
	var triggersRaw []byte
	var conditionsRaw []byte
	var actionsRaw []byte

	if err := row.Scan(
		&scenario.ID,
		&scenario.EdgeID,
		&scenario.Name,
		&scenario.Enabled,
		&scenario.Priority,
		&scenario.OfflineEligible,
		&triggersRaw,
		&conditionsRaw,
		&actionsRaw,
		&scenario.CreatedAt,
		&scenario.UpdatedAt,
	); err != nil {
		return model.Scenario{}, err
	}

	if err := json.Unmarshal(triggersRaw, &scenario.Triggers); err != nil {
		return model.Scenario{}, err
	}
	if err := json.Unmarshal(conditionsRaw, &scenario.Conditions); err != nil {
		return model.Scenario{}, err
	}
	if err := json.Unmarshal(actionsRaw, &scenario.Actions); err != nil {
		return model.Scenario{}, err
	}
	return scenario, nil
}

func zeroToNilTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
