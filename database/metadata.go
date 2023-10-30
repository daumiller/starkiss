package database

import (
  "database/sql"
  "encoding/json"
)

type FileStreamType string
const (
  FileStreamTypeVideo    FileStreamType = "video"
  FileStreamTypeAudio    FileStreamType = "audio"
  FileStreamTypeSubtitle FileStreamType = "subtitle"
)
type FileStream struct {
  StreamType FileStreamType `json:"stream_type"`
  Index      int64          `json:"index"`
  Codec      string         `json:"codec"`
  Width      int64          `json:"width"`
  Height     int64          `json:"height"`
  Fps        int64          `json:"fps"`
  Channels   int64          `json:"channels"`
  Language   string         `json:"language"`
}
func (stream *FileStream) Copy() (*FileStream) {
  copy := FileStream{}
  copy.StreamType = stream.StreamType
  copy.Index      = stream.Index
  copy.Codec      = stream.Codec
  copy.Width      = stream.Width
  copy.Height     = stream.Height
  copy.Fps        = stream.Fps
  copy.Channels   = stream.Channels
  copy.Language   = stream.Language
  return &copy
}

// ============================================================================

type MetadataMediaType string
const (
  MetadataMediaTypeFileVideo MetadataMediaType = "file-video"
  MetadataMediaTypeFileAudio MetadataMediaType = "file-audio"
  MetadataMediaTypeSeries    MetadataMediaType = "series"
  MetadataMediaTypeSeason    MetadataMediaType = "season"
  MetadataMediaTypeArtist    MetadataMediaType = "artist"
  MetadataMediaTypeAlbum     MetadataMediaType = "album"
)
type Metadata struct {
  Id          string            `json:"id"`
  ParentId    string            `json:"parent_id"`
  MediaType   MetadataMediaType `json:"media_type"`
  NameDisplay string            `json:"name_display"`
  NameSort    string            `json:"name_sort"`
  Streams     []FileStream      `json:"streams"`
  Duration    int64             `json:"duration"`
  Size        int64             `json:"size"`
  Hidden      bool              `json:"hidden"`
}

// ============================================================================

func (md *Metadata) Copy() (*Metadata) {
  copy := Metadata {}
  copy.Id          = md.Id
  copy.ParentId    = md.ParentId
  copy.MediaType   = md.MediaType
  copy.NameDisplay = md.NameDisplay
  copy.NameSort    = md.NameSort
  copy.Streams     = make([]FileStream, len(md.Streams))
  copy.Duration    = md.Duration
  copy.Size        = md.Size
  copy.Hidden      = md.Hidden

  for index, stream := range md.Streams {
    stream_copy := stream.Copy()
    copy.Streams[index] = *stream_copy
  }

  return &copy
}

func (md *Metadata) Create (db *sql.DB) (err error) { return TableCreate(db, md) }
func (md *Metadata) Replace(db *sql.DB, new_values *Metadata) (err error) { return TableReplace(db, md, new_values) }
func (md *Metadata) Patch  (db *sql.DB, new_values map[string]any) (err error) { return TablePatch(db, md, new_values) }
func (md *Metadata) Delete (db *sql.DB) (err error) { return TableDelete(db, md) }

func MetadataRead(db *sql.DB, id string) (md *Metadata, err error) {
  md = &Metadata{}
  err = TableRead(db, md, id)
  return md, err
}
func MetadataWhere(db *sql.DB, where_string string, where_args ...any) (md_list []Metadata, err error) {
  tables, err := TableWhere(db, &Metadata{}, where_string, where_args...)
  if err != nil { return nil, err }

  md_list = make([]Metadata, len(tables))
  for index, table := range tables { md_list[index] = *(table.(*Metadata)) }
  return md_list, nil
}

// ============================================================================
// Table interface

func (md *Metadata) TableName() string {
  return "metadata"
}

func (md *Metadata) CopyRecord() (Table, error) {
  return md.Copy(), nil
}

