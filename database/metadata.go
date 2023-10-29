package database

import (
  "database/sql"
  "encoding/json"
)

type StreamType string
const (
  StreamTypeVideo    StreamType = "video"
  StreamTypeAudio    StreamType = "audio"
  StreamTypeSubtitle StreamType = "subtitle"
)
type Stream struct {
  Type     StreamType `json:"type"`
  Index    int64      `json:"index"`
  Codec    string     `json:"codec"`
  Width    int64      `json:"width"`
  Height   int64      `json:"height"`
  Fps      int64      `json:"fps"`
  Channels int64      `json:"channels"`
  Language string     `json:"language"`
}
func (stream *Stream) Copy() (copy Stream) {
  copy = Stream{}
  copy.Type     = stream.Type
  copy.Index    = stream.Index
  copy.Codec    = stream.Codec
  copy.Width    = stream.Width
  copy.Height   = stream.Height
  copy.Fps      = stream.Fps
  copy.Channels = stream.Channels
  copy.Language = stream.Language
  return copy
}

type MetadataType string
const (
  MetadataTypeFileVideo MetadataType = "file-video"
  MetadataTypeFileAudio MetadataType = "file-audio"
  MetadataTypeSeries    MetadataType = "series"
  MetadataTypeSeason    MetadataType = "season"
  MetadataTypeArtist    MetadataType = "artist"
  MetadataTypeAlbum     MetadataType = "album"
)
type Metadata struct {
  Id                string       `json:"id"`
  CategoryId        string       `json:"category_id"`
  CategoryType      CategoryType `json:"category_type"`
  ParentId          string       `json:"parent_id"`
  GrandparentId     string       `json:"grandparent_id"`
  Type              MetadataType `json:"type"`
  TitleUser         string       `json:"title_user"`
  TitleSort         string       `json:"title_sort"`
  MatchData         string       `json:"match_data"` // JSON match data
  DescriptionShort  string       `json:"description_short"`
  DescriptionLong   string       `json:"description_long"`
  Genre             string       `json:"genre"`
  ReleaseYear       int64        `json:"release_year"`
  ReleaseMonth      int64        `json:"release_month"`
  ReleaseDay        int64        `json:"release_day"`
  SiblingIndex      int64        `json:"sibling_index"` // Season/Episode/Track number
  HasPoster         bool         `json:"has_poster"`
  Location          string       `json:"-"`             // Location of the file, not included in JSON
  Size              int64        `json:"size"`
  Duration          int64        `json:"duration"`
  Streams           []Stream     `json:"streams"`
}

// ============================================================================

func (md *Metadata) Create(db *sql.DB) (err error) { return TableCreate(db, md) }
func (md *Metadata) Update(db *sql.DB, new_values *Metadata) (err error) { return TableUpdate(db, md, new_values) }
func (md *Metadata) Delete(db *sql.DB) (err error) { return TableDelete(db, md) }

func (md *Metadata) Copy() (copy *Metadata) {
  copy = &Metadata{}
  copy.Id                = md.Id
  copy.CategoryId        = md.CategoryId
  copy.CategoryType      = md.CategoryType
  copy.ParentId          = md.ParentId
  copy.GrandparentId     = md.GrandparentId
  copy.Type              = md.Type
  copy.TitleUser         = md.TitleUser
  copy.TitleSort         = md.TitleSort
  copy.MatchData         = md.MatchData
  copy.DescriptionShort  = md.DescriptionShort
  copy.DescriptionLong   = md.DescriptionLong
  copy.Genre             = md.Genre
  copy.ReleaseYear       = md.ReleaseYear
  copy.ReleaseMonth      = md.ReleaseMonth
  copy.ReleaseDay        = md.ReleaseDay
  copy.SiblingIndex      = md.SiblingIndex
  copy.HasPoster         = md.HasPoster
  copy.Location          = md.Location
  copy.Size              = md.Size
  copy.Duration          = md.Duration
  copy.Streams           = make([]Stream, len(md.Streams))
  for index, stream := range md.Streams { copy.Streams[index] = stream.Copy() }
  return copy
}

func MetadataRead(db *sql.DB, id string) (md Metadata, err error) { err = TableRead(db, &md, id) ;  return md, err }

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

func (md *Metadata) CreateFrom(fields map[string]any) (instance Table, err error) {
  new_instance := Metadata {}
  err = new_instance.FieldsWrite(fields)
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
  has_poster := int64(0) ; if md.HasPoster { has_poster = 1 }
  streams_bytes, err := json.Marshal(md.Streams) ; if err != nil { return nil, err } ; streams_string := string(streams_bytes)

  fields["id"               ] = md.Id
  fields["category_id"      ] = md.CategoryId
  fields["category_type"    ] = string(md.CategoryType)
  fields["parent_id"        ] = md.ParentId
  fields["grandparent_id"   ] = md.GrandparentId
  fields["type"             ] = string(md.Type)
  fields["title_user"       ] = md.TitleUser
  fields["title_sort"       ] = md.TitleSort
  fields["match_data"       ] = md.MatchData
  fields["description_short"] = md.DescriptionShort
  fields["description_long" ] = md.DescriptionLong
  fields["genre"            ] = md.Genre
  fields["release_year"     ] = md.ReleaseYear
  fields["release_month"    ] = md.ReleaseMonth
  fields["release_day"      ] = md.ReleaseDay
  fields["sibling_index"    ] = md.SiblingIndex
  fields["has_poster"       ] = has_poster
  fields["location"         ] = md.Location
  fields["size"             ] = md.Size
  fields["duration"         ] = md.Duration
  fields["streams"          ] = streams_string

  return fields, nil
}

