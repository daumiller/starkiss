package library

import (
  "fmt"
  "regexp"
  "strings"
  "database/sql"
)

var db *sql.DB = nil
func SetDatabase(database *sql.DB) { db = database }

// ============================================================================
// General Utilities

// characters not allowed in Category.Name, or Metadata.NameSort fields (characters disallowed by exfat, plus some extras)
var InvalidSortCharacters *regexp.Regexp = regexp.MustCompile(`[<>:"/\\|?*$!%\` + "`" + `~]`)

func SortName(name string) string {
  // lowercase string
  // remove all characters disallowed by exfat (plus a couple)
  sort_name := strings.ToLower(name)
  sort_name = InvalidSortCharacters.ReplaceAllString(sort_name, "")
  return sort_name
}

var ErrInvalidName = fmt.Errorf("invalid name")

func NameValidForDisk(name string) bool {
  // test if this name is valid as a NameSort, CategoryName, or anything that should exist on disk
  // (not the same as comparing to SortName(name), because we don't lowercase)
  sort_name := InvalidSortCharacters.ReplaceAllString(name, "")
  return sort_name == name
}
