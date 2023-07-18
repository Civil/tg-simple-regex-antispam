package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ActionCfg struct {
	Name      string                 `yaml:"name"`
	Arguments map[string]interface{} `yaml:"arguments"`
}

type StatefulFilterConfig struct {
	Name      string                 `yaml:"name"`
	Arguments map[string]interface{} `yaml:"arguments"`
	// Order matters
	StatelessFilters []StatelessFilteringRules `yaml:"stateless_filtering_rules"`
	Actions          []ActionCfg               `yaml:"actions"`
}

type StatelessFilteringRules struct {
	Name      string                 `yaml:"name"`
	Arguments map[string]interface{} `yaml:"arguments"`
}

type Config struct {
	TelegramToken  string  `yaml:"telegram_token"`
	AllowedChatIDs []int64 `yaml:"allowed_chat_ids"`
	AdminIDs       []int64 `yaml:"admin_ids"`
	// Order matters
	StatefulFilters []StatefulFilterConfig `yaml:"stateful_filters"`
	BannedDBConfig  map[string]interface{} `yaml:"banned_db_config"`
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
	return res, nil
}
