package library

import (
  "os"
  "fmt"
  "path/filepath"
)

type CategoryMediaType string
const (
  CategoryMediaTypeMovie  CategoryMediaType = "movie"
  CategoryMediaTypeSeries CategoryMediaType = "series"
  CategoryMediaTypeMusic  CategoryMediaType = "music"
)
type Category struct {
  Id        string            `json:"id"`
  MediaType CategoryMediaType `json:"media_type"`
  Name      string            `json:"name"`
  SortIndex int64             `json:"sort_index"`
}

var ErrInvalidMediaType = fmt.Errorf("invalid media type")

// ============================================================================
// Public Interface

func (cat *Category) Copy() (*Category) {
  copy := Category {}
  copy.Id        = cat.Id
  copy.MediaType = cat.MediaType
  copy.Name      = cat.Name
  copy.SortIndex = cat.SortIndex
  return &copy
}

func (cat *Category) DiskPath() string {
  return filepath.Join(mediaPath, cat.Name)
}

func CategoryList() ([]Category, error) {
  records, err := dbRecordWhere(&Category{}, `(id <> '') ORDER BY sort_index ASC, name ASC`)
  if err != nil { return nil, ErrQueryFailed }
  categories := make([]Category, len(records))
  for index, record := range records { categories[index] = *(record.(*Category)) }
  return categories, nil
}

func CategoryRead(id string) (*Category, error) {
  cat := Category {}
  err := dbRecordRead(&cat, id)
  if err != nil { return nil, err }
  return &cat, nil
}

func CategoryCreate(name string, media_type CategoryMediaType) (*Category, error) {
  // validate inputs
  if !nameValidForDisk(name) { return nil, ErrInvalidName }
  if !categoryMediaTypeValid(media_type) { return nil, ErrInvalidMediaType }

  // in DB, verify category doesn't already exist
  db_exists := CategoryNameExists(name)
  if db_exists { return nil, fmt.Errorf("category named \"%s\" already exists in database", name) }

  // on FS, verify category doesn't already exist
  disk_path := filepath.Join(mediaPath, name)
  if pathExists(disk_path) { return nil, fmt.Errorf("category named \"%s\" already exists on disk", name) }

  // create category on FS
  err := os.Mkdir(disk_path, 0770)
  if err != nil { return nil, fmt.Errorf("error creating category \"%s\" on disk: %s", name, err.Error()) }

  // create category in DB
  cat := Category { Name:name, MediaType:media_type, SortIndex:9999 }
  err = dbRecordCreate(&cat)
  if err != nil { return nil, ErrQueryFailed }
  return &cat, nil
}

func CategoryDelete(cat *Category) error {
  // get children
  children, err := MetadataForParent(cat.Id)
  if err != nil { return err }

  // verify all children can be unparented (moving to root won't result in name collisions)
  for _, child := range children {
    if metadataCanMoveFilesToPath(&child, mediaPath) == false {
      return fmt.Errorf("cannot delete category \"%s\": child \"%s\" files cannot be moved to media root", cat.Name, child.NameDisplay)
    }
  }

  // move all children fs objects to root, unparenting DB records as we go
  for _, child := range children {
    err := metadataMoveFilesToPath(&child, mediaPath)
    if err != nil { return fmt.Errorf("category deletion stopped, error moving files for child \"%s\" to media root: %s", child.NameDisplay, err.Error()) }
    child.ParentId = ""
    err = dbRecordPatch(&child, map[string]any { "parent_id":"" })
    if err != nil { return fmt.Errorf("category deletion stopped, error removing child \"%s\" from database parent: %s", child.NameDisplay, err.Error()) }
  }

  // delete category in DB
  err = dbRecordDelete(cat)
  if err != nil { return fmt.Errorf("error deleting category \"%s\" in database: %s", cat.Name, err.Error()) }

  // delete category on FS
  err = os.RemoveAll(cat.DiskPath())
  if err != nil { return fmt.Errorf("error removing category \"%s\" on disk: %s", cat.Name, err.Error()) }

  return nil
}