func (md *Metadata) FieldsWrite(fields map[string]any) (err error) {
  has_poster := fields["has_poster"].(int64) == 1
  streams_string := fields["streams"].(string) ; var streams []Stream ; err = json.Unmarshal([]byte(streams_string), &streams) ; if err != nil { return err }

  md.Id               = fields["id"               ].(string)
  md.CategoryId       = fields["category_id"      ].(string)
  md.CategoryType     = fields["category_type"    ].(CategoryType)
  md.ParentId         = fields["parent_id"        ].(string)
  md.GrandparentId    = fields["grandparent_id"   ].(string)
  md.Type             = fields["type"             ].(MetadataType)
  md.TitleUser        = fields["title_user"       ].(string)
  md.TitleSort        = fields["title_sort"       ].(string)
  md.MatchData        = fields["match_data"       ].(string)
  md.DescriptionShort = fields["description_short"].(string)
  md.DescriptionLong  = fields["description_long" ].(string)
  md.Genre            = fields["genre"            ].(string)
  md.ReleaseYear      = fields["release_year"     ].(int64)
  md.ReleaseMonth     = fields["release_month"    ].(int64)
  md.ReleaseDay       = fields["release_day"      ].(int64)
  md.SiblingIndex     = fields["sibling_index"    ].(int64)
  md.HasPoster        = has_poster
  md.Location         = fields["location"         ].(string)
  md.Size             = fields["size"             ].(int64)
  md.Duration         = fields["duration"         ].(int64)
  md.Streams          = streams

  return nil
}

func (md_a *Metadata) FieldsDifference(other Table) (diff map[string]any, err error) {
  diff = make(map[string]any)
  md_b, b_is_md := other.(*Metadata)
  if b_is_md == false { return diff, ErrInvalidType }

  b_has_poster := int64(0) ; if md_b.HasPoster { b_has_poster = 1 }
  a_streams_bytes, err := json.Marshal(md_a.Streams) ; if err != nil { return nil, err } ; a_streams_string := string(a_streams_bytes)
  b_streams_bytes, err := json.Marshal(md_b.Streams) ; if err != nil { return nil, err } ; b_streams_string := string(b_streams_bytes)

  if md_a.Id               != md_b.Id               { diff["id"               ] = md_b.Id                   }
  if md_a.CategoryId       != md_b.CategoryId       { diff["category_id"      ] = md_b.CategoryId           }
  if md_a.CategoryType     != md_b.CategoryType     { diff["category_type"    ] = string(md_b.CategoryType) }
  if md_a.ParentId         != md_b.ParentId         { diff["parent_id"        ] = md_b.ParentId             }
  if md_a.GrandparentId    != md_b.GrandparentId    { diff["grandparent_id"   ] = md_b.GrandparentId        }
  if md_a.Type             != md_b.Type             { diff["type"             ] = string(md_b.Type)         }
  if md_a.TitleUser        != md_b.TitleUser        { diff["title_user"       ] = md_b.TitleUser            }
  if md_a.TitleSort        != md_b.TitleSort        { diff["title_sort"       ] = md_b.TitleSort            }
  if md_a.MatchData        != md_b.MatchData        { diff["match_data"       ] = md_b.MatchData            }
  if md_a.DescriptionShort != md_b.DescriptionShort { diff["description_short"] = md_b.DescriptionShort     }
  if md_a.DescriptionLong  != md_b.DescriptionLong  { diff["description_long" ] = md_b.DescriptionLong      }
  if md_a.Genre            != md_b.Genre            { diff["genre"            ] = md_b.Genre                }
  if md_a.ReleaseYear      != md_b.ReleaseYear      { diff["release_year"     ] = md_b.ReleaseYear          }
  if md_a.ReleaseMonth     != md_b.ReleaseMonth     { diff["release_month"    ] = md_b.ReleaseMonth         }
  if md_a.ReleaseDay       != md_b.ReleaseDay       { diff["release_day"      ] = md_b.ReleaseDay           }
  if md_a.SiblingIndex     != md_b.SiblingIndex     { diff["sibling_index"    ] = md_b.SiblingIndex         }
  if md_a.HasPoster        != md_b.HasPoster        { diff["has_poster"       ] = b_has_poster              }
  if md_a.Location         != md_b.Location         { diff["location"         ] = md_b.Location             }
  if md_a.Size             != md_b.Size             { diff["size"             ] = md_b.Size                 }
  if md_a.Duration         != md_b.Duration         { diff["duration"         ] = md_b.Duration             }
  if a_streams_string      != b_streams_string      { diff["streams"          ] = b_streams_string          }

  return diff, nil
}

func (md *Metadata) ValidCreate(db *sql.DB) (valid bool, err error) {
  return true, nil
}

func (md *Metadata) ValidUpdate(db *sql.DB, other Table) (valid bool, err error) {
  return true, nil
}

func (md *Metadata) ValidDelete(db *sql.DB) (valid bool, err error) {
  return true, nil
}
