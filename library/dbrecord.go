package library

import (
  "fmt"
  "strings"
  "database/sql"
  "github.com/google/uuid"
)

var ErrInvalidType = fmt.Errorf("invalid type")

type dbRecord interface {
  TableName() string                                                 // name of the table
  GetId() string                                                     // get the id
  SetId(id string)                                                   // set the id
  RecordCreate(fields map[string]any) (instance dbRecord, err error) // create record from full set of field:values
  RecordCopy() (instance dbRecord, err error)                        // create record copy

  FieldsRead() (fields map[string]any, err error)                   // read field:values from struct
  FieldsReplace(fields map[string]any) (err error)                  // write full set of field:values to struct
  FieldsPatch(fields map[string]any) (err error)                    // write partial set of field:values to struct
  FieldsDifference(other dbRecord) (diff map[string]any, err error) // compare field:values
}

// ============================================================================

func dbRecordCreate(record dbRecord) (err error) {
  fields, err := record.FieldsRead()
  if err != nil { return err }

  id, id_present := fields["id"]
  if (id_present == false) || (id == "") {
    fields["id"] = uuid.NewString()
    record.SetId(fields["id"].(string))
  }

  columns := []string {}
  params  := []string {}
  values  := []any {}
  for column, value := range fields {
    columns = append(columns, column)
    params  = append(params, "?")
    values  = append(values, value)
  }

  query_string := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s);`, record.TableName(), strings.Join(columns, ", "), strings.Join(params, ", "))
  dbLock.Lock()
  defer dbLock.Unlock()
  result, err := dbHandle.Exec(query_string, values...)
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  return nil
}

func dbRecordDelete(record dbRecord) (err error) {
  dbLock.Lock()
  defer dbLock.Unlock()
  query_string := fmt.Sprintf(`DELETE FROM %s WHERE id = ?;`, record.TableName())
  result, err := dbHandle.Exec(query_string, record.GetId())
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  return nil
}

func dbRecordReplace(current dbRecord, proposed dbRecord) (err error) {
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

  dbLock.Lock()
  defer dbLock.Unlock()
  query_string := fmt.Sprintf(`UPDATE %s SET %s WHERE id = ?;`, current.TableName(), strings.Join(columns, ", "))
  result, err := dbHandle.Exec(query_string, values...)
  if err != nil { return err }
  count, err := result.RowsAffected()
  if err != nil { return err }
  if count != 1 { return ErrQueryFailed }

  current.FieldsPatch(difference)
  return nil
}

func dbRecordPatch(current dbRecord, patch map[string]any) (err error) {
  if len(patch) == 0 { return nil }

  proposed, err := current.RecordCopy()
  if err != nil { return err }
  proposed.FieldsPatch(patch)

  columns := []string {}
  values  := []any {}
  for column, value := range patch {
    columns = append(columns, fmt.Sprintf("%s = ?", column))
    values  = append(values, value)
  }
  values = append(values, current.GetId())

  dbLock.Lock()
  defer dbLock.Unlock()
  query_string := fmt.Sprintf(`UPDATE %s SET %s WHERE id = ?;`, current.TableName(), strings.Join(columns, ", "))
  result, err := dbHandle.Exec(query_string, values...)
  if err != nil { return err }
  count, err := result.RowsAffected()
  if err != nil { return err }
  if count != 1 { return ErrQueryFailed }

  current.FieldsPatch(patch)
  return nil
}

func dbRecordRead(record dbRecord, id string) (err error) {
  fields, err := record.FieldsRead()
  if err != nil { return err }

  columns := []string {}
  values  := []any {}
  for column, value := range fields {
    columns = append(columns, column)
    values  = append(values, value)
  }
  addresses := make([]any, len(values))
  for index := range values { addresses[index] = &(values[index]) }

  dbLock.RLock()
  defer dbLock.RUnlock()
  query_string := fmt.Sprintf(`SELECT %s FROM %s WHERE id = ?;`, strings.Join(columns, ", "), record.TableName())
  result := dbHandle.QueryRow(query_string, id)
  err = result.Scan(addresses...)
  if err == sql.ErrNoRows { return ErrNotFound }
  if err != nil { return err }

  for index := range values { fields[columns[index]] = values[index] }
  return record.FieldsReplace(fields)
}

func dbRecordWhere(record dbRecord, where_string string, where_values ...any) (results []dbRecord, err error) {
  results = make([]dbRecord, 0)

  fields, err := record.FieldsRead()
  if err != nil { return results, err }

  columns := []string {}
  values  := []any {}
  for column, value := range fields {
    columns = append(columns, column)
    values  = append(values, value)
  }
  addresses := make([]any, len(values))
  for index := range values { addresses[index] = &(values[index]) }

  dbLock.RLock()
  defer dbLock.RUnlock()
  if where_string != "" { where_string = fmt.Sprintf(`WHERE %s`, where_string) }
  query_string := fmt.Sprintf(`SELECT %s FROM %s %s ;`, strings.Join(columns, ", "), record.TableName(), where_string)
  rows, err := dbHandle.Query(query_string, where_values...)
  if err != nil { return results, err }
  defer rows.Close()

  for rows.Next() {
    err = rows.Scan(addresses...)
    if err != nil { return results, err }
    for index := range values { fields[columns[index]] = values[index] }
    this_result, err := record.RecordCreate(fields)
    if err != nil { return results, err }
    results = append(results, this_result)
  }

  return results, nil
}
