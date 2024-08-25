package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type ActionCfg struct {
	Name      string         `yaml:"name"`
	Arguments map[string]any `yaml:"arguments"`
}

type StatefulFilterConfig struct {
	Name      string         `yaml:"name"`
	Arguments map[string]any `yaml:"arguments"`
	// Order matters
	StatelessFilters []StatelessFilteringRules `yaml:"stateless_filtering_rules"`
	Actions          []ActionCfg               `yaml:"actions"`
}

type StatelessFilteringRules struct {
	Name      string         `yaml:"name"`
	Arguments map[string]any `yaml:"arguments"`
}

type Config struct {
	TelegramToken          string  `yaml:"telegram_token"`
	AllowedChatIDs         []int64 `yaml:"allowed_chat_ids"`
	AdminIDs               []int64 `yaml:"admin_ids"`
	DatabaseStateDirectory string  `yaml:"database_state_directory"`
	// Order matters
	StatefulFilters []StatefulFilterConfig `yaml:"stateful_filters"`
	BannedDBConfig  map[string]any         `yaml:"banned_db_config"`
}

func (c *Config) Validate() error {
	if c.DatabaseStateDirectory == "" {
		return errors.New("database_state_directory is required")
	}
	if c.TelegramToken == "" {
		return errors.New("telegram_token is required")
	}
	if len(c.AllowedChatIDs) == 0 {
		return errors.New("allowed_chat_ids is required")
	}
	if len(c.AdminIDs) == 0 {
		return errors.New("admin_ids is required")
	}
	return nil
}

func (c *Config) FillDefaults() error {
	for i := range c.StatefulFilters {
		if c.StatefulFilters[i].Arguments == nil || c.StatefulFilters[i].Arguments["state_dir"] == nil {
			if c.StatefulFilters[i].Arguments == nil {
				c.StatefulFilters[i].Arguments = map[string]any{}
			}
			c.StatefulFilters[i].Arguments["state_dir"] = c.DatabaseStateDirectory + "/" + c.StatefulFilters[i].Name
		}
	}

	if c.BannedDBConfig == nil {
		c.BannedDBConfig = map[string]any{}
	}

	if c.BannedDBConfig["state_dir"] == "" {
		c.BannedDBConfig["state_dir"] = c.DatabaseStateDirectory + "/BannedDB"
	}

	return nil
}

func Load(file string) (*Config, error) {
	res := &Config{}

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, res)
	if err != nil {
		return nil, err
	}
	err = res.FillDefaults()
	if err != nil {
		return nil, err
	}
	err = res.Validate()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DefaultConfig() *Config {
	return &Config{}
}

func SampleConfig() *Config {
	res := &Config{}
	res.TelegramToken = "your_telegram_bot_token"
	res.DatabaseStateDirectory = "/path/to/common/database/state"
	res.AllowedChatIDs = []int64{1234567890, 2345678901}
	res.AdminIDs = []int64{1, 2, 3}
	res.StatefulFilters = []StatefulFilterConfig{
		{
			Name:      "spam_filter",
			Arguments: map[string]any{"threshold": 3},
			StatelessFilters: []StatelessFilteringRules{
				{
					Name:      "contains_regex",
					Arguments: map[string]any{"regex": ".*[Ss]pam.*"},
				},
			},
			Actions: []ActionCfg{
				{
					Name:      "deleteAndBan",
					Arguments: nil,
				},
			},
		},
		{
			Name:      "moderation_filter",
			Arguments: map[string]any{"threshold": 2},
			StatelessFilters: []StatelessFilteringRules{
				{
					Name:      "contains_regex",
					Arguments: map[string]any{"regex": ".*[Bb]ad.*"},
				},
			},
			Actions: []ActionCfg{
				{
					Name:      "deleteAndBan",
					Arguments: nil,
				},
			},
		},
		{
			Name:      "report",
			Arguments: map[string]any{"threshold": 2},
			StatelessFilters: []StatelessFilteringRules{
				{
					Name:      "contains_regex",
					Arguments: map[string]any{"regex": "^/report"},
				},
			},
			Actions: []ActionCfg{
				{
					Name:      "addReportButton",
					Arguments: nil,
				},
			},
		},
	}
	_ = res.FillDefaults()
	res.BannedDBConfig = map[string]any{
		"state_dir": "/path/to/banned_db_state/dir",
	}
	return res
}
