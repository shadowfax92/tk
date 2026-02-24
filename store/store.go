package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nickhudkins/tk/config"
	"github.com/nickhudkins/tk/model"
	"gopkg.in/yaml.v3"
)

type Store struct {
	Root string
	Cfg  *config.Config
}

type Meta struct {
	NextID int `yaml:"next_id"`
}

func New(cfg *config.Config) (*Store, error) {
	if err := os.MkdirAll(cfg.Root, 0755); err != nil {
		return nil, err
	}
	return &Store{Root: cfg.Root, Cfg: cfg}, nil
}

func (s *Store) metaPath() string {
	return filepath.Join(s.Root, ".meta.yaml")
}

func (s *Store) focusPath() string {
	return filepath.Join(s.Root, ".focus.md")
}

func (s *Store) taskPath(id int) string {
	return filepath.Join(s.Root, fmt.Sprintf("%03d.md", id))
}

func (s *Store) loadMeta() (*Meta, error) {
	m := &Meta{NextID: 1}

	data, err := os.ReadFile(s.metaPath())
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Store) saveMeta(m *Meta) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(s.metaPath(), data, 0644)
}

func (s *Store) NextID() (int, error) {
	m, err := s.loadMeta()
	if err != nil {
		return 0, err
	}
	id := m.NextID
	m.NextID++
	return id, s.saveMeta(m)
}

func (s *Store) Add(title, body string) (*model.Task, error) {
	id, err := s.NextID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	t := &model.Task{
		ID:      id,
		Title:   title,
		Status:  model.StatusInbox,
		Created: now,
		Updated: now,
		Body:    body,
	}

	return t, t.SaveTo(s.taskPath(id))
}

func (s *Store) Get(id int) (*model.Task, error) {
	return model.ParseFile(s.taskPath(id))
}

func (s *Store) Save(t *model.Task) error {
	t.Updated = time.Now()
	return t.SaveTo(s.taskPath(t.ID))
}

func (s *Store) Delete(id int) error {
	return os.Remove(s.taskPath(id))
}

func (s *Store) TaskFilePath(id int) string {
	return s.taskPath(id)
}

func (s *Store) List(filter func(*model.Task) bool) ([]*model.Task, error) {
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return nil, err
	}

	var tasks []*model.Task
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		// Extract ID from filename
		name := strings.TrimSuffix(e.Name(), ".md")
		if _, err := strconv.Atoi(name); err != nil {
			continue
		}

		t, err := model.ParseFile(filepath.Join(s.Root, e.Name()))
		if err != nil {
			continue
		}

		if filter == nil || filter(t) {
			tasks = append(tasks, t)
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	return tasks, nil
}

func (s *Store) ReadFocus() (string, error) {
	data, err := os.ReadFile(s.focusPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (s *Store) FocusFilePath() string {
	return s.focusPath()
}

func (s *Store) EnsureFocus() error {
	if _, err := os.Stat(s.focusPath()); err == nil {
		return nil
	}
	return os.WriteFile(s.focusPath(), []byte("- [your focus here]\n"), 0644)
}
