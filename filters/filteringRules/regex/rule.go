package regex

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	"github.com/Civil/tg-simple-regex-antispam/filters/types/regexConfig"
	"github.com/Civil/tg-simple-regex-antispam/helper/badger/badgerOpts"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
	"github.com/Civil/tg-simple-regex-antispam/helper/tg"
)

type Filter struct {
	sync.RWMutex
	logger        *zap.Logger
	chainName     string
	regex         []*regexp.Regexp
	isFinal       bool
	caseSensitive bool

	configDB *badger.DB
	reConfig regexConfig.Config

	tg.TGHaveAdminCommands
}

func (r *Filter) Score(msg *telego.Message) int {
	var (
		text    string
		caption string
	)
	if r.caseSensitive {
		text = msg.Text
		caption = msg.Caption
	} else {
		text = strings.ToLower(msg.Text)
		caption = strings.ToLower(msg.Caption)
	}
	r.RLock()
	defer r.RUnlock()
	if len(r.regex) == 0 {
		return 0
	}

	for _, re := range r.regex {
		if re.MatchString(caption) || re.MatchString(text) {
			r.logger.Debug("regex match found", zap.String("regex", re.String()))
			return 100
		}
	}
	return 0
}

func (r *Filter) IsStateful() bool {
	return false
}

func (r *Filter) GetName() string {
	return "regex"
}

func (r *Filter) GetFilterName() string {
	return ""
}

func (r *Filter) IsFinal() bool {
	return r.isFinal
}

var (
	ErrRequiresRegexParameter = errors.New(
		"regex filter requires `regex` parameter to work properly",
	)
	ErrRegexEmpty     = errors.New("regex cannot be empty")
	ErrConfigDirEmpty = errors.New("config_dir cannot be empty")
)

func New(logger *zap.Logger, config map[string]any, chainName string) (interfaces.FilteringRule, error) {
	logger = logger.With(zap.String("filter", chainName), zap.String("filter_type", "regex"))
	configDir, err := config2.GetOptionString(config, "config_dir")
	if err != nil {
		return nil, err
	}
	if configDir == "" {
		return nil, ErrConfigDirEmpty
	}
	configDB, err := badger.Open(badgerOpts.GetBadgerOptions(logger, chainName+"_DB", configDir))
	if err != nil {
		return nil, err
	}

	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", false)
	if err != nil {
		return nil, err
	}

	caseSensitive, err := config2.GetOptionBoolWithDefault(config, "caseSensetive", false)
	if err != nil {
		return nil, err
	}

	res := Filter{
		logger:              logger,
		chainName:           chainName,
		isFinal:             isFinal,
		caseSensitive:       caseSensitive,
		regex:               make([]*regexp.Regexp, 0),
		TGHaveAdminCommands: tg.TGHaveAdminCommands{},
		configDB:            configDB,
	}

	err = res.loadConfig()
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return nil, err
	}
	uniqueRegex := make(map[string]struct{})
	for _, regex := range res.reConfig.Regex {
		if !caseSensitive {
			regex = strings.ToLower(regex)
		}
		if _, ok := uniqueRegex[regex]; !ok {
			uniqueRegex[regex] = struct{}{}
		} else {
			continue
		}
		re, err := regexp.Compile(regex)
		if err != nil {
			continue
		}
		res.regex = append(res.regex, re)
	}

	res.TGHaveAdminCommands.Handlers = map[string]tg.AdminCMDHandlerFunc{
		"help": res.tgHelp,
		"list": res.tgListRegex,
		"add":  res.tgAddRegex,
		"del":  res.tgDelRegex,
	}

	return &res, nil
}

func (r *Filter) tgHelp(logger *zap.Logger, bot *telego.Bot, message *telego.Message, _ []string) error {
	logger.Debug("sending help message")
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("Commands allows to add, list or remove filtering regex (syntax is re2):\n\n")
	for prefix := range r.TGHaveAdminCommands.Handlers {
		buf.WriteString("   " + prefix + "\n")
	}

	err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
	if err != nil {
		logger.Error("failed to send message", zap.Error(err))
	}
	return err
}

func (r *Filter) tgListRegex(logger *zap.Logger, bot *telego.Bot, message *telego.Message, _ []string) error {
	r.RLock()
	defer r.RUnlock()
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString("List of configured regexes:\n\n")
	for _, regex := range r.reConfig.Regex {
		buf.WriteString("   " + regex + "\n")
	}

	err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, buf.String())
	if err != nil {
		logger.Error("failed to send message", zap.Error(err))
	}
	return tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, "End of list")
}

func (r *Filter) tgAddRegex(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error {
	r.Lock()
	defer r.Unlock()
	logger.Debug("adding regex", zap.String("regex", strings.Join(tokens, " ")))
	newRegex := strings.Join(tokens, " ")

	if len(newRegex) == 0 {
		err := tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, "Regex cannot be empty")
		if err != nil {
			return err
		}
		return ErrRegexEmpty
	}

	if !r.caseSensitive {
		newRegex = strings.ToLower(newRegex)
	}

	re, err := regexp.Compile(newRegex)
	if err != nil {
		return tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, fmt.Sprintf("Invalid regex: %v", err))
	}

	// Check if regex already exists
	for _, regex := range r.reConfig.Regex {
		if regex == newRegex {
			return tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, fmt.Sprintf("Regex already exists: %s", newRegex))
		}
	}

	r.reConfig.Regex = append(r.reConfig.Regex, newRegex)
	err = r.saveConfig()
	if err != nil {
		return tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, fmt.Sprintf("Failed to save config: %v", err))
	}
	r.regex = append(r.regex, re)

	return tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, "Done")
}

func (r *Filter) tgDelRegex(logger *zap.Logger, bot *telego.Bot, message *telego.Message, tokens []string) error {
	r.Lock()
	defer r.Unlock()
	logger.Debug("deleting regex", zap.String("regex", strings.Join(tokens, " ")))

	reToDel := strings.Join(tokens, " ")
	index := -1
	for i, regex := range r.reConfig.Regex {
		if regex == reToDel {
			index = i
			break
		}
	}

	if index == -1 {
		return tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, fmt.Sprintf("Regex not found: %s", reToDel))
	}

	r.reConfig.Regex = append(r.reConfig.Regex[:index], r.reConfig.Regex[index+1:]...)
	err := r.saveConfig()
	if err != nil {
		return tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, fmt.Sprintf("Failed to save config: %v", err))
	}

	for i, re := range r.regex {
		if re.String() == reToDel {
			r.regex = append(r.regex[:i], r.regex[i+1:]...)
			break
		}
	}

	return tg.SendMessage(bot, message.Chat.ChatID(), &message.MessageID, "Done")
}

func (r *Filter) saveConfig() error {
	err := r.configDB.Update(func(txn *badger.Txn) error {
		buf, err := proto.Marshal(&r.reConfig)
		if err != nil {
			return err
		}
		return txn.Set([]byte("config"), buf)
	})
	if err != nil {
		return err
	}
	return r.configDB.Sync()
}

func (r *Filter) loadConfig() error {
	err := r.configDB.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("config"))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return proto.Unmarshal(val, &r.reConfig)
		})
	})
	return err
}

func Help() string {
	return "regex requires `config_dir` parameter"
}

func (r *Filter) Close() error {
	return r.configDB.Close()
}

func (r *Filter) TGAdminPrefix() string {
	return r.chainName
}
