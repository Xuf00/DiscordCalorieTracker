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

func InitDatabase() {
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

func FetchUserByID(id string) (User, error) {
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

func SetUserCalories(user *User) (sql.Result, error) {
	log.Printf("Setting the calories in the database for user %v", user.id)
	result, err := db.ExecContext(
		context.Background(),
		`INSERT INTO user (id, daily_calories) VALUES (?,?) ON CONFLICT (id) DO UPDATE SET daily_calories=excluded.daily_calories`,
		user.id, user.daily_calories,
	)
	return result, err
}

func AddUserFoodLog(foodLog *FoodLog) (int64, error) {
	log.Printf("Adding a food log to the database for user %v", foodLog.user_id)
	result, err := db.ExecContext(
		context.Background(),
		`INSERT INTO food_log (user_id, food_item, calories) VALUES (?, ?, ?)`,
		foodLog.user_id, foodLog.food_item, foodLog.calories,
	)

	id, err := result.LastInsertId()

	return id, err
}

func UpdateUserFoodLog(foodLog *FoodLog) (int64, error) {
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

func DeleteUserFoodLog(userId string, logId int64) (int64, error) {
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

func FetchDailyFoodLogs(userId string, date time.Time) ([]FoodLog, error) {
	dateStr := date.Format("2006-01-02")
	var foodLogs []FoodLog
	rows, err := db.QueryContext(
		context.Background(),
		`SELECT * FROM food_log WHERE user_id=? AND DATE(date_time)=?`,
		userId, dateStr,
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

func FetchConsumedCaloriesForDate(userId string, date time.Time) (int64, error) {
	dateStr := date.Format("2006-01-02")

	row := db.QueryRowContext(
		context.Background(),
		`SELECT SUM(calories) consumed FROM food_log WHERE user_id=? AND DATE(date_time)=?`,
		userId, dateStr,
	)

	var consumedCalories int64

	err := row.Scan(&consumedCalories)

	if err != nil && err == sql.ErrNoRows {
		return 0, err
	}

	if err != nil {
		return consumedCalories, err
	}

	return consumedCalories, nil
}

func FetchAverageConsumedCalories(userId string, date string) (int64, error) {
	row := db.QueryRowContext(
		context.Background(),
		`SELECT AVG(daily_sum) average_calories
		FROM (
			SELECT SUM(calories) daily_sum
			FROM food_log 
			WHERE user_id=?
			AND DATE(date_time) BETWEEN ? AND CURRENT_DATE
			GROUP BY DATE(date_time)
		) AS daily_calories`,
		userId, date,
	)

	var averageCalories int64

	err := row.Scan(&averageCalories)
	if err != nil && err == sql.ErrNoRows {
		return 0, err
	}

	if err != nil {
		return averageCalories, err
	}

	return averageCalories, nil
}

func FetchRemainingCalories(userId string, date time.Time) (int64, error) {
	dateStr := date.Format("2006-01-02")
	row := db.QueryRowContext(
		context.Background(),
		`SELECT user.daily_calories - COALESCE(SUM(food_log.calories), 0) AS remaining_calories
		FROM user
		LEFT JOIN food_log ON user.id = food_log.user_id AND DATE(food_log.date_time)=?
		WHERE user.id=?
		GROUP BY user.id, user.daily_calories;`,
		dateStr, userId,
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

func FetchFoodLogDaysCount(userId string) (int64, error) {
	row := db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(DISTINCT DATE(date_time)) AS days_count
		FROM food_log
		WHERE user_id=?`,
		userId,
	)

	var daysDataCount int64

	err := row.Scan(&daysDataCount)
	if err != nil && err == sql.ErrNoRows {
		return 0, err
	}

	if err != nil {
		return daysDataCount, err
	}

	return daysDataCount, nil
}
