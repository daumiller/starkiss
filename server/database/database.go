package database

import (
  "os"
  "io"
  "fmt"
  "database/sql"
  _ "modernc.org/sqlite"
)

var Location       = "./starkiss.db"
var ErrNotFound    = fmt.Errorf("not found")
var ErrQueryFailed = fmt.Errorf("database query failed")

// ============================================================================

// Open the database (creating new db, if necessary).
func Open() (db *sql.DB, err error) {
  db, err = sql.Open("sqlite", Location)
  if err == nil {
    // sqlite won't complain about an invalid file until you actually attempt to write to it...
    _, err = db.Exec(`CREATE TABLE is_connection_valid (id INTEGER PRIMARY KEY, name TEXT);`)
    if err != nil { return nil, err }
    _, err = db.Exec(`DROP TABLE is_connection_valid;`)
  }
  return db, nil
}

// Close the database.
func Close(db *sql.DB) (err error) {
  return db.Close()
}

// Copy database to backup file.
func BackupCreate() (err error) {
  location_backup := fmt.Sprintf("%s.bak", Location)
  if (fileExists(Location) == false) { return nil }
  if (fileExists(location_backup) == true) {
    err = os.Remove(location_backup)
    if err != nil { return err }
  }

  return fileCopy(Location, location_backup)
}

// Restore database from backup file.
func BackupRestore() (err error) {
  location_backup := fmt.Sprintf("%s.bak", Location)
  if (fileExists(location_backup) == false) { return ErrNotFound }
  if (fileExists(Location) == true) {
    err = os.Remove(Location)
    if err != nil { return err }
  }

  return fileCopy(location_backup, Location)
}

// Begin a transaction.
func TransactionBegin(db *sql.DB) (err error) {
  _, err = db.Exec(`BEGIN TRANSACTION;`)
  return err
}

// Commit a transaction (that has begun, and not been rolled back).
func TransactionCommit(db *sql.DB) (err error) {
  _, err = db.Exec(`COMMIT;`)
  return err
}

// Rollback a transaction (that has begun, and not been committed).
func TransactionRollback(db *sql.DB) (err error) {
  _, err = db.Exec(`ROLLBACK;`)
  return err
}

// ============================================================================

// Does something exist at this path?
func fileExists(path string) bool {
  _, err := os.Stat(path)
  return err == nil
}

// Copy file from source path to destination path.
func fileCopy(source string, destination string) (err error) {
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
