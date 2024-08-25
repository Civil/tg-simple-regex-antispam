package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type ActionCfg struct {
	Name       string         `yaml:"name"`
	ActionName string         `yaml:"action_name"`
	Arguments  map[string]any `yaml:"arguments"`
}

type StatefulFilterConfig struct {
	Name       string         `yaml:"name"`
	FilterName string         `yaml:"filter_name"`
	Arguments  map[string]any `yaml:"arguments"`
	// Order matters
	StatelessFilters []StatelessFilteringRules `yaml:"stateless_filtering_rules"`
	Actions          []ActionCfg               `yaml:"actions"`
}

type StatelessFilteringRules struct {
	Name       string         `yaml:"name"`
	FilterName string         `yaml:"filter_name"`
	Arguments  map[string]any `yaml:"arguments"`
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
			c.StatefulFilters[i].Arguments["state_dir"] = c.DatabaseStateDirectory + "/" + c.StatefulFilters[i].FilterName
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
			FilterName: "spam_filter",
			Name:       "checkNevents",
			Arguments: map[string]any{
				"n":       5,
				"isFinal": false,
			},
			StatelessFilters: []StatelessFilteringRules{
				{
					Name:       "partial_match",
					FilterName: "match bad substring",
					Arguments: map[string]any{
						"match":         "bad_substring",
						"isFinal":       true,
						"caseSensitive": false,
					},
				},
				{
					Name:       "regex",
					FilterName: "contains_regex",
					Arguments:  map[string]any{"regex": ".*[Ss]pam.*"},
				},
			},
			Actions: []ActionCfg{
				{
					Name:       "deleteAndBan",
					ActionName: "ban user and delete their messages",
					Arguments:  nil,
				},
			},
		},
		{
			FilterName: "moderation_filter",
			Name:       "checkNevents",
			Arguments: map[string]any{
				"n":       2,
				"isFinal": false,
			},
			StatelessFilters: []StatelessFilteringRules{
				{
					Name:       "regex",
					FilterName: "contains a regex",
					Arguments:  map[string]any{"regex": ".*[Bb]ad.*"},
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
			FilterName: "report",
			Name:       "report",
			Arguments: map[string]any{
				"is_anonymous_report": false,
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
