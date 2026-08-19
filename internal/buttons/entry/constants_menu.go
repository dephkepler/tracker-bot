package entry

// Reply menu buttons for the entry screen.
const (
	EntryButtonProfile      = "👤My account"
	EntryButtonTrack        = "📈Track"
	EntryButtonLearning     = "🧠Learning"
	EntryButtonSubscription = "💳Subscription"
	// EntryButtonAdmin is only ever rendered into the entry keyboard for the
	// configured admin (see handlers.Module.IsAdmin) — never shown to
	// regular users.
	EntryButtonAdmin = "👑 Admin"
)
