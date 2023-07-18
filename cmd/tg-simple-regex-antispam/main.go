package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/actions"
	actionsInterfaces "github.com/Civil/tg-simple-regex-antispam/actions/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/bannedDB"
	"github.com/Civil/tg-simple-regex-antispam/config"
	"github.com/Civil/tg-simple-regex-antispam/filters"
	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/tg"
)

func main() {
	logger := zap.Must(zap.NewProduction())
	zap.ReplaceGlobals(logger)
	defer logger.Sync() // flushes buffer, if any

	filters.New()
	actions.New()

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
						logger.Fatal("failed to load configuration", zap.Error(err))
					}

					banDB, err := bannedDB.New(logger, cfg.BannedDBConfig)
					defer banDB.Close()

					statefulFilters := make([]interfaces.StatefulFilter, 0)
					statelessFilters := filters.GetFilteringRules()

					tbot, err := tg.NewTelego(logger, cfg.TelegramToken, statefulFilters)
					if err != nil {
						logger.Error("error creating bot", zap.Error(err))
					}
					defer tbot.Stop()

					for _, filter := range cfg.StatefulFilters {
						f, err := filters.GetStatefulFilter(filter.Name)
						if err != nil {
							logger.Fatal("error creating stateful filter", zap.Error(err))
							return err
						}

						filteringRules := make([]interfaces.FilteringRule, 0)
						for _, rule := range filter.StatelessFilters {
							fInit, ok := statelessFilters[rule.Name]
							if !ok {
								logger.Fatal("unsupported filtering rule", zap.String("rule", rule.Name), zap.Error(err))
								return err
							}

							f, err := fInit(rule.Arguments)
							if err != nil {
								logger.Fatal("error initializing filtering rule", zap.Error(err))
								return err
							}

							filteringRules = append(filteringRules, f)
						}

						actionsObjs := make([]actionsInterfaces.Action, 0)
						for _, action := range filter.Actions {
							actionInit, err := actions.GetAction(action.Name)
							if err != nil {
								logger.Fatal("error creating action", zap.Error(err))
								return err
							}

							actionObj, err := actionInit(logger, tbot.GetBot(), action.Arguments)
							if err != nil {
								logger.Fatal("error initializing action", zap.Error(err))
								return err
							}

							actionsObjs = append(actionsObjs, actionObj)

						}

						statefulFilter, err := f(logger, banDB, filter.Arguments, filteringRules, actionsObjs)
						if err != nil {
							logger.Fatal("error initializing stateful filter", zap.Error(err))
						}
						defer statefulFilter.Close()
						statefulFilters = append(statefulFilters, statefulFilter)
					}

					logger.Info("starting bot", zap.Any("config", cfg))

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
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
