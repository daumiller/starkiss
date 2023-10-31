package library

import (
  "os"
  "fmt"
  "path/filepath"
  "github.com/daumiller/starkiss/database"
)

var ErrInvalidMediaType = fmt.Errorf("invalid media type")

func CategoryRead(id string) (*database.Category, error) {
  return database.CategoryRead(db, id)
}

func CategoryList() ([]database.Category, error) {
  return database.CategoryList(db)
}

func CategoryDiskPath(cat *database.Category) (string, error) {
  media_path, err := MediaPathGet()
  if err != nil { return "", err }
  return filepath.Join(media_path, cat.Name), nil
}

func CategoryCreate(name string, media_type database.CategoryMediaType) (*database.Category, error) {
  // validate inputs
  if !NameValidForDisk(name) { return nil, ErrInvalidName }
  if !database.CategoryMediaTypeValid(media_type) { return nil, ErrInvalidMediaType }

  // in DB, verify category doesn't already exist
  exists, err := database.CategoryNameExists(db, name)
  if err != nil { return nil, err }
  if exists { return nil, fmt.Errorf("category named \"%s\" already exists in database", name) }

  // on FS, verify category doesn't already exist
  media_path, err := MediaPathGet()
  if err != nil { return nil, err }
  disk_path := filepath.Join(media_path, name)
  if _, err := os.Stat(disk_path); err == nil {
    return nil, fmt.Errorf("category named \"%s\" already exists on disk", name)
  }

  // create category on FS
  err = os.Mkdir(disk_path, 0770)
  if err != nil { return nil, fmt.Errorf("error creating category \"%s\" on disk: %s", name, err.Error()) }

  // create category in DB
  cat := database.Category { Name:name, MediaType:media_type }
  err = database.CategoryCreate(db, &cat)
  if err != nil { return nil, err }
  return &cat, nil
}

func CategoryDelete(cat *database.Category) error {
  media_path, err := MediaPathGet()
  if err != nil { return err }

  // get children
  children, err := database.MetadataWhere(db, `parent_id = ?`, cat.Id)
  if err != nil { return err }

  // verify all children can be unparented (moving to root won't result in name collisions)
  for _, child := range children {
    if metadataCanMoveFilesToPath(&child, media_path) == false {
      return fmt.Errorf("cannot delete category \"%s\": child \"%s\" files cannot be moved to media root", cat.Name, child.NameDisplay)
    }
  }

  // move all children fs objects to root, unparenting DB records as we go
  for _, child := range children {
    err := metadataMoveFilesToPath(&child, media_path)
    if err != nil { return fmt.Errorf("category deletion stopped, error moving files for child \"%s\" to media root: %s", child.NameDisplay, err.Error()) }
    child.ParentId = ""
    err = database.MetadataPatch(db, &child, map[string]any { "parent_id":"" })
    if err != nil { return fmt.Errorf("category deletion stopped, error removing child \"%s\" from database parent: %s", child.NameDisplay, err.Error()) }
  }

  // delete category in DB
  err = database.CategoryDelete(db, cat)
  if err != nil { return fmt.Errorf("error deleting category \"%s\" in database: %s", cat.Name, err.Error()) }

  // delete category on FS
  disk_path := filepath.Join(media_path, cat.Name)
  err = os.RemoveAll(disk_path)
  if err != nil { return fmt.Errorf("error removing category \"%s\" on disk: %s", cat.Name, err.Error()) }

  return nil
}

func CategoryUpdate(cat *database.Category, name string, media_type string) error {
  media_path, err := MediaPathGet()
  if err != nil { return err }

  // valid inputs?
  media_type_enum := database.CategoryMediaType(media_type)
  if !database.CategoryMediaTypeValid(media_type_enum) { return ErrInvalidMediaType }
  if !NameValidForDisk(name) { return ErrInvalidName }

  // if changing type, verify empty
  if media_type_enum != cat.MediaType {
    // verify no children exist
    empty, err := database.CategoryIsEmpty(db, cat.Id)
    if err != nil { return err }
    if empty == false { return fmt.Errorf("cannot change media type of category \"%s\": children exist", cat.Name) }
  }

  // if changing name, verify no collisions; then move on disk
  if name != cat.Name {
    exists, err := database.CategoryNameExists(db, name)
    if err != nil { return err }
    if exists { return fmt.Errorf("cannot rename category \"%s\" to \"%s\": name already exists", cat.Name, name) }
    disk_path := filepath.Join(media_path, name)
    if _, err := os.Stat(disk_path); err == nil {
      return fmt.Errorf("cannot rename category \"%s\" to \"%s\": name already exists on disk", cat.Name, name)
    }
    old_path := filepath.Join(media_path, cat.Name)
    err = os.Rename(old_path, disk_path)
    if err != nil { return fmt.Errorf("error renaming category \"%s\" on disk: %s", cat.Name, err.Error()) }
  }

  // update record
  return database.CategoryUpdate(db, cat, name, media_type_enum)
}
