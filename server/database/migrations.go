package database

import (
  "fmt"
  "strconv"
)

type Migration interface {
  Up() (error)
  Down() (error)
}
var Migrations []Migration = []Migration {
  &Migration0000{},
  &Migration0001{},
}

// ============================================================================

// Get current migration level.
func MigrationGetCurrent() (level uint32) {
  if (fileExists(Location) == false) { return 0 }

  db, err := Open()
  if err != nil { return 0 }
  defer db.Close()

  level_string, err := PropertyRead(db, "migration_level")
  if err != nil { return 0 }

  level64, err := strconv.ParseUint(level_string, 10, 32)
  if err != nil { return 0 }
  return uint32(level64)
}

// Migrate to specified level.
// NOTE: Creates a backup of db before starting, and rolls back if migration fails.
// NOTE: This will overwrite the current backup file.
func MigrateTo(level_target uint32) (err error) {
  level_current := MigrationGetCurrent()
  if level_current == level_target { return nil }

  // if db doesn't exist, create it before attempting backups
  if (fileExists(Location) == false) {
    db, err := Open()
    if err == nil { Close(db) }
  }

  err = BackupCreate()
  if err != nil { return fmt.Errorf("Migration backup failed: %s", err.Error()) }

  if level_target > level_current {
    for index := level_current + 1; index <= level_target; index++ {
      err = Migrations[index].Up()
      if err != nil {
        err_new := BackupRestore()
        if err_new != nil { return fmt.Errorf("Migration and rollback failed: %s; %s", err.Error(), err_new.Error()) }
        return fmt.Errorf("Migration failed (rollback successful): %s", err.Error())
      }
    }
  } else {
    for index := level_current; index > level_target; index-- {
      err = Migrations[index].Down()
      if err != nil {
        err_new := BackupRestore()
        if err_new != nil { return fmt.Errorf("Migration and rollback failed: %s; %s", err.Error(), err_new.Error()) }
        return fmt.Errorf("Migration failed (rollback successful): %s", err.Error())
      }
    }
  }

  err = migrationSetCurrent(level_target)
  if err != nil {
    return fmt.Errorf("Migration succeeded, but setting current migration level failed: %s", err.Error())
  }

  return nil
}

// Migrate to the latest level.
func MigrateToLatest() (err error) {
  level_latest := uint32(len(Migrations) - 1)
  return MigrateTo(level_latest)
}

// Set current migration level.
func migrationSetCurrent(level uint32) (err error) {
  db, err := Open()
  if err != nil { return err }
  defer db.Close()

  err = PropertyUpsert(db, "migration_level", strconv.FormatUint(uint64(level), 10))
  if err != nil { return err }
  return nil
}
