package i18n

// catalogCommon holds strings shared across many screens rather than owned
// by one feature — navigation (Back/Home/Cancel), and the generic
// error/fallback messages the dispatcher sends regardless of which screen
// the user was on.
var catalogCommon = map[string]entry{
	KeyCommonBack: {
		RU: "◀ Назад",
		EN: "◀ Back",
		DE: "◀ Zurück",
		UK: "◀ Назад",
		AR: "◀ رجوع",
	},
	KeyCommonHome: {
		RU: "🏠 Домой",
		EN: "🏠 Home",
		DE: "🏠 Start",
		UK: "🏠 Додому",
		AR: "🏠 الرئيسية",
	},
	KeyCommonCancel: {
		RU: "✖️ Отмена",
		EN: "✖️ Cancel",
		DE: "✖️ Abbrechen",
		UK: "✖️ Скасувати",
		AR: "✖️ إلغاء",
	},
	KeyCommonCancelled: {
		RU: "Отменено.",
		EN: "Cancelled.",
		DE: "Abgebrochen.",
		UK: "Скасовано.",
		AR: "تم الإلغاء.",
	},
	KeyCommonAdmin: {
		RU: "👑 Админ",
		EN: "👑 Admin",
		DE: "👑 Admin",
		UK: "👑 Адмін",
		AR: "👑 المشرف",
	},
	KeyCommonChallenges: {
		RU: "🎯 Челленджи",
		EN: "🎯 Challenges",
		DE: "🎯 Challenges",
		UK: "🎯 Челенджі",
		AR: "🎯 التحديات",
	},
	KeyCommonOnboarding: {
		RU: "🎓 Как это работает",
		EN: "🎓 How it works",
		DE: "🎓 So funktioniert's",
		UK: "🎓 Як це працює",
		AR: "🎓 كيف يعمل هذا",
	},
	KeyCommonUnknownCommand: {
		RU: "Неизвестная команда.",
		EN: "Unknown command.",
		DE: "Unbekannter Befehl.",
		UK: "Невідома команда.",
		AR: "أمر غير معروف.",
	},
	KeyCommonHelpText: {
		RU: "Доступные команды: /start, /help",
		EN: "Available commands: /start, /help",
		DE: "Verfügbare Befehle: /start, /help",
		UK: "Доступні команди: /start, /help",
		AR: "الأوامر المتاحة: /start، /help",
	},
	KeyCommonFallback: {
		RU: "Я тебя понял, но не знаю что с этим сделать. Напиши /help",
		EN: "Got that, but I don't know what to do with it. Try /help",
		DE: "Verstanden, aber ich weiß nicht, was ich damit anfangen soll. Versuch's mit /help",
		UK: "Я тебе зрозумів, але не знаю, що з цим робити. Напиши /help",
		AR: "فهمت، لكن لا أعرف ماذا أفعل بهذا. جرّب /help",
	},
	KeyCommonGenericError: {
		RU: "⚠️ Ошибка. Попробуй ещё раз.",
		EN: "⚠️ Error. Please try again.",
		DE: "⚠️ Fehler. Bitte versuch es noch einmal.",
		UK: "⚠️ Помилка. Спробуй ще раз.",
		AR: "⚠️ حدث خطأ. حاول مرة أخرى.",
	},
	KeyCommonUseButtons: {
		RU: "Используй кнопки меню.",
		EN: "Use buttons from menu.",
		DE: "Benutze die Menü-Buttons.",
		UK: "Використовуй кнопки меню.",
		AR: "استخدم أزرار القائمة.",
	},
	KeyCommonCancelX: {
		RU: "❌ Отмена", EN: "❌ Cancel", DE: "❌ Abbrechen", UK: "❌ Скасувати", AR: "❌ إلغاء",
	},
	KeyCommonDone: {
		RU: "✅ Готово", EN: "✅ Done", DE: "✅ Fertig", UK: "✅ Готово", AR: "✅ تم",
	},
	KeyCommonNameSingleLineInvalid: {
		RU: "⚠️ Имя должно быть в одну строку, 2-60 символов. Попробуй ещё раз:", EN: "⚠️ Name must be a single line, 2-60 characters. Try again:", DE: "⚠️ Der Name muss einzeilig sein, 2-60 Zeichen. Versuch es noch einmal:", UK: "⚠️ Ім'я має бути в один рядок, 2-60 символів. Спробуй ще раз:", AR: "⚠️ يجب أن يكون الاسم في سطر واحد، 2-60 حرفًا. حاول مرة أخرى:",
	},
}
