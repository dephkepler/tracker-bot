package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"
	"tracker-bot/internal/config"
	"tracker-bot/internal/utils/pgclient"
)

func main() {
	var tgUserID int64
	flag.Int64Var(&tgUserID, "tg-user-id", 0, "telegram user id")
	flag.Parse()
	if tgUserID <= 0 {
		log.Fatal("pass -tg-user-id")
	}

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("parse config: %v", err)
	}
	db, err := pgclient.New(context.Background(), cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("pg connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	var userID int64
	if err := db.Pool().QueryRow(ctx, "SELECT id FROM users WHERE tg_user_id=$1", tgUserID).Scan(&userID); err != nil {
		log.Fatalf("find user by tg_user_id: %v", err)
	}

	rows, err := db.Pool().Query(ctx, "SELECT id FROM activities WHERE user_id=$1 AND is_archived=FALSE ORDER BY id", userID)
	if err != nil {
		log.Fatalf("load activities: %v", err)
	}
	defer rows.Close()
	var activityIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("scan activity id: %v", err)
		}
		activityIDs = append(activityIDs, id)
	}
	if len(activityIDs) < 10 {
		base := []struct {
			name  string
			emoji string
		}{
			{"Go", "🦫"}, {"English", "📚"}, {"Workout", "🏋️"}, {"Reading", "📖"}, {"Coding", "💻"},
			{"Design", "🎨"}, {"Walking", "🚶"}, {"Meditation", "🧘"}, {"Writing", "✍️"}, {"Music", "🎵"},
		}
		for i := len(activityIDs); i < 10; i++ {
			item := base[i%len(base)]
			var id int64
			err := db.Pool().QueryRow(
				ctx,
				`INSERT INTO activities (user_id,name,emoji,is_archived) VALUES ($1,$2,$3,FALSE)
                 ON CONFLICT DO NOTHING
                 RETURNING id`,
				userID, item.name, item.emoji,
			).Scan(&id)
			if err == nil {
				activityIDs = append(activityIDs, id)
				continue
			}
			// no RETURNING row means the insert hit ON CONFLICT — fall back to loading the existing id
			_ = db.Pool().QueryRow(ctx, "SELECT id FROM activities WHERE user_id=$1 AND lower(name)=lower($2) LIMIT 1", userID, item.name).Scan(&id)
			if id > 0 {
				activityIDs = append(activityIDs, id)
			}
		}
	}
	if len(activityIDs) == 0 {
		log.Fatal("no active activities for user after ensure")
	}

	r := rand.New(rand.NewSource(42))
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Now().UTC().AddDate(0, 0, 1)

	insQ := `
	INSERT INTO activity_sessions (user_id, activity_id, start_at, end_at, planned_min, source)
	VALUES ($1,$2,$3,$4,$5,'seed');
	`

	inserted := 0
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		monthWeight := 1 + int(d.Month())/4
		sessionsPerDay := monthWeight + r.Intn(3)
		for i := 0; i < sessionsPerDay; i++ {
			actID := activityIDs[r.Intn(len(activityIDs))]
			minutes := 15 * (1 + r.Intn(8))
			hour := 7 + r.Intn(15)
			min := []int{0, 15, 30, 45}[r.Intn(4)]
			startAt := time.Date(d.Year(), d.Month(), d.Day(), hour, min, 0, 0, time.UTC)
			endAt := startAt.Add(time.Duration(minutes) * time.Minute)
			if _, err := db.Pool().Exec(ctx, insQ, userID, actID, startAt, endAt, minutes); err != nil {
				log.Fatalf("insert seed session: %v", err)
			}
			inserted++
		}
	}

	fmt.Printf("seed completed: user_id=%d sessions=%d range=%s..%s\n", userID, inserted, start.Format("2006-01-02"), end.AddDate(0, 0, -1).Format("2006-01-02"))
}
