package i18n

// catalogTrackReports holds the Track Reports screens: Today (chart +
// select-activities), Period (menu, text report, chart report, calendar).
var catalogTrackReports = map[string]entry{
	KeyTrackButtonToday: {
		RU: "📊 Сегодня", EN: "📊 Today", DE: "📊 Heute", UK: "📊 Сьогодні", AR: "📊 اليوم",
	},
	KeyTrackButtonCalendar: {
		RU: "📅 Календарь", EN: "📅 Calendar", DE: "📅 Kalender", UK: "📅 Календар", AR: "📅 التقويم",
	},
	KeyTrackButtonHeatmap: {
		RU: "🔥 Карта активности", EN: "🔥 Heatmap", DE: "🔥 Heatmap", UK: "🔥 Карта активності", AR: "🔥 خريطة النشاط",
	},
	KeyTrackHeatmapTitle: {
		RU: "🔥 *Карта активности* — последние 8 недель", EN: "🔥 *Heatmap* — last 8 weeks", DE: "🔥 *Heatmap* — letzte 8 Wochen", UK: "🔥 *Карта активності* — останні 8 тижнів", AR: "🔥 *خريطة النشاط* — آخر 8 أسابيع",
	},
	KeyTrackHeatmapLegend: {
		RU: "🟩 трекал  ⬛ пропуск  ⬜ впереди", EN: "🟩 tracked  ⬛ missed  ⬜ upcoming", DE: "🟩 erfasst  ⬛ verpasst  ⬜ bevorstehend", UK: "🟩 трекав  ⬛ пропуск  ⬜ попереду", AR: "🟩 تم التتبع  ⬛ تفويت  ⬜ قادم",
	},
	KeyTrackHeatmapDaysTracked: {
		RU: "%d из %d дней с трекингом", EN: "%d of %d days tracked", DE: "%d von %d Tagen erfasst", UK: "%d з %d днів з трекінгом", AR: "%d من %d أيام تم تتبعها",
	},
	KeyTrackHeatmapHint: {
		RU: "Тапни на квадратик, чтобы посмотреть, что было в этот день", EN: "Tap a square to see what happened that day", DE: "Tippe auf ein Feld, um zu sehen, was an diesem Tag war", UK: "Тапни на квадратик, щоб побачити, що було цього дня", AR: "اضغط على مربع لترى ما حدث في ذلك اليوم",
	},
	KeyTrackHeatmapDayTrackedHeader: {
		RU: "⏱ Трекинг:", EN: "⏱ Tracked:", DE: "⏱ Erfasst:", UK: "⏱ Трекінг:", AR: "⏱ تم التتبع:",
	},
	KeyTrackHeatmapDayNoActivity: {
		RU: "В этот день ничего не трекалось.", EN: "No tracked time this day.", DE: "An diesem Tag wurde nichts erfasst.", UK: "Цього дня нічого не трекалося.", AR: "لم يتم تتبع أي وقت في هذا اليوم.",
	},
	KeyTrackHeatmapDayActivityLine: {
		RU: "• %s — %s (%d)\n", EN: "• %s — %s (%d)\n", DE: "• %s — %s (%d)\n", UK: "• %s — %s (%d)\n", AR: "• %s — %s (%d)\n",
	},
	KeyTrackHeatmapDayReviewsHeader: {
		RU: "🧠 Слова:", EN: "🧠 Words:", DE: "🧠 Wörter:", UK: "🧠 Слова:", AR: "🧠 الكلمات:",
	},
	KeyTrackHeatmapDayNoReviews: {
		RU: "В этот день слова не повторялись.", EN: "No word reviews this day.", DE: "An diesem Tag wurden keine Wörter wiederholt.", UK: "Цього дня слова не повторювалися.", AR: "لم تتم مراجعة أي كلمات في هذا اليوم.",
	},
	KeyTrackHeatmapDayReviewLine: {
		RU: "%s %s → %s\n", EN: "%s %s → %s\n", DE: "%s %s → %s\n", UK: "%s %s → %s\n", AR: "%s %s → %s\n",
	},
	KeyTrackLabelBackToReports: {
		RU: "↩️ Назад к отчётам", EN: "↩️ Back to Reports", DE: "↩️ Zurück zu Berichten", UK: "↩️ Назад до звітів", AR: "↩️ العودة إلى التقارير",
	},
	KeyTrackLabelSelectedActivities: {
		RU: "Выбранные активности", EN: "Selected activities", DE: "Ausgewählte Aktivitäten", UK: "Вибрані активності", AR: "الأنشطة المحددة",
	},
	KeyTrackLabelTextReport: {
		RU: "📄 Текстовый отчёт", EN: "📄 Text report", DE: "📄 Textbericht", UK: "📄 Текстовий звіт", AR: "📄 تقرير نصي",
	},
	KeyTrackLabelChartReport: {
		RU: "📉 Отчёт-график", EN: "📉 Chart report", DE: "📉 Diagrammbericht", UK: "📉 Звіт-графік", AR: "📉 تقرير بياني",
	},
	KeyTrackLabelSelectActivities: {
		RU: "🧩 Выбрать активности", EN: "🧩 Select activities", DE: "🧩 Aktivitäten auswählen", UK: "🧩 Обрати активності", AR: "🧩 اختيار الأنشطة",
	},
	KeyTrackLabelBuildChart: {
		RU: "✅ Построить график", EN: "✅ Build chart", DE: "✅ Diagramm erstellen", UK: "✅ Побудувати графік", AR: "✅ إنشاء الرسم البياني",
	},
	KeyTrackLabelRangePrefix: {
		RU: "🗓 Период: %s", EN: "🗓 Range: %s", DE: "🗓 Zeitraum: %s", UK: "🗓 Період: %s", AR: "🗓 النطاق: %s",
	},
	KeyTrackLabelConfirmRange: {
		RU: "✅ Подтвердить период", EN: "✅ Confirm range", DE: "✅ Zeitraum bestätigen", UK: "✅ Підтвердити період", AR: "✅ تأكيد النطاق",
	},
	KeyTrackLabelSelectEndDate: {
		RU: "Выбери дату окончания", EN: "Select end date", DE: "Enddatum wählen", UK: "Обери дату завершення", AR: "اختر تاريخ الانتهاء",
	},
	KeyTrackCalendarCancel: {
		RU: "Отмена", EN: "Cancel", DE: "Abbrechen", UK: "Скасувати", AR: "إلغاء",
	},
	KeyTrackCalendarMonth: {
		RU: "Месяц", EN: "Month", DE: "Monat", UK: "Місяць", AR: "الشهر",
	},

	KeyTrackCalendarMon: {RU: "Пн", EN: "Mo", DE: "Mo", UK: "Пн", AR: "ن"},
	KeyTrackCalendarTue: {RU: "Вт", EN: "Tu", DE: "Di", UK: "Вт", AR: "ث"},
	KeyTrackCalendarWed: {RU: "Ср", EN: "We", DE: "Mi", UK: "Ср", AR: "ر"},
	KeyTrackCalendarThu: {RU: "Чт", EN: "Th", DE: "Do", UK: "Чт", AR: "خ"},
	KeyTrackCalendarFri: {RU: "Пт", EN: "Fr", DE: "Fr", UK: "Пт", AR: "ج"},
	KeyTrackCalendarSat: {RU: "Сб", EN: "Sa", DE: "Sa", UK: "Сб", AR: "س"},
	KeyTrackCalendarSun: {RU: "Вс", EN: "Su", DE: "So", UK: "Нд", AR: "ح"},

	KeyTrackCalendarMonth01: {RU: "Январь", EN: "January", DE: "Januar", UK: "Січень", AR: "يناير"},
	KeyTrackCalendarMonth02: {RU: "Февраль", EN: "February", DE: "Februar", UK: "Лютий", AR: "فبراير"},
	KeyTrackCalendarMonth03: {RU: "Март", EN: "March", DE: "März", UK: "Березень", AR: "مارس"},
	KeyTrackCalendarMonth04: {RU: "Апрель", EN: "April", DE: "April", UK: "Квітень", AR: "أبريل"},
	KeyTrackCalendarMonth05: {RU: "Май", EN: "May", DE: "Mai", UK: "Травень", AR: "مايو"},
	KeyTrackCalendarMonth06: {RU: "Июнь", EN: "June", DE: "Juni", UK: "Червень", AR: "يونيو"},
	KeyTrackCalendarMonth07: {RU: "Июль", EN: "July", DE: "Juli", UK: "Липень", AR: "يوليو"},
	KeyTrackCalendarMonth08: {RU: "Август", EN: "August", DE: "August", UK: "Серпень", AR: "أغسطس"},
	KeyTrackCalendarMonth09: {RU: "Сентябрь", EN: "September", DE: "September", UK: "Вересень", AR: "سبتمبر"},
	KeyTrackCalendarMonth10: {RU: "Октябрь", EN: "October", DE: "Oktober", UK: "Жовтень", AR: "أكتوبر"},
	KeyTrackCalendarMonth11: {RU: "Ноябрь", EN: "November", DE: "November", UK: "Листопад", AR: "نوفمبر"},
	KeyTrackCalendarMonth12: {RU: "Декабрь", EN: "December", DE: "Dezember", UK: "Грудень", AR: "ديسمبر"},

	KeyTrackReportsHubTitle: {
		RU: "📈 Отчёты\n\nВыбери тип отчёта:", EN: "📈 Reports\n\nChoose a report type:", DE: "📈 Berichte\n\nWähle einen Berichtstyp:", UK: "📈 Звіти\n\nОбери тип звіту:", AR: "📈 التقارير\n\nاختر نوع التقرير:",
	},

	KeyTrackTodayChartLoadFailed: {
		RU: "⚠️ Не удалось загрузить данные графика.", EN: "⚠️ Failed to load chart data.", DE: "⚠️ Diagrammdaten konnten nicht geladen werden.", UK: "⚠️ Не вдалося завантажити дані графіка.", AR: "⚠️ تعذّر تحميل بيانات الرسم البياني.",
	},
	KeyTrackTodayChartEmpty: {
		RU: "📉 Пока нет данных для графика.", EN: "📉 No data for chart yet.", DE: "📉 Noch keine Daten für das Diagramm.", UK: "📉 Поки немає даних для графіка.", AR: "📉 لا توجد بيانات للرسم البياني بعد.",
	},
	KeyTrackTodayChartTitle: {
		RU: "📉 График за сегодня\n\n", EN: "📉 Today Chart\n\n", DE: "📉 Heutiges Diagramm\n\n", UK: "📉 Графік за сьогодні\n\n", AR: "📉 مخطط اليوم\n\n",
	},
	// Pure formatting, no natural-language words — identical across
	// languages (same reasoning as KeyTrackPeriodTextActivityLine).
	KeyTrackTodayChartActivityLine: {
		RU: "%s\n%s %s (%s)\n\n", EN: "%s\n%s %s (%s)\n\n", DE: "%s\n%s %s (%s)\n\n", UK: "%s\n%s %s (%s)\n\n", AR: "%s\n%s %s (%s)\n\n",
	},

	KeyTrackTodaySelectTitle: {
		RU: "🧩 Выбери активности для графика за сегодня", EN: "🧩 Select activities for today chart", DE: "🧩 Wähle Aktivitäten für das heutige Diagramm", UK: "🧩 Обери активності для графіка за сьогодні", AR: "🧩 اختر الأنشطة لمخطط اليوم",
	},

	KeyTrackPeriodMenuLoadFailed: {
		RU: "⚠️ Не удалось загрузить активности для периода.", EN: "⚠️ Failed to load activities for period.", DE: "⚠️ Aktivitäten für den Zeitraum konnten nicht geladen werden.", UK: "⚠️ Не вдалося завантажити активності для періоду.", AR: "⚠️ تعذّر تحميل الأنشطة للفترة.",
	},
	KeyTrackPeriodMenuTitle: {
		RU: "📅 Отчёт за период\nВыбрано: %d активностей\nПериод: %s",
		EN: "📅 Period Report\nSelected: %d activities\nRange: %s",
		DE: "📅 Zeitraumbericht\nAusgewählt: %d Aktivitäten\nZeitraum: %s",
		UK: "📅 Звіт за період\nВибрано: %d активностей\nПеріод: %s",
		AR: "📅 تقرير الفترة\nتم اختيار: %d نشاطًا\nالنطاق: %s",
	},

	KeyTrackPeriodTextFailed: {
		RU: "⚠️ Не удалось сформировать отчёт за период.", EN: "⚠️ Failed to build period report.", DE: "⚠️ Zeitraumbericht konnte nicht erstellt werden.", UK: "⚠️ Не вдалося сформувати звіт за період.", AR: "⚠️ تعذّر إنشاء تقرير الفترة.",
	},
	KeyTrackPeriodTextTitle: {
		RU: "📄 Отчёт за период\n\n", EN: "📄 Period Report\n\n", DE: "📄 Zeitraumbericht\n\n", UK: "📄 Звіт за період\n\n", AR: "📄 تقرير الفترة\n\n",
	},
	KeyTrackPeriodRangeLine: {
		RU: "Период: %s..%s\n", EN: "Range: %s..%s\n", DE: "Zeitraum: %s..%s\n", UK: "Період: %s..%s\n", AR: "النطاق: %s..%s\n",
	},
	KeyTrackPeriodScopeSelected: {
		RU: "Охват: выбранные активности\n", EN: "Scope: selected activities\n", DE: "Umfang: ausgewählte Aktivitäten\n", UK: "Охоплення: вибрані активності\n", AR: "النطاق: الأنشطة المحددة\n",
	},
	KeyTrackPeriodScopeAll: {
		RU: "Охват: все выбранные в меню\n", EN: "Scope: all selected in menu\n", DE: "Umfang: alle im Menü ausgewählten\n", UK: "Охоплення: усі вибрані в меню\n", AR: "النطاق: كل ما تم اختياره في القائمة\n",
	},
	KeyTrackPeriodTotalsLine: {
		RU: "Итого: %s\nСессий: %d\n\n", EN: "Total: %s\nSessions: %d\n\n", DE: "Gesamt: %s\nSitzungen: %d\n\n", UK: "Разом: %s\nСесій: %d\n\n", AR: "الإجمالي: %s\nالجلسات: %d\n\n",
	},
	KeyTrackPeriodNoSessions: {
		RU: "Нет сессий за этот период.", EN: "No sessions for this period.", DE: "Keine Sitzungen für diesen Zeitraum.", UK: "Немає сесій за цей період.", AR: "لا توجد جلسات لهذه الفترة.",
	},
	// Pure formatting, no natural-language words — identical across
	// languages by design (same reasoning as duration/percent formatting
	// elsewhere in this codebase, e.g. track.TrackingMenuText's "4h 30m").
	KeyTrackPeriodTextActivityLine: {
		RU: "%d) %s - %s (%s, %d)\n", EN: "%d) %s - %s (%s, %d)\n", DE: "%d) %s - %s (%s, %d)\n", UK: "%d) %s - %s (%s, %d)\n", AR: "%d) %s - %s (%s, %d)\n",
	},

	KeyTrackPeriodChartFailed: {
		RU: "⚠️ Не удалось построить график за период.", EN: "⚠️ Failed to build period chart.", DE: "⚠️ Zeitraumdiagramm konnte nicht erstellt werden.", UK: "⚠️ Не вдалося побудувати графік за період.", AR: "⚠️ تعذّر إنشاء الرسم البياني للفترة.",
	},
	KeyTrackPeriodChartEmpty: {
		RU: "📉 Нет данных за выбранный период.", EN: "📉 No data for selected period.", DE: "📉 Keine Daten für den ausgewählten Zeitraum.", UK: "📉 Немає даних за обраний період.", AR: "📉 لا توجد بيانات للفترة المحددة.",
	},
	KeyTrackPeriodChartTitle: {
		RU: "📉 График за период\n\n", EN: "📉 Period Chart\n\n", DE: "📉 Zeitraumdiagramm\n\n", UK: "📉 Графік за період\n\n", AR: "📉 مخطط الفترة\n\n",
	},
	KeyTrackPeriodChartActivityLine: {
		RU: "%s\n%s %s (%s, %d)\n\n", EN: "%s\n%s %s (%s, %d)\n\n", DE: "%s\n%s %s (%s, %d)\n\n", UK: "%s\n%s %s (%s, %d)\n\n", AR: "%s\n%s %s (%s, %d)\n\n",
	},

	KeyTrackCalendarPickTitle: {
		RU: "📅 Выбери дни периода\nОт: %s\nДо: %s", EN: "📅 Pick period days\nFrom: %s\nTo: %s", DE: "📅 Wähle die Tage des Zeitraums\nVon: %s\nBis: %s", UK: "📅 Обери дні періоду\nВід: %s\nДо: %s", AR: "📅 اختر أيام الفترة\nمن: %s\nإلى: %s",
	},

	KeyTrackGranularityByMonths: {
		RU: "\nПо месяцам:\n", EN: "\nBy months:\n", DE: "\nNach Monaten:\n", UK: "\nЗа місяцями:\n", AR: "\nحسب الأشهر:\n",
	},
	KeyTrackGranularityByDays: {
		RU: "\nПо дням:\n", EN: "\nBy days:\n", DE: "\nNach Tagen:\n", UK: "\nЗа днями:\n", AR: "\nحسب الأيام:\n",
	},
	KeyTrackGranularityByHours: {
		RU: "\nПо часам:\n", EN: "\nBy hours:\n", DE: "\nNach Stunden:\n", UK: "\nЗа годинами:\n", AR: "\nحسب الساعات:\n",
	},
	KeyTrackGranularityBucketLine: {
		RU: "- %s: %s\n", EN: "- %s: %s\n", DE: "- %s: %s\n", UK: "- %s: %s\n", AR: "- %s: %s\n",
	},

	KeyTrackPeriodRangeInvalidFmt: {
		RU: "Используй формат: YYYY-MM-DD..YYYY-MM-DD", EN: "Use format: YYYY-MM-DD..YYYY-MM-DD", DE: "Verwende das Format: YYYY-MM-DD..YYYY-MM-DD", UK: "Використовуй формат: YYYY-MM-DD..YYYY-MM-DD", AR: "استخدم الصيغة: YYYY-MM-DD..YYYY-MM-DD",
	},
	KeyTrackPeriodRangeSetConfirm: {
		RU: "Период установлен: %s..%s", EN: "Range set: %s..%s", DE: "Zeitraum festgelegt: %s..%s", UK: "Період встановлено: %s..%s", AR: "تم ضبط النطاق: %s..%s",
	},
	KeyTrackCalendarPickBothDays: {
		RU: "Выбери даты ОТ и ДО.", EN: "Pick FROM and TO days.", DE: "Wähle das Start- und Enddatum.", UK: "Обери дати ВІД і ДО.", AR: "اختر تاريخي البداية والنهاية.",
	},
}
