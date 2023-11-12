package library

import (
  "fmt"
)

type Migration interface {
  Up() (error)
  Down() (error)
}
var Migrations []Migration = []Migration {
  &migration0000{},
  &migration0001{},
  &migration0002{},
}

// ============================================================================

// Migrate to specified level.
// NOTE: Creates a backup of db before starting, and rolls back if migration fails.
// NOTE: This will overwrite the current backup file.
func MigrateTo(level_target uint32) (err error) {
  level_current := MigrationLevelGet()
  if level_current == level_target { return nil }

  err = dbBackupCreate()
  if err != nil { return fmt.Errorf("Migration backup failed: %s", err.Error()) }

  if level_target > level_current {
    for index := level_current + 1; index <= level_target; index++ {
      err = Migrations[index].Up()
      if err != nil {
        err_new := dbBackupRestore()
        if err_new != nil { return fmt.Errorf("Migration step %d, and rollback failed: %s; %s", index, err.Error(), err_new.Error()) }
        return fmt.Errorf("Migration step %d failed (rollback successful): %s", index, err.Error())
      }
    }
  } else {
    for index := level_current; index > level_target; index-- {
      err = Migrations[index].Down()
      if err != nil {
        err_new := dbBackupRestore()
        if err_new != nil { return fmt.Errorf("Migration step %d, and rollback failed: %s; %s", index, err.Error(), err_new.Error()) }
        return fmt.Errorf("Migration step %d failed (rollback successful): %s", index, err.Error())
      }
    }
  }

  err = migrationLevelSet(level_target)
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
