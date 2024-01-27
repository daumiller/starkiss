package library

import (
  "database/sql"
  "encoding/json"
)

type UserMetadataView struct {
  UserId      string            `json:"user_id"`
  MetadataId  string            `json:"metadata_id"`

  ParentId    string            `json:"parent_id"`
  MediaType   MetadataMediaType `json:"media_type"`
  NameDisplay string            `json:"name_display"`
  NameSort    string            `json:"name_sort"`
  Streams     []FileStream      `json:"streams"`
  Duration    int64             `json:"duration"`
  Size        int64             `json:"size"`

  Started     bool              `json:"started"`
  Timestamp   int64             `json:"timestamp"`
}

// ============================================================================
// Public Interface
func UserMetadataViewForParent(user_id string, parent_id string) ([]UserMetadataView, error) {
  results := make([]UserMetadataView, 0)

  dbLock.RLock()
  defer dbLock.RUnlock()
  query_string := `SELECT
      metadata.id AS metadata_id,
      metadata.parent_id,
      metadata.media_type,
      metadata.name_display,
      metadata.name_sort,
      metadata.streams,
      metadata.duration,
      metadata.size,
      user_metadata.user_id,
      user_metadata.started,
      user_metadata.timestamp
    FROM metadata
    LEFT JOIN user_metadata
      ON  (user_metadata.metadata_id = metadata.id)
      AND (user_metadata.user_id = ?)
    WHERE (metadata.parent_id = ?)
    ORDER BY metadata.name_sort ASC;`
  rows, err := dbHandle.Query(query_string, user_id, parent_id)
  if err != nil { return results, err }

  defer rows.Close()
  for rows.Next() {
    row_result, err := ScanUserMetadataViewRow(rows)
    if err != nil { return results, err }
    results = append(results, row_result)
  }

  return results, nil
}

// ============================================================================
// private utilities
func ScanUserMetadataViewRow(rows *sql.Rows) (UserMetadataView, error) {
  var (
    metadata_id, parent_id, media_type, name_display, name_sort, streams string
    duration, size int64
    null_user_id sql.NullString
    null_timestamp, null_started sql.NullInt64
  )
  err := rows.Scan(&metadata_id, &parent_id, &media_type, &name_display, &name_sort, &streams, &duration, &size, &null_user_id, &null_started, &null_timestamp)
  if err != nil { return UserMetadataView {}, err }

  var streams_array []FileStream
  err = json.Unmarshal([]byte(streams), &streams_array)
  if err != nil { return UserMetadataView {}, err }

  user_id   := ""       ; if null_user_id.Valid   { user_id   = null_user_id.String      }
  started   := false    ; if null_started.Valid   { started   = (null_started.Int64 > 0) }
  timestamp := int64(0) ; if null_timestamp.Valid { timestamp = null_timestamp.Int64     }

  return UserMetadataView {
    UserId:      user_id,
    MetadataId:  metadata_id,
    ParentId:    parent_id,
    MediaType:   MetadataMediaType(media_type),
    NameDisplay: name_display,
    NameSort:    name_sort,
    Streams:     streams_array,
    Duration:    duration,
    Size:        size,
    Started:     started,
    Timestamp:   timestamp,
  }, nil
}
