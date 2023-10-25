package database

import (
  "fmt"
  "database/sql"
  _ "modernc.org/sqlite"
)

// Read a key/value from properties.
func PropertyRead(db *sql.DB, key string) (value string, err error) {
  row := db.QueryRow(`SELECT value FROM properties WHERE key = ?;`, key)
  err = row.Scan(&value)
  if err == sql.ErrNoRows { return "", ErrNotFound }
  if err != nil { return "", err }
  return value, nil
}

// Insert/Update a key/value in properties.
func PropertyUpsert(db *sql.DB, key string, value string) (err error) {
  result, err := db.Exec(`INSERT INTO properties (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?;`, key, value, value)
  if err != nil { return err }
  affected, err := result.RowsAffected()
  if err != nil { return err }
  if affected != 1 { return fmt.Errorf("Upserting property failed") }
  return nil
}

// Delete a key/value from properties.
func PropertyDelete(db *sql.DB, key string) (err error) {
  result, err := db.Exec(`DELETE FROM properties WHERE key = ?;`, key)
  if err != nil { return err }
  affected, err := result.RowsAffected()
  if err != nil { return err }
  if affected != 1 { return ErrNotFound }
  return nil
}

// Read all key/values from properties.
func PropertyList(db *sql.DB) (properties map[string]string, err error) {
  properties = map[string]string {}
  rows, err := db.Query(`SELECT key, value FROM properties;`)
  if err != nil { return nil, err }
  defer rows.Close()
  for rows.Next() {
    var key string
    var value string
    err = rows.Scan(&key, &value)
    if err != nil { return nil, err }
    properties[key] = value
  }
  return properties, nil
}
