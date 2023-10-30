package database

import (
  "fmt"
  "database/sql"
)

var ErrCategoryNotEmpty = fmt.Errorf("category is not empty")

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
}

// ============================================================================

func (cat *Category) Copy() (*Category) {
  copy := Category {}
  copy.Id        = cat.Id
  copy.MediaType = cat.MediaType
  copy.Name      = cat.Name
  return &copy
}

func (cat *Category) Create(db *sql.DB) (err error) { return TableCreate(db, cat) }
func (cat *Category) Delete(db *sql.DB) (err error) { return TableDelete(db, cat) }

func (cat *Category) Update(db *sql.DB, new_name string, new_type CategoryMediaType) (err error) {
  return TablePatch(db, cat, map[string]any { "name":new_name, "media_type":string(new_type) })
}

func CategoryRead(db *sql.DB, id string) (cat *Category, err error) {
  cat = &Category{}
  err = TableRead(db, cat, id)
  return cat, err
}

func CategoryList(db *sql.DB) (cat_list []Category, err error) {
  tables, err := TableWhere(db, &Category{}, "(id <> '') ORDER BY name ASC")
  if err != nil { return nil, err }
  cat_list = make([]Category, len(tables))
  for index, table := range tables { cat_list[index] = *(table.(*Category)) }
  return cat_list, nil
}

// ============================================================================
// Table interface

func (cat *Category) TableName() string { return "categories"}
func (cat *Category) GetId() string { return cat.Id }
func (cat *Category) SetId(id string) { cat.Id = id }
func (cat *Category) CopyRecord() (Table, error) {
  return cat.Copy(), nil
}

func (cat *Category) CreateFrom(fields map[string]any) (instance Table, err error) {
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
  return fields, nil
}

func (cat *Category) FieldsReplace(fields map[string]any) (err error) {
  cat.Id        = fields["id"].(string)
  cat.MediaType = CategoryMediaType(fields["media_type"].(string))
  cat.Name      = fields["name"].(string)
  return nil
}

func (cat *Category) FieldsPatch(fields map[string]any) (err error) {
  if id,         ok := fields["id"]         ; ok { cat.Id        = id.(string)                            }
  if media_type, ok := fields["media_type"] ; ok { cat.MediaType = CategoryMediaType(media_type.(string)) }
  if name,       ok := fields["name"]       ; ok { cat.Name      = name.(string)                          }
  return nil
}

func (cat_a *Category) FieldsDifference(other Table) (diff map[string]any, err error) {
  diff = make(map[string]any)
  cat_b, b_is_cat := other.(*Category)
  if b_is_cat == false { return diff, ErrInvalidType }

  if cat_a.Id        != cat_b.Id        { diff["id"]         = cat_b.Id                }
  if cat_a.MediaType != cat_b.MediaType { diff["media_type"] = string(cat_b.MediaType) }
  if cat_a.Name      != cat_b.Name      { diff["name"]       = cat_b.Name              }

  return diff, nil
}

func (cat *Category) ValidCreate(db *sql.DB) (valid bool, err error) {
  // ensure a valid type
  switch cat.MediaType {
    case CategoryMediaTypeMovie, CategoryMediaTypeSeries, CategoryMediaTypeMusic: break
    default: return false, nil
  }
  return true, nil
}

func (cat_a *Category) ValidUpdate(db *sql.DB, other Table) (valid bool, err error) {
  cat_b, b_is_cat := other.(*Category)
  if b_is_cat == false { return false, ErrInvalidType }

  // only need to check for type changes
  if cat_a.MediaType == cat_b.MediaType { return true, nil }
  // first, ensure a valid type
  switch cat_b.MediaType {
    case CategoryMediaTypeMovie, CategoryMediaTypeSeries, CategoryMediaTypeMusic: break
    default: return false, nil
  }
  // then, only allow changing type if no items already assigned
  empty, err := categoryIsEmpty(db, cat_a.Id)
  if err != nil { return false, ErrQueryFailed }
  if (empty == false) { return false, nil }

  return true, nil
}

func (cat *Category) ValidDelete(db *sql.DB) (valid bool, err error) {
  // ensure category is empty before deleting
  empty, err := categoryIsEmpty(db, cat.Id)
  if err != nil { return false, ErrQueryFailed }
  if (empty == false) { return false, nil }
  return true, nil
}

// ============================================================================

func categoryIsEmpty(db *sql.DB, id string) (empty bool, err error) {
  var dummy string

  row := db.QueryRow(`SELECT id FROM metadata WHERE parent_id = ? LIMIT 1;`, id)
  err = row.Scan(&dummy)
  any_rows := (err != sql.ErrNoRows)
  if any_rows { return false, nil }

  return true, nil
}
