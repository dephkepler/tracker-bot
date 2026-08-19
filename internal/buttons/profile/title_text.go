package profile

import (
	"fmt"
	"tracker-bot/internal/i18n"
	"tracker-bot/internal/models"
	"tracker-bot/pkg/textbuilder"
)

func ProfileMenuText(lang i18n.Lang, stats *models.ProfileStats) string {
	return fmt.Sprintf(
		"%s\n\n"+
			"%s %d\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s\n"+
			"%s %s",
		i18n.T(lang, i18n.KeyProfileTitle),
		i18n.T(lang, i18n.KeyProfileLabelID), stats.TgUserID,
		i18n.T(lang, i18n.KeyProfileLabelName), textbuilder.StrOrDashMD(stats.UserName),
		i18n.T(lang, i18n.KeyProfileLabelLanguage), textbuilder.StrOrDashMD(stats.Language),
		i18n.T(lang, i18n.KeyProfileLabelTimezone), textbuilder.StrOrDashMD(stats.TimeZone),
		i18n.T(lang, i18n.KeyProfileLabelEmail), textbuilder.StrOrDashMD(stats.Email),
	)
}
