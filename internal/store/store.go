package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/danzelVash/diplom_mirea_vps/internal/model"
)

type Store struct {
	mu       sync.RWMutex
	dataPath string
	snapshot model.Snapshot
}

func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data dir: %w", err)
	}

	s := &Store{
		dataPath: filepath.Join(dataDir, "snapshot.json"),
		snapshot: model.Snapshot{
			Commands: map[string][]model.Command{},
		},
	}

	if err := s.load(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw, err := os.ReadFile(s.dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read snapshot: %w", err)
	}

	if len(raw) == 0 {
		return nil
	}

	if err := json.Unmarshal(raw, &s.snapshot); err != nil {
		return fmt.Errorf("unmarshal snapshot: %w", err)
	}
	if s.snapshot.Commands == nil {
		s.snapshot.Commands = map[string][]model.Command{}
	}
	return nil
}

func (s *Store) Save(mutator func(snapshot *model.Snapshot) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := mutator(&s.snapshot); err != nil {
		return err
	}
	if s.snapshot.Commands == nil {
		s.snapshot.Commands = map[string][]model.Command{}
	}

	raw, err := json.MarshalIndent(s.snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	tmpPath := s.dataPath + ".tmp"
	if err := os.WriteFile(tmpPath, raw, 0o644); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	if err := os.Rename(tmpPath, s.dataPath); err != nil {
		return fmt.Errorf("rename snapshot: %w", err)
	}

	return nil
}

func (s *Store) Snapshot() model.Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	raw, _ := json.Marshal(s.snapshot)
	var copy model.Snapshot
	_ = json.Unmarshal(raw, &copy)
	if copy.Commands == nil {
		copy.Commands = map[string][]model.Command{}
	}
	return copy
}
