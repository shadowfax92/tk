package model

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Goal struct {
	Title    string `yaml:"title"`
	Deadline string `yaml:"deadline"`
	Metric   string `yaml:"metric,omitempty"`
	Current  int    `yaml:"current,omitempty"`
	Target   int    `yaml:"target,omitempty"`
}

func (g *Goal) DeadlineTime() (time.Time, error) {
	return time.Parse("2006-01-02", g.Deadline)
}

func (g *Goal) DaysLeft() int {
	dl, err := g.DeadlineTime()
	if err != nil {
		return 0
	}
	now := time.Now().Truncate(24 * time.Hour)
	dl = dl.Truncate(24 * time.Hour)
	return int(dl.Sub(now).Hours() / 24)
}

func (g *Goal) HasMetric() bool {
	return g.Metric != "" && g.Target > 0
}

func ParseGoals(path string) ([]Goal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var goals []Goal
	if err := yaml.Unmarshal(data, &goals); err != nil {
		return nil, err
	}
	return goals, nil
}
