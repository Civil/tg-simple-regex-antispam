package isForward

import (
	"github.com/mymmrac/telego"

	"github.com/Civil/tg-simple-regex-antispam/filters/interfaces"
	config2 "github.com/Civil/tg-simple-regex-antispam/helper/config"
)

type Filter struct {
	isFinal bool
}

func (r *Filter) Score(msg *telego.Message) int {
	if msg.ForwardOrigin != nil {
		return 100
	}
	return 0
}

func (r *Filter) IsStateful() bool {
	return false
}

func (r *Filter) GetName() string {
	return "isForward"
}

func (r *Filter) GetFilterName() string {
	return ""
}

func (r *Filter) IsFinal() bool {
	return r.isFinal
}

func New(config map[string]any) (interfaces.FilteringRule, error) {
	isFinal, err := config2.GetOptionBoolWithDefault(config, "isFinal", false)
	if err != nil {
		return nil, err
	}

	return &Filter{
		isFinal: isFinal,
	}, nil
}

func Help() string {
	return "isForward checks if the message is forwarded"
}
