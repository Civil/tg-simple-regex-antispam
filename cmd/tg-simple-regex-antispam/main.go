package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"

	"github.com/Civil/tg-simple-regex-antispam/actions"
	actionsInterfaces "github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/config"
	"github.com/Civil/tg-simple-regex-antispam/filters"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/helper/logs"
	"github.com/Civil/tg-simple-regex-antispam/tg"
)

func main() {
	atom := zap.NewAtomicLevel()
	encoderCfg := zap.NewProductionEncoderConfig()

	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			atom,
		))
	zap.ReplaceGlobals(logger)
	_, err := zap.RedirectStdLogAt(logger, zap.ErrorLevel)
	if err != nil {
		logger.Fatal("failed to redirect std logger", zap.Error(err))
	}
	defer func() {
		_ = logger.Sync() // flushes buffer, if any
	}()

	app := &cli.App{
		Name:  "tg-simple-regex-antispam",
		Usage: "A simple telegram bot to filter messages based on regex and other rules",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the bot",
				Action: func(c *cli.Context) error {
					cfg, err := config.Load("config.yaml")
					if err != nil {
						logger.Error("failed to load configuration", zap.Error(err))
						return err
					}

					atom.SetLevel(cfg.LogLevel)

					banDB, err := bannedDB.New(logger, cfg.BannedDBConfig)
					if err != nil {
						logs.ErrNST(logger, "failed initializing bannedDB", err)
						return err
					}
					defer func(logger *zap.Logger) {
						err = banDB.Close()
						if err != nil {
							logger.Error("failed to close bannedDB", zap.Error(err))
						}
					}(logger)

					statefulFilters := make([]interfaces.StatefulFilter, 0)
					statelessFilters := filters.GetFilteringRules()

					tbot, err := tg.New(logger, cfg.TelegramToken, &statefulFilters, cfg.AdminIDs, cfg.AllowedChatIDs, cfg.AdminUsernames, banDB)
					if err != nil {
						logger.Error("error creating bot", zap.Error(err))
						return err
					}
					defer tbot.Stop()

					for _, filter := range cfg.StatefulFilters {
						sfLogger := logger.With(zap.String("filter", filter.Name))
						f, err := filters.GetStatefulFilter(filter.Name)
						if err != nil {
							sfLogger.Error("error creating stateful filter", zap.String("name", filter.Name), zap.Error(err))
							return err
						}

						filteringRules := make([]interfaces.FilteringRule, 0)
						for _, rule := range filter.StatelessFilters {
							fInit, ok := statelessFilters[rule.Name]
							if !ok {
								sfLogger.Error("unsupported filtering rule", zap.String("rule", rule.Name), zap.Error(err))
								return err
							}

							f, err := fInit(sfLogger, rule.Arguments, rule.Name)
							if err != nil {
								sfLogger.Error("error initializing filtering rule", zap.Error(err))
								return err
							}

							filteringRules = append(filteringRules, f)
						}

						actionsObjs := make([]actionsInterfaces.Action, 0)
						for _, action := range filter.Actions {
							actionInit, err := actions.GetAction(action.Name)
							if err != nil {
								logger.Error("error creating action", zap.Error(err))
								return err
							}

							actionObj, err := actionInit(sfLogger, tbot.GetBot(), action.Arguments)
							if err != nil {
								logger.Error("error initializing action", zap.Error(err))
								return err
							}

							actionsObjs = append(actionsObjs, actionObj)

						}

						statefulFilter, err := f(logger, filter.FilterName, banDB, tbot.GetBot(), filter.Arguments, filteringRules, actionsObjs)
						if err != nil {
							logger.Error("error initializing stateful filter", zap.Error(err))
							return err
						}
						defer func() { _ = statefulFilter.Close() }()
						statefulFilters = append(statefulFilters, statefulFilter)
					}

					tbot.UpdatePrefixes()

					logger.Info("starting bot",
						zap.Int64s("allowed_chat_ids", cfg.AllowedChatIDs),
						zap.Int64s("admin_ids", cfg.AdminIDs),
						zap.String("database_state_directory", cfg.DatabaseStateDirectory),
						zap.Any("banned_db_config", cfg.BannedDBConfig),
						zap.String("log_level", cfg.LogLevel.String()),
						zap.Any("stateful_filters", cfg.StatefulFilters),
					)

					tbot.Start()

					err = banDB.SaveState()
					if err != nil {
						logger.Error("error saving banDB state", zap.Error(err))
					}

					for _, statefulFilter := range statefulFilters {
						err = statefulFilter.SaveState()
						if err != nil {
							logger.Error("error saving stateful filter state", zap.Error(err))
						}
					}
					return nil
				},
			},
			{
				Name:  "rules",
				Usage: "List available stateless filtering rules",
				Action: func(c *cli.Context) error {
					help := filters.GetFilteringRulesHelp()
					for filter := range help {
						fmt.Printf("filter `%v`: %v\n", filter, help[filter]())
					}
					return nil
				},
			},
			{
				Name:  "filters",
				Usage: "List available stateful filtering rules",
				Action: func(c *cli.Context) error {
					help := filters.GetStatefulFiltersHelp()
					for filter := range help {
						fmt.Printf("filter `%v`: %v\n", filter, help[filter]())
					}
					return nil
				},
			},
			{
				Name:  "actions",
				Usage: "List available actions",
				Action: func(c *cli.Context) error {
					help := actions.GetHelp()
					for action := range help {
						fmt.Printf("action `%v`: %v\n", action, help[action]())
					}
					return nil
				},
			},
			{
				Name:  "config",
				Usage: "Prints current configuration, as parsed from config.yaml",
				Action: func(c *cli.Context) error {
					cfg, err := config.Load("config.yaml")
					if err != nil {
						cfg = config.DefaultConfig()
					}
					encoder := yaml.NewEncoder(os.Stdout)
					encoder.SetIndent(4)
					err = encoder.Encode(cfg)
					if err != nil {
						logger.Fatal("error encoding config", zap.Error(err))
					}
					return nil
				},
			},
			{
				Name:  "sample_config",
				Usage: "Prints sample configuration",
				Action: func(c *cli.Context) error {
					cfg := config.SampleConfig()
					encoder := yaml.NewEncoder(os.Stdout)
					encoder.SetIndent(4)
					_ = encoder.Encode(cfg)
					return nil
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		logs.ErrNST(logger, "failed to start app", err)
	}
}
