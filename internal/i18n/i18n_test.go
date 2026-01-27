package i18n

import (
	"encoding/json"
	"testing"
)

func TestLocalesHaveSameKeysAsEnglish(t *testing.T) {
	en := loadMap(t, "en")
	for _, lang := range []string{"es", "de"} {
		m := loadMap(t, lang)
		for k := range en {
			if _, ok := m[k]; !ok {
				t.Fatalf("%s missing key: %s", lang, k)
			}
		}
	}
}

func loadMap(t *testing.T, lang string) map[string]string {
	t.Helper()
	raw, err := fs.ReadFile("locales/" + lang + ".json")
	if err != nil {
		t.Fatalf("read %s: %v", lang, err)
	}
	m := map[string]string{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse %s: %v", lang, err)
	}
	return m
}
