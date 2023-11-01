package library

import (
  "os"
  "fmt"
  "strconv"
  "path/filepath"
  "database/sql"
)

var ErrInvalidProperty = fmt.Errorf("invalid property")

var excluded_properties = map[string]bool {
  "migration_level" : true,
  "media_path"      : true,
  "jwt_key"         : true,
}

// ============================================================================
// General Properties

// List all properties that can be freely edited.
func PropertyList() (map[string]string, error) {
  full_list, err := dbPropertyList()
  if err != nil { return nil, err }

  for key := range full_list {
    if excluded_properties[key] == true {
      delete(full_list, key)
    }
  }
  return full_list, nil
}

func PropertyGet(key string) (string, error) {
  if excluded_properties[key] == true { return "", ErrInvalidProperty }
  return dbPropertyRead(key)
}

func PropertySet(key string, value string) error {
  if excluded_properties[key] == true { return ErrInvalidProperty }
  return dbPropertyUpsert(key, value)
}

func PropertyDelete(key string) error {
  if excluded_properties[key] == true { return ErrInvalidProperty }
  return dbPropertyDelete(key)
}

// ============================================================================
// Special Properties

func MediaPathGet() (string, error) {
  if mediaPath != "" { return mediaPath, nil }

  var err error
  mediaPath, err = dbPropertyRead("media_path")
  return mediaPath, err
}

func MediaPathValid() bool {
  media_path, err := MediaPathGet()
  if err != nil { return false }
  if media_path == "" { return false }
  if pathIsDirectory(media_path) == false { return false }
  return true
}

// Set the media path.
// This involves moving the media directory, and may take a while.
// This call should not be allowed from a web interface.
func MediaPathSet(new_path string) error {
  if !nameValidForDisk(filepath.Base(new_path)) { return ErrInvalidName }

  new_library := false
  curr_path, err := MediaPathGet()
  if err != nil { return err }
  if curr_path == "" {
    new_library = true
  } else {
    if !pathExists(curr_path) { return fmt.Errorf("cannot move media library, existing path \"%s\" not found: %s", curr_path, err.Error()) }
  }

  if pathExists(new_path) { return fmt.Errorf("cannot move media to \"%s\": path already exists", new_path) }

  if new_library {
    err = os.MkdirAll(new_path, 0770)
    if err != nil { return fmt.Errorf("cannot create media library at \"%s\": %s", new_path, err.Error()) }
  } else {
    err = os.Rename(curr_path, new_path)
    if err != nil { return fmt.Errorf("cannot move media library from \"%s\" to \"%s\": %s", curr_path, new_path, err.Error()) }
  }

  err = dbPropertyUpsert("media_path", new_path)
  if err != nil { return fmt.Errorf("cannot update database with new media path \"%s\": %s", new_path, err.Error()) }

  mediaPath = new_path
  return nil
}


// Get current migration level.
func MigrationLevelGet() (level uint32) {
  level_string, err := dbPropertyRead("migration_level")
  if err != nil { return 0 }

  level64, err := strconv.ParseUint(level_string, 10, 32)
  if err != nil { return 0 }
  return uint32(level64)
}

// Set current migration level.
func migrationLevelSet(level uint32) (err error) {
  err = dbPropertyUpsert("migration_level", strconv.FormatUint(uint64(level), 10))
  if err != nil { return err }
  return nil
}


// ============================================================================
// database interface

// Read a key/value from properties.
func dbPropertyRead(key string) (value string, err error) {
  row := dbHandle.QueryRow(`SELECT value FROM properties WHERE key = ?;`, key)
  err = row.Scan(&value)
  if err == sql.ErrNoRows { return "", ErrNotFound }
  if err != nil { return "", ErrQueryFailed }
  return value, nil
}

// Insert/Update a key/value in properties.
func dbPropertyUpsert(key string, value string) (err error) {
  result, err := dbHandle.Exec(`INSERT INTO properties (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?;`, key, value, value)
  if err != nil { return ErrQueryFailed }
  affected, err := result.RowsAffected()
  if err != nil { return ErrQueryFailed }
  if affected != 1 { return ErrQueryFailed}
  return nil
}

// Delete a key/value from properties.
func dbPropertyDelete(key string) (err error) {
  result, err := dbHandle.Exec(`DELETE FROM properties WHERE key = ?;`, key)
  if err != nil { return ErrQueryFailed }
  affected, err := result.RowsAffected()
  if err != nil { return ErrQueryFailed }
  if affected != 1 { return ErrNotFound }
  return nil
}

// Read all key/values from properties.
func dbPropertyList() (properties map[string]string, err error) {
  properties = map[string]string {}
  rows, err := dbHandle.Query(`SELECT key, value FROM properties;`)
  if err != nil { return nil, ErrQueryFailed }
  defer rows.Close()
  for rows.Next() {
    var key string
    var value string
    err = rows.Scan(&key, &value)
    if err != nil { return nil, ErrQueryFailed }
    properties[key] = value
  }
  return properties, nil
}
