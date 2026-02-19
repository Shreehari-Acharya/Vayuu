package memory

import (
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

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

type ProfileEntry struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

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

func (d *Database) SetProfile(key, value string) error {
	_, err := d.db.Exec(`
		INSERT INTO user_profile (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = ?
	`, key, value, time.Now().Format(time.RFC3339), value, time.Now().Format(time.RFC3339))
	return err
}

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

type Preference struct {
	ID         int
	Key        string
	Value      string
	Category   string
	Confidence float64
	UpdatedAt  time.Time
}

func (d *Database) SetPreference(key, value, category string) error {
	_, err := d.db.Exec(`
		INSERT INTO preferences (key, value, category, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET value = ?, updated_at = ?, confidence = MIN(confidence + 0.1, 1.0)
	`, key, value, category, time.Now().Format(time.RFC3339), value, time.Now().Format(time.RFC3339))
	return err
}

func (d *Database) GetPreferences(category string) ([]Preference, error) {
	rows, err := d.db.Query(`
		SELECT id, key, value, category, confidence, updated_at 
		FROM preferences 
		WHERE category = ?
		ORDER BY confidence DESC
	`, category)
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

func (d *Database) GetAllPreferences() ([]Preference, error) {
	rows, err := d.db.Query(`
		SELECT id, key, value, category, confidence, updated_at 
		FROM preferences 
		ORDER BY confidence DESC
	`)
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

func (d *Database) GetTopTopics(limit int) ([]struct {
	Name     string
	Mentions int
}, error) {
	rows, err := d.db.Query(`
		SELECT name, mentions FROM topics 
		ORDER BY mentions DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []struct {
		Name     string
		Mentions int
	}
	for rows.Next() {
		var t struct {
			Name     string
			Mentions int
		}
		if err := rows.Scan(&t.Name, &t.Mentions); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

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
