package main

import (
  "os"
  "fmt"
  "strconv"
  "github.com/chzyer/readline"
  "github.com/daumiller/starkiss/library"
)

// Perform database migration, either:
// 1) if server started with "migration" command line argument, may specify level to migrate to (or "latest")
// 2) automatically during server startup to "latest"
// Returns 0 on success, -1 otherwise.
func startupMigration(target string) (exit_code int) {
  // Migrate database.
  var err error = nil
  if target == "latest" {
    err = library.MigrateToLatest()
  } else {
    level := uint64(0)
    level, err = strconv.ParseUint(target, 10, 32)
    if err != nil { fmt.Printf("Invalid migration level: \"%s\"\n", target); return -1 }
    err = library.MigrateTo(uint32(level))
  }
  if err != nil {
    fmt.Printf("Migration failed: \"%s\"\n", err.Error())
    return -1
  }

  fmt.Printf("Migrated database to \"%s\"...\n", target)
  return 0
}

// Read properties from database.
func startupProperties() {
  var err error
  JWT_KEY, err = library.JwtKeyGet()
  if err != nil { fmt.Printf("JWT Error: %s\n", err.Error()) ; os.Exit(-1) }

  // validate media_path
  for library.MediaPathValid() == false {
    media_path, _ := library.MediaPathGet()
    if media_path != "" {
      fmt.Printf("Existing media library is set to \"%s\".\n", media_path)
      fmt.Printf("Currently unable to read from this location.\n")
      fmt.Printf("Use the \"edit-library-path\" command line argument to set a different location, if your library has moved.\n")
      fmt.Printf("Use the \"reset-library\" command line argument to start over with an empty library.\n")
      os.Exit(-1)
    }
    fmt.Printf("Media library not found.\n")
    fmt.Printf("Enter the path you'd like to store your library at.\n")

    // if no path set, prompt to set one
    line_reader, err := readline.New("path: ")
    if err != nil { fmt.Printf("Error reading media path: \"%s\"\n", err.Error()); os.Exit(-1) }
    defer line_reader.Close()
    line, err := line_reader.Readline()
    if err != nil { fmt.Printf("Error reading media path: \"%s\"\n", err.Error()); os.Exit(-1) }
    err = library.MediaPathSet(line)
    if err != nil { fmt.Printf("Error setting media path: \"%s\"\n", err.Error()); os.Exit(-1) }
  }
}
