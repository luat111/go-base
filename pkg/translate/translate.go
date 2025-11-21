package translate

import (
	"embed"

	"gopkg.in/yaml.v3"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	content   embed.FS
	bundle    *i18n.Bundle
	localizer map[string]*i18n.Localizer
)

func InitTranslate(supportLanguage map[string]string) {
	// load language
	// Create a new i18n bundle with default language.
	bundle = i18n.NewBundle(language.English)
	// Register a toml unmarshal function for i18n bundle.
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	// Load translations from toml files for non-default languages.
	localizer = make(map[string]*i18n.Localizer)
	for lang, path := range supportLanguage {
		bundle.LoadMessageFileFS(content, path)
		localizer[lang] = i18n.NewLocalizer(bundle, lang)
	}
}

func TryLocalize(l *i18n.Localizer, lc *i18n.LocalizeConfig) (string, error) {
	localized, err := l.Localize(lc)
	if err != nil {
		return "", err
	}
	return localized, nil
}

func GetMessage(code string, lang string, param map[string]interface{}) string {
	if loc, found := localizer[lang]; found {
		str, err := TryLocalize(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID: code,
			},
			TemplateData: param,
		})
		if err == nil && str != "" {
			return str
		}
		return code
	}
	return code
}

func GetListMessage(codes []string, lang string) []string {
	listMessage := []string{}
	for _, code := range codes {
		if loc, found := localizer[lang]; found {
			str, err := TryLocalize(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID: code,
				},
			})
			if err == nil && str != "" {
				listMessage = append(listMessage, str)
			} else {
				listMessage = append(listMessage, code)
			}
		} else {
			listMessage = append(listMessage, code)
		}
	}
	return listMessage
}
