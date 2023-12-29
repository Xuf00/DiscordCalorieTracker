package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type User struct {
	id             string
	daily_calories int16
}

type FoodLog struct {
	id        int64
	user_id   string
	food_item string
	calories  int16
	date_time time.Time
}

var db *sql.DB

func initDb() {
	var err error
	db, err = sql.Open("sqlite", "app.db")
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	_, err = db.ExecContext(
		context.Background(),
		`CREATE TABLE IF NOT EXISTS user (
			id TEXT PRIMARY KEY,
			daily_calories INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS food_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			food_item TEXT NOT NULL,
			calories INTEGER NOT NULL,
			date_time DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user(id)
		)`,
	)
	if err != nil {
		log.Fatalf("Could not create schema: %v", err)
	}
	log.Printf("Connected to the DB")
}

func userById(id string) (User, error) {
	var user User

	row := db.QueryRowContext(
		context.Background(),
		`SELECT * FROM user WHERE id=?`, id,
	)

	err := row.Scan(&user.id, &user.daily_calories)

	if err != nil && err != sql.ErrNoRows {
		return user, err
	}

	return user, nil
}

func setUserCalories(user *User) (sql.Result, error) {
	log.Printf("Setting the calories in the database for user %v", user.id)
	result, err := db.ExecContext(
		context.Background(),
		`INSERT INTO user (id, daily_calories) VALUES (?,?) ON CONFLICT (id) DO UPDATE SET daily_calories=excluded.daily_calories`,
		user.id, user.daily_calories,
	)
	return result, err
}

func addUserFoodLog(foodLog *FoodLog) (sql.Result, error) {
	log.Printf("Adding a food log to the database for user %v", foodLog.user_id)
	result, err := db.ExecContext(
		context.Background(),
		`INSERT INTO food_log (user_id, food_item, calories) VALUES (?, ?, ?)`,
		foodLog.user_id, foodLog.food_item, foodLog.calories,
	)
	return result, err
}

func updateUserFoodLog(foodLog *FoodLog) (int64, error) {
	result, err := db.ExecContext(
		context.Background(),
		`UPDATE food_log SET food_item=?, calories=? WHERE id=? AND user_id=?`,
		foodLog.food_item, foodLog.calories, foodLog.id, foodLog.user_id,
	)
	if err != nil {
		return 0, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return n, nil
}

func deleteUserFoodLog(userId string, logId int64) (int64, error) {
	result, err := db.ExecContext(
		context.Background(),
		`DELETE FROM food_log WHERE user_id=? AND id=?`,
		userId, logId,
	)
	if err != nil {
		return 0, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return n, nil
}

func fetchDailyFoodLogs(userId string) ([]FoodLog, error) {
	var foodLogs []FoodLog
	rows, err := db.QueryContext(
		context.Background(),
		`SELECT * FROM food_log WHERE user_id=? AND DATE(date_time) = CURRENT_DATE`,
		userId,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var foodLog FoodLog

		if err := rows.Scan(
			&foodLog.id, &foodLog.user_id, &foodLog.food_item, &foodLog.calories, &foodLog.date_time,
		); err != nil {
			return nil, err
		}
		foodLogs = append(foodLogs, foodLog)
	}
	return foodLogs, err
}

func fetchRemainingCalories(userId string) (int64, error) {
	row := db.QueryRowContext(
		context.Background(),
		`SELECT user.daily_calories - COALESCE(SUM(food_log.calories), 0) AS remaining_calories
		FROM user
		LEFT JOIN food_log ON user.id = food_log.user_id AND DATE(food_log.date_time) = CURRENT_DATE
		WHERE user.id=?
		GROUP BY user.id, user.daily_calories;`,
		userId,
	)

	var remainingCalories int64

	err := row.Scan(&remainingCalories)

	if err != nil && err == sql.ErrNoRows {
		return 10000, err
	}

	if err != nil {
		return remainingCalories, err
	}

	return remainingCalories, nil
}
