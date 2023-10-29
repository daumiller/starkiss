package database

import (
  "fmt"
  "database/sql"
)

var ErrCategoryNotEmpty = fmt.Errorf("category is not empty")

type CategoryType string
const (
  CategoryTypeMovie  CategoryType = "movie"
  CategoryTypeSeries CategoryType = "series"
  CategoryTypeMusic  CategoryType = "music"
)
type Category struct {
  Id   string       `json:"id"`
  Type CategoryType `json:"type"`
  Name string       `json:"name"`
}

// ============================================================================

func (cat *Category) Create(db *sql.DB) (err error) { return TableCreate(db, cat) }
func (cat *Category) Delete(db *sql.DB) (err error) { return TableDelete(db, cat) }

func (cat *Category) Rename(db *sql.DB, new_name string) (err error) {
  cat_proposed := Category{ Id:cat.Id, Type:cat.Type, Name: new_name }
  return TableUpdate(db, cat, &cat_proposed)
}

func (cat *Category) SetType(db *sql.DB, new_type CategoryType) (err error) {
  cat_proposed := Category{ Id:cat.Id, Name:cat.Name, Type:new_type }
  return TableUpdate(db, cat, &cat_proposed)
}

func CategoryRead(db *sql.DB, id string) (cat Category, err error) {
  err = TableRead(db, &cat, id)
  return cat, err
}

func CategoryList(db *sql.DB) (cat_list []Category, err error) {
  tables, err := TableWhere(db, &Category{}, "")
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

func (cat *Category) CreateFrom(fields map[string]any) (instance Table, err error) {
  new_instance := Category {}
  err = new_instance.FieldsWrite(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (cat *Category) FieldsRead() (fields map[string]any, err error) {
  fields = make(map[string]any)
  fields["id"]   = cat.Id
  fields["type"] = string(cat.Type)
  fields["name"] = cat.Name
  return fields, nil
}

func (cat *Category) FieldsWrite(fields map[string]any) (err error) {
  cat.Id   = fields["id"].(string)
  cat.Type = CategoryType(fields["type"].(string))
  cat.Name = fields["name"].(string)
  return nil
}

func (cat_a *Category) FieldsDifference(other Table) (diff map[string]any, err error) {
  diff = make(map[string]any)
  cat_b, b_is_cat := other.(*Category)
  if b_is_cat == false { return diff, ErrInvalidType }

  if cat_a.Id   != cat_b.Id   { diff["id"]   = cat_b.Id           }
  if cat_a.Type != cat_b.Type { diff["type"] = string(cat_b.Type) }
  if cat_a.Name != cat_b.Name { diff["name"] = cat_b.Name         }

  return diff, nil
}

func (cat *Category) ValidCreate(db *sql.DB) (valid bool, err error) {
  // ensure a valid type
  switch cat.Type {
    case CategoryTypeMovie, CategoryTypeSeries, CategoryTypeMusic: break
    default: return false, nil
  }
  return true, nil
}

func (cat_a *Category) ValidUpdate(db *sql.DB, other Table) (valid bool, err error) {
  cat_b, b_is_cat := other.(*Category)
  if b_is_cat == false { return false, ErrInvalidType }

  // only need to check for type changes
  if cat_a.Type == cat_b.Type { return true, nil }
  // first, ensure a valid type
  switch cat_b.Type {
    case CategoryTypeMovie, CategoryTypeSeries, CategoryTypeMusic: break
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

/*
func (cat *Category) Create(db *sql.DB) (err error) {
  if cat.Id == "" { cat.Id = uuid.NewString() }
  result, err := db.Exec(`INSERT INTO categories (id, name, type) VALUES (?, ?, ?);`, cat.Id, cat.Name, string(cat.Type))
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  return nil
}

func (cat *Category) Rename(db *sql.DB, new_name string) (err error) {
  result, err := db.Exec(`UPDATE categories SET name = ? WHERE id = ?;`, new_name, cat.Id)
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  cat.Name = new_name
  return nil
}

func (cat *Category) SetType(db *sql.DB, cat_type CategoryType) (err error) {
  empty, err := categoryIsEmpty(db, cat.Id)
  if err != nil { return ErrQueryFailed }
  if (empty == false) { return ErrCategoryNotEmpty }

  result, err := db.Exec(`UPDATE categories SET type = ? WHERE id = ?;`, string(cat_type), cat.Id)
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  cat.Type = cat_type
  return nil
}

func (cat *Category) Delete(db *sql.DB) (err error) {
  empty, err := categoryIsEmpty(db, cat.Id)
  if err != nil { return ErrQueryFailed }
  if (empty == false) { return ErrCategoryNotEmpty }

  result, err := db.Exec(`DELETE FROM categories WHERE id = ?;`, cat.Id)
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  return nil
}

func CategoryCreate(db *sql.DB, name string, cat_type CategoryType) (id string, err error) {
  cat := Category{ Name: name, Type: cat_type }
  err = cat.Create(db)
  return cat.Id, err
}
func CategoryRename(db *sql.DB, id string, new_name string) (err error) {
  cat := Category{ Id: id }
  return cat.Rename(db, new_name)
}
func CategorySetType(db *sql.DB, id string, cat_type CategoryType) (err error) {
  cat := Category{ Id: id }
  return cat.SetType(db, cat_type)
}
func CategoryDelete(db *sql.DB, id string) (err error) {
  cat := Category{ Id: id }
  return cat.Delete(db)
}

func CategoryList(db *sql.DB) (categories []Category, err error) {
  rows, err := db.Query(`SELECT id, type, name FROM categories ORDER BY name ASC;`)
  if err != nil { return nil, err }
  defer rows.Close()

  for rows.Next() {
    var category Category
    err = rows.Scan(&category.Id, &category.Type, &category.Name)
    if err != nil { return nil, err }
    categories = append(categories, category)
  }

  return categories, nil
}
*/

func categoryIsEmpty(db *sql.DB, id string) (empty bool, err error) {
  var dummy string

  row := db.QueryRow(`SELECT id FROM metadata WHERE category_id = ? LIMIT 1;`, id)
  err = row.Scan(&dummy)
  any_rows := (err != sql.ErrNoRows)
  if any_rows { return false, nil }

  return true, nil
}
