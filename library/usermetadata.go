package library

type UserMetadata struct {
  Id          string  `json:"id"`
  UserId      string  `json:"user_id"`
  MetadataId  string  `json:"metadata_id"`
  Started     bool    `json:"started"`
  Timestamp   int64   `json:"timestamp"`
}

// ============================================================================
// Public Interface
func (umd *UserMetadata) Copy() (*UserMetadata) {
  copy := UserMetadata {}
  copy.Id          = umd.Id
  copy.UserId      = umd.UserId
  copy.MetadataId  = umd.MetadataId
  copy.Started     = umd.Started
  copy.Timestamp   = umd.Timestamp

  return &copy
}

// ============================================================================
// public utilities

func UserMetadataExists(user_id string, metadata_id string) bool {
  dbLock.RLock()
  defer dbLock.RUnlock()
  var id string
  queryRow := dbHandle.QueryRow(`SELECT id FROM user_metadata WHERE (user_id = ?) AND (metadata_id = ?) LIMIT 1;`, user_id, metadata_id)
  err := queryRow.Scan(&id)
  return (err == nil)
}

func UserMetadataSetWatchStatus(user_id string, metadata_id string, started bool, timestamp int64) error {
  dbLock.RLock()
  records, err := dbRecordWhere(&UserMetadata{}, `(user_id = ?) AND (metadata_id = ?) LIMIT 1;`, user_id, metadata_id)
  dbLock.RUnlock()
  if err != nil { return ErrQueryFailed }
  if len(records) > 1 { return ErrQueryFailed }
  
  dbLock.Lock()
  defer dbLock.Unlock()

  if len(records) == 1 {
    umd := *(records[0].(*UserMetadata))
    started_int := 0 ; if started { started_int = 1 }
    err := dbRecordPatch(&umd, map[string]any { "started":started_int, "timestamp":timestamp })
    if err != nil { return ErrQueryFailed }
    return nil
  }

  // len(records) == 0
  umd := UserMetadata {}
  umd.UserId     = user_id
  umd.MetadataId = metadata_id
  umd.Started    = started
  umd.Timestamp  = timestamp
  err = dbRecordCreate(&umd)
  if err != nil { return ErrQueryFailed }
  return nil
}

// ============================================================================
// dbRecord interface

func (umd *UserMetadata) TableName() string {
  return "user_metadata"
}

func (umd *UserMetadata) RecordCopy() (dbRecord, error) {
  return umd.Copy(), nil
}

func (umd *UserMetadata) RecordCreate(fields map[string]any) (instance dbRecord, err error) {
  new_instance := UserMetadata {}
  err = new_instance.FieldsReplace(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (umd *UserMetadata) GetId() string {
  return umd.Id
}
func (umd *UserMetadata) SetId(id string) {
  umd.Id = id
}

func (umd *UserMetadata) FieldsRead() (fields map[string]any, err error) {
  fields = make(map[string]any)
  started := 0 ; if umd.Started { started = 1 }

  fields["id"           ] = umd.Id
  fields["user_id"      ] = umd.UserId
  fields["metadata_id"  ] = umd.MetadataId
  fields["started"      ] = started
  fields["timestamp"    ] = umd.Timestamp

  return fields, nil
}

func (umd *UserMetadata) FieldsReplace(fields map[string]any) (err error) {
  umd.Id         = fields["id"          ].(string)
  umd.UserId     = fields["user_id"     ].(string)
  umd.MetadataId = fields["metadata_id" ].(string)
  umd.Started    = fields["started"     ].(int64) > 0
  umd.Timestamp  = fields["timestamp"   ].(int64)

  return nil
}

func (umd *UserMetadata) FieldsPatch(fields map[string]any) (err error) {
  if id,          ok := fields["id"]          ; ok { umd.Id         = id.(string)           }
  if user_id,     ok := fields["user_id"]     ; ok { umd.UserId     = user_id.(string)      }
  if metadata_id, ok := fields["metadata_id"] ; ok { umd.MetadataId = metadata_id.(string)  }
  if started,     ok := fields["started"]     ; ok { umd.Started    = started.(int64) > 0   }
  if timestamp,   ok := fields["timestamp"]   ; ok { umd.Timestamp  = timestamp.(int64)     }

  return nil
}

func (umd_a *UserMetadata) FieldsDifference(other dbRecord) (diff map[string]any, err error) {
  diff = make(map[string]any)
  umd_b, b_is_umd := other.(*UserMetadata)
  if b_is_umd == false { return diff, ErrInvalidType }
  b_started_int := 0 ; if umd_b.Started { b_started_int = 1 }

  if umd_a.Id           != umd_b.Id           { diff["id"          ] = umd_b.Id         }
  if umd_a.UserId       != umd_b.UserId       { diff["user_id"     ] = umd_b.UserId     }
  if umd_a.MetadataId   != umd_b.MetadataId   { diff["metadata_id" ] = umd_b.MetadataId }
  if umd_a.Started      != umd_b.Started      { diff["started"     ] = b_started_int    }
  if umd_a.Timestamp    != umd_b.Timestamp    { diff["timestamp"   ] = umd_b.Timestamp  }

  return diff, nil
}
