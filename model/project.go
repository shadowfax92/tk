package model

import (
	"os"

	"gopkg.in/yaml.v3"
)

const (
	ProjectStatusTodo     = "todo"
	ProjectStatusNext     = "next"
	ProjectStatusDone     = "done"
	ProjectStatusArchived = "archived"
)

type Project struct {
	Slug   string `yaml:"slug"`
	Title  string `yaml:"title"`
	Status string `yaml:"status"`
}

func (p *Project) IsActive() bool {
	return p.Status != ProjectStatusDone && p.Status != ProjectStatusArchived
}

func ParseProjects(path string) ([]Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var projects []Project
	if err := yaml.Unmarshal(data, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func SaveProjects(path string, projects []Project) error {
	data, err := yaml.Marshal(projects)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
