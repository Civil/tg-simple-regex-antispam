package similarlyLookingCharsToRegex

import (
	"bytes"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"

	"github.com/Civil/tg-simple-regex-antispam/transofrmations/interfaces"
)

var (
	charMap map[string][]string = map[string][]string{
		"а": {"a", "α"},
		"б": {"6", "b"},
		"в": {"b"},
		"г": {"r"},
		"д": {},
		"е": {"e", "ё"},
		"ё": {"е", "e"},
		"ж": {},
		"з": {"z", "3"},
		"и": {"u"},
		"й": {"u"},
		"к": {"k"},
		"л": {},
		"м": {"m"},
		"н": {"h"},
		"о": {"o"},
		"п": {"n"},
		"р": {"p", "ρ"},
		"с": {"c"},
		"т": {"t", "m"},
		"у": {"y"},
		"ф": {},
		"х": {"x"},
		"ц": {},
		"ч": {"4"},
		"ш": {},
		"щ": {},
		"ъ": {},
		"ы": {"bi"},
		"ь": {"b"},
		"э": {},
		"ю": {},
		"я": {},
	}
)

type similarlyLookingCharsToRegex struct {
}

func New() interfaces.TransformationRule {
	return &similarlyLookingCharsToRegex{}
}

func (t *similarlyLookingCharsToRegex) Transform(s string) string {
	newStr := bytes.NewBufferString("")
	for _, c := range s {
		if repl, ok := charMap[string(c)]; ok {
			newStr.WriteString("(")
			for _, r := range repl {
				newStr.WriteString(r)
				newStr.WriteString("|")
			}
			newStr.WriteString(string(c))
			newStr.WriteString(")")
			continue
		}
		newStr.WriteString(string(c))
	}
	return s
}

func (t *similarlyLookingCharsToRegex) IsStateful() bool {
	return true
}

func (t *similarlyLookingCharsToRegex) GetName() string {
	return "similarlyLookingCharsToRegex"
}

func (t *similarlyLookingCharsToRegex) GetTransformationName() string {
	return "similarlyLookingCharsToRegex"
}

func (t *similarlyLookingCharsToRegex) TGAdminPrefix() string {
	return "similarlyLookingCharsToRegex"
}

func (t *similarlyLookingCharsToRegex) HandleTGCommands(logger *zap.Logger, bot *telego.Bot, msg *telego.Message,
	tokens []string) error {
	return nil
}
