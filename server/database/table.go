package database

import (
  "fmt"
  "strings"
  "database/sql"
  "github.com/google/uuid"
)

var ErrInvalidType       = fmt.Errorf("invalid type")
var ErrValidationFailed  = fmt.Errorf("validation failed")

type Table interface {
  TableName() string                                              // name of the table
  GetId() string                                                  // get the id
  CreateFrom(fields map[string]any) (instance Table, err error)   // create from field:values

  FieldsRead() (fields map[string]any, err error)                 // read field:values from struct
  FieldsWrite(fields map[string]any) (err error)                  // write field:values to struct
  FieldsDifference(other Table) (diff map[string]any, err error)  // compare field:values

  ValidCreate(db *sql.DB) (valid bool, err error)                 // is this a valid create?
  ValidUpdate(db *sql.DB, other Table) (valid bool, err error)    // is this a valid update?
  ValidDelete(db *sql.DB) (valid bool, err error)                 // is this a valid delete?
}

// ============================================================================

func TableCreate(db *sql.DB, table Table) (err error) {
  valid, err := table.ValidCreate(db)
  if err != nil { return err }
  if valid == false { return ErrValidationFailed }

  fields, err := table.FieldsRead()
  if err != nil { return err }

  id, id_present := fields["id"]
  if (id_present == false) || (id == "") { fields["id"] = uuid.NewString() }

  columns := []string {}
  params  := []string {}
  values  := []any {}
  for column, value := range fields {
    columns = append(columns, column)
    params  = append(params, "?")
    values  = append(values, value)
  }

  query_string := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s);`, table.TableName(), strings.Join(columns, ", "), strings.Join(params, ", "))
  result, err := db.Exec(query_string, values...)
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  return nil
}

func TableDelete(db *sql.DB, table Table) (err error) {
  valid, err := table.ValidDelete(db)
  if err != nil { return err }
  if valid == false { return ErrValidationFailed }

  query_string := fmt.Sprintf(`DELETE FROM %s WHERE id = ?;`, table.TableName())
  result, err := db.Exec(query_string, table.GetId())
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  return nil
}

func TableUpdate(db *sql.DB, current Table, proposed Table) (err error) {
  valid, err := current.ValidUpdate(db, proposed)
  if err != nil { return err }
  if valid == false { return ErrValidationFailed }

  difference, err := current.FieldsDifference(proposed)
  if err != nil { return err }

  delete(difference, "id")
  if len(difference) == 0 { return nil }

  columns := []string {}
  values  := []any {}
  for column, value := range difference {
    columns = append(columns, fmt.Sprintf("%s = ?", column))
    values  = append(values, value)
  }
  values = append(values, current.GetId())

  query_string := fmt.Sprintf(`UPDATE %s SET %s WHERE id = ?;`, current.TableName(), strings.Join(columns, ", "))
  result, err := db.Exec(query_string, values...)
  if err != nil { return err }
  count, err := result.RowsAffected()
  if err != nil { return err }
  if count != 1 { return ErrQueryFailed }
  return nil
}

func TableRead(db *sql.DB, table Table, id string) (err error) {
  fields, err := table.FieldsRead()
  if err != nil { return err }

  columns := []string {}
  values  := []any {}
  for column, value := range fields {
    columns = append(columns, column)
    values  = append(values, value)
  }
  values = append(values, table.GetId())

  query_string := fmt.Sprintf(`SELECT %s FROM %s WHERE id = ?;`, strings.Join(columns, ", "), table.TableName())
  result := db.QueryRow(query_string, values...)
  err = result.Scan(values...)
  if err == sql.ErrNoRows { return ErrNotFound }
  if err != nil { return err }

  return table.FieldsWrite(fields)
}

func TableWhere(db *sql.DB, table Table, where_string string, where_values ...any) (results []Table, err error) {
  fields, err := table.FieldsRead()
  if err != nil { return results, err }

  columns := []string {}
  values  := []any {}
  for column, value := range fields {
    columns = append(columns, column)
    values  = append(values, value)
  }
  values = append(values, where_values...)

  if where_string != "" { where_string = fmt.Sprintf(`WHERE %s`, where_string) }
  query_string := fmt.Sprintf(`SELECT %s FROM %s %s ;`, strings.Join(columns, ", "), table.TableName(), where_string)
  rows, err := db.Query(query_string, where_values...)
  if err != nil { return results, err }
  defer rows.Close()

  for rows.Next() {
    err = rows.Scan(values...)
    if err != nil { return results, err }
    this_result, err := table.CreateFrom(fields)
    if err != nil { return results, err }
    results = append(results, this_result)
  }

  return results, nil
}
