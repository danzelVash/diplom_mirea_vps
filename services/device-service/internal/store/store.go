package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"device-service/internal/model"

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
CREATE TABLE IF NOT EXISTS device_rooms (
    room_id TEXT PRIMARY KEY,
    edge_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    floor TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS device_devices (
    device_id TEXT PRIMARY KEY,
    edge_id TEXT NOT NULL DEFAULT '',
    room_id TEXT REFERENCES device_rooms(room_id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    device_type TEXT NOT NULL,
    entity_id TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'unknown',
    offline_capable BOOLEAN NOT NULL DEFAULT FALSE,
    last_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS device_commands (
    command_id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL DEFAULT '',
    entity_id TEXT NOT NULL DEFAULT '',
    target_state TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'created',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_rooms_edge_id ON device_rooms(edge_id);
CREATE INDEX IF NOT EXISTS idx_device_devices_edge_id ON device_devices(edge_id);
CREATE INDEX IF NOT EXISTS idx_device_devices_room_id ON device_devices(room_id);
CREATE INDEX IF NOT EXISTS idx_device_devices_entity_id ON device_devices(entity_id);
`

	_, err := s.pool.Exec(ctx, query)
	return err
}

func (s *Store) ListRooms(ctx context.Context, edgeID string) ([]model.Room, error) {
	query := `SELECT room_id, edge_id, name, floor, created_at, updated_at FROM device_rooms`
	args := []any{}
	if edgeID != "" {
		query += ` WHERE edge_id = $1`
		args = append(args, edgeID)
	}
	query += ` ORDER BY created_at ASC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]model.Room, 0)
	for rows.Next() {
		var room model.Room
		if err := rows.Scan(&room.ID, &room.EdgeID, &room.Name, &room.Floor, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (s *Store) GetRoom(ctx context.Context, id string) (model.Room, error) {
	var room model.Room
	err := s.pool.QueryRow(ctx, `
SELECT room_id, edge_id, name, floor, created_at, updated_at
FROM device_rooms WHERE room_id = $1
`, id).Scan(&room.ID, &room.EdgeID, &room.Name, &room.Floor, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Room{}, ErrNotFound
		}
		return model.Room{}, err
	}
	return room, nil
}

func (s *Store) UpsertRoom(ctx context.Context, room model.Room) (model.Room, error) {
	err := s.pool.QueryRow(ctx, `
INSERT INTO device_rooms (room_id, edge_id, name, floor, created_at, updated_at)
VALUES ($1, $2, $3, $4, COALESCE($5, NOW()), NOW())
ON CONFLICT (room_id) DO UPDATE
SET edge_id = EXCLUDED.edge_id,
    name = EXCLUDED.name,
    floor = EXCLUDED.floor,
    updated_at = NOW()
RETURNING room_id, edge_id, name, floor, created_at, updated_at
`, room.ID, room.EdgeID, room.Name, room.Floor, zeroToNilTime(room.CreatedAt)).
		Scan(&room.ID, &room.EdgeID, &room.Name, &room.Floor, &room.CreatedAt, &room.UpdatedAt)
	return room, err
}

func (s *Store) DeleteRoom(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM device_rooms WHERE room_id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListDevices(ctx context.Context, edgeID, roomID string) ([]model.Device, error) {
	query := `
SELECT device_id, edge_id, COALESCE(room_id, ''), name, device_type, entity_id, state, offline_capable, last_changed_at, updated_at
FROM device_devices`
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if edgeID != "" {
		args = append(args, edgeID)
		clauses = append(clauses, fmt.Sprintf("edge_id = $%d", len(args)))
	}
	if roomID != "" {
		args = append(args, roomID)
		clauses = append(clauses, fmt.Sprintf("room_id = $%d", len(args)))
	}
	if len(clauses) > 0 {
		query += ` WHERE ` + strings.Join(clauses, " AND ")
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]model.Device, 0)
	for rows.Next() {
		var device model.Device
		if err := rows.Scan(
			&device.ID,
			&device.EdgeID,
			&device.RoomID,
			&device.Name,
			&device.DeviceType,
			&device.EntityID,
			&device.State,
			&device.OfflineCapable,
			&device.LastChangedAt,
			&device.UpdatedAt,
		); err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, rows.Err()
}

func (s *Store) GetDevice(ctx context.Context, id string) (model.Device, error) {
	var device model.Device
	err := s.pool.QueryRow(ctx, `
SELECT device_id, edge_id, COALESCE(room_id, ''), name, device_type, entity_id, state, offline_capable, last_changed_at, updated_at
FROM device_devices WHERE device_id = $1
`, id).Scan(
		&device.ID,
		&device.EdgeID,
		&device.RoomID,
		&device.Name,
		&device.DeviceType,
		&device.EntityID,
		&device.State,
		&device.OfflineCapable,
		&device.LastChangedAt,
		&device.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Device{}, ErrNotFound
		}
		return model.Device{}, err
	}
	return device, nil
}

func (s *Store) UpsertDevice(ctx context.Context, device model.Device) (model.Device, error) {
	err := s.pool.QueryRow(ctx, `
INSERT INTO device_devices (
    device_id, edge_id, room_id, name, device_type, entity_id, state, offline_capable, last_changed_at, updated_at
)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, COALESCE($9, NOW()), NOW())
ON CONFLICT (device_id) DO UPDATE
SET edge_id = EXCLUDED.edge_id,
    room_id = EXCLUDED.room_id,
    name = EXCLUDED.name,
    device_type = EXCLUDED.device_type,
    entity_id = EXCLUDED.entity_id,
    state = EXCLUDED.state,
    offline_capable = EXCLUDED.offline_capable,
    last_changed_at = EXCLUDED.last_changed_at,
    updated_at = NOW()
RETURNING device_id, edge_id, COALESCE(room_id, ''), name, device_type, entity_id, state, offline_capable, last_changed_at, updated_at
`, device.ID, device.EdgeID, device.RoomID, device.Name, device.DeviceType, device.EntityID, device.State, device.OfflineCapable, zeroToNilTime(device.LastChangedAt)).
		Scan(
			&device.ID,
			&device.EdgeID,
			&device.RoomID,
			&device.Name,
			&device.DeviceType,
			&device.EntityID,
			&device.State,
			&device.OfflineCapable,
			&device.LastChangedAt,
			&device.UpdatedAt,
		)
	return device, err
}

func (s *Store) DeleteDevice(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM device_devices WHERE device_id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SyncInventory(ctx context.Context, edgeID string, rooms []model.Room, devices []model.Device) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, room := range rooms {
		room.EdgeID = edgeID
		if room.ID == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO device_rooms (room_id, edge_id, name, floor, created_at, updated_at)
VALUES ($1, $2, $3, $4, COALESCE($5, NOW()), NOW())
ON CONFLICT (room_id) DO UPDATE
SET edge_id = EXCLUDED.edge_id,
    name = EXCLUDED.name,
    floor = EXCLUDED.floor,
    updated_at = NOW()
`, room.ID, edgeID, room.Name, room.Floor, zeroToNilTime(room.CreatedAt)); err != nil {
			return err
		}
	}

	for _, device := range devices {
		device.EdgeID = edgeID
		if device.ID == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO device_devices (
    device_id, edge_id, room_id, name, device_type, entity_id, state, offline_capable, last_changed_at, updated_at
)
VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, COALESCE($9, NOW()), NOW())
ON CONFLICT (device_id) DO UPDATE
SET edge_id = EXCLUDED.edge_id,
    room_id = EXCLUDED.room_id,
    name = EXCLUDED.name,
    device_type = EXCLUDED.device_type,
    entity_id = EXCLUDED.entity_id,
    state = EXCLUDED.state,
    offline_capable = EXCLUDED.offline_capable,
    last_changed_at = EXCLUDED.last_changed_at,
    updated_at = NOW()
`, device.ID, edgeID, device.RoomID, device.Name, device.DeviceType, device.EntityID, device.State, device.OfflineCapable, zeroToNilTime(device.LastChangedAt)); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) UpdateDeviceState(ctx context.Context, edgeID, deviceID, entityID, state string, changedAt time.Time) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE device_devices
SET edge_id = CASE WHEN $1 <> '' THEN $1 ELSE edge_id END,
    state = $4,
    last_changed_at = $5,
    updated_at = NOW()
WHERE (device_id = $2 AND $2 <> '') OR (entity_id = $3 AND $3 <> '')
`, edgeID, deviceID, entityID, state, changedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ExecuteCommand(ctx context.Context, command model.Command) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
UPDATE device_devices
SET state = $3, last_changed_at = NOW(), updated_at = NOW()
WHERE (device_id = $1 AND $1 <> '') OR (entity_id = $2 AND $2 <> '')
`, command.DeviceID, command.EntityID, command.TargetState)
	if err != nil {
		return "", err
	}
	if tag.RowsAffected() == 0 {
		return "", ErrNotFound
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO device_commands (command_id, device_id, entity_id, target_state, source, status, created_at, executed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, command.ID, command.DeviceID, command.EntityID, command.TargetState, command.Source, command.Status, command.CreatedAt, command.ExecutedAt); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return command.ID, nil
}

func zeroToNilTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
