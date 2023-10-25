package database

import (
  "strings"
  "database/sql"
  "encoding/json"
)

type Unprocessed struct {
  Id                 string      `json:"id"`
  NeedsStreamMap     bool        `json:"needs_stream_map"`
  NeedsTranscoding   bool        `json:"needs_transcoding"`
  NeedsMetadata      bool        `json:"needs_metadata"`
  SourceLocation     string      `json:"source_location"`
  SourceStreams      []Stream    `json:"source_streams"`
  TranscodedLocation string      `json:"transcoded_location"`
  TranscodedStreams  []Stream    `json:"transcoded_streams"` // for this list only, index is index from the source stream
  MatchData          string      `json:"match_data"`         // JSON match data
  ProvisionalId      string      `json:"provisional_id"`
  CreatedAt          uint64      `json:"created_at"`
}

// ============================================================================

func (unp *Unprocessed) Create(db *sql.DB) (err error) {
  return TableCreate(db, unp)
}

func (unp *Unprocessed) Update(db *sql.DB, new_values *Unprocessed) (err error) {
  return TableUpdate(db, unp, new_values)
}

func (unp *Unprocessed) Delete(db *sql.DB) (err error) {
  return TableDelete(db, unp)
}

func UnprocessedRead(db *sql.DB, id string) (unp Unprocessed, err error) {
  err = TableRead(db, &unp, id)
  return unp, err
}

func UnprocessedList(db *sql.DB, needs_stream_map bool, needs_transcoding bool, needs_metadata bool) (unp_list []Unprocessed, err error) {
  where_strings := []string {}
  where_params  := []any {}
  if needs_stream_map  == true { where_strings = append(where_strings, `needs_stream_map = ?`)  ; where_params = append(where_params, uint64(1)) }
  if needs_transcoding == true { where_strings = append(where_strings, `needs_transcoding = ?`) ; where_params = append(where_params, uint64(1)) }
  if needs_metadata    == true { where_strings = append(where_strings, `needs_metadata = ?`)    ; where_params = append(where_params, uint64(1)) }
  where_string := strings.Join(where_strings, " AND ")

  tables, err := TableWhere(db, &Unprocessed{}, where_string, where_params...)
  if err != nil { return nil, err }

  for _, table := range tables {
    unp_list = append(unp_list, *(table.(*Unprocessed)))
  }
  return unp_list, nil
}

// ============================================================================
// Table interface

func (unp *Unprocessed) TableName() string { return "unprocessed" }
func (unp *Unprocessed) GetId() string { return unp.Id }

