package database

import (
  "encoding/json"
  "database/sql"
)

type InputFile struct {
  Id                       string       `json:"id"`                        // Metadata.Id == InputFile.Id
  SourceLocation           string       `json:"source_location"`
  TranscodedLocation       string       `json:"transcoded_location"`       // empty == needs_transcoded
  SourceStreams            []FileStream `json:"source_streams"`
  StreamMap                []int64      `json:"stream_map"`                // empty == needs_map
  SourceDuration           int64        `json:"source_duration"`           // length of media in seconds
  TimeScanned              int64        `json:"time_scanned"`
  TranscodingCommand       string       `json:"transcoding_command"`       // ffmpeg command line
  TranscodingTimeStarted   int64        `json:"transcoding_time_started"`  // time transcoding was started
  TranscodingTimeElapsed   int64        `json:"transcoding_time_elapsed"`  // seconds elapsed during transcoding
  TranscodingError         string       `json:"transcoding_error"`         // error message from transcoding process
}

func (inp *InputFile) Copy() (*InputFile) {
  copy := InputFile {}
  copy.Id                       = inp.Id
  copy.SourceLocation           = inp.SourceLocation
  copy.TranscodedLocation       = inp.TranscodedLocation
  copy.SourceStreams            = make([]FileStream, len(inp.SourceStreams))
  copy.StreamMap                = make([]int64, len(inp.StreamMap))
  copy.SourceDuration           = inp.SourceDuration
  copy.TimeScanned              = inp.TimeScanned
  copy.TranscodingCommand       = inp.TranscodingCommand
  copy.TranscodingTimeStarted   = inp.TranscodingTimeStarted
  copy.TranscodingTimeElapsed   = inp.TranscodingTimeElapsed
  copy.TranscodingError         = inp.TranscodingError

  for index, stream := range inp.SourceStreams {
    stream_copy := stream.Copy()
    copy.SourceStreams[index] = *stream_copy
  }
  for index, stream := range inp.StreamMap {
    copy.StreamMap[index] = stream
  }
  return &copy
}

func (inp *InputFile) Create(db *sql.DB) (err error) { return TableCreate(db, inp) }
func (inp *InputFile) Update(db *sql.DB, new_values *InputFile) (err error) { return TableUpdate(db, inp, new_values) }
func (inp *InputFile) Delete(db *sql.DB) (err error) { return TableDelete(db, inp) }

func InputFileRead(db *sql.DB, id string) (inp *InputFile, err error) { err = TableRead(db, inp, id) ;  return inp, err }
func InputFileWhere(db *sql.DB, where_string string, where_args ...any) (inp_list []InputFile, err error) {
  tables, err := TableWhere(db, &InputFile{}, where_string, where_args...)
  if err != nil { return nil, err }

  inp_list = make([]InputFile, len(tables))
  for index, table := range tables { inp_list[index] = *(table.(*InputFile)) }
  return inp_list, nil
}

// ============================================================================
// Table interface

func (inp *InputFile) TableName() string { return "input_files"}
func (inp *InputFile) GetId() string { return inp.Id }
func (inp *InputFile) SetId(id string) { inp.Id = id }

func (inp *InputFile) CopyRecord() (Table, error) {
  return inp.Copy(), nil
}