func CategoryUpdate(cat *Category, name string, media_type string) error {
  // valid inputs?
  media_type_enum := CategoryMediaType(media_type)
  if !categoryMediaTypeValid(media_type_enum) { return ErrInvalidMediaType }
  if !nameValidForDisk(name) { return ErrInvalidName }

  // if changing type, verify empty
  if media_type_enum != cat.MediaType {
    // verify no children exist
    empty := CategoryIsEmpty(cat.Id)
    if empty == false { return fmt.Errorf("cannot change media type of category \"%s\": children exist", cat.Name) }
  }

  // if changing name, verify no collisions; then move on disk
  if name != cat.Name {
    exists := CategoryNameExists(name)
    if exists { return fmt.Errorf("cannot rename category \"%s\" to \"%s\": name already exists", cat.Name, name) }
    new_disk_path := filepath.Join(mediaPath, name)
    if pathExists(new_disk_path) { return fmt.Errorf("cannot rename category \"%s\" to \"%s\": name already exists on disk", cat.Name, name) }
    err := os.Rename(cat.DiskPath(), new_disk_path)
    if err != nil { return fmt.Errorf("error renaming category \"%s\" on disk: %s", cat.Name, err.Error()) }
  }

  // update record
  err := dbRecordPatch(cat, map[string]any { "name":name, "media_type":string(media_type_enum) })
  if err != nil { return ErrQueryFailed }
  return nil
}

func CategoryReindex(cat *Category, index int64) error {
  err := dbRecordPatch(cat, map[string]any { "sort_index":index })
  if err != nil { return ErrQueryFailed }
  return nil
}

// ============================================================================
// public utilities

func CategoryIdExists(id string) bool {
  dbLock.RLock()
  defer dbLock.RUnlock()
  queryRow := dbHandle.QueryRow(`SELECT id FROM categories WHERE id = ? LIMIT 1;`, id)
  err := queryRow.Scan(&id)
  return (err == nil)
}

func CategoryNameExists(name string) bool {
  dbLock.RLock()
  defer dbLock.RUnlock()
  queryRow := dbHandle.QueryRow(`SELECT id FROM categories WHERE name = ? LIMIT 1;`, name)
  err := queryRow.Scan(&name)
  return (err == nil)
}

func CategoryIsEmpty(id string) bool {
  dbLock.RLock()
  defer dbLock.RUnlock()
  row := dbHandle.QueryRow(`SELECT id FROM metadata WHERE parent_id = ? LIMIT 1;`, id)
  err := row.Scan(&id)
  found := (err == nil)
  return !found
}

// ============================================================================
// private utilities

func categoryMediaTypeValidString(media_type string) bool {
  cat_media_type := CategoryMediaType(media_type)
  return categoryMediaTypeValid(cat_media_type)
}
func categoryMediaTypeValid(media_type CategoryMediaType) bool {
  switch(media_type) {
    case CategoryMediaTypeMovie  : fallthrough
    case CategoryMediaTypeSeries : fallthrough
    case CategoryMediaTypeMusic  : return true
    default: return false
  }
}

// ============================================================================
// dbRecord interface

func (cat *Category) TableName() string { return "categories"}
func (cat *Category) GetId() string { return cat.Id }
func (cat *Category) SetId(id string) { cat.Id = id }
func (cat *Category) RecordCopy() (dbRecord, error) {
  return cat.Copy(), nil
}

func (cat *Category) RecordCreate(fields map[string]any) (instance dbRecord, err error) {
  new_instance := Category {}
  err = new_instance.FieldsReplace(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (cat *Category) FieldsRead() (fields map[string]any, err error) {
  fields = make(map[string]any)
  fields["id"]         = cat.Id
  fields["media_type"] = string(cat.MediaType)
  fields["name"]       = cat.Name
  fields["sort_index"] = cat.SortIndex
  return fields, nil
}

func (cat *Category) FieldsReplace(fields map[string]any) (err error) {
  cat.Id        = fields["id"].(string)
  cat.MediaType = CategoryMediaType(fields["media_type"].(string))
  cat.Name      = fields["name"].(string)
  cat.SortIndex = fields["sort_index"].(int64)
  return nil
}

func (cat *Category) FieldsPatch(fields map[string]any) (err error) {
  if id,         ok := fields["id"]         ; ok { cat.Id        = id.(string)                            }
  if media_type, ok := fields["media_type"] ; ok { cat.MediaType = CategoryMediaType(media_type.(string)) }
  if name,       ok := fields["name"]       ; ok { cat.Name      = name.(string)                          }
  if sort_index, ok := fields["sort_index"] ; ok { cat.SortIndex = sort_index.(int64)                     }
  return nil
}

func (cat_a *Category) FieldsDifference(other dbRecord) (diff map[string]any, err error) {
  diff = make(map[string]any)
  cat_b, b_is_cat := other.(*Category)
  if b_is_cat == false { return diff, ErrInvalidType }

  if cat_a.Id        != cat_b.Id        { diff["id"]         = cat_b.Id                }
  if cat_a.MediaType != cat_b.MediaType { diff["media_type"] = string(cat_b.MediaType) }
  if cat_a.Name      != cat_b.Name      { diff["name"]       = cat_b.Name              }
  if cat_a.SortIndex != cat_b.SortIndex { diff["sort_index"] = cat_b.SortIndex         }

  return diff, nil
}
