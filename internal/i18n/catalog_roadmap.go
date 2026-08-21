package i18n

// catalogRoadmap holds the Roadmap screens: goals, the technologies under a
// goal, a technology's card checklist, archive, reminder scheduling, digest
// pushes and progress. See the Roadmap key block in keys.go for which shared
// keys are reused instead of being repeated here.
var catalogRoadmap = map[string]entry{
	KeyRoadmapButtonCreateGoal: {
		RU: "➕ Новая цель", EN: "➕ New goal", DE: "➕ Neues Ziel", UK: "➕ Нова ціль", AR: "➕ هدف جديد",
	},
	KeyRoadmapButtonAddTech: {
		RU: "➕ Добавить технологию", EN: "➕ Add a technology", DE: "➕ Technologie hinzufügen", UK: "➕ Додати технологію", AR: "➕ إضافة تقنية",
	},
	KeyRoadmapButtonGoals: {
		RU: "🎯 Мои цели", EN: "🎯 My goals", DE: "🎯 Meine Ziele", UK: "🎯 Мої цілі", AR: "🎯 أهدافي",
	},
	KeyRoadmapButtonOrphans: {
		RU: "🧭 Без цели", EN: "🧭 Without a goal", DE: "🧭 Ohne Ziel", UK: "🧭 Без цілі", AR: "🧭 بدون هدف",
	},
	KeyRoadmapButtonAssignGoal: {
		RU: "🎯 Привязать к цели", EN: "🎯 Attach to a goal", DE: "🎯 Einem Ziel zuordnen", UK: "🎯 Прив'язати до цілі", AR: "🎯 إلحاق بهدف",
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
		RU: "🔁 Архив роадмап", EN: "🔁 Roadmap archive", DE: "🔁 Roadmap-Archiv", UK: "🔁 Архів роадмап", AR: "🔁 أرشيف الخطط",
	},
	KeyRoadmapButtonProgress: {
		RU: "📈 Прогресс", EN: "📈 Progress", DE: "📈 Fortschritt", UK: "📈 Прогрес", AR: "📈 التقدّم",
	},
	KeyRoadmapButtonAddCards: {
		RU: "➕ Добавить карточки", EN: "➕ Add cards", DE: "➕ Karten hinzufügen", UK: "➕ Додати картки", AR: "➕ إضافة بطاقات",
	},
	KeyRoadmapButtonSetCriteria: {
		RU: "📌 Критерий", EN: "📌 Criteria", DE: "📌 Kriterium", UK: "📌 Критерій", AR: "📌 المعيار",
	},
	KeyRoadmapButtonArchiveGoal: {
		RU: "📦 Архивировать цель", EN: "📦 Archive goal", DE: "📦 Ziel archivieren", UK: "📦 Архівувати ціль", AR: "📦 أرشفة الهدف",
	},
	KeyRoadmapButtonArchiveThis: {
		RU: "📦 Архивировать технологию", EN: "📦 Archive technology", DE: "📦 Technologie archivieren", UK: "📦 Архівувати технологію", AR: "📦 أرشفة التقنية",
	},
	KeyRoadmapButtonSkipCriteria: {
		RU: "⏭ Пропустить", EN: "⏭ Skip", DE: "⏭ Überspringen", UK: "⏭ Пропустити", AR: "⏭ تخطٍّ",
	},

	KeyRoadmapGoalItemFmt: {
		RU: "🎯 %s — %d/%d (%d%%)", EN: "🎯 %s — %d/%d (%d%%)", DE: "🎯 %s — %d/%d (%d%%)", UK: "🎯 %s — %d/%d (%d%%)", AR: "🎯 %s — %d/%d (%d%%)",
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
	KeyRoadmapArchiveGoalItemFmt: {
		RU: "📦 🎯 %s — технологий: %d", EN: "📦 🎯 %s — %d technology(ies)", DE: "📦 🎯 %s — %d Technologie(n)", UK: "📦 🎯 %s — технологій: %d", AR: "📦 🎯 %s — %d تقنية",
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
	KeyRoadmapMenuTotalGoals: {
		RU: "🎯 Целей: *%d* из *%d*", EN: "🎯 Goals: *%d* of *%d*", DE: "🎯 Ziele: *%d* von *%d*", UK: "🎯 Цілей: *%d* з *%d*", AR: "🎯 الأهداف: *%d* من *%d*",
	},
	KeyRoadmapMenuTotalRoadmaps: {
		RU: "🧭 Технологий: *%d*", EN: "🧭 Technologies: *%d*", DE: "🧭 Technologien: *%d*", UK: "🧭 Технологій: *%d*", AR: "🧭 التقنيات: *%d*",
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

	KeyRoadmapGoalsTitle: {
		RU: "🎯 *Мои цели* — %d из %d\n\nТапни на цель, чтобы открыть её технологии.", EN: "🎯 *My goals* — %d of %d\n\nTap a goal to open its technologies.", DE: "🎯 *Meine Ziele* — %d von %d\n\nTippe auf ein Ziel, um seine Technologien zu öffnen.", UK: "🎯 *Мої цілі* — %d з %d\n\nТапни на ціль, щоб відкрити її технології.", AR: "🎯 *أهدافي* — %d من %d\n\nاضغط على هدف لعرض تقنياته.",
	},
	KeyRoadmapGoalsEmpty: {
		RU: "🎯 Пока нет ни одной цели. Начни с большой — например, «выйти на уровень мидла».", EN: "🎯 No goals yet. Start with a big one — say, \"reach mid-level\".", DE: "🎯 Noch keine Ziele. Fang mit einem großen an — etwa \"Mid-Level erreichen\".", UK: "🎯 Поки немає жодної цілі. Почни з великої — наприклад, «вийти на рівень мідла».", AR: "🎯 لا توجد أهداف بعد. ابدأ بهدف كبير — مثل \"الوصول إلى المستوى المتوسط\".",
	},
	KeyRoadmapGoalsLoadFailed: {
		RU: "⚠️ Не удалось загрузить цели.", EN: "⚠️ Failed to load goals.", DE: "⚠️ Ziele konnten nicht geladen werden.", UK: "⚠️ Не вдалося завантажити цілі.", AR: "⚠️ تعذّر تحميل الأهداف.",
	},

	KeyRoadmapGoalDetailTitle: {
		RU: "🎯 *%s*", EN: "🎯 *%s*", DE: "🎯 *%s*", UK: "🎯 *%s*", AR: "🎯 *%s*",
	},
	KeyRoadmapGoalDetailCounts: {
		RU: "🧭 Технологий: %d из %d · 🗂 %d/%d (%d%%)", EN: "🧭 Technologies: %d of %d · 🗂 %d/%d (%d%%)", DE: "🧭 Technologien: %d von %d · 🗂 %d/%d (%d%%)", UK: "🧭 Технологій: %d з %d · 🗂 %d/%d (%d%%)", AR: "🧭 التقنيات: %d من %d · 🗂 %d/%d (%d%%)",
	},
	KeyRoadmapGoalDetailNoTech: {
		RU: "Пока нет технологий — добавь первую кнопкой ниже.", EN: "No technologies yet — add the first one with the button below.", DE: "Noch keine Technologien — füge die erste über die Schaltfläche unten hinzu.", UK: "Поки немає технологій — додай першу кнопкою нижче.", AR: "لا توجد تقنيات بعد — أضِف الأولى بالزر أدناه.",
	},
	KeyRoadmapGoalDetailHint: {
		RU: "Тапни на технологию, чтобы открыть её карточки.", EN: "Tap a technology to open its cards.", DE: "Tippe auf eine Technologie, um ihre Karten zu öffnen.", UK: "Тапни на технологію, щоб відкрити її картки.", AR: "اضغط على تقنية لعرض بطاقاتها.",
	},
	KeyRoadmapGoalCreatePrompt: {
		RU: "🎯 Куда ты идёшь? Опиши цель одной строкой — например, «выйти на уровень мидла». Технологии добавим внутрь неё.",
		EN: "🎯 Where are you heading? Describe the goal in one line — say, \"reach mid-level\". Technologies go inside it next.",
		DE: "🎯 Wohin willst du? Beschreibe das Ziel in einer Zeile — etwa \"Mid-Level erreichen\". Die Technologien kommen dann hinein.",
		UK: "🎯 Куди ти йдеш? Опиши ціль одним рядком — наприклад, «вийти на рівень мідла». Технології додамо всередину.",
		AR: "🎯 إلى أين تتجه؟ اكتب الهدف في سطر واحد — مثل \"الوصول إلى المستوى المتوسط\". ثم نضيف التقنيات داخله.",
	},
	KeyRoadmapGoalExists: {
		RU: "⚠️ Цель с таким названием уже есть. Выбери другое:", EN: "⚠️ A goal with that name already exists. Pick another one:", DE: "⚠️ Ein Ziel mit diesem Namen existiert bereits. Wähle einen anderen:", UK: "⚠️ Ціль з такою назвою вже є. Обери іншу:", AR: "⚠️ يوجد هدف بهذا الاسم بالفعل. اختر اسمًا آخر:",
	},
	KeyRoadmapGoalCreateFailed: {
		RU: "⚠️ Не удалось создать цель. Попробуй позже.", EN: "⚠️ Failed to create the goal. Please try again later.", DE: "⚠️ Ziel konnte nicht erstellt werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося створити ціль. Спробуй пізніше.", AR: "⚠️ فشل إنشاء الهدف. حاول لاحقًا.",
	},
	KeyRoadmapGoalCreated: {
		RU: "🎯 Цель *%s* создана! Теперь добавь технологии, из которых она складывается.", EN: "🎯 Goal *%s* created! Now add the technologies it's made of.", DE: "🎯 Ziel *%s* erstellt! Füge jetzt die Technologien hinzu, aus denen es besteht.", UK: "🎯 Ціль *%s* створено! Тепер додай технології, з яких вона складається.", AR: "🎯 تم إنشاء الهدف *%s*! أضِف الآن التقنيات التي يتكوّن منها.",
	},
	KeyRoadmapGoalLimitReached: {
		RU: "⚠️ Одновременно можно вести не больше %d целей. Заархивируй одну, чтобы освободить место.", EN: "⚠️ You can chase at most %d goals at once. Archive one to free a slot.", DE: "⚠️ Du kannst höchstens %d Ziele gleichzeitig verfolgen. Archiviere eins, um Platz zu schaffen.", UK: "⚠️ Одночасно можна вести не більше %d цілей. Заархівуй одну, щоб звільнити місце.", AR: "⚠️ يمكنك متابعة %d أهداف كحد أقصى في الوقت نفسه. أرشِف واحدًا لتحرير مكان.",
	},
	KeyRoadmapGoalNotFound: {
		RU: "⚠️ Цель не найдена.", EN: "⚠️ Goal not found.", DE: "⚠️ Ziel nicht gefunden.", UK: "⚠️ Ціль не знайдено.", AR: "⚠️ الهدف غير موجود.",
	},
	KeyRoadmapGoalRenamePrompt: {
		RU: "✏️ Отправь новое название для цели *%s* (2-60 символов, одна строка):", EN: "✏️ Send a new name for the *%s* goal (2-60 characters, single line):", DE: "✏️ Sende einen neuen Namen für das Ziel *%s* (2-60 Zeichen, eine Zeile):", UK: "✏️ Надішли нову назву для цілі *%s* (2-60 символів, один рядок):", AR: "✏️ أرسل اسمًا جديدًا للهدف *%s* (2-60 حرفًا، سطر واحد):",
	},
	KeyRoadmapGoalRenamed: {
		RU: "✅ Цель теперь называется *%s*.", EN: "✅ The goal is now called *%s*.", DE: "✅ Das Ziel heißt jetzt *%s*.", UK: "✅ Ціль тепер називається *%s*.", AR: "✅ الهدف الآن باسم *%s*.",
	},
	KeyRoadmapGoalRenameFailed: {
		RU: "⚠️ Не удалось переименовать цель.", EN: "⚠️ Failed to rename the goal.", DE: "⚠️ Ziel konnte nicht umbenannt werden.", UK: "⚠️ Не вдалося перейменувати ціль.", AR: "⚠️ تعذّرت إعادة تسمية الهدف.",
	},

	KeyRoadmapListTitle: {
		RU: "🧭 *Технологии* — %d из %d", EN: "🧭 *Technologies* — %d of %d", DE: "🧭 *Technologien* — %d von %d", UK: "🧭 *Технології* — %d з %d", AR: "🧭 *التقنيات* — %d من %d",
	},
	KeyRoadmapListEmpty: {
		RU: "🧭 В этой цели пока нет технологий.", EN: "🧭 This goal has no technologies yet.", DE: "🧭 Dieses Ziel hat noch keine Technologien.", UK: "🧭 У цій цілі поки немає технологій.", AR: "🧭 لا تحتوي هذه الخطة على تقنيات بعد.",
	},
	KeyRoadmapOrphansTitle: {
		RU: "🧭 *Без цели* — %d\n\nЭти технологии ни к чему не привязаны. Открой любую и привяжи к цели.", EN: "🧭 *Without a goal* — %d\n\nThese technologies aren't attached to anything. Open one and attach it to a goal.", DE: "🧭 *Ohne Ziel* — %d\n\nDiese Technologien hängen an keinem Ziel. Öffne eine und ordne sie zu.", UK: "🧭 *Без цілі* — %d\n\nЦі технології ні до чого не прив'язані. Відкрий будь-яку і прив'яжи до цілі.", AR: "🧭 *بدون هدف* — %d\n\nهذه التقنيات غير مرتبطة بأي هدف. افتح إحداها وألحقها بهدف.",
	},
	KeyRoadmapOrphansEmpty: {
		RU: "🧭 Все технологии привязаны к целям.", EN: "🧭 Every technology sits under a goal.", DE: "🧭 Jede Technologie gehört zu einem Ziel.", UK: "🧭 Усі технології прив'язані до цілей.", AR: "🧭 كل تقنية مرتبطة بهدف.",
	},
	KeyRoadmapDetailTitle: {
		RU: "🧭 *%s* — %d/%d закрыто", EN: "🧭 *%s* — %d/%d done", DE: "🧭 *%s* — %d/%d erledigt", UK: "🧭 *%s* — %d/%d закрито", AR: "🧭 *%s* — %d/%d مكتملة",
	},
	KeyRoadmapDetailCriteria: {
		RU: "📌 Критерий: %s", EN: "📌 Criteria: %s", DE: "📌 Kriterium: %s", UK: "📌 Критерій: %s", AR: "📌 المعيار: %s",
	},
	KeyRoadmapDetailNoCriteria: {
		RU: "📌 Критерий освоения не указан.", EN: "📌 No mastery criteria set.", DE: "📌 Kein Beherrschungskriterium festgelegt.", UK: "📌 Критерій опанування не вказано.", AR: "📌 لم يُحدَّد معيار الإتقان.",
	},
	KeyRoadmapDetailHint: {
		RU: "Тапни карточку — отметить. 🎚 — сложность (🟢→🟡→🔴).", EN: "Tap a card to tick it. 🎚 cycles its difficulty (🟢→🟡→🔴).", DE: "Tippe eine Karte an, um sie abzuhaken. 🎚 wechselt die Schwierigkeit (🟢→🟡→🔴).", UK: "Тапни картку — позначити. 🎚 — складність (🟢→🟡→🔴).", AR: "اضغط بطاقة لتمييزها. 🎚 يغيّر صعوبتها (🟢→🟡→🔴).",
	},
	KeyRoadmapDetailNoCards: {
		RU: "Пока нет карточек — добавь их кнопкой ниже.", EN: "No cards yet — add some with the button below.", DE: "Noch keine Karten — füge welche über die Schaltfläche unten hinzu.", UK: "Поки немає карток — додай їх кнопкою нижче.", AR: "لا توجد بطاقات بعد — أضِف بعضها بالزر أدناه.",
	},
	KeyRoadmapNotFound: {
		RU: "⚠️ Технология не найдена.", EN: "⚠️ Technology not found.", DE: "⚠️ Technologie nicht gefunden.", UK: "⚠️ Технологію не знайдено.", AR: "⚠️ التقنية غير موجودة.",
	},
	KeyRoadmapCardsLoadFailed: {
		RU: "⚠️ Не удалось загрузить карточки.", EN: "⚠️ Failed to load cards.", DE: "⚠️ Karten konnten nicht geladen werden.", UK: "⚠️ Не вдалося завантажити картки.", AR: "⚠️ تعذّر تحميل البطاقات.",
	},

	KeyRoadmapCreatePrompt: {
		RU: "🧭 Какую технологию добавить в эту цель? Например, «Kafka». Критерий и карточки — на следующих шагах.",
		EN: "🧭 Which technology goes into this goal? \"Kafka\", for example. Criteria and cards come next.",
		DE: "🧭 Welche Technologie kommt in dieses Ziel? Zum Beispiel \"Kafka\". Kriterium und Karten folgen.",
		UK: "🧭 Яку технологію додати в цю ціль? Наприклад, «Kafka». Критерій і картки — на наступних кроках.",
		AR: "🧭 أي تقنية تضاف إلى هذا الهدف؟ \"Kafka\" مثلًا. المعيار والبطاقات في الخطوات التالية.",
	},
	KeyRoadmapCreateExists: {
		RU: "⚠️ Технология с таким названием уже есть. Выбери другое:", EN: "⚠️ A technology with that name already exists. Pick another one:", DE: "⚠️ Eine Technologie mit diesem Namen existiert bereits. Wähle einen anderen:", UK: "⚠️ Технологія з такою назвою вже є. Обери іншу:", AR: "⚠️ توجد تقنية بهذا الاسم بالفعل. اختر اسمًا آخر:",
	},
	KeyRoadmapCreateFailed: {
		RU: "⚠️ Не удалось добавить технологию. Попробуй позже.", EN: "⚠️ Failed to add the technology. Please try again later.", DE: "⚠️ Technologie konnte nicht hinzugefügt werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося додати технологію. Спробуй пізніше.", AR: "⚠️ فشلت إضافة التقنية. حاول لاحقًا.",
	},
	KeyRoadmapCreateConfirmed: {
		RU: "🧭 Технология *%s* добавлена!", EN: "🧭 Technology *%s* added!", DE: "🧭 Technologie *%s* hinzugefügt!", UK: "🧭 Технологію *%s* додано!", AR: "🧭 تمت إضافة التقنية *%s*!",
	},
	KeyRoadmapLimitReached: {
		RU: "⚠️ В одной цели не больше %d технологий. Заархивируй одну или создай отдельную цель.", EN: "⚠️ At most %d technologies per goal. Archive one, or make a separate goal.", DE: "⚠️ Höchstens %d Technologien pro Ziel. Archiviere eine oder erstelle ein eigenes Ziel.", UK: "⚠️ В одній цілі не більше %d технологій. Заархівуй одну або створи окрему ціль.", AR: "⚠️ %d تقنيات كحد أقصى لكل هدف. أرشِف واحدة أو أنشئ هدفًا منفصلًا.",
	},

	KeyRoadmapAssignPrompt: {
		RU: "🎯 К какой цели привязать *%s*?", EN: "🎯 Which goal should *%s* go under?", DE: "🎯 Zu welchem Ziel soll *%s* gehören?", UK: "🎯 До якої цілі прив'язати *%s*?", AR: "🎯 إلى أي هدف تُلحَق *%s*؟",
	},
	KeyRoadmapAssignNoGoals: {
		RU: "⚠️ Сначала создай хотя бы одну цель.", EN: "⚠️ Create at least one goal first.", DE: "⚠️ Erstelle zuerst mindestens ein Ziel.", UK: "⚠️ Спочатку створи хоча б одну ціль.", AR: "⚠️ أنشئ هدفًا واحدًا على الأقل أولًا.",
	},
	KeyRoadmapAssigned: {
		RU: "✅ *%s* теперь в цели *%s*.", EN: "✅ *%s* now sits under *%s*.", DE: "✅ *%s* gehört jetzt zu *%s*.", UK: "✅ *%s* тепер у цілі *%s*.", AR: "✅ *%s* الآن تحت *%s*.",
	},
	KeyRoadmapAssignFailed: {
		RU: "⚠️ Не удалось привязать технологию к цели.", EN: "⚠️ Failed to attach the technology to that goal.", DE: "⚠️ Technologie konnte dem Ziel nicht zugeordnet werden.", UK: "⚠️ Не вдалося прив'язати технологію до цілі.", AR: "⚠️ تعذّر إلحاق التقنية بذلك الهدف.",
	},

	KeyRoadmapCriteriaPrompt: {
		RU: "📌 Что для тебя значит «я знаю *%s*»? Напиши критерий одним сообщением (до %d символов) — или нажми «Пропустить».",
		EN: "📌 What does \"I know *%s*\" mean for you? Write the criteria in one message (up to %d characters) — or tap Skip.",
		DE: "📌 Was bedeutet \"Ich kann *%s*\" für dich? Schreib das Kriterium in einer Nachricht (bis zu %d Zeichen) — oder tippe auf Überspringen.",
		UK: "📌 Що для тебе означає «я знаю *%s*»? Напиши критерій одним повідомленням (до %d символів) — або натисни «Пропустити».",
		AR: "📌 ماذا يعني لك \"أنا أعرف *%s*\"؟ اكتب المعيار في رسالة واحدة (حتى %d حرفًا) — أو اضغط تخطٍّ.",
	},
	KeyRoadmapCriteriaTooLong: {
		RU: "⚠️ Слишком длинный критерий — максимум %d символов. Попробуй короче:", EN: "⚠️ That's too long — %d characters max. Try a shorter one:", DE: "⚠️ Das ist zu lang — maximal %d Zeichen. Versuch es kürzer:", UK: "⚠️ Занадто довгий критерій — максимум %d символів. Спробуй коротше:", AR: "⚠️ هذا طويل جدًا — %d حرفًا كحد أقصى. اكتب نصًا أقصر:",
	},
	KeyRoadmapCriteriaSaved: {
		RU: "📌 Критерий сохранён.", EN: "📌 Criteria saved.", DE: "📌 Kriterium gespeichert.", UK: "📌 Критерій збережено.", AR: "📌 تم حفظ المعيار.",
	},
	KeyRoadmapCriteriaSkipped: {
		RU: "⏭ Без критерия — его можно указать позже.", EN: "⏭ No criteria for now — you can set it later.", DE: "⏭ Vorerst kein Kriterium — du kannst es später festlegen.", UK: "⏭ Без критерію — його можна вказати пізніше.", AR: "⏭ بلا معيار حاليًا — يمكنك تحديده لاحقًا.",
	},
	KeyRoadmapCriteriaFailed: {
		RU: "⚠️ Не удалось сохранить критерий. Попробуй позже.", EN: "⚠️ Failed to save the criteria. Please try again later.", DE: "⚠️ Kriterium konnte nicht gespeichert werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося зберегти критерій. Спробуй пізніше.", AR: "⚠️ تعذّر حفظ المعيار. حاول لاحقًا.",
	},

	KeyRoadmapRenamePrompt: {
		RU: "✏️ Отправь новое название для технологии *%s* (2-60 символов, одна строка):", EN: "✏️ Send a new name for the *%s* technology (2-60 characters, single line):", DE: "✏️ Sende einen neuen Namen für die Technologie *%s* (2-60 Zeichen, eine Zeile):", UK: "✏️ Надішли нову назву для технології *%s* (2-60 символів, один рядок):", AR: "✏️ أرسل اسمًا جديدًا للتقنية *%s* (2-60 حرفًا، سطر واحد):",
	},
	KeyRoadmapRenamed: {
		RU: "✅ Теперь это *%s*.", EN: "✅ Now called *%s*.", DE: "✅ Jetzt *%s*.", UK: "✅ Тепер це *%s*.", AR: "✅ الاسم الآن *%s*.",
	},
	KeyRoadmapRenameFailed: {
		RU: "⚠️ Не удалось переименовать технологию.", EN: "⚠️ Failed to rename the technology.", DE: "⚠️ Technologie konnte nicht umbenannt werden.", UK: "⚠️ Не вдалося перейменувати технологію.", AR: "⚠️ تعذّرت إعادة تسمية التقنية.",
	},

	KeyRoadmapAddCardsPromptFirst: {
		RU: "🗂 Теперь пришли, что нужно изучить — по одному пункту на строку (тема, статья со ссылкой, книга, лекция). Можно сразу списком.\n\nНеобязательные теги в строке: `#article` `#book` `#lecture` `#topic` и `!easy` `!mid` `!hard` — по ним бот ведёт тебя от простого к сложному.\n\nНажми ✅ Готово, когда закончишь.",
		EN: "🗂 Now send what you need to learn — one item per line (a topic, an article with its link, a book, a lecture). A whole list at once is fine.\n\nOptional inline tags: `#article` `#book` `#lecture` `#topic` and `!easy` `!mid` `!hard` — they're what lets the bot walk you from easy to hard.\n\nTap ✅ Done when finished.",
		DE: "🗂 Sende jetzt, was du lernen willst — ein Punkt pro Zeile (Thema, Artikel mit Link, Buch, Vorlesung). Eine ganze Liste auf einmal ist in Ordnung.\n\nOptionale Tags in der Zeile: `#article` `#book` `#lecture` `#topic` und `!easy` `!mid` `!hard` — damit führt dich der Bot von leicht zu schwer.\n\nTippe auf ✅ Fertig, wenn du fertig bist.",
		UK: "🗂 Тепер надішли, що потрібно вивчити — по одному пункту на рядок (тема, стаття з посиланням, книга, лекція). Можна одразу списком.\n\nНеобов'язкові теги в рядку: `#article` `#book` `#lecture` `#topic` та `!easy` `!mid` `!hard` — за ними бот веде тебе від простого до складного.\n\nНатисни ✅ Готово, коли закінчиш.",
		AR: "🗂 أرسل الآن ما تريد تعلّمه — عنصر واحد في كل سطر (موضوع، مقالة مع رابطها، كتاب، محاضرة). يمكنك لصق قائمة كاملة.\n\nوسوم اختيارية داخل السطر: `#article` `#book` `#lecture` `#topic` و`!easy` `!mid` `!hard` — بها ينقلك البوت من الأسهل إلى الأصعب.\n\nاضغط ✅ تم عند الانتهاء.",
	},
	KeyRoadmapAddCardsPromptMore: {
		RU: "🗂 Пришли новые пункты — по одному на строку. Теги `#book` `!hard` и т.п. работают так же. Нажми ✅ Готово, когда закончишь.",
		EN: "🗂 Send more items — one per line. Tags like `#book` and `!hard` work the same. Tap ✅ Done when finished.",
		DE: "🗂 Sende weitere Punkte — einen pro Zeile. Tags wie `#book` und `!hard` funktionieren genauso. Tippe auf ✅ Fertig, wenn du fertig bist.",
		UK: "🗂 Надішли нові пункти — по одному на рядок. Теги `#book` `!hard` тощо працюють так само. Натисни ✅ Готово, коли закінчиш.",
		AR: "🗂 أرسل عناصر إضافية — عنصرًا واحدًا في كل سطر. الوسوم مثل `#book` و`!hard` تعمل بالطريقة نفسها. اضغط ✅ تم عند الانتهاء.",
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
		RU: "🔁 *Архив*", EN: "🔁 *Archive*", DE: "🔁 *Archiv*", UK: "🔁 *Архів*", AR: "🔁 *الأرشيف*",
	},
	KeyRoadmapArchiveEmpty: {
		RU: "🔁 В архиве пусто.", EN: "🔁 The archive is empty.", DE: "🔁 Das Archiv ist leer.", UK: "🔁 В архіві порожньо.", AR: "🔁 الأرشيف فارغ.",
	},
	KeyRoadmapArchiveGoalsHdr: {
		RU: "\n🎯 Цели:", EN: "\n🎯 Goals:", DE: "\n🎯 Ziele:", UK: "\n🎯 Цілі:", AR: "\n🎯 الأهداف:",
	},
	KeyRoadmapArchiveTechHdr: {
		RU: "\n🧭 Технологии:", EN: "\n🧭 Technologies:", DE: "\n🧭 Technologien:", UK: "\n🧭 Технології:", AR: "\n🧭 التقنيات:",
	},

	KeyRoadmapPushIntervalPrompt: {
		RU: "⏰ Как часто присылать напоминание с самым простым из того, что осталось?", EN: "⏰ How often should the bot send the easiest of what's left?", DE: "⏰ Wie oft soll der Bot das Leichteste vom Rest schicken?", UK: "⏰ Як часто надсилати нагадування з найпростішим із того, що залишилося?", AR: "⏰ كم مرة يرسل البوت أسهل ما تبقّى؟",
	},
	KeyRoadmapPushActivateFailed: {
		RU: "⚠️ Не удалось включить напоминания. Попробуй позже.", EN: "⚠️ Failed to turn on reminders. Please try again later.", DE: "⚠️ Erinnerungen konnten nicht aktiviert werden. Bitte versuch es später noch einmal.", UK: "⚠️ Не вдалося увімкнути нагадування. Спробуй пізніше.", AR: "⚠️ تعذّر تشغيل التذكيرات. حاول لاحقًا.",
	},
	KeyRoadmapPushActivated: {
		RU: "🔔 Напоминания включены — дайджест каждые %d мин.", EN: "🔔 Reminders on — a digest every %d min.", DE: "🔔 Erinnerungen aktiv — eine Übersicht alle %d Min.", UK: "🔔 Нагадування увімкнено — дайджест кожні %d хв.", AR: "🔔 التذكيرات مُفعّلة — ملخّص كل %d دقيقة.",
	},
	KeyRoadmapPushNeedOne: {
		RU: "⚠️ Нужна хотя бы одна активная технология с незакрытыми карточками.", EN: "⚠️ You need at least one active technology with unfinished cards.", DE: "⚠️ Du brauchst mindestens eine aktive Technologie mit offenen Karten.", UK: "⚠️ Потрібна хоча б одна активна технологія з незакритими картками.", AR: "⚠️ تحتاج إلى تقنية نشطة واحدة على الأقل تحتوي بطاقات غير مكتملة.",
	},

	KeyRoadmapDigestTitle: {
		RU: "🗺 *Следующие шаги*", EN: "🗺 *Next steps*", DE: "🗺 *Nächste Schritte*", UK: "🗺 *Наступні кроки*", AR: "🗺 *الخطوات التالية*",
	},
	KeyRoadmapDigestRoadmapLine: {
		RU: "\n*%s* — %d/%d\n", EN: "\n*%s* — %d/%d\n", DE: "\n*%s* — %d/%d\n", UK: "\n*%s* — %d/%d\n", AR: "\n*%s* — %d/%d\n",
	},
	KeyRoadmapDigestEmpty: {
		RU: "🎉 Все карточки закрыты — отличная работа!", EN: "🎉 Every card is done — great work!", DE: "🎉 Alle Karten sind erledigt — großartig!", UK: "🎉 Усі картки закриті — чудова робота!", AR: "🎉 كل البطاقات مكتملة — عمل رائع!",
	},

	KeyRoadmapStatsTitle: {
		RU: "📈 *Прогресс*", EN: "📈 *Progress*", DE: "📈 *Fortschritt*", UK: "📈 *Прогрес*", AR: "📈 *التقدّم*",
	},
	KeyRoadmapStatsGoalLine: {
		RU: "\n🎯 *%s* — %d/%d (%d%%)\n", EN: "\n🎯 *%s* — %d/%d (%d%%)\n", DE: "\n🎯 *%s* — %d/%d (%d%%)\n", UK: "\n🎯 *%s* — %d/%d (%d%%)\n", AR: "\n🎯 *%s* — %d/%d (%d%%)\n",
	},
	KeyRoadmapStatsRoadmapLine: {
		RU: "   • %s — %d/%d (%d%%)\n", EN: "   • %s — %d/%d (%d%%)\n", DE: "   • %s — %d/%d (%d%%)\n", UK: "   • %s — %d/%d (%d%%)\n", AR: "   • %s — %d/%d (%d%%)\n",
	},
	KeyRoadmapStatsCriteriaLine: {
		RU: "     📌 %s\n", EN: "     📌 %s\n", DE: "     📌 %s\n", UK: "     📌 %s\n", AR: "     📌 %s\n",
	},
	KeyRoadmapStatsNoGoalHeader: {
		RU: "\n🧭 *Без цели*\n", EN: "\n🧭 *Without a goal*\n", DE: "\n🧭 *Ohne Ziel*\n", UK: "\n🧭 *Без цілі*\n", AR: "\n🧭 *بدون هدف*\n",
	},
	KeyRoadmapStatsEmpty: {
		RU: "📈 Пока нечего показать — создай первую цель.", EN: "📈 Nothing to show yet — create your first goal.", DE: "📈 Noch nichts zu zeigen — erstelle dein erstes Ziel.", UK: "📈 Поки нічого показати — створи першу ціль.", AR: "📈 لا شيء لعرضه بعد — أنشئ هدفك الأول.",
	},
	KeyRoadmapStatsLoadFailed: {
		RU: "⚠️ Не удалось загрузить прогресс.", EN: "⚠️ Failed to load progress.", DE: "⚠️ Fortschritt konnte nicht geladen werden.", UK: "⚠️ Не вдалося завантажити прогрес.", AR: "⚠️ تعذّر تحميل التقدّم.",
	},

	// Roadmap AI. The added-cards count deliberately reuses
	// KeyRoadmapAddCardsAdded — the confirmation is identical whether the
	// cards came from a paste or from the model, and two keys rendering the
	// same string collide in i18n's reverse index (see
	// TestCatalog_NoTextCollisionsWithinLanguage).
	KeyRoadmapButtonAIPlan: {
		RU: "✨ План от ИИ", EN: "✨ AI plan", DE: "✨ KI-Plan", UK: "✨ План від ШІ", AR: "✨ خطة بالذكاء الاصطناعي",
	},
	KeyRoadmapButtonAIPaste: {
		RU: "🤖 Разметить вставку", EN: "🤖 Tag a paste", DE: "🤖 Einfügung einordnen", UK: "🤖 Розмітити вставку", AR: "🤖 تصنيف نص ملصق",
	},
	KeyRoadmapButtonAIQuiz: {
		RU: "❓", EN: "❓", DE: "❓", UK: "❓", AR: "❓",
	},
	KeyRoadmapButtonQuizDone: {
		RU: "✅ Отметить выученным", EN: "✅ Mark as learned", DE: "✅ Als gelernt markieren", UK: "✅ Позначити вивченим", AR: "✅ اعتبارها متعلَّمة",
	},
	KeyRoadmapAIWorking: {
		RU: "✨ Составляю план, это займёт до минуты…", EN: "✨ Building the plan, this can take up to a minute…", DE: "✨ Der Plan entsteht, das kann bis zu einer Minute dauern…", UK: "✨ Складаю план, це займе до хвилини…", AR: "✨ يجري إعداد الخطة، قد يستغرق ذلك دقيقة…",
	},
	KeyRoadmapAIQuizWorking: {
		RU: "❓ Придумываю вопрос…", EN: "❓ Coming up with a question…", DE: "❓ Eine Frage entsteht…", UK: "❓ Придумую питання…", AR: "❓ يجري تحضير سؤال…",
	},
	KeyRoadmapAIFailed: {
		RU: "🤖 ИИ не ответил. Попробуй ещё раз чуть позже.", EN: "🤖 The AI didn't answer. Try again in a bit.", DE: "🤖 Die KI hat nicht geantwortet. Versuche es später erneut.", UK: "🤖 ШІ не відповів. Спробуй ще раз трохи пізніше.", AR: "🤖 لم يستجب الذكاء الاصطناعي. حاول لاحقًا.",
	},
	KeyRoadmapAIEmpty: {
		RU: "🤖 ИИ не вернул ничего пригодного. Попробуй ещё раз.", EN: "🤖 The AI returned nothing usable. Try again.", DE: "🤖 Die KI hat nichts Brauchbares geliefert. Versuche es erneut.", UK: "🤖 ШІ не повернув нічого придатного. Спробуй ще раз.", AR: "🤖 لم يعد الذكاء الاصطناعي بشيء صالح. حاول مرة أخرى.",
	},
	KeyRoadmapAIDisabled: {
		RU: "🤖 ИИ не подключён.", EN: "🤖 AI is not connected.", DE: "🤖 KI ist nicht angebunden.", UK: "🤖 ШІ не підключений.", AR: "🤖 الذكاء الاصطناعي غير موصول.",
	},
	KeyRoadmapAIRejectedFmt: {
		RU: " (отброшено: %d — не прошли проверку.)", EN: " (%d dropped — failed validation.)", DE: " (%d verworfen — Prüfung nicht bestanden.)", UK: " (відкинуто: %d — не пройшли перевірку.)", AR: " (تم استبعاد %d — لم تجتز التحقق.)",
	},
	KeyRoadmapAIPastePrompt: {
		RU: "🤖 Пришли пункты — по одному на строку. Теги ставить не нужно: вид и сложность определит ИИ.\n\nНажми ✅ Готово, когда закончишь.", EN: "🤖 Send the items, one per line. No tags needed — the AI decides kind and difficulty.\n\nTap ✅ Done when finished.", DE: "🤖 Sende die Punkte, einen pro Zeile. Keine Tags nötig — Art und Schwierigkeit bestimmt die KI.\n\nTippe ✅ Fertig, wenn du durch bist.", UK: "🤖 Надішли пункти — по одному на рядок. Теги не потрібні: вид і складність визначить ШІ.\n\nНатисни ✅ Готово, коли закінчиш.", AR: "🤖 أرسل العناصر، عنصرًا في كل سطر. لا حاجة للوسوم — الذكاء الاصطناعي يحدد النوع والصعوبة.\n\nاضغط ✅ تم عند الانتهاء.",
	},
	KeyRoadmapAIQuizPromptFmt: {
		RU: "❓ *Вопрос*\n\n%s\n\nОтветь текстом — или нажми ❌ Отмена.", EN: "❓ *Question*\n\n%s\n\nAnswer in a message — or tap ❌ Cancel.", DE: "❓ *Frage*\n\n%s\n\nAntworte mit einer Nachricht — oder tippe ❌ Abbrechen.", UK: "❓ *Питання*\n\n%s\n\nВідповідай текстом — або натисни ❌ Скасувати.", AR: "❓ *سؤال*\n\n%s\n\nأجب برسالة — أو اضغط ❌ إلغاء.",
	},
	KeyRoadmapAIQuizCorrect: {
		RU: "✅ Верно", EN: "✅ Correct", DE: "✅ Richtig", UK: "✅ Правильно", AR: "✅ صحيح",
	},
	KeyRoadmapAIQuizPartial: {
		RU: "🟡 Почти", EN: "🟡 Almost", DE: "🟡 Fast", UK: "🟡 Майже", AR: "🟡 تقريبًا",
	},
	KeyRoadmapAIQuizWrong: {
		RU: "🔴 Мимо", EN: "🔴 Not quite", DE: "🔴 Daneben", UK: "🔴 Не те", AR: "🔴 غير صحيح",
	},
	KeyRoadmapAIQuizGradeFmt: {
		RU: "%s\n\n%s", EN: "%s\n\n%s", DE: "%s\n\n%s", UK: "%s\n\n%s", AR: "%s\n\n%s",
	},
	KeyRoadmapAIDigestHintFmt: {
		RU: "\n\n💡 %s", EN: "\n\n💡 %s", DE: "\n\n💡 %s", UK: "\n\n💡 %s", AR: "\n\n💡 %s",
	},
}
