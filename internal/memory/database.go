package memory

import (
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database provides SQLite-based storage for structured user data.
// It stores user profile, preferences, and conversation topics.
type Database struct {
	db *sql.DB
}

// NewDatabase opens or creates a SQLite database in the work directory.
// Returns error if database cannot be opened or initialized.
func NewDatabase(workDir string) (*Database, error) {
	dbPath := filepath.Join(workDir, "vayuu.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	d := &Database{db: db}
	if err := d.init(); err != nil {
		return nil, err
	}

	slog.Info("database initialized", "path", dbPath)
	return d, nil
}

// init creates the database schema if it doesn't exist.
// Tables: user_profile, preferences, topics
func (d *Database) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS user_profile (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS preferences (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		category TEXT NOT NULL,
		confidence REAL DEFAULT 1.0,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		mentions INTEGER DEFAULT 1,
		last_mentioned TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_preferences_key ON preferences(key);
	CREATE INDEX IF NOT EXISTS idx_topics_name ON topics(name);
	`

	_, err := d.db.Exec(schema)
	return err
}

// GetProfile retrieves a single profile value by key.
// Returns empty string if key doesn't exist.
func (d *Database) GetProfile(key string) (string, error) {
	var value string
	err := d.db.QueryRow("SELECT value FROM user_profile WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetProfile stores or updates a profile key-value pair.
func (d *Database) SetProfile(key, value string) error {
	_, err := d.db.Exec(`
		INSERT INTO user_profile (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`, key, value, time.Now().Format(time.RFC3339), value, time.Now().Format(time.RFC3339))
	return err
}

// GetAllProfile returns all profile key-value pairs as a map.
func (d *Database) GetAllProfile() (map[string]string, error) {
	rows, err := d.db.Query("SELECT key, value FROM user_profile")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

// Preference represents a user preference with confidence score.
// Confidence increases with each mention (0.0 to 1.0).
type Preference struct {
	ID         int
	Key        string
	Value      string
	Category   string
	Confidence float64
	UpdatedAt  time.Time
}

// SetPreference stores or updates a user preference.
// Confidence increases by 0.1 on each update (max 1.0).
func (d *Database) SetPreference(key, value, category string) error {
	_, err := d.db.Exec(`
		INSERT INTO preferences (key, value, category, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET value = ?, updated_at = ?, confidence = MIN(confidence + 0.1, 1.0)
	`, key, value, category, time.Now().Format(time.RFC3339), value, time.Now().Format(time.RFC3339))
	return err
}

// GetPreferences returns all preferences for a specific category,
// sorted by confidence (highest first).
func (d *Database) GetPreferences(category string) ([]Preference, error) {
	return d.queryPreferences("WHERE category = ?", category)
}

// GetAllPreferences returns all preferences sorted by confidence.
func (d *Database) GetAllPreferences() ([]Preference, error) {
	return d.queryPreferences("", nil)
}

// queryPreferences is a helper to query preferences with optional filter.
func (d *Database) queryPreferences(where string, args ...any) ([]Preference, error) {
	query := "SELECT id, key, value, category, confidence, updated_at FROM preferences"
	if where != "" {
		query += " " + where
	}
	query += " ORDER BY confidence DESC"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Preference
	for rows.Next() {
		var p Preference
		var updatedAt string
		if err := rows.Scan(&p.ID, &p.Key, &p.Value, &p.Category, &p.Confidence, &updatedAt); err != nil {
			return nil, err
		}
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		result = append(result, p)
	}
	return result, nil
}

// IncrementTopic increments the mention count for a topic.
// Creates the topic if it doesn't exist.
func (d *Database) IncrementTopic(name string) error {
	_, err := d.db.Exec(`
		INSERT INTO topics (name, last_mentioned)
		VALUES (?, ?)
		ON CONFLICT(name) DO UPDATE SET 
			mentions = mentions + 1, 
			last_mentioned = ?
	`, name, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	return err
}

// Topic represents a conversation topic with mention count.
type Topic struct {
	Name     string
	Mentions int
}

// GetTopTopics returns the most mentioned topics.
func (d *Database) GetTopTopics(limit int) ([]Topic, error) {
	rows, err := d.db.Query(`
		SELECT name, mentions FROM topics 
		ORDER BY mentions DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Topic
	for rows.Next() {
		var t Topic
		if err := rows.Scan(&t.Name, &t.Mentions); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

// Close releases the database connection.
func (d *Database) Close() error {
	return d.db.Close()
}

// GetUserSummary returns a human-readable summary of user data.
// Used to include in LLM context.
func (d *Database) GetUserSummary() string {
	profile, _ := d.GetAllProfile()
	prefs, _ := d.GetAllPreferences()
	topics, _ := d.GetTopTopics(5)

	var summary string

	if len(profile) > 0 {
		summary += "User Profile:\n"
		for k, v := range profile {
			summary += fmt.Sprintf("  - %s: %s\n", k, v)
		}
	}

	if len(prefs) > 0 {
		summary += "\nKnown Preferences:\n"
		for _, p := range prefs {
			summary += fmt.Sprintf("  - %s: %s (%.0f%%)\n", p.Key, p.Value, p.Confidence*100)
		}
	}

	if len(topics) > 0 {
		summary += "\nRecent Topics:\n"
		for _, t := range topics {
			summary += fmt.Sprintf("  - %s (%d)\n", t.Name, t.Mentions)
		}
	}

	return summary
}
