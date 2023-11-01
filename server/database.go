package main

import (
  "os"
  "fmt"
  "strconv"
  "crypto/rand"
  "encoding/base64"
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
    err = database.MigrateToLatest()
  } else {
    level := uint64(0)
    level, err = strconv.ParseUint(target, 10, 32)
    if err != nil { fmt.Printf("Invalid migration level: \"%s\"\n", target); return -1 }
    err = database.MigrateTo(uint32(level))
  }
  if err != nil {
    fmt.Printf("Migration failed: \"%s\"\n", err.Error())
    return -1
  }

  fmt.Printf("Migrated database to \"%s\"...\n", target)
  return 0
}

// Read properties from database.
// If not present, create them with defaults.
// If present, use them rather than defaults.
func startupProperties() {
  // Ensure we have a JWT key.
  key_base64, err := database.PropertyRead(DB, "jwtkey")
  if err == nil {
    // key in database (stored base64), decode it
    key_bytes := make([]byte, base64.StdEncoding.DecodedLen(len(key_base64)))
    wrote_length, err := base64.StdEncoding.Decode(key_bytes, []byte(key_base64))
    if err != nil { fmt.Printf("Error decoding jwtkey: \"%s\"\n", err.Error()); os.Exit(-1) }
    JWTKEY = key_bytes[:wrote_length]
  } else {
    // no jwtkey in database, create one
    fmt.Printf("Creating JWT key...\n")
    key_bytes := make([]byte, 32)
    _, err := rand.Read(key_bytes)
    if err != nil { fmt.Printf("Error creating JWT key: \"%s\"\n", err.Error()); os.Exit(-1) }
    key_base64 := base64.StdEncoding.EncodeToString(key_bytes)
    err = database.PropertyUpsert(DB, "jwtkey", key_base64)
    if err != nil { fmt.Printf("Error creating JWT key: \"%s\"\n", err.Error()); os.Exit(-1) }
    JWTKEY = key_bytes
  }

  // cache media path
  media_path, err  := database.PropertyRead(DB, "media_path")
  if err == nil { MEDIA_PATH  = media_path  }
}
