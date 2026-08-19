package i18n

// catalogTrack holds the Track main screen, activity management, timer, and
// archive. Reports (Today/Calendar/Period) is a separate later phase — not
// in this file.
var catalogTrack = map[string]entry{
	KeyTrackButtonSelectActivity: {
		RU: "📂 Активности", EN: "📂 Activities", DE: "📂 Aktivitäten", UK: "📂 Активності", AR: "📂 الأنشطة",
	},
	KeyTrackButtonCreateActivity: {
		RU: "➕ Новая активность", EN: "➕ New Activity", DE: "➕ Neue Aktivität", UK: "➕ Нова активність", AR: "➕ نشاط جديد",
	},
	KeyTrackButtonViewReports: {
		RU: "📈 Отчёты", EN: "📈 Reports", DE: "📈 Berichte", UK: "📈 Звіти", AR: "📈 التقارير",
	},
	KeyTrackButtonViewArchive: {
		RU: "🗄 Архив", EN: "🗄 Archive", DE: "🗄 Archiv", UK: "🗄 Архів", AR: "🗄 الأرشيف",
	},
	KeyTrackButtonActivityActivate: {
		RU: "📳 Активировать", EN: "📳 Activate", DE: "📳 Aktivieren", UK: "📳 Активувати", AR: "📳 تفعيل",
	},
	KeyTrackButtonActivityDelete: {
		RU: "🗑 Удалить", EN: "🗑 Delete", DE: "🗑 Löschen", UK: "🗑 Видалити", AR: "🗑 حذف",
	},
	KeyTrackButtonTimerCreate: {
		RU: "➕ Свой таймер", EN: "➕ Custom Timer", DE: "➕ Eigener Timer", UK: "➕ Свій таймер", AR: "➕ مؤقت مخصص",
	},
	KeyTrackButtonTimerDelete: {
		RU: "🗑 Удалить таймер", EN: "🗑 Delete Timer", DE: "🗑 Timer löschen", UK: "🗑 Видалити таймер", AR: "🗑 حذف المؤقت",
	},

	KeyTrackLabelBack: {
		RU: "↩️ Назад", EN: "↩️ Back", DE: "↩️ Zurück", UK: "↩️ Назад", AR: "↩️ رجوع",
	},
	KeyTrackLabelOpenActivities: {
		RU: "📂 Открыть активности", EN: "📂 Open Activities", DE: "📂 Aktivitäten öffnen", UK: "📂 Відкрити активності", AR: "📂 فتح الأنشطة",
	},
	KeyTrackLabelOpenArchive: {
		RU: "🗄 Открыть архив", EN: "🗄 Open Archive", DE: "🗄 Archiv öffnen", UK: "🗄 Відкрити архів", AR: "🗄 فتح الأرشيف",
	},
	KeyTrackLabelCreateAnother: {
		RU: "➕ Создать ещё", EN: "➕ Create Another", DE: "➕ Weitere erstellen", UK: "➕ Створити ще", AR: "➕ إنشاء آخر",
	},
	KeyTrackLabelArchiveSelected: {
		RU: "🛒 В архив выбранные", EN: "🛒 Archive selected", DE: "🛒 Auswahl archivieren", UK: "🛒 В архів вибрані", AR: "🛒 أرشفة المحدد",
	},
	KeyTrackLabelActiveActivities: {
		RU: "📂 Активные активности", EN: "📂 Active activities", DE: "📂 Aktive Aktivitäten", UK: "📂 Активні активності", AR: "📂 الأنشطة النشطة",
	},
	KeyTrackLabelRestore: {
		RU: "♻ Восстановить", EN: "♻ Restore", DE: "♻ Wiederherstellen", UK: "♻ Відновити", AR: "♻ استعادة",
	},
	KeyTrackLabelDeleteForever: {
		RU: "🗑 Удалить навсегда", EN: "🗑 Delete forever", DE: "🗑 Endgültig löschen", UK: "🗑 Видалити назавжди", AR: "🗑 حذف نهائي",
	},
	KeyTrackLabelStopTimer: {
		RU: "⏹ Остановить таймер", EN: "⏹ Stop Timer", DE: "⏹ Timer stoppen", UK: "⏹ Зупинити таймер", AR: "⏹ إيقاف المؤقت",
	},

	KeyTrackMainTitle: {
		RU: "📈 Трекинг", EN: "📈 Tracking", DE: "📈 Tracking", UK: "📈 Трекінг", AR: "📈 التتبع",
	},
	KeyTrackMainCurrentActivity: {
		RU: "📌 Текущая активность:", EN: "📌 Current activity:", DE: "📌 Aktuelle Aktivität:", UK: "📌 Поточна активність:", AR: "📌 النشاط الحالي:",
	},
	KeyTrackMainTodayTime: {
		RU: "⏱ Отслежено сегодня:", EN: "⏱ Tracked today:", DE: "⏱ Heute erfasst:", UK: "⏱ Відстежено сьогодні:", AR: "⏱ تم تتبعه اليوم:",
	},
	KeyTrackMainStreak: {
		RU: "🔥 Серия:", EN: "🔥 Streak:", DE: "🔥 Serie:", UK: "🔥 Серія:", AR: "🔥 السلسلة:",
	},
	KeyTrackMainTodayCount: {
		RU: "✅ Сессий сегодня:", EN: "✅ Sessions today:", DE: "✅ Sitzungen heute:", UK: "✅ Сесій сьогодні:", AR: "✅ الجلسات اليوم:",
	},
	KeyTrackMainProgress: {
		RU: "Прогресс: %s (%d%%, цель %s)", EN: "Progress: %s (%d%%, target %s)", DE: "Fortschritt: %s (%d%%, Ziel %s)", UK: "Прогрес: %s (%d%%, ціль %s)", AR: "التقدم: %s (%d%%، الهدف %s)",
	},
	KeyTrackLoadFailed: {
		RU: "⚠️ Не удалось загрузить данные трекера. Попробуй ещё раз.", EN: "⚠️ Failed to load tracking data. Please try again.", DE: "⚠️ Tracking-Daten konnten nicht geladen werden. Bitte versuch es noch einmal.", UK: "⚠️ Не вдалося завантажити дані трекера. Спробуй ще раз.", AR: "⚠️ تعذّر تحميل بيانات التتبع. حاول مرة أخرى.",
	},

	KeyTrackCreatePrompt: {
		RU: "📌 *Новая активность*\n\nВведи название активности:", EN: "📌 *Create New Activity*\n\nEnter activity name:", DE: "📌 *Neue Aktivität erstellen*\n\nGib den Namen der Aktivität ein:", UK: "📌 *Нова активність*\n\nВведи назву активності:", AR: "📌 *إنشاء نشاط جديد*\n\nأدخل اسم النشاط:",
	},
	KeyTrackCreatePromptBlocked: {
		RU: "Используй кнопки меню. Введи название активности обычным текстом.", EN: "Use buttons from menu. Enter activity name as plain text.", DE: "Benutze die Menü-Buttons. Gib den Aktivitätsnamen als reinen Text ein.", UK: "Використовуй кнопки меню. Введи назву активності звичайним текстом.", AR: "استخدم أزرار القائمة. أدخل اسم النشاط كنص عادي.",
	},
	KeyTrackCreateEmptyName: {
		RU: "Название активности не может быть пустым.", EN: "Activity name cannot be empty.", DE: "Der Aktivitätsname darf nicht leer sein.", UK: "Назва активності не може бути порожньою.", AR: "لا يمكن أن يكون اسم النشاط فارغًا.",
	},
	KeyTrackCreateAlreadyExists: {
		RU: "Такая активность уже существует.", EN: "Activity already exists.", DE: "Diese Aktivität existiert bereits.", UK: "Така активність вже існує.", AR: "هذا النشاط موجود بالفعل.",
	},
	KeyTrackCreateFailed: {
		RU: "⚠️ Не удалось создать активность.", EN: "⚠️ Failed to create activity.", DE: "⚠️ Aktivität konnte nicht erstellt werden.", UK: "⚠️ Не вдалося створити активність.", AR: "⚠️ تعذّر إنشاء النشاط.",
	},
	KeyTrackCreateConfirmed: {
		RU: "Создано: %s", EN: "Created: %s", DE: "Erstellt: %s", UK: "Створено: %s", AR: "تم الإنشاء: %s",
	},

	KeyTrackManageMenuClosed: {
		RU: "Меню активностей закрыто. Открой Активности заново из Трекера.", EN: "Activities menu is closed. Open Activities again from Track.", DE: "Das Aktivitätenmenü ist geschlossen. Öffne Aktivitäten erneut über Tracker.", UK: "Меню активностей закрито. Відкрий Активності знову з Трекера.", AR: "قائمة الأنشطة مغلقة. افتح الأنشطة مرة أخرى من التتبع.",
	},
	KeyTrackManageLoadFailed: {
		RU: "⚠️ Не удалось загрузить активности.", EN: "⚠️ Failed to load activities.", DE: "⚠️ Aktivitäten konnten nicht geladen werden.", UK: "⚠️ Не вдалося завантажити активності.", AR: "⚠️ تعذّر تحميل الأنشطة.",
	},
	KeyTrackManageEmpty: {
		RU: "Активностей пока нет. Сначала создай одну.", EN: "No activities yet. Create one first.", DE: "Noch keine Aktivitäten. Erstelle zuerst eine.", UK: "Активностей поки немає. Спочатку створи одну.", AR: "لا توجد أنشطة بعد. أنشئ واحدًا أولاً.",
	},
	KeyTrackManageSelectTitle: {
		RU: "📂 Выбор активности\n\nВыбрано: %d из %d", EN: "📂 Select Activity\n\nSelected: %d of %d", DE: "📂 Aktivität auswählen\n\nAusgewählt: %d von %d", UK: "📂 Вибір активності\n\nВибрано: %d із %d", AR: "📂 اختيار النشاط\n\nتم اختيار: %d من %d",
	},
	KeyTrackInvalidActivityID: {
		RU: "Некорректный ID активности.", EN: "Invalid activity id.", DE: "Ungültige Aktivitäts-ID.", UK: "Некоректний ID активності.", AR: "معرّف نشاط غير صالح.",
	},
	KeyTrackManageToggleFailed: {
		RU: "⚠️ Не удалось изменить выбор активности.", EN: "⚠️ Failed to update activity selection.", DE: "⚠️ Auswahl der Aktivität konnte nicht geändert werden.", UK: "⚠️ Не вдалося змінити вибір активності.", AR: "⚠️ تعذّر تحديث اختيار النشاط.",
	},
	KeyTrackManageRefreshFailed: {
		RU: "⚠️ Не удалось обновить активности.", EN: "⚠️ Failed to refresh activities.", DE: "⚠️ Aktivitäten konnten nicht aktualisiert werden.", UK: "⚠️ Не вдалося оновити активності.", AR: "⚠️ تعذّر تحديث الأنشطة.",
	},
	KeyTrackManageDeleteFailed: {
		RU: "⚠️ Не удалось удалить выбранные активности.", EN: "⚠️ Failed to delete selected activities.", DE: "⚠️ Ausgewählte Aktivitäten konnten nicht gelöscht werden.", UK: "⚠️ Не вдалося видалити вибрані активності.", AR: "⚠️ تعذّر حذف الأنشطة المحددة.",
	},
	KeyTrackManageDeleted: {
		RU: "🗑 Удалено: %d", EN: "🗑 Deleted: %d", DE: "🗑 Gelöscht: %d", UK: "🗑 Видалено: %d", AR: "🗑 تم الحذف: %d",
	},

	KeyTrackArchiveFailed: {
		RU: "⚠️ Не удалось отправить выбранные активности в архив.", EN: "⚠️ Failed to archive selected activities.", DE: "⚠️ Ausgewählte Aktivitäten konnten nicht archiviert werden.", UK: "⚠️ Не вдалося відправити вибрані активності в архів.", AR: "⚠️ تعذّر أرشفة الأنشطة المحددة.",
	},
	KeyTrackArchiveNoneSelected: {
		RU: "Нет выбранных активностей для архивации.", EN: "No selected activities to archive.", DE: "Keine ausgewählten Aktivitäten zum Archivieren.", UK: "Немає вибраних активностей для архівації.", AR: "لا توجد أنشطة محددة للأرشفة.",
	},
	KeyTrackArchived: {
		RU: "📦 В архиве: %d", EN: "📦 Archived: %d", DE: "📦 Archiviert: %d", UK: "📦 В архіві: %d", AR: "📦 تمت الأرشفة: %d",
	},
	KeyTrackArchiveLoadFailed: {
		RU: "⚠️ Не удалось загрузить архив.", EN: "⚠️ Failed to load archive.", DE: "⚠️ Archiv konnte nicht geladen werden.", UK: "⚠️ Не вдалося завантажити архів.", AR: "⚠️ تعذّر تحميل الأرشيف.",
	},
	KeyTrackArchiveEmpty: {
		RU: "Архив пуст.", EN: "Archive is empty.", DE: "Das Archiv ist leer.", UK: "Архів порожній.", AR: "الأرشيف فارغ.",
	},
	KeyTrackArchiveTitle: {
		RU: "🗄 Архив\n\nВсего в архиве: %d", EN: "🗄 Archive\n\nTotal archived: %d", DE: "🗄 Archiv\n\nInsgesamt archiviert: %d", UK: "🗄 Архів\n\nУсього в архіві: %d", AR: "🗄 الأرشيف\n\nإجمالي المؤرشف: %d",
	},
	KeyTrackArchiveInvalidItem: {
		RU: "Некорректная активность.", EN: "Invalid activity.", DE: "Ungültige Aktivität.", UK: "Некоректна активність.", AR: "نشاط غير صالح.",
	},
	KeyTrackArchiveRestoreFailed: {
		RU: "⚠️ Не удалось восстановить активность.", EN: "⚠️ Failed to restore activity.", DE: "⚠️ Aktivität konnte nicht wiederhergestellt werden.", UK: "⚠️ Не вдалося відновити активність.", AR: "⚠️ تعذّر استعادة النشاط.",
	},
	KeyTrackArchiveRestored: {
		RU: "♻ Активность восстановлена: %s", EN: "♻ Activity restored: %s", DE: "♻ Aktivität wiederhergestellt: %s", UK: "♻ Активність відновлено: %s", AR: "♻ تمت استعادة النشاط: %s",
	},
	KeyTrackArchiveDeleteForeverFailed: {
		RU: "⚠️ Не удалось удалить активность навсегда.", EN: "⚠️ Failed to delete activity forever.", DE: "⚠️ Aktivität konnte nicht endgültig gelöscht werden.", UK: "⚠️ Не вдалося видалити активність назавжди.", AR: "⚠️ تعذّر حذف النشاط نهائيًا.",
	},
	KeyTrackArchiveDeletedForever: {
		RU: "🗑 Удалено навсегда: %s", EN: "🗑 Deleted forever: %s", DE: "🗑 Endgültig gelöscht: %s", UK: "🗑 Видалено назавжди: %s", AR: "🗑 تم الحذف نهائيًا: %s",
	},

	KeyTrackMinutesUnit: {
		RU: "мин", EN: "min", DE: "Min.", UK: "хв", AR: "د",
	},
	KeyTrackTimerPickerTitle: {
		RU: "⏱ Выбери интервал отслеживания:", EN: "⏱ Select tracking interval:", DE: "⏱ Wähle das Tracking-Intervall:", UK: "⏱ Обери інтервал відстеження:", AR: "⏱ اختر فاصل التتبع:",
	},
	KeyTrackTimerCustomPrompt: {
		RU: "Введи свой интервал в минутах (1-360):", EN: "Enter custom interval in minutes (1-360):", DE: "Gib ein eigenes Intervall in Minuten ein (1-360):", UK: "Введи свій інтервал у хвилинах (1-360):", AR: "أدخل فاصلاً مخصصًا بالدقائق (1-360):",
	},
	KeyTrackTimerPromptBlocked: {
		RU: "Используй кнопки меню. Введи количество минут обычным текстом.", EN: "Use buttons from menu. Enter minutes as a plain number.", DE: "Benutze die Menü-Buttons. Gib die Minutenzahl als reinen Text ein.", UK: "Використовуй кнопки меню. Введи кількість хвилин звичайним текстом.", AR: "استخدم أزرار القائمة. أدخل عدد الدقائق كنص عادي.",
	},
	KeyTrackTimerNoneToDelete: {
		RU: "Пока нет своих таймеров для удаления.", EN: "No custom timers to delete yet.", DE: "Noch keine eigenen Timer zum Löschen.", UK: "Поки немає власних таймерів для видалення.", AR: "لا توجد مؤقتات مخصصة لحذفها بعد.",
	},
	KeyTrackTimerPickToDelete: {
		RU: "Выбери таймер для удаления:", EN: "Pick a custom timer to delete:", DE: "Wähle einen Timer zum Löschen:", UK: "Обери таймер для видалення:", AR: "اختر مؤقتًا لحذفه:",
	},
	KeyTrackTimerNotANumber: {
		RU: "Введи целое число минут, например 45.", EN: "Enter a whole number of minutes, e.g. 45.", DE: "Gib eine ganze Zahl von Minuten ein, z. B. 45.", UK: "Введи ціле число хвилин, наприклад 45.", AR: "أدخل عددًا صحيحًا من الدقائق، مثل 45.",
	},
	KeyTrackTimerOutOfRange: {
		RU: "Интервал должен быть от %d до %d минут.", EN: "Interval must be between %d and %d minutes.", DE: "Das Intervall muss zwischen %d und %d Minuten liegen.", UK: "Інтервал має бути від %d до %d хвилин.", AR: "يجب أن يكون الفاصل بين %d و%d دقيقة.",
	},
	KeyTrackTimerLimitReached: {
		RU: "Можно хранить до %d своих таймеров. Удали один, прежде чем добавить новый.", EN: "You can keep up to %d custom timers. Delete one before adding a new one.", DE: "Du kannst bis zu %d eigene Timer speichern. Lösche einen, bevor du einen neuen hinzufügst.", UK: "Можна зберігати до %d власних таймерів. Видали один, перш ніж додати новий.", AR: "يمكنك الاحتفاظ بحتى %d مؤقتات مخصصة. احذف واحدًا قبل إضافة مؤقت جديد.",
	},
	KeyTrackTimerSaveFailed: {
		RU: "⚠️ Не удалось сохранить свой таймер.", EN: "⚠️ Failed to save custom timer.", DE: "⚠️ Eigener Timer konnte nicht gespeichert werden.", UK: "⚠️ Не вдалося зберегти власний таймер.", AR: "⚠️ تعذّر حفظ المؤقت المخصص.",
	},
	KeyTrackTimerAdded: {
		RU: "✅ Добавлен свой таймер: %d мин", EN: "✅ Custom timer added: %d min", DE: "✅ Eigener Timer hinzugefügt: %d Min.", UK: "✅ Додано власний таймер: %d хв", AR: "✅ تمت إضافة مؤقت مخصص: %d دقيقة",
	},
	KeyTrackTimerDeleteFailed: {
		RU: "⚠️ Не удалось удалить свой таймер.", EN: "⚠️ Failed to delete custom timer.", DE: "⚠️ Eigener Timer konnte nicht gelöscht werden.", UK: "⚠️ Не вдалося видалити власний таймер.", AR: "⚠️ تعذّر حذف المؤقت المخصص.",
	},
	KeyTrackTimerRemoved: {
		RU: "🗑 Свой таймер удалён: %d мин", EN: "🗑 Custom timer removed: %d min", DE: "🗑 Eigener Timer entfernt: %d Min.", UK: "🗑 Власний таймер видалено: %d хв", AR: "🗑 تمت إزالة المؤقت المخصص: %d دقيقة",
	},
	KeyTrackTimerLoadSelectedFailed: {
		RU: "⚠️ Не удалось загрузить выбранные активности.", EN: "⚠️ Failed to load selected activities.", DE: "⚠️ Ausgewählte Aktivitäten konnten nicht geladen werden.", UK: "⚠️ Не вдалося завантажити вибрані активності.", AR: "⚠️ تعذّر تحميل الأنشطة المحددة.",
	},
	KeyTrackTimerNoneSelected: {
		RU: "Выбери хотя бы одну активность перед запуском таймера.", EN: "Select at least one activity before activating timer.", DE: "Wähle mindestens eine Aktivität, bevor du den Timer aktivierst.", UK: "Обери хоча б одну активність перед запуском таймера.", AR: "اختر نشاطًا واحدًا على الأقل قبل تفعيل المؤقت.",
	},
	KeyTrackTimerActivateFailed: {
		RU: "⚠️ Не удалось запустить таймер.", EN: "⚠️ Failed to activate timer.", DE: "⚠️ Timer konnte nicht aktiviert werden.", UK: "⚠️ Не вдалося запустити таймер.", AR: "⚠️ تعذّر تفعيل المؤقت.",
	},
	KeyTrackTimerActivated: {
		RU: "✅ Таймер запущен: каждые %d мин", EN: "✅ Timer activated: every %d min", DE: "✅ Timer aktiviert: alle %d Min.", UK: "✅ Таймер запущено: кожні %d хв", AR: "✅ تم تفعيل المؤقت: كل %d دقيقة",
	},
	KeyTrackTimerStopFailed: {
		RU: "⚠️ Не удалось остановить таймер.", EN: "⚠️ Failed to stop timer.", DE: "⚠️ Timer konnte nicht gestoppt werden.", UK: "⚠️ Не вдалося зупинити таймер.", AR: "⚠️ تعذّر إيقاف المؤقت.",
	},
	KeyTrackTimerStopped: {
		RU: "⏹ Таймер остановлен", EN: "⏹ Timer stopped", DE: "⏹ Timer gestoppt", UK: "⏹ Таймер зупинено", AR: "⏹ تم إيقاف المؤقت",
	},

	KeyTrackPromptQuestion: {
		RU: "Чем ты сейчас занимаешься?", EN: "What are you doing now?", DE: "Was machst du gerade?", UK: "Чим ти зараз займаєшся?", AR: "ماذا تفعل الآن؟",
	},
	KeyTrackPromptInvalidPayload: {
		RU: "Некорректные данные выбора.", EN: "Invalid selection payload.", DE: "Ungültige Auswahldaten.", UK: "Некоректні дані вибору.", AR: "بيانات اختيار غير صالحة.",
	},
	KeyTrackPromptInvalidInterval: {
		RU: "Некорректный интервал.", EN: "Invalid interval.", DE: "Ungültiges Intervall.", UK: "Некоректний інтервал.", AR: "فاصل غير صالح.",
	},
	KeyTrackPromptSaveFailed: {
		RU: "⚠️ Не удалось сохранить активность.", EN: "⚠️ Failed to save activity.", DE: "⚠️ Aktivität konnte nicht gespeichert werden.", UK: "⚠️ Не вдалося зберегти активність.", AR: "⚠️ تعذّر حفظ النشاط.",
	},
	KeyTrackPromptSaved: {
		RU: "Сохранено ✅\nАктивность: %s\nВремя: %s-%s (%d мин)", EN: "Saved ✅\nActivity: %s\nTime: %s-%s (%d min)", DE: "Gespeichert ✅\nAktivität: %s\nZeit: %s-%s (%d Min.)", UK: "Збережено ✅\nАктивність: %s\nЧас: %s-%s (%d хв)", AR: "تم الحفظ ✅\nالنشاط: %s\nالوقت: %s-%s (%d دقيقة)",
	},
}
