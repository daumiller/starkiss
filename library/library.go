package library

import (
  "os"
  "io"
  "fmt"
  "regexp"
  "strings"
  "database/sql"
  _ "modernc.org/sqlite"
)

// ============================================================================
// Common Errors

var ErrNotFound        = fmt.Errorf("not found")
var ErrPathExists      = fmt.Errorf("path already exists")
var ErrQueryFailed     = fmt.Errorf("database query failed")
var ErrInvalidName     = fmt.Errorf("invalid name")
var ErrDbPathNotSet    = fmt.Errorf("database path not set")
var ErrDbNotOpened     = fmt.Errorf("database not opened")
var ErrMediaPathNotSet = fmt.Errorf("media path not set")
var ErrMediaPathNotDir = fmt.Errorf("media path not a valid directory")

// ============================================================================
// Shared Globals

var dbPath    string  = ""
var dbHandle  *sql.DB = nil
var mediaPath string  = ""

// ============================================================================
// Library Lifecycle

func LibraryStartup(database_path string) error {
  dbPath = database_path
  err := dbOpen()
  if err != nil { return err }

  MediaPathGet()
  return nil
}

func LibraryReady() error {
  if dbPath    == ""                     { return ErrDbPathNotSet    }
  if dbHandle  == nil                    { return ErrDbNotOpened     }
  if mediaPath == ""                     { return ErrMediaPathNotSet }
  if pathIsDirectory(mediaPath) == false { return ErrMediaPathNotDir }

  return nil
}

func LibraryShutdown() {
  if dbHandle != nil { dbClose() }
}

// ============================================================================
// Naming Utilities

// characters not allowed in Category.Name, or Metadata.NameSort fields (characters disallowed by exfat, plus some extras)
var InvalidDiskCharacters *regexp.Regexp = regexp.MustCompile(`[<>:"/\\|?*$!%\'` + "`" + `~]`)

// for Metadata, get NameSort from a NameDisplay
func nameGetSortForDisplay(name string) string {
  // lowercase & remove invalid characters
  sort_name := strings.ToLower(name)
  sort_name = InvalidDiskCharacters.ReplaceAllString(sort_name, "")
  sort_name = strings.TrimSpace(sort_name)
  return sort_name
}

// for Categories, test if Name is valid (for use as a directory name)
func nameValidForDisk(name string) bool {
  // test if this name is valid as a NameSort, CategoryName, or anything that should exist on disk
  // (not the same as comparing to SortName(name), because we don't lowercase)
  sort_name := InvalidDiskCharacters.ReplaceAllString(name, "")
  return sort_name == name
}

// ============================================================================
// Path Utilities

// Does something exist at this path?
func pathExists(path string) bool {
  _, err := os.Stat(path)
  return err == nil
}

func pathIsDirectory(path string) bool {
  stat, err := os.Stat(path)
  if err != nil { return false }
  return stat.IsDir()
}

// Copy file from source path to destination path.
func fileCopy(source string, destination string) (err error) {
  if pathExists(source)      == false { return ErrNotFound   }
  if pathExists(destination) == true  { return ErrPathExists }

  source_file, err := os.Open(source)
  if err != nil { return err }
  defer source_file.Close()

  destination_file, err := os.Create(destination)
  if err != nil { return err }
  defer destination_file.Close()

  io.Copy(destination_file, source_file)
  if err != nil { return err }

  return nil
}

// Move path from source to destination (move/rename).
func pathMove(source string, destination string) (err error) {
  if pathExists(source)      == false { return ErrNotFound   }
  if pathExists(destination) == true  { return ErrPathExists }
  return os.Rename(source, destination)
}

// ============================================================================
// Database Utilities

// Open the database (creating new db, if necessary).
func dbOpen() error {
  var err error
  dbHandle, err = sql.Open("sqlite", dbPath)
  if err != nil { return err }

  if err == nil {
    // sqlite won't complain about an invalid file until you actually attempt to write to it...
    _, err = dbHandle.Exec(`CREATE TABLE is_connection_valid (id INTEGER PRIMARY KEY, name TEXT);`)
    if err != nil { dbClose() ; return err }
    _, err = dbHandle.Exec(`DROP TABLE is_connection_valid;`)
    if err != nil { dbClose() ; return err }
  }

  return nil
}

// Close the database.
func dbClose() error {
  return dbHandle.Close()
}

// Copy database to backup file.
func dbBackupCreate() error {
  backup_path := fmt.Sprintf("%s.bak", dbPath)
  if (pathExists(dbPath) == false) { return ErrNotFound }
  if (pathExists(backup_path) == true) {
    err := os.Remove(backup_path)
    if err != nil { return err }
  }

  return fileCopy(dbPath, backup_path)
}

// Restore database from backup file.
func dbBackupRestore() error {
  backup_path := fmt.Sprintf("%s.bak", dbPath)
  if (pathExists(backup_path) == false) { return ErrNotFound }
  if (pathExists(dbPath) == true) {
    err := os.Remove(dbPath)
    if err != nil { return err }
  }

  return fileCopy(backup_path, dbPath)
}

// Begin a transaction.
func dbTransactionBegin() error {
  _, err := dbHandle.Exec(`BEGIN TRANSACTION;`)
  return err
}

// Commit a transaction (that has begun, and not been rolled back).
func dbTransactionCommit() error {
  _, err := dbHandle.Exec(`COMMIT;`)
  return err
}

// Rollback a transaction (that has begun, and not been committed).
func dbTransactionRollback() error {
  _, err := dbHandle.Exec(`ROLLBACK;`)
  return err
}