func (md *Metadata) CreateFrom(fields map[string]any) (instance Table, err error) {
  new_instance := Metadata {}
  err = new_instance.FieldsReplace(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (md *Metadata) GetId() string {
  return md.Id
}
func (md *Metadata) SetId(id string) {
  md.Id = id
}

func (md *Metadata) FieldsRead() (fields map[string]any, err error) {
  fields = make(map[string]any)
  hidden := int64(0) ; if md.Hidden { hidden = 1 }
  streams_bytes, err := json.Marshal(md.Streams) ; if err != nil { return nil, err } ; streams_string := string(streams_bytes)

  fields["id"           ] = md.Id
  fields["parent_id"    ] = md.ParentId
  fields["media_type"   ] = string(md.MediaType)
  fields["name_display" ] = md.NameDisplay
  fields["name_sort"    ] = md.NameSort
  fields["streams"      ] = streams_string
  fields["duration"     ] = md.Duration
  fields["size"         ] = md.Size
  fields["hidden"       ] = hidden

  return fields, nil
}

func (md *Metadata) FieldsReplace(fields map[string]any) (err error) {
  hidden := fields["hidden"].(int64) == 1
  streams_string := fields["streams"].(string) ; var streams []FileStream ; err = json.Unmarshal([]byte(streams_string), &streams) ; if err != nil { return err }

  md.Id               = fields["id"               ].(string)
  md.ParentId         = fields["parent_id"        ].(string)
  md.MediaType        = fields["media_type"       ].(MetadataMediaType)
  md.NameDisplay      = fields["name_display"     ].(string)
  md.NameSort         = fields["name_sort"        ].(string)
  md.Streams          = streams
  md.Duration         = fields["duration"         ].(int64)
  md.Size             = fields["size"             ].(int64)
  md.Hidden           = hidden

  return nil
}

func (md *Metadata) FieldsPatch(fields map[string]any) (err error) {
  if id,           ok := fields["id"]           ; ok { md.Id          = id.(string)                    }
  if parent_id,    ok := fields["parent_id"]    ; ok { md.ParentId    = parent_id.(string)             }
  if media_type,   ok := fields["media_type"]   ; ok { md.MediaType   = media_type.(MetadataMediaType) }
  if name_display, ok := fields["name_display"] ; ok { md.NameDisplay = name_display.(string)          }
  if name_sort,    ok := fields["name_sort"]    ; ok { md.NameSort    = name_sort.(string)             }
  if duration,     ok := fields["duration"]     ; ok { md.Duration    = duration.(int64)               }
  if size,         ok := fields["size"]         ; ok { md.Size        = size.(int64)                   }

  if streams, ok := fields["streams"] ; ok {
    streams_string := streams.(string)
    var streams []FileStream
    err = json.Unmarshal([]byte(streams_string), &streams)
    if err != nil { return err }
    md.Streams = streams
  }

  if hidden, ok := fields["hidden"] ; ok {
    md.Hidden = (hidden.(int64) == 1)
  }

  return nil
}

func (md_a *Metadata) FieldsDifference(other Table) (diff map[string]any, err error) {
  diff = make(map[string]any)
  md_b, b_is_md := other.(*Metadata)
  if b_is_md == false { return diff, ErrInvalidType }

  b_hidden := int64(0) ; if md_b.Hidden { b_hidden = 1 }
  a_streams_bytes, err := json.Marshal(md_a.Streams) ; if err != nil { return nil, err } ; a_streams_string := string(a_streams_bytes)
  b_streams_bytes, err := json.Marshal(md_b.Streams) ; if err != nil { return nil, err } ; b_streams_string := string(b_streams_bytes)

  if md_a.Id          != md_b.Id          { diff["id"           ] = md_b.Id                }
  if md_a.ParentId    != md_b.ParentId    { diff["parent_id"    ] = md_b.ParentId          }
  if md_a.MediaType   != md_b.MediaType   { diff["media_type"   ] = string(md_b.MediaType) }
  if md_a.NameDisplay != md_b.NameDisplay { diff["name_display" ] = md_b.NameDisplay       }
  if md_a.NameSort    != md_b.NameSort    { diff["name_sort"    ] = md_b.NameSort          }
  if a_streams_string != b_streams_string { diff["streams"      ] = b_streams_string       }
  if md_a.Duration    != md_b.Duration    { diff["duration"     ] = md_b.Duration          }
  if md_a.Size        != md_b.Size        { diff["size"         ] = md_b.Size              }
  if md_a.Hidden      != md_b.Hidden      { diff["hidden"       ] = b_hidden               }

  return diff, nil
}

func (md *Metadata) ValidCreate(db *sql.DB) (valid bool, err error) {
  return true, nil
}

func (md *Metadata) ValidUpdate(db *sql.DB, other Table) (valid bool, err error) {
  return true, nil
}

func (md *Metadata) ValidDelete(db *sql.DB) (valid bool, err error) {
  // TODO: ensure no children
  return true, nil
}
