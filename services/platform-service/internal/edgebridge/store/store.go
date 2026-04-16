package store

import (
	"context"
	"errors"
	"time"

	"platform-service/internal/edgebridge/model"

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
CREATE TABLE IF NOT EXISTS edge_bridge_edges (
    edge_id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    public_addr TEXT NOT NULL DEFAULT '',
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_inventory_sync TIMESTAMPTZ,
    last_event_at TIMESTAMPTZ,
    last_poll_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS edge_bridge_last_events (
    edge_id TEXT PRIMARY KEY REFERENCES edge_bridge_edges(edge_id) ON DELETE CASCADE,
    event_type TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT '',
    device_id TEXT NOT NULL DEFAULT '',
    entity_id TEXT NOT NULL DEFAULT '',
    room_id TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS edge_bridge_commands (
    command_id TEXT PRIMARY KEY,
    edge_id TEXT NOT NULL,
    device_id TEXT NOT NULL DEFAULT '',
    entity_id TEXT NOT NULL DEFAULT '',
    target_state TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ,
    acked_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'queued',
    last_error TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_edge_bridge_commands_edge_status_created
ON edge_bridge_commands(edge_id, status, created_at);
`

	_, err := s.pool.Exec(ctx, query)
	return err
}

func (s *Store) RegisterEdge(ctx context.Context, req model.EdgeRegistration) (model.EdgeStatus, error) {
	_, err := s.pool.Exec(ctx, `
INSERT INTO edge_bridge_edges (edge_id, name, public_addr, registered_at, last_seen_at, last_error)
VALUES ($1, $2, $3, NOW(), NOW(), '')
ON CONFLICT (edge_id) DO UPDATE
SET name = EXCLUDED.name,
    public_addr = EXCLUDED.public_addr,
    last_seen_at = NOW(),
    last_error = ''
`, req.EdgeID, req.Name, req.PublicAddr)
	if err != nil {
		return model.EdgeStatus{}, err
	}
	return s.GetEdgeStatus(ctx, req.EdgeID)
}

func (s *Store) MarkInventorySynced(ctx context.Context, edgeID string) error {
	_, err := s.pool.Exec(ctx, `
UPDATE edge_bridge_edges
SET last_seen_at = NOW(),
    last_inventory_sync = NOW(),
    last_error = ''
WHERE edge_id = $1
`, edgeID)
	return err
}

func (s *Store) RecordEvent(ctx context.Context, event model.Event) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
INSERT INTO edge_bridge_edges (edge_id, registered_at, last_seen_at, last_event_at, last_error)
VALUES ($1, NOW(), NOW(), $2, '')
ON CONFLICT (edge_id) DO UPDATE
SET last_seen_at = NOW(),
    last_event_at = $2,
    last_error = ''
`, event.EdgeID, event.OccurredAt.UTC())
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
INSERT INTO edge_bridge_last_events (edge_id, event_type, state, device_id, entity_id, room_id, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (edge_id) DO UPDATE
SET event_type = EXCLUDED.event_type,
    state = EXCLUDED.state,
    device_id = EXCLUDED.device_id,
    entity_id = EXCLUDED.entity_id,
    room_id = EXCLUDED.room_id,
    occurred_at = EXCLUDED.occurred_at
`, event.EdgeID, event.EventType, event.State, event.DeviceID, event.EntityID, event.RoomID, event.OccurredAt.UTC())
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) EnqueueCommands(ctx context.Context, edgeID string, commands []model.Command) error {
	if len(commands) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, command := range commands {
		createdAt := command.CreatedAt.UTC()
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		_, err := tx.Exec(ctx, `
INSERT INTO edge_bridge_commands (
    command_id, edge_id, device_id, entity_id, target_state, source, created_at, status, last_error
)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'queued', '')
ON CONFLICT (command_id) DO NOTHING
`, command.CommandID, edgeID, command.DeviceID, command.EntityID, command.TargetState, command.Source, createdAt)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
UPDATE edge_bridge_edges
SET last_seen_at = NOW(),
    last_error = ''
WHERE edge_id = $1
`, edgeID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) PollCommands(ctx context.Context, edgeID string, retryAfter time.Duration) ([]model.Command, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	cutoff := time.Now().UTC().Add(-retryAfter)
	rows, err := tx.Query(ctx, `
WITH due AS (
    SELECT command_id
    FROM edge_bridge_commands
    WHERE edge_id = $1
      AND acked_at IS NULL
      AND (
          status = 'queued'
          OR (status = 'delivered' AND delivered_at <= $2)
      )
    ORDER BY created_at ASC
    FOR UPDATE SKIP LOCKED
),
updated AS (
    UPDATE edge_bridge_commands
    SET status = 'delivered',
        delivered_at = NOW(),
        last_error = ''
    WHERE command_id IN (SELECT command_id FROM due)
    RETURNING command_id, edge_id, device_id, entity_id, target_state, source, created_at
)
SELECT command_id, edge_id, device_id, entity_id, target_state, source, created_at
FROM updated
ORDER BY created_at ASC
`, edgeID, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commands := make([]model.Command, 0)
	for rows.Next() {
		var command model.Command
		if err := rows.Scan(
			&command.CommandID,
			&command.EdgeID,
			&command.DeviceID,
			&command.EntityID,
			&command.TargetState,
			&command.Source,
			&command.CreatedAt,
		); err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
UPDATE edge_bridge_edges
SET last_seen_at = NOW(),
    last_poll_at = NOW()
WHERE edge_id = $1
`, edgeID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return commands, nil
}

func (s *Store) AckCommand(ctx context.Context, edgeID, commandID string) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE edge_bridge_commands
SET acked_at = NOW(),
    status = 'acked',
    last_error = ''
WHERE edge_id = $1 AND command_id = $2
`, edgeID, commandID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	_, err = s.pool.Exec(ctx, `
UPDATE edge_bridge_edges
SET last_seen_at = NOW(),
    last_error = ''
WHERE edge_id = $1
`, edgeID)
	return err
}

func (s *Store) SetEdgeError(ctx context.Context, edgeID, message string) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO edge_bridge_edges (edge_id, registered_at, last_seen_at, last_error)
VALUES ($1, NOW(), NOW(), $2)
ON CONFLICT (edge_id) DO UPDATE
SET last_seen_at = NOW(),
    last_error = $2
`, edgeID, message)
	return err
}

func (s *Store) GetEdgeStatus(ctx context.Context, edgeID string) (model.EdgeStatus, error) {
	var status model.EdgeStatus
	err := s.pool.QueryRow(ctx, `
SELECT
    edge_id,
    name,
    public_addr,
    registered_at,
    last_seen_at,
    COALESCE(last_inventory_sync, 'epoch'::timestamptz),
    COALESCE(last_event_at, 'epoch'::timestamptz),
    COALESCE(last_poll_at, 'epoch'::timestamptz),
    last_error,
    (
        SELECT COUNT(*)
        FROM edge_bridge_commands
        WHERE edge_id = $1 AND acked_at IS NULL
    ) AS pending_commands
FROM edge_bridge_edges
WHERE edge_id = $1
`, edgeID).Scan(
		&status.EdgeID,
		&status.Name,
		&status.PublicAddr,
		&status.RegisteredAt,
		&status.LastSeenAt,
		&status.LastInventorySync,
		&status.LastEventAt,
		&status.LastPollAt,
		&status.LastError,
		&status.PendingCommands,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EdgeStatus{}, ErrNotFound
		}
		return model.EdgeStatus{}, err
	}

	status.LastInventorySync = normalizeTime(status.LastInventorySync)
	status.LastEventAt = normalizeTime(status.LastEventAt)
	status.LastPollAt = normalizeTime(status.LastPollAt)
	return status, nil
}

func (s *Store) GetLastEvent(ctx context.Context, edgeID string) (model.EventSnapshot, error) {
	var snapshot model.EventSnapshot
	err := s.pool.QueryRow(ctx, `
SELECT event_type, state, device_id, entity_id, room_id, occurred_at
FROM edge_bridge_last_events
WHERE edge_id = $1
`, edgeID).Scan(
		&snapshot.EventType,
		&snapshot.State,
		&snapshot.DeviceID,
		&snapshot.EntityID,
		&snapshot.RoomID,
		&snapshot.OccurredAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EventSnapshot{}, ErrNotFound
		}
		return model.EventSnapshot{}, err
	}
	return snapshot, nil
}

func normalizeTime(value time.Time) time.Time {
	if value.Year() <= 1970 {
		return time.Time{}
	}
	return value.UTC()
}
