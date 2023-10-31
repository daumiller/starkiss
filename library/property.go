package library

import (
	"os"
  "fmt"
  "io/fs"
  "errors"
	"path/filepath"
	"github.com/daumiller/starkiss/database"
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
  full_list, err := database.PropertyList(db)
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
  return database.PropertyRead(db, key)
}

func PropertySet(key string, value string) error {
  if excluded_properties[key] == true { return ErrInvalidProperty }
  return database.PropertyUpsert(db, key, value)
}

func PropertyDelete(key string) error {
  if excluded_properties[key] == true { return ErrInvalidProperty }
  return database.PropertyDelete(db, key)
}

// ============================================================================
// Special Properties

var media_path_cache string = ""

func MediaPathGet() (string, error) {
  if media_path_cache != "" { return media_path_cache, nil }

  var err error
  media_path_cache, err = database.PropertyRead(db, "media_path")
  return media_path_cache, err
}

func MediaPathValid() bool {
  media_path, err := MediaPathGet()
  if err != nil { return false }
  if media_path == "" { return false }
  _, err = os.Stat(media_path)
  if err != nil { return false }
  return true
}

// Set the media path.
// This involves moving the media directory, and may take a while.
// This call should not be allowed from a web interface.
func MediaPathSet(new_path string) error {
  if !NameValidForDisk(filepath.Base(new_path)) { return ErrInvalidName }

  new_library := false
  curr_path, err := MediaPathGet()
  if err != nil { return err }
  if curr_path == "" {
    new_library = true
  } else {
    _, err := os.Stat(curr_path)
    if err != nil { return fmt.Errorf("cannot move media library, existing path \"%s\" not found: %s", curr_path, err.Error()) }
  }

  _, err = os.Stat(new_path)
  if err == nil { return fmt.Errorf("cannot move media to \"%s\": path already exists", new_path) }
  if !errors.Is(err, fs.ErrNotExist) { return fmt.Errorf("cannot move media to \"%s\": %s", new_path, err.Error()) }

  if new_library {
    err = os.MkdirAll(new_path, 0770)
    if err != nil { return fmt.Errorf("cannot create media library at \"%s\": %s", new_path, err.Error()) }
  } else {
    err = os.Rename(curr_path, new_path)
    if err != nil { return fmt.Errorf("cannot move media library from \"%s\" to \"%s\": %s", curr_path, new_path, err.Error()) }
  }

  err = database.PropertyUpsert(db, "media_path", new_path)
  if err != nil { return fmt.Errorf("cannot update database with new media path \"%s\": %s", new_path, err.Error()) }

  media_path_cache = new_path
  return nil
}
