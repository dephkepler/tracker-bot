package i18n

// catalogChallenge holds the Challenges screens: list, day-grid, day-detail
// (status + progress donut + trend + streak), archive, create flow, and the
// evening push. Shared buttons are not repeated here — see the comment on
// the Challenge key block in keys.go for which common/Track keys are reused
// instead.
var catalogChallenge = map[string]entry{
	KeyChallengeButtonCreate: {
		RU: "➕ Новый челлендж", EN: "➕ New challenge", DE: "➕ Neue Challenge", UK: "➕ Новий челендж", AR: "➕ تحدٍّ جديد",
	},
	KeyChallengeButtonArchive: {
		RU: "🔁 Архив", EN: "🔁 Archive", DE: "🔁 Archiv", UK: "🔁 Архів", AR: "🔁 الأرشيف",
	},
	KeyChallengeButtonArchiveThis: {
		RU: "📦 Архивировать этот", EN: "📦 Archive this", DE: "📦 Diese archivieren", UK: "📦 Архівувати цей", AR: "📦 أرشفة هذا",
	},
	KeyChallengeButtonSkip: {
		RU: "❌ Пропущено", EN: "❌ Skipped", DE: "❌ Übersprungen", UK: "❌ Пропущено", AR: "❌ تم التخطي",
	},
	KeyChallengeListTitleFmt: {
		RU: "🎯 *Челленджи* — %d активных", EN: "🎯 *Challenges* — %d active", DE: "🎯 *Challenges* — %d aktiv", UK: "🎯 *Челенджі* — %d активних", AR: "🎯 *التحديات* — %d نشطة",
	},
	KeyChallengeListEmpty: {
		RU: "🎯 *Челленджи*\n\nПока нет ни одного. Начни первый — например, «100 дней чтения».",
		EN: "🎯 *Challenges*\n\nNo challenges yet. Start one — e.g. \"100 days of reading\".",
		DE: "🎯 *Challenges*\n\nNoch keine Challenge. Starte eine — z. B. \"100 Tage lesen\".",
		UK: "🎯 *Челенджі*\n\nПоки немає жодного. Почни перший — наприклад, «100 днів читання».",
		AR: "🎯 *التحديات*\n\nلا توجد تحديات بعد. ابدأ واحدًا — مثل \"100 يوم من القراءة\".",
	},
	KeyChallengeListLoadFailed: {
		RU: "⚠️ Не удалось загрузить челленджи.", EN: "⚠️ Failed to load challenges.", DE: "⚠️ Challenges konnten nicht geladen werden.", UK: "⚠️ Не вдалося завантажити челенджі.", AR: "⚠️ تعذّر تحميل التحديات.",
	},
	KeyChallengeItemLabelFmt: {
		RU: "🎯 %s — %d%% (%d/%d)", EN: "🎯 %s — %d%% (%d/%d)", DE: "🎯 %s — %d%% (%d/%d)", UK: "🎯 %s — %d%% (%d/%d)", AR: "🎯 %s — %d%% (%d/%d)",
	},
	KeyChallengeArchiveItemFmt: {
		RU: "📦 %s — %d/%d готово", EN: "📦 %s — %d/%d done", DE: "📦 %s — %d/%d erledigt", UK: "📦 %s — %d/%d готово", AR: "📦 %s — %d/%d مكتملة",
	},

	KeyChallengeNotFound: {
		RU: "⚠️ Челлендж не найден.", EN: "⚠️ Challenge not found.", DE: "⚠️ Challenge nicht gefunden.", UK: "⚠️ Челендж не знайдено.", AR: "⚠️ التحدي غير موجود.",
	},
	KeyChallengeLoadFailed: {
		RU: "⚠️ Не удалось загрузить челлендж.", EN: "⚠️ Failed to load challenge.", DE: "⚠️ Challenge konnte nicht geladen werden.", UK: "⚠️ Не вдалося завантажити челендж.", AR: "⚠️ تعذّر تحميل التحدي.",
	},
	KeyChallengeDayNotFound: {
		RU: "⚠️ День не найден.", EN: "⚠️ Day not found.", DE: "⚠️ Tag nicht gefunden.", UK: "⚠️ День не знайдено.", AR: "⚠️ اليوم غير موجود.",
	},
	KeyChallengeGridTitleFmt: {
		RU: "🎯 *%s*\n%s → %s (%d дн.)\n\n✅ %d готово · ❌ %d пропущено · 🔲 %d осталось\n📈 %d%% выполнено\n\nТапни по квадрату, чтобы отметить день.",
		EN: "🎯 *%s*\n%s → %s (%d days)\n\n✅ %d done · ❌ %d skipped · 🔲 %d pending\n📈 %d%% complete\n\nTap a square to mark that day.",
		DE: "🎯 *%s*\n%s → %s (%d Tage)\n\n✅ %d erledigt · ❌ %d übersprungen · 🔲 %d offen\n📈 %d%% abgeschlossen\n\nTippe auf ein Quadrat, um den Tag zu markieren.",
		UK: "🎯 *%s*\n%s → %s (%d дн.)\n\n✅ %d готово · ❌ %d пропущено · 🔲 %d залишилось\n📈 %d%% виконано\n\nТапни на квадрат, щоб позначити день.",
		AR: "🎯 *%s*\n%s → %s (%d يومًا)\n\n✅ %d مكتمل · ❌ %d متخطّى · 🔲 %d متبقٍّ\n📈 %d%% مكتمل\n\nاضغط على مربع لتمييز ذلك اليوم.",
	},

	KeyChallengeDayStatusUnmarked: {
		RU: "ещё не отмечен", EN: "not marked yet", DE: "noch nicht markiert", UK: "ще не позначено", AR: "لم يُحدَّد بعد",
	},
	KeyChallengeDayStatusDoneText: {
		RU: "отмечен как ✅ Готово", EN: "currently marked ✅ Done", DE: "aktuell markiert als ✅ Erledigt", UK: "позначено як ✅ Готово", AR: "مُحدَّد حاليًا كـ ✅ مكتمل",
	},
	KeyChallengeDayStatusSkippedText: {
		RU: "отмечен как ❌ Пропущено", EN: "currently marked ❌ Skipped", DE: "aktuell markiert als ❌ Übersprungen", UK: "позначено як ❌ Пропущено", AR: "مُحدَّد حاليًا كـ ❌ تم التخطي",
	},
	KeyChallengeDayHeaderFmt: {
		RU: "🎯 *%s*\n\n%s — %s.", EN: "🎯 *%s*\n\n%s — %s.", DE: "🎯 *%s*\n\n%s — %s.", UK: "🎯 *%s*\n\n%s — %s.", AR: "🎯 *%s*\n\n%s — %s.",
	},
	KeyChallengeDayProportionsFmt: {
		RU: "✅ Готово %d%%   ❌ Пропущено %d%%   🔲 Осталось %d%%",
		EN: "✅ Done %d%%   ❌ Skipped %d%%   🔲 Left %d%%",
		DE: "✅ Erledigt %d%%   ❌ Übersprungen %d%%   🔲 Offen %d%%",
		UK: "✅ Готово %d%%   ❌ Пропущено %d%%   🔲 Залишилось %d%%",
		AR: "✅ مكتمل %d%%   ❌ متخطّى %d%%   🔲 متبقٍّ %d%%",
	},
	KeyChallengeDayTrendLabelFmt: {
		RU: "📈 Тренд (последние %d дн.):", EN: "📈 Trend (last %d days):", DE: "📈 Trend (letzte %d Tage):", UK: "📈 Тренд (останні %d дн.):", AR: "📈 الاتجاه (آخر %d يومًا):",
	},
	KeyChallengeDayStreakFmt: {
		RU: "🔥 Текущий стрик: %d дн.   🏆 Лучший: %d дн.", EN: "🔥 Current streak: %d days   🏆 Best: %d days", DE: "🔥 Aktuelle Serie: %d Tage   🏆 Beste: %d Tage", UK: "🔥 Поточний стрик: %d дн.   🏆 Найкращий: %d дн.", AR: "🔥 السلسلة الحالية: %d يومًا   🏆 الأفضل: %d يومًا",
	},
	KeyChallengeDayMarkPrompt: {
		RU: "Отметь этот день:", EN: "Mark this day:", DE: "Markiere diesen Tag:", UK: "Познач цей день:", AR: "حدِّد هذا اليوم:",
	},

	KeyChallengeArchiveTitleFmt: {
		RU: "🔁 *Архивные челленджи* — %d", EN: "🔁 *Archived challenges* — %d", DE: "🔁 *Archivierte Challenges* — %d", UK: "🔁 *Архівні челенджі* — %d", AR: "🔁 *التحديات المؤرشفة* — %d",
	},
	KeyChallengeArchiveEmpty: {
		RU: "🔁 Архивных челленджей нет.", EN: "🔁 No archived challenges.", DE: "🔁 Keine archivierten Challenges.", UK: "🔁 Архівних челенджів немає.", AR: "🔁 لا توجد تحديات مؤرشفة.",
	},

	KeyChallengeCreateNamePrompt: {
		RU: "✏️ Назови свой челлендж (например, «100 дней чтения»), 2-60 символов, одна строка:",
		EN: "✏️ Name your challenge (e.g. \"100 days of reading\"), 2-60 characters, single line:",
		DE: "✏️ Benenne deine Challenge (z. B. \"100 Tage lesen\"), 2-60 Zeichen, eine Zeile:",
		UK: "✏️ Назви свій челендж (наприклад, «100 днів читання»), 2-60 символів, один рядок:",
		AR: "✏️ سمِّ تحديك (مثل \"100 يوم من القراءة\")، 2-60 حرفًا، سطر واحد:",
	},
	KeyChallengeCreateRangeIntro: {
		RU: "📅 Выбери дату начала, затем дату окончания (максимум 100 дней всего):",
		EN: "📅 Pick the start date, then the end date (max 100 days total):",
		DE: "📅 Wähle das Startdatum, dann das Enddatum (max. 100 Tage insgesamt):",
		UK: "📅 Обери дату початку, потім дату завершення (максимум 100 днів загалом):",
		AR: "📅 اختر تاريخ البدء، ثم تاريخ الانتهاء (100 يوم كحد أقصى إجمالًا):",
	},
	KeyChallengeCreateCalendarHeader: {
		RU: "📅 Выбери начало и конец периода:", EN: "📅 Pick start and end date:", DE: "📅 Wähle Start- und Enddatum:", UK: "📅 Обери початок і кінець періоду:", AR: "📅 اختر تاريخ البدء والانتهاء:",
	},
	KeyChallengeCreateExists: {
		RU: "⚠️ У тебя уже есть челлендж с таким названием.", EN: "⚠️ You already have a challenge with that name.", DE: "⚠️ Du hast bereits eine Challenge mit diesem Namen.", UK: "⚠️ У тебе вже є челендж із такою назвою.", AR: "⚠️ لديك بالفعل تحدٍّ بهذا الاسم.",
	},
	KeyChallengeCreateInvalidRange: {
		RU: "⚠️ Челлендж должен длиться 1-100 дней, дата окончания не раньше даты начала.",
		EN: "⚠️ Challenge must be 1-100 days, end date on or after start date.",
		DE: "⚠️ Die Challenge muss 1-100 Tage dauern, Enddatum nicht vor dem Startdatum.",
		UK: "⚠️ Челендж має тривати 1-100 днів, дата завершення не раніше дати початку.",
		AR: "⚠️ يجب أن يكون التحدي بين 1 و100 يوم، وتاريخ الانتهاء لا يسبق تاريخ البدء.",
	},
	KeyChallengeCreateFailed: {
		RU: "⚠️ Не удалось создать челлендж. Попробуй ещё раз.", EN: "⚠️ Failed to create challenge. Please try again.", DE: "⚠️ Challenge konnte nicht erstellt werden. Bitte versuch es noch einmal.", UK: "⚠️ Не вдалося створити челендж. Спробуй ще раз.", AR: "⚠️ فشل إنشاء التحدي. حاول مرة أخرى.",
	},
	KeyChallengeCreatedFmt: {
		RU: "🎯 Челлендж *%s* создан — %d дн. Первая отметка сегодня в 21:00.",
		EN: "🎯 Challenge *%s* created — %d days. First check-in tonight at 21:00.",
		DE: "🎯 Challenge *%s* erstellt — %d Tage. Erster Check-in heute Abend um 21:00.",
		UK: "🎯 Челендж *%s* створено — %d дн. Перша відмітка сьогодні о 21:00.",
		AR: "🎯 تم إنشاء التحدي *%s* — %d يومًا. أول تسجيل الليلة الساعة 21:00.",
	},

	KeyChallengePushTextFmt: {
		RU: "🎯 *%s* — День %d/%d\n\nСделал сегодня?", EN: "🎯 *%s* — Day %d/%d\n\nDid you do it today?", DE: "🎯 *%s* — Tag %d/%d\n\nHast du es heute gemacht?", UK: "🎯 *%s* — День %d/%d\n\nЗробив сьогодні?", AR: "🎯 *%s* — اليوم %d/%d\n\nهل فعلتها اليوم؟",
	},
	KeyChallengePushMarkedDone: {
		RU: "✅ Отмечено как готово — отличная работа!", EN: "✅ Marked as done — nice work!", DE: "✅ Als erledigt markiert — gut gemacht!", UK: "✅ Позначено як готово — чудова робота!", AR: "✅ تم التحديد كمكتمل — عمل رائع!",
	},
	KeyChallengePushMarkedSkipped: {
		RU: "❌ Отмечено как пропущено.", EN: "❌ Marked as skipped.", DE: "❌ Als übersprungen markiert.", UK: "❌ Позначено як пропущено.", AR: "❌ تم التحديد كمتخطّى.",
	},
}
