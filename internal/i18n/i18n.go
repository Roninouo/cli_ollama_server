package i18n

import (
	"embed"
	"encoding/json"
	"os"
	"strings"
)

//go:embed locales/*.json
var fs embed.FS

type Bundle struct {
	lang string
	msg  map[string]string
}

func New(lang string) *Bundle {
	lang = DetectPreferredLang(lang)
	b := &Bundle{lang: lang, msg: map[string]string{}}
	_ = b.load("en")
	if lang != "en" {
		_ = b.load(lang)
	}
	return b
}

func (b *Bundle) Lang() string { return b.lang }

func (b *Bundle) load(lang string) error {
	raw, err := fs.ReadFile("locales/" + lang + ".json")
	if err != nil {
		return err
	}
	m := map[string]string{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	for k, v := range m {
		b.msg[k] = v
	}
	return nil
}

func (b *Bundle) Sprintf(key string, kv ...string) string {
	s, ok := b.msg[key]
	if !ok {
		return "[" + key + "]"
	}
	if len(kv)%2 != 0 {
		return s
	}
	for i := 0; i < len(kv); i += 2 {
		s = strings.ReplaceAll(s, "{"+kv[i]+"}", kv[i+1])
	}
	return s
}

func DetectPreferredLang(in string) string {
	in = strings.TrimSpace(strings.ToLower(in))
	if in == "" || in == "auto" {
		env := os.Getenv("LC_ALL")
		if env == "" {
			env = os.Getenv("LANG")
		}
		env = strings.ToLower(env)
		switch {
		case strings.HasPrefix(env, "es"):
			return "es"
		case strings.HasPrefix(env, "de"):
			return "de"
		default:
			return "en"
		}
	}
	if strings.HasPrefix(in, "en") {
		return "en"
	}
	if strings.HasPrefix(in, "es") {
		return "es"
	}
	if strings.HasPrefix(in, "de") {
		return "de"
	}
	return "en"
}