func (inp *InputFile) CreateFrom(fields map[string]any) (instance Table, err error) {
  new_instance := InputFile {}
  err = new_instance.FieldsWrite(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (inp *InputFile) FieldsRead() (fields map[string]any, err error) {
  streams_bytes, err := json.Marshal(inp.SourceStreams) ; if err != nil { return nil, err } ; streams_string := string(streams_bytes)
  map_bytes, err := json.Marshal(inp.StreamMap) ; if err != nil { return nil, err } ; map_string := string(map_bytes)

  fields = make(map[string]any)
  fields["id"]                       = inp.Id
  fields["source_location"]          = inp.SourceLocation
  fields["transcoded_location"]      = inp.TranscodedLocation
  fields["source_streams"]           = streams_string
  fields["stream_map"]               = map_string
  fields["source_duration"]          = inp.SourceDuration
  fields["time_scanned"]             = inp.TimeScanned
  fields["transcoding_command"]      = inp.TranscodingCommand
  fields["transcoding_time_started"] = inp.TranscodingTimeStarted
  fields["transcoding_time_elapsed"] = inp.TranscodingTimeElapsed
  fields["transcoding_error"]        = inp.TranscodingError

  return fields, nil
}

func (inp *InputFile) FieldsWrite(fields map[string]any) (err error) {
  streams_string := fields["streams"].(string) ; var source_streams []FileStream ; err = json.Unmarshal([]byte(streams_string), &source_streams) ; if err != nil { return err }
  map_string := fields["stream_map"].(string) ; var stream_map []int64 ; err = json.Unmarshal([]byte(map_string), &stream_map) ; if err != nil { return err }

  inp.Id                     = fields["id"].(string)
  inp.SourceLocation         = fields["source_location"].(string)
  inp.TranscodedLocation     = fields["transcoded_location"].(string)
  inp.SourceStreams          = source_streams
  inp.StreamMap              = stream_map
  inp.SourceDuration         = fields["source_duration"].(int64)
  inp.TimeScanned            = fields["time_scanned"].(int64)
  inp.TranscodingCommand     = fields["transcoding_command"].(string)
  inp.TranscodingTimeStarted = fields["transcoding_time_started"].(int64)
  inp.TranscodingTimeElapsed = fields["transcoding_time_elapsed"].(int64)
  inp.TranscodingError       = fields["transcoding_error"].(string)
  return nil
}

func (inp_a *InputFile) FieldsDifference(other Table) (diff map[string]any, err error) {
  diff = make(map[string]any)
  inp_b, b_is_inp := other.(*InputFile)
  if b_is_inp == false { return diff, ErrInvalidType }

  a_streams_bytes, err := json.Marshal(inp_a.SourceStreams) ; if err != nil { return nil, err } ; a_streams_string := string(a_streams_bytes)
  b_streams_bytes, err := json.Marshal(inp_b.SourceStreams) ; if err != nil { return nil, err } ; b_streams_string := string(b_streams_bytes)
  a_map_bytes, err := json.Marshal(inp_a.StreamMap) ; if err != nil { return nil, err } ; a_map_string := string(a_map_bytes)
  b_map_bytes, err := json.Marshal(inp_b.StreamMap) ; if err != nil { return nil, err } ; b_map_string := string(b_map_bytes)

  if inp_a.Id                       != inp_b.Id                       { diff["id"]                       = inp_b.Id                       }
  if inp_a.SourceLocation           != inp_b.SourceLocation           { diff["source_location"]          = inp_b.SourceLocation           }
  if inp_a.TranscodedLocation       != inp_b.TranscodedLocation       { diff["transcoded_location"]      = inp_b.TranscodedLocation       }
  if a_streams_string               != b_streams_string               { diff["source_streams"]           = b_streams_string               }
  if a_map_string                   != b_map_string                   { diff["stream_map"]               = b_map_string                   }
  if inp_a.SourceDuration           != inp_b.SourceDuration           { diff["source_duration"]          = inp_b.SourceDuration           }
  if inp_a.TimeScanned              != inp_b.TimeScanned              { diff["time_scanned"]             = inp_b.TimeScanned              }
  if inp_a.TranscodingCommand       != inp_b.TranscodingCommand       { diff["transcoding_command"]      = inp_b.TranscodingCommand       }
  if inp_a.TranscodingTimeStarted   != inp_b.TranscodingTimeStarted   { diff["transcoding_time_started"] = inp_b.TranscodingTimeStarted   }
  if inp_a.TranscodingTimeElapsed   != inp_b.TranscodingTimeElapsed   { diff["transcoding_time_elapsed"] = inp_b.TranscodingTimeElapsed   }
  if inp_a.TranscodingError         != inp_b.TranscodingError         { diff["transcoding_error"]        = inp_b.TranscodingError         }

  return diff, nil
}

func (inp *InputFile) ValidCreate(db *sql.DB) (valid bool, err error) {
  return true, nil
}

func (inp_a *InputFile) ValidUpdate(db *sql.DB, other Table) (valid bool, err error) {
  return true, nil
}

func (inp *InputFile) ValidDelete(db *sql.DB) (valid bool, err error) {
  return true, nil
}
