package i18n

// catalogRoadmap holds the Roadmap screens: main menu, roadmap list,
// roadmap detail (card checklist), goal/rename prompts, archive, reminder
// scheduling, digest pushes, and progress statistics.
//
// Shared buttons are not repeated here — see the comment on the Roadmap key
// block in keys.go for which common/Learning/Track keys are reused instead.
var catalogRoadmap = map[string]entry{
	KeyRoadmapButtonCreate: {
		RU: "➕ Новая технология", EN: "➕ New technology", DE: "➕ Neue Technologie", UK: "➕ Нова технологія", AR: "➕ تقنية جديدة",
	},
	KeyRoadmapButtonList: {
		RU: "🗺 Мои роадмапы", EN: "🗺 My roadmaps", DE: "🗺 Meine Roadmaps", UK: "🗺 Мої роадмапи", AR: "🗺 خطط التعلّم",
	},
	KeyRoadmapButtonStartReminders: {
		RU: "🔔 Включить напоминания", EN: "🔔 Turn on reminders", DE: "🔔 Erinnerungen aktivieren", UK: "🔔 Увімкнути нагадування", AR: "🔔 تشغيل التذكيرات",
	},
	KeyRoadmapButtonManageReminders: {
		RU: "🔧 Настроить напоминания", EN: "🔧 Manage reminders", DE: "🔧 Erinnerungen verwalten", UK: "🔧 Налаштувати нагадування", AR: "🔧 إدارة التذكيرات",
	},
	KeyRoadmapButtonStopReminders: {
		RU: "⏹ Остановить напоминания", EN: "⏹ Stop reminders", DE: "⏹ Erinnerungen stoppen", UK: "⏹ Зупинити нагадування", AR: "⏹ إيقاف التذكيرات",
	},
	KeyRoadmapButtonArchive: {
		RU: "🔁 Архив роадмап", EN: "🔁 Archive of roadmaps", DE: "🔁 Archiv der Roadmaps", UK: "🔁 Архів роадмап", AR: "🔁 أرشيف الخطط",
	},
	KeyRoadmapButtonProgress: {
		RU: "📈 Прогресс", EN: "📈 Progress", DE: "📈 Fortschritt", UK: "📈 Прогрес", AR: "📈 التقدّم",
	},
	KeyRoadmapButtonAddCards: {
		RU: "➕ Добавить карточки", EN: "➕ Add cards", DE: "➕ Karten hinzufügen", UK: "➕ Додати картки", AR: "➕ إضافة بطاقات",
	},
	KeyRoadmapButtonSetGoal: {
		RU: "🎯 Цель", EN: "🎯 Goal", DE: "🎯 Ziel", UK: "🎯 Ціль", AR: "🎯 الهدف",
	},
	KeyRoadmapButtonArchiveThis: {
		RU: "📦 Архивировать роадмапу", EN: "📦 Archive roadmap", DE: "📦 Roadmap archivieren", UK: "📦 Архівувати роадмапу", AR: "📦 أرشفة الخطة",
	},
	KeyRoadmapButtonSkipGoal: {
		RU: "⏭ Пропустить", EN: "⏭ Skip", DE: "⏭ Überspringen", UK: "⏭ Пропустити", AR: "⏭ تخطٍّ",
	},
	KeyRoadmapButtonToggleOnFmt: {
		RU: "🟢 %s — %d/%d", EN: "🟢 %s — %d/%d", DE: "🟢 %s — %d/%d", UK: "🟢 %s — %d/%d", AR: "🟢 %s — %d/%d",
	},
	KeyRoadmapButtonToggleOffFmt: {
		RU: "⚪ %s — %d/%d", EN: "⚪ %s — %d/%d", DE: "⚪ %s — %d/%d", UK: "⚪ %s — %d/%d", AR: "⚪ %s — %d/%d",
	},
	KeyRoadmapArchiveItemFmt: {
		RU: "📦 %s — %d/%d", EN: "📦 %s — %d/%d", DE: "📦 %s — %d/%d", UK: "📦 %s — %d/%d", AR: "📦 %s — %d/%d",
	},
	KeyRoadmapLabelIncludedInReminder: {
		RU: "🟢 Участвует в напоминаниях", EN: "🟢 Included in reminders", DE: "🟢 In Erinnerungen enthalten", UK: "🟢 Бере участь у нагадуваннях", AR: "🟢 مشمولة في التذكيرات",
	},
	KeyRoadmapLabelExcludedReminder: {
		RU: "⚪ Исключена из напоминаний", EN: "⚪ Excluded from reminders", DE: "⚪ Von Erinnerungen ausgeschlossen", UK: "⚪ Виключена з нагадувань", AR: "⚪ مستبعدة من التذكيرات",
	},

	KeyRoadmapMenuTitle: {
		RU: "🗺 *Роадмапы*", EN: "🗺 *Roadmaps*", DE: "🗺 *Roadmaps*", UK: "🗺 *Роадмапи*", AR: "🗺 *خطط التعلّم*",
	},
	KeyRoadmapMenuTotalRoadmaps: {
		RU: "🧭 Технологий: *%d* из *%d*", EN: "🧭 Technologies: *%d* of *%d*", DE: "🧭 Technologien: *%d* von *%d*", UK: "🧭 Технологій: *%d* з *%d*", AR: "🧭 التقنيات: *%d* من *%d*",
	},
	KeyRoadmapMenuTotalCards: {
		RU: "🗂 Карточек: *%d*", EN: "🗂 Cards: *%d*", DE: "🗂 Karten: *%d*", UK: "🗂 Карток: *%d*", AR: "🗂 البطاقات: *%d*",
	},
	KeyRoadmapMenuDone: {
		RU: "✅ Закрыто: *%d*", EN: "✅ Completed: *%d*", DE: "✅ Abgeschlossen: *%d*", UK: "✅ Закрито: *%d*", AR: "✅ مكتملة: *%d*",
	},
	KeyRoadmapMenuPending: {
		RU: "📌 Осталось: *%d*", EN: "📌 Remaining: *%d*", DE: "📌 Verbleibend: *%d*", UK: "📌 Залишилось: *%d*", AR: "📌 متبقية: *%d*",
	},
	KeyRoadmapMenuRemindersActive: {
		RU: "🕐 Напоминания: каждые *%d* мин (следующее через %s)", EN: "🕐 Reminders: every *%d* min (next in %s)", DE: "🕐 Erinnerungen: alle *%d* Min (nächste in %s)", UK: "🕐 Нагадування: кожні *%d* хв (наступне через %s)", AR: "🕐 التذكيرات: كل *%d* دقيقة (التالي خلال %s)",
	},
	KeyRoadmapMenuRemindersInactive: {
		RU: "🕐 Напоминания: не активны", EN: "🕐 Reminders: not active", DE: "🕐 Erinnerungen: nicht aktiv", UK: "🕐 Нагадування: не активні", AR: "🕐 التذكيرات: غير نشطة",
	},
	KeyRoadmapLoadFailed: {
		RU: "⚠️ Не удалось загрузить роадмапы. Попробуй позже.", EN: "⚠️ Failed to load roadmaps. Please try again later.", DE: "⚠️ Roadmaps konnten nicht geladen werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося завантажити роадмапи. Спробуй пізніше.", AR: "⚠️ تعذّر تحميل خطط التعلّم. حاول لاحقًا.",
	},

	KeyRoadmapListTitle: {
		RU: "🗺 *Мои роадмапы* — %d из %d\n\nТапни на технологию, чтобы открыть её карточки.", EN: "🗺 *My roadmaps* — %d of %d\n\nTap a technology to open its cards.", DE: "🗺 *Meine Roadmaps* — %d von %d\n\nTippe auf eine Technologie, um ihre Karten zu öffnen.", UK: "🗺 *Мої роадмапи* — %d з %d\n\nТапни на технологію, щоб відкрити її картки.", AR: "🗺 *خطط التعلّم* — %d من %d\n\nاضغط على تقنية لعرض بطاقاتها.",
	},
	KeyRoadmapListEmpty: {
		RU: "🗺 Пока нет ни одной роадмапы. Создай первую в меню роадмап.", EN: "🗺 No roadmaps yet. Create your first one from the Roadmap menu.", DE: "🗺 Noch keine Roadmaps. Erstelle deine erste im Roadmap-Menü.", UK: "🗺 Поки немає жодної роадмапи. Створи першу в меню роадмап.", AR: "🗺 لا توجد خطط بعد. أنشئ أول خطة من قائمة خطط التعلّم.",
	},
	KeyRoadmapListLoadFailed: {
		RU: "⚠️ Не удалось загрузить список роадмап.", EN: "⚠️ Failed to load the roadmap list.", DE: "⚠️ Roadmap-Liste konnte nicht geladen werden.", UK: "⚠️ Не вдалося завантажити список роадмап.", AR: "⚠️ تعذّر تحميل قائمة الخطط.",
	},
	KeyRoadmapDetailTitle: {
		RU: "🧭 *%s* — %d/%d закрыто", EN: "🧭 *%s* — %d/%d done", DE: "🧭 *%s* — %d/%d erledigt", UK: "🧭 *%s* — %d/%d закрито", AR: "🧭 *%s* — %d/%d مكتملة",
	},
	KeyRoadmapDetailGoal: {
		RU: "🎯 Цель: %s", EN: "🎯 Goal: %s", DE: "🎯 Ziel: %s", UK: "🎯 Ціль: %s", AR: "🎯 الهدف: %s",
	},
	KeyRoadmapDetailNoGoal: {
		RU: "🎯 Цель пока не указана.", EN: "🎯 No goal set yet.", DE: "🎯 Noch kein Ziel festgelegt.", UK: "🎯 Ціль поки не вказана.", AR: "🎯 لم يُحدَّد هدف بعد.",
	},
	KeyRoadmapDetailHint: {
		RU: "Тапни на карточку, чтобы отметить её закрытой или снова открыть.", EN: "Tap a card to mark it done, or tap again to reopen it.", DE: "Tippe auf eine Karte, um sie als erledigt zu markieren — oder erneut, um sie wieder zu öffnen.", UK: "Тапни на картку, щоб позначити її закритою або знову відкрити.", AR: "اضغط على بطاقة لتمييزها كمكتملة، أو اضغط مرة أخرى لإعادة فتحها.",
	},
	KeyRoadmapDetailNoCards: {
		RU: "Пока нет карточек — добавь их кнопкой ниже.", EN: "No cards yet — add some with the button below.", DE: "Noch keine Karten — füge welche über die Schaltfläche unten hinzu.", UK: "Поки немає карток — додай їх кнопкою нижче.", AR: "لا توجد بطاقات بعد — أضِف بعضها بالزر أدناه.",
	},
	KeyRoadmapNotFound: {
		RU: "⚠️ Роадмапа не найдена.", EN: "⚠️ Roadmap not found.", DE: "⚠️ Roadmap nicht gefunden.", UK: "⚠️ Роадмапу не знайдено.", AR: "⚠️ الخطة غير موجودة.",
	},
	KeyRoadmapCardsLoadFailed: {
		RU: "⚠️ Не удалось загрузить карточки.", EN: "⚠️ Failed to load cards.", DE: "⚠️ Karten konnten nicht geladen werden.", UK: "⚠️ Не вдалося завантажити картки.", AR: "⚠️ تعذّر تحميل البطاقات.",
	},

	KeyRoadmapCreatePrompt: {
		RU: "🧭 Отправь название технологии, которую хочешь освоить (например, «Kafka»). Цель и карточки добавим на следующих шагах.",
		EN: "🧭 Send the name of the technology you want to learn (e.g. \"Kafka\"). You'll add the goal and the cards on the next steps.",
		DE: "🧭 Sende den Namen der Technologie, die du lernen willst (z. B. \"Kafka\"). Ziel und Karten kommen in den nächsten Schritten.",
		UK: "🧭 Надішли назву технології, яку хочеш опанувати (наприклад, «Kafka»). Ціль і картки додамо на наступних кроках.",
		AR: "🧭 أرسل اسم التقنية التي تريد تعلّمها (مثل \"Kafka\"). ستضيف الهدف والبطاقات في الخطوات التالية.",
	},
	KeyRoadmapCreateExists: {
		RU: "⚠️ Роадмапа с таким названием уже есть. Выбери другое:", EN: "⚠️ A roadmap with that name already exists. Pick another one:", DE: "⚠️ Eine Roadmap mit diesem Namen existiert bereits. Wähle einen anderen:", UK: "⚠️ Роадмапа з такою назвою вже є. Обери іншу:", AR: "⚠️ توجد خطة بهذا الاسم بالفعل. اختر اسمًا آخر:",
	},
	KeyRoadmapCreateFailed: {
		RU: "⚠️ Не удалось создать роадмапу. Попробуй позже.", EN: "⚠️ Failed to create the roadmap. Please try again later.", DE: "⚠️ Roadmap konnte nicht erstellt werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося створити роадмапу. Спробуй пізніше.", AR: "⚠️ فشل إنشاء الخطة. حاول لاحقًا.",
	},
	KeyRoadmapCreateConfirmed: {
		RU: "🧭 Роадмапа *%s* создана!", EN: "🧭 Roadmap *%s* created!", DE: "🧭 Roadmap *%s* erstellt!", UK: "🧭 Роадмапу *%s* створено!", AR: "🧭 تم إنشاء الخطة *%s*!",
	},
	KeyRoadmapLimitReached: {
		RU: "⚠️ Одновременно можно вести не больше %d технологий. Заархивируй одну, чтобы освободить место.", EN: "⚠️ You can track at most %d technologies at once. Archive one to free a slot.", DE: "⚠️ Du kannst höchstens %d Technologien gleichzeitig verfolgen. Archiviere eine, um Platz zu schaffen.", UK: "⚠️ Одночасно можна вести не більше %d технологій. Заархівуй одну, щоб звільнити місце.", AR: "⚠️ يمكنك متابعة %d تقنيات كحد أقصى في الوقت نفسه. أرشِف واحدة لتحرير مكان.",
	},

	KeyRoadmapGoalPrompt: {
		RU: "🎯 Что для тебя значит «я знаю *%s*»? Опиши цель одним сообщением (до %d символов) — или нажми «Пропустить».",
		EN: "🎯 What does \"I know *%s*\" mean for you? Describe the goal in one message (up to %d characters) — or tap Skip.",
		DE: "🎯 Was bedeutet \"Ich kann *%s*\" für dich? Beschreibe das Ziel in einer Nachricht (bis zu %d Zeichen) — oder tippe auf Überspringen.",
		UK: "🎯 Що для тебе означає «я знаю *%s*»? Опиши ціль одним повідомленням (до %d символів) — або натисни «Пропустити».",
		AR: "🎯 ماذا يعني لك \"أنا أعرف *%s*\"؟ اكتب الهدف في رسالة واحدة (حتى %d حرفًا) — أو اضغط تخطٍّ.",
	},
	KeyRoadmapGoalTooLong: {
		RU: "⚠️ Слишком длинная цель — максимум %d символов. Попробуй короче:", EN: "⚠️ That goal is too long — %d characters max. Try a shorter one:", DE: "⚠️ Das Ziel ist zu lang — maximal %d Zeichen. Versuch es kürzer:", UK: "⚠️ Занадто довга ціль — максимум %d символів. Спробуй коротше:", AR: "⚠️ الهدف طويل جدًا — %d حرفًا كحد أقصى. اكتب هدفًا أقصر:",
	},
	KeyRoadmapGoalSaved: {
		RU: "🎯 Цель сохранена.", EN: "🎯 Goal saved.", DE: "🎯 Ziel gespeichert.", UK: "🎯 Ціль збережено.", AR: "🎯 تم حفظ الهدف.",
	},
	KeyRoadmapGoalSkipped: {
		RU: "⏭ Без цели — её можно указать позже.", EN: "⏭ No goal for now — you can set it later.", DE: "⏭ Vorerst kein Ziel — du kannst es später festlegen.", UK: "⏭ Без цілі — її можна вказати пізніше.", AR: "⏭ بلا هدف حاليًا — يمكنك تحديده لاحقًا.",
	},
	KeyRoadmapGoalFailed: {
		RU: "⚠️ Не удалось сохранить цель. Попробуй позже.", EN: "⚠️ Failed to save the goal. Please try again later.", DE: "⚠️ Ziel konnte nicht gespeichert werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося зберегти ціль. Спробуй пізніше.", AR: "⚠️ تعذّر حفظ الهدف. حاول لاحقًا.",
	},

	KeyRoadmapRenamePrompt: {
		RU: "✏️ Отправь новое название для *%s* (2-60 символов, одна строка):", EN: "✏️ Send a new name for the *%s* roadmap (2-60 characters, single line):", DE: "✏️ Sende einen neuen Namen für die Roadmap *%s* (2-60 Zeichen, eine Zeile):", UK: "✏️ Надішли нову назву для *%s* (2-60 символів, один рядок):", AR: "✏️ أرسل اسمًا جديدًا لخطة *%s* (2-60 حرفًا، سطر واحد):",
	},
	KeyRoadmapRenamed: {
		RU: "✅ Теперь это *%s*.", EN: "✅ Now called *%s*.", DE: "✅ Jetzt *%s*.", UK: "✅ Тепер це *%s*.", AR: "✅ الاسم الآن *%s*.",
	},
	KeyRoadmapRenameFailed: {
		RU: "⚠️ Не удалось переименовать роадмапу.", EN: "⚠️ Failed to rename the roadmap.", DE: "⚠️ Roadmap konnte nicht umbenannt werden.", UK: "⚠️ Не вдалося перейменувати роадмапу.", AR: "⚠️ تعذّرت إعادة تسمية الخطة.",
	},

	KeyRoadmapAddCardsPromptFirst: {
		RU: "🗂 Теперь пришли, что нужно изучить — по одному пункту на строку (тема, задача, статья). Можно вставить сразу список. Нажми ✅ Готово, когда закончишь.",
		EN: "🗂 Now send what you need to learn — one item per line (a topic, a task, an article). You can paste a whole list at once. Tap ✅ Done when finished.",
		DE: "🗂 Sende jetzt, was du lernen willst — ein Punkt pro Zeile (Thema, Aufgabe, Artikel). Du kannst eine ganze Liste einfügen. Tippe auf ✅ Fertig, wenn du fertig bist.",
		UK: "🗂 Тепер надішли, що потрібно вивчити — по одному пункту на рядок (тема, задача, стаття). Можна вставити одразу список. Натисни ✅ Готово, коли закінчиш.",
		AR: "🗂 أرسل الآن ما تريد تعلّمه — عنصر واحد في كل سطر (موضوع، مهمة، مقالة). يمكنك لصق قائمة كاملة. اضغط ✅ تم عند الانتهاء.",
	},
	KeyRoadmapAddCardsPromptMore: {
		RU: "🗂 Пришли новые пункты — по одному на строку. Нажми ✅ Готово, когда закончишь.",
		EN: "🗂 Send more items — one per line. Tap ✅ Done when finished.",
		DE: "🗂 Sende weitere Punkte — einen pro Zeile. Tippe auf ✅ Fertig, wenn du fertig bist.",
		UK: "🗂 Надішли нові пункти — по одному на рядок. Натисни ✅ Готово, коли закінчиш.",
		AR: "🗂 أرسل عناصر إضافية — عنصرًا واحدًا في كل سطر. اضغط ✅ تم عند الانتهاء.",
	},
	KeyRoadmapAddCardsNoneParsed: {
		RU: "⚠️ В сообщении нет ни одной непустой строки. Напиши хотя бы один пункт:", EN: "⚠️ There isn't a single non-empty line in that message. Write at least one item:", DE: "⚠️ In dieser Nachricht ist keine einzige nicht-leere Zeile. Schreib mindestens einen Punkt:", UK: "⚠️ У повідомленні немає жодного непорожнього рядка. Напиши хоча б один пункт:", AR: "⚠️ لا يوجد أي سطر غير فارغ في تلك الرسالة. اكتب عنصرًا واحدًا على الأقل:",
	},
	KeyRoadmapAddCardsFailed: {
		RU: "⚠️ Не удалось добавить карточки. Попробуй позже.", EN: "⚠️ Failed to add the cards. Please try again later.", DE: "⚠️ Karten konnten nicht hinzugefügt werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося додати картки. Спробуй пізніше.", AR: "⚠️ فشلت إضافة البطاقات. حاول لاحقًا.",
	},
	KeyRoadmapAddCardsAdded: {
		RU: "✅ Добавлено карточек: %d.", EN: "✅ Added %d card(s).", DE: "✅ %d Karte(n) hinzugefügt.", UK: "✅ Додано карток: %d.", AR: "✅ تمت إضافة %d بطاقة.",
	},
	KeyRoadmapAddCardsSkipped: {
		RU: " (пропущено строк: %d — слишком длинные.)", EN: " (%d line(s) skipped — too long.)", DE: " (%d Zeile(n) übersprungen — zu lang.)", UK: " (пропущено рядків: %d — занадто довгі.)", AR: " (تم تخطي %d سطر — طويلة جدًا.)",
	},
	KeyRoadmapAddCardsDoneNotice: {
		RU: "🗂 Карточки сохранены.", EN: "🗂 Cards saved.", DE: "🗂 Karten gespeichert.", UK: "🗂 Картки збережено.", AR: "🗂 تم حفظ البطاقات.",
	},

	KeyRoadmapArchiveTitle: {
		RU: "🔁 *Архивные роадмапы* — %d", EN: "🔁 *Archived roadmaps* — %d", DE: "🔁 *Archivierte Roadmaps* — %d", UK: "🔁 *Архівні роадмапи* — %d", AR: "🔁 *الخطط المؤرشفة* — %d",
	},
	KeyRoadmapArchiveEmpty: {
		RU: "🔁 Архивных роадмап нет.", EN: "🔁 No archived roadmaps.", DE: "🔁 Keine archivierten Roadmaps.", UK: "🔁 Архівних роадмап немає.", AR: "🔁 لا توجد خطط مؤرشفة.",
	},

	KeyRoadmapPushIntervalPrompt: {
		RU: "⏰ Как часто присылать напоминание о том, что осталось изучить?", EN: "⏰ How often should the bot remind you what's still left to learn?", DE: "⏰ Wie oft soll dich der Bot daran erinnern, was noch zu lernen ist?", UK: "⏰ Як часто надсилати нагадування про те, що залишилося вивчити?", AR: "⏰ كم مرة يجب أن يذكّرك البوت بما تبقّى لتتعلّمه؟",
	},
	KeyRoadmapPushActivateFailed: {
		RU: "⚠️ Не удалось включить напоминания. Попробуй позже.", EN: "⚠️ Failed to turn on reminders. Please try again later.", DE: "⚠️ Erinnerungen konnten nicht aktiviert werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося увімкнути нагадування. Спробуй пізніше.", AR: "⚠️ تعذّر تشغيل التذكيرات. حاول لاحقًا.",
	},
	KeyRoadmapPushActivated: {
		RU: "🔔 Напоминания включены — дайджест каждые %d мин.", EN: "🔔 Reminders on — a digest every %d min.", DE: "🔔 Erinnerungen aktiv — eine Übersicht alle %d Min.", UK: "🔔 Нагадування увімкнено — дайджест кожні %d хв.", AR: "🔔 التذكيرات مُفعّلة — ملخّص كل %d دقيقة.",
	},
	KeyRoadmapPushNeedOne: {
		RU: "⚠️ Нужна хотя бы одна активная роадмапа с незакрытыми карточками.", EN: "⚠️ You need at least one active roadmap with unfinished cards.", DE: "⚠️ Du brauchst mindestens eine aktive Roadmap mit offenen Karten.", UK: "⚠️ Потрібна хоча б одна активна роадмапа з незакритими картками.", AR: "⚠️ تحتاج إلى خطة نشطة واحدة على الأقل تحتوي بطاقات غير مكتملة.",
	},

	KeyRoadmapDigestTitle: {
		RU: "🗺 *Что осталось изучить*", EN: "🗺 *Still left to learn*", DE: "🗺 *Noch zu lernen*", UK: "🗺 *Що залишилося вивчити*", AR: "🗺 *ما تبقّى لتتعلّمه*",
	},
	KeyRoadmapDigestRoadmapLine: {
		RU: "\n*%s* — %d/%d\n", EN: "\n*%s* — %d/%d\n", DE: "\n*%s* — %d/%d\n", UK: "\n*%s* — %d/%d\n", AR: "\n*%s* — %d/%d\n",
	},
	KeyRoadmapDigestEmpty: {
		RU: "🎉 Все карточки закрыты — отличная работа!", EN: "🎉 Every card is done — great work!", DE: "🎉 Alle Karten sind erledigt — großartig!", UK: "🎉 Усі картки закриті — чудова робота!", AR: "🎉 كل البطاقات مكتملة — عمل رائع!",
	},

	KeyRoadmapStatsTitle: {
		RU: "📈 *Прогресс по технологиям*", EN: "📈 *Progress by technology*", DE: "📈 *Fortschritt nach Technologie*", UK: "📈 *Прогрес за технологіями*", AR: "📈 *التقدّم حسب التقنية*",
	},
	KeyRoadmapStatsRoadmapLine: {
		RU: "• %s — %d/%d (%d%%)\n", EN: "• %s — %d/%d (%d%%)\n", DE: "• %s — %d/%d (%d%%)\n", UK: "• %s — %d/%d (%d%%)\n", AR: "• %s — %d/%d (%d%%)\n",
	},
	KeyRoadmapStatsGoalLine: {
		RU: "   🎯 %s\n", EN: "   🎯 %s\n", DE: "   🎯 %s\n", UK: "   🎯 %s\n", AR: "   🎯 %s\n",
	},
	KeyRoadmapStatsEmpty: {
		RU: "📈 Пока нечего показать — создай первую роадмапу.", EN: "📈 Nothing to show yet — create your first roadmap.", DE: "📈 Noch nichts zu zeigen — erstelle deine erste Roadmap.", UK: "📈 Поки нічого показати — створи першу роадмапу.", AR: "📈 لا شيء لعرضه بعد — أنشئ أول خطة لك.",
	},
	KeyRoadmapStatsLoadFailed: {
		RU: "⚠️ Не удалось загрузить прогресс.", EN: "⚠️ Failed to load progress.", DE: "⚠️ Fortschritt konnte nicht geladen werden.", UK: "⚠️ Не вдалося завантажити прогрес.", AR: "⚠️ تعذّر تحميل التقدّم.",
	},
}