func (unp *Unprocessed) CreateFrom(fields map[string]any) (instance Table, err error) {
  new_instance := Unprocessed {}
  err = new_instance.FieldsWrite(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (unp *Unprocessed) FieldsRead() (fields map[string]any, err error) {
  needs_stream_map  := uint64(0) ; if (unp.NeedsStreamMap   == true) { needs_stream_map  = 1 }
  needs_transcoding := uint64(0) ; if (unp.NeedsTranscoding == true) { needs_transcoding = 1 }
  needs_metadata    := uint64(0) ; if (unp.NeedsMetadata    == true) { needs_metadata    = 1 }
  source_bytes,     err := json.Marshal(unp.SourceStreams)     ; if err != nil { return fields, err } ; source_streams     := string(source_bytes)
  transcoded_bytes, err := json.Marshal(unp.TranscodedStreams) ; if err != nil { return fields, err } ; transcoded_streams := string(transcoded_bytes)

  fields["id"]                  = unp.Id
  fields["needs_stream_map"]    = needs_stream_map
  fields["needs_transcoding"]   = needs_transcoding
  fields["needs_metadata"]      = needs_metadata
  fields["source_location"]     = unp.SourceLocation
  fields["source_streams"]      = source_streams
  fields["transcoded_location"] = unp.TranscodedLocation
  fields["transcoded_streams"]  = transcoded_streams
  fields["match_data"]          = unp.MatchData
  fields["provisional_id"]      = unp.ProvisionalId
  fields["created_at"]          = unp.CreatedAt

  return fields, nil
}

func (unp *Unprocessed) FieldsWrite(fields map[string]any) (err error) {
  needs_stream_map  := false ; if (fields["needs_stream_map"]   == 1) { needs_stream_map  = true }
  needs_transcoding := false ; if (fields["needs_transcoding"]  == 1) { needs_transcoding = true }
  needs_metadata    := false ; if (fields["needs_metadata"]     == 1) { needs_metadata    = true }
  source_streams     := []Stream {} ; err = json.Unmarshal([]byte(fields["source_streams"    ].(string)), &source_streams)     ; if err != nil { return err }
  transcoded_streams := []Stream {} ; err = json.Unmarshal([]byte(fields["transcoded_streams"].(string)), &transcoded_streams) ; if err != nil { return err }

  unp.Id                  = fields["id"].(string)
  unp.NeedsStreamMap      = needs_stream_map
  unp.NeedsTranscoding    = needs_transcoding
  unp.NeedsMetadata       = needs_metadata
  unp.SourceLocation      = fields["source_location"].(string)
  unp.SourceStreams       = source_streams
  unp.TranscodedLocation  = fields["transcoded_location"].(string)
  unp.TranscodedStreams   = transcoded_streams
  unp.MatchData           = fields["match_data"].(string)
  unp.ProvisionalId       = fields["provisional_id"].(string)
  unp.CreatedAt           = fields["created_at"].(uint64)

  return nil
}

func (unp_a *Unprocessed) FieldsDifference(other Table) (diff map[string]any, err error) {
  unp_b, b_is_unp := other.(*Unprocessed)
  if b_is_unp == false { return diff, ErrInvalidType }

  b_needs_stream_map  := uint64(0) ; if (unp_b.NeedsStreamMap   == true) { b_needs_stream_map  = 1 }
  b_needs_transcoding := uint64(0) ; if (unp_b.NeedsTranscoding == true) { b_needs_transcoding = 1 }
  b_needs_metadata    := uint64(0) ; if (unp_b.NeedsMetadata    == true) { b_needs_metadata    = 1 }
  a_source_bytes,     err := json.Marshal(unp_a.SourceStreams)     ; if err != nil { return diff, err } ; a_source_streams     := string(a_source_bytes)
  a_transcoded_bytes, err := json.Marshal(unp_a.TranscodedStreams) ; if err != nil { return diff, err } ; a_transcoded_streams := string(a_transcoded_bytes)
  b_source_bytes,     err := json.Marshal(unp_b.SourceStreams)     ; if err != nil { return diff, err } ; b_source_streams     := string(b_source_bytes)
  b_transcoded_bytes, err := json.Marshal(unp_b.TranscodedStreams) ; if err != nil { return diff, err } ; b_transcoded_streams := string(b_transcoded_bytes)

  if unp_b.Id               != unp_a.Id               { diff["id"]                 = unp_b.Id             }
  if unp_b.NeedsStreamMap   != unp_a.NeedsStreamMap   { diff["needs_stream_map"]   = b_needs_stream_map   }
  if unp_b.NeedsTranscoding != unp_a.NeedsTranscoding { diff["needs_transcoding"]  = b_needs_transcoding  }
  if unp_b.NeedsMetadata    != unp_a.NeedsMetadata    { diff["needs_metadata"]     = b_needs_metadata     }
  if unp_b.SourceLocation   != unp_a.SourceLocation   { diff["source_location"]    = unp_b.SourceLocation }
  if b_source_streams       != a_source_streams       { diff["source_streams"]     = b_source_streams     }
  if b_transcoded_streams   != a_transcoded_streams   { diff["transcoded_streams"] = b_transcoded_streams }
  if unp_b.MatchData        != unp_a.MatchData        { diff["match_data"]         = unp_b.MatchData      }
  if unp_b.ProvisionalId    != unp_a.ProvisionalId    { diff["provisional_id"]     = unp_b.ProvisionalId  }
  if unp_b.CreatedAt        != unp_a.CreatedAt        { diff["created_at"]         = unp_b.CreatedAt      }

  return diff, nil
}

func (unp *Unprocessed) ValidCreate(db *sql.DB) (valid bool, err error) {
  return true, nil
}

func (unp *Unprocessed) ValidUpdate(db *sql.DB, other Table) (valid bool, err error) {
  return true, nil
}

func (unp *Unprocessed) ValidDelete(db *sql.DB) (valid bool, err error) {
  return true, nil
}

// ============================================================================

/*
type unprocessedFields struct {
  needs_stream_map   uint64
  needs_transcoding  uint64
  needs_metadata     uint64
  source_streams     string
  transcoded_streams string
}
func (fields *unprocessedFields) fromUnprocessed(unp *Unprocessed) (err error) {
  fields.needs_stream_map  = uint64(0) ; if (unp.NeedsStreamMap   == true) { fields.needs_stream_map  = 1 }
  fields.needs_transcoding = uint64(0) ; if (unp.NeedsTranscoding == true) { fields.needs_transcoding = 1 }
  fields.needs_metadata    = uint64(0) ; if (unp.NeedsMetadata    == true) { fields.needs_metadata    = 1 }
  source_bytes,     err := json.Marshal(unp.SourceStreams)     ; if err != nil { return err } ; fields.source_streams     = string(source_bytes)
  transcoded_bytes, err := json.Marshal(unp.TranscodedStreams) ; if err != nil { return err } ; fields.transcoded_streams = string(transcoded_bytes)
  return nil
}
func (unp *Unprocessed) fromFields(fields *unprocessedFields) (err error) {
  unp.NeedsStreamMap   = false ; if (fields.needs_stream_map   == 1) { unp.NeedsStreamMap   = true }
  unp.NeedsTranscoding = false ; if (fields.needs_transcoding  == 1) { unp.NeedsTranscoding = true }
  unp.NeedsMetadata    = false ; if (fields.needs_metadata     == 1) { unp.NeedsMetadata    = true }
  err = json.Unmarshal([]byte(fields.source_streams),     &unp.SourceStreams)     ; if err != nil { return err }
  err = json.Unmarshal([]byte(fields.transcoded_streams), &unp.TranscodedStreams) ; if err != nil { return err }
  return nil
}

func (unp *Unprocessed) Create(db *sql.DB) (err error) {
  if unp.Id == "" { unp.Id = uuid.NewString() }
  fields := unprocessedFields {} ; err = fields.fromUnprocessed(unp) ; if err != nil { return err }

  result, err := db.Exec(`INSERT INTO unprocessed (
      id, needs_stream_map, needs_transcoding, needs_metadata, 
      source_location, source_streams, transcoded_location, transcoded_streams, 
      match_data, provisional_id, created_at
    ) VALUES (?,?,?,?, ?,?,?,?, ?,?,?)`,
    unp.Id, fields.needs_stream_map, fields.needs_transcoding, fields.needs_metadata,
    unp.SourceLocation, fields.source_streams, unp.TranscodedLocation, fields.transcoded_streams,
    unp.MatchData, unp.ProvisionalId, unp.CreatedAt,
  )
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  return nil
}

func (unp *Unprocessed) Update(db *sql.DB, new_values *Unprocessed) (err error) {
  unp_fields := unprocessedFields {} ; err = unp_fields.fromUnprocessed(unp)        ; if err != nil { return err }
  new_fields := unprocessedFields {} ; err = new_fields.fromUnprocessed(new_values) ; if err != nil { return err }

  update_columns := []string {}
  update_values  := []any {}

  // NOTE: cannot update Id or CreatedAt
  if unp.NeedsStreamMap            != new_values.NeedsStreamMap     { update_columns = append(update_columns, `needs_stream_map = ?`)    ; update_values = append(update_values, new_fields.needs_stream_map)   }
  if unp.NeedsTranscoding          != new_values.NeedsTranscoding   { update_columns = append(update_columns, `needs_transcoding = ?`)   ; update_values = append(update_values, new_fields.needs_transcoding)  }
  if unp.NeedsMetadata             != new_values.NeedsMetadata      { update_columns = append(update_columns, `needs_metadata = ?`)      ; update_values = append(update_values, new_fields.needs_metadata)     }
  if unp.SourceLocation            != new_values.SourceLocation     { update_columns = append(update_columns, `source_location = ?`)     ; update_values = append(update_values, new_values.SourceLocation)     }
  if unp_fields.source_streams     != new_fields.source_streams     { update_columns = append(update_columns, `source_streams = ?`)      ; update_values = append(update_values, new_fields.source_streams)     }
  if unp.TranscodedLocation        != new_values.TranscodedLocation { update_columns = append(update_columns, `transcoded_location = ?`) ; update_values = append(update_values, new_values.TranscodedLocation) }
  if unp_fields.transcoded_streams != new_fields.transcoded_streams { update_columns = append(update_columns, `transcoded_streams = ?`)  ; update_values = append(update_values, new_fields.transcoded_streams) }
  if unp.MatchData                 != new_values.MatchData          { update_columns = append(update_columns, `match_data = ?`)          ; update_values = append(update_values, new_values.MatchData)          }
  if unp.ProvisionalId             != new_values.ProvisionalId      { update_columns = append(update_columns, `provisional_id = ?`)      ; update_values = append(update_values, new_values.ProvisionalId)      }

  if len(update_columns) == 0 { return nil }
  query_string := `UPDATE unprocessed SET ` + strings.Join(update_columns, ", ") + ` WHERE id = ?;`
  update_values = append(update_values, unp.Id)

  result, err := db.Exec(query_string, update_values...)
  if err != nil { return err }
  count, err := result.RowsAffected()
  if err != nil { return err }
  if count != 1 { return ErrQueryFailed }
  return nil
}

func (unp *Unprocessed) Delete(db *sql.DB) (err error) {
  result, err := db.Exec(`DELETE FROM unprocessed WHERE id = ?;`, unp.Id)
  if err != nil { return err }
  if rows, _ := result.RowsAffected(); rows != 1 { return ErrQueryFailed }
  return nil
}

func UnprocessedRead(db *sql.DB, id string) (unp Unprocessed, err error) {
  fields := unprocessedFields {}
  row := db.QueryRow(`SELECT 
      id, needs_stream_map, needs_transcoding, needs_metadata, 
      source_location, source_streams, transcoded_location, transcoded_streams, 
      match_data, provisional_id, created_at
    FROM unprocessed WHERE id = ?;`,
    id,
  )
  err = row.Scan(
    &unp.Id, &fields.needs_stream_map, &fields.needs_transcoding, &fields.needs_metadata,
    &unp.SourceLocation, &fields.source_streams, &unp.TranscodedLocation, &fields.transcoded_streams,
    &unp.MatchData, &unp.ProvisionalId, &unp.CreatedAt,
  )
  if err != nil { return unp, err }
  err = unp.fromFields(&fields)
  if err != nil { return unp, err }
  return unp, nil
}

func UnprocessedList(db *sql.DB, needs_stream_map bool, needs_transcoding bool, needs_metadata bool) (unp_list []Unprocessed, err error) {
  fields := unprocessedFields {}
  query_strings := []string {}
  query_params  := []any {}
  if needs_stream_map  == true { query_strings = append(query_strings, `needs_stream_map = ?`)  ; query_params = append(query_params, uint64(1)) }
  if needs_transcoding == true { query_strings = append(query_strings, `needs_transcoding = ?`) ; query_params = append(query_params, uint64(1)) }
  if needs_metadata    == true { query_strings = append(query_strings, `needs_metadata = ?`)    ; query_params = append(query_params, uint64(1)) }
  query_string := `SELECT 
      id, needs_stream_map, needs_transcoding, needs_metadata, 
      source_location, source_streams, transcoded_location, transcoded_streams, 
      match_data, provisional_id, created_at
    FROM unprocessed`
  if len(query_strings) > 0 { query_string += ` WHERE ` + strings.Join(query_strings, " AND ") }
  query_string += `;`

  rows, err := db.Query(query_string, query_params...)
  if err != nil { return nil, err }
  defer rows.Close()

  for rows.Next() {
    var unp Unprocessed
    err = rows.Scan(
      &unp.Id, &fields.needs_stream_map, &fields.needs_transcoding, &fields.needs_metadata,
      &unp.SourceLocation, &fields.source_streams, &unp.TranscodedLocation, &fields.transcoded_streams,
      &unp.MatchData, &unp.ProvisionalId, &unp.CreatedAt,
    )
    if err != nil { return nil, err }
    err = unp.fromFields(&fields)
    if err != nil { return nil, err }
    unp_list = append(unp_list, unp)
  }

  return unp_list, nil
}
*/
