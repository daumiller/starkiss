package library

import (
  "os"
  "fmt"
  "strings"
  "encoding/json"
  "path/filepath"
)

type InputFile struct {
  Id                       string       `json:"id"`                        // Metadata.Id == InputFile.Id
  SourceLocation           string       `json:"source_location"`           // path to source file
  SourceStreams            []FileStream `json:"source_streams"`
  StreamMap                []int64      `json:"stream_map"`                // empty == needs_map
  SourceDuration           int64        `json:"source_duration"`           // length of media in seconds
  TimeScanned              int64        `json:"time_scanned"`
  TranscodingCommand       string       `json:"transcoding_command"`       // ffmpeg command line
  TranscodingTimeStarted   int64        `json:"transcoding_time_started"`  // time transcoding was started
  TranscodingTimeElapsed   int64        `json:"transcoding_time_elapsed"`  // seconds elapsed during transcoding
  TranscodingError         string       `json:"transcoding_error"`         // error message from transcoding process
}

var ErrInvalidStreamIndex = fmt.Errorf("invalid stream index")
var ErrMissingVideoStream = fmt.Errorf("missing video stream")
var ErrMissingAudioStream = fmt.Errorf("missing audio stream")

// ============================================================================
// Public Interface

func (inp *InputFile) Copy() (*InputFile) {
  copy := InputFile {}
  copy.Id                       = inp.Id
  copy.SourceLocation           = inp.SourceLocation
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

func (inp *InputFile) Remap(source_stream_map []int64) error {
  has_video := false
  has_audio := false
  stream_index_map := map[int64]*(FileStream) {}
  for source_array_index, input_stream := range inp.SourceStreams {
    if input_stream.StreamType == FileStreamTypeVideo { has_video = true }
    if input_stream.StreamType == FileStreamTypeAudio { has_audio = true }
    stream_index_map[input_stream.Index] = &(inp.SourceStreams[source_array_index])
  }
  matched_video := false
  matched_audio := false
  for _, stream_index := range source_stream_map {
    stream, ok := stream_index_map[stream_index]
    if !ok { return ErrInvalidStreamIndex }
    if stream.StreamType == FileStreamTypeVideo { matched_video = true }
    if stream.StreamType == FileStreamTypeAudio { matched_audio = true }
  }
  if has_video && !matched_video { return ErrMissingVideoStream }
  if has_audio && !matched_audio { return ErrMissingAudioStream }

  inp_update := inp.Copy()
  inp_update.StreamMap = source_stream_map

  err := dbRecordReplace(inp, inp_update)
  if err != nil { return ErrQueryFailed }
  return nil
}

func (inp *InputFile) OutputNames() (name_display string, name_sort string, path string) {
  output_type := inp.OutputType()
  output_extension := ""
  if output_type == FileStreamTypeVideo { output_extension = ".mp4" }
  if output_type == FileStreamTypeAudio { output_extension = ".mp3" }

  path_base   := filepath.Base(inp.SourceLocation)
  name_display = strings.TrimSuffix(path_base, filepath.Ext(path_base))
  name_sort    = nameGetSortForDisplay(name_display)
  name_sort    = strings.TrimPrefix(name_sort, "the ") // very basic cleanup, first time InputFile->Metadata only
  path = filepath.Join(mediaPath, name_sort) + output_extension

  return name_display, name_sort, path
}

func (inp *InputFile) OutputType() FileStreamType {
  has_video := false
  has_audio := false
  stream_index_map := map[int64]*(FileStream) {}
  for index := range inp.SourceStreams {
    if inp.SourceStreams[index].StreamType == FileStreamTypeVideo { has_video = true }
    if inp.SourceStreams[index].StreamType == FileStreamTypeAudio { has_audio = true }
    stream_index_map[inp.SourceStreams[index].Index] = &(inp.SourceStreams[index])
  }
  matched_video := false
  matched_audio := false
  for _, stream_index := range inp.StreamMap {
    stream, ok := stream_index_map[stream_index]
    if !ok { continue }
    if stream.StreamType == FileStreamTypeVideo { matched_video = true }
    if stream.StreamType == FileStreamTypeAudio { matched_audio = true }
  }
  if has_video && matched_video { return FileStreamTypeVideo }
  if has_audio && matched_audio { return FileStreamTypeAudio }
  return FileStreamTypeSubtitle // error state
}

func (inp *InputFile) StatusSetStarted(time int64, command string) error {
  inp_update := inp.Copy()
  inp_update.TranscodingTimeStarted = time
  inp_update.TranscodingCommand     = command
  err := dbRecordReplace(inp, inp_update)
  if err != nil { return ErrQueryFailed }
  return nil
}

func (inp *InputFile) StatusSetFailed(time int64, error string) error {
  inp_update := inp.Copy()
  inp_update.TranscodingError       = error
  inp_update.TranscodingTimeElapsed = time - inp.TranscodingTimeStarted
  err := dbRecordReplace(inp, inp_update)
  if err != nil { return ErrQueryFailed }
  return nil
}

func (inp *InputFile) StatusSetSucceeded(time int64) error {
  inp_update := inp.Copy()
  inp_update.TranscodingError = ""
  inp_update.TranscodingTimeElapsed = time - inp.TranscodingTimeStarted
  if inp_update.TranscodingTimeElapsed < 1 { inp_update.TranscodingTimeElapsed = 1 }
  err := dbRecordReplace(inp, inp_update)
  if err != nil { return ErrQueryFailed }
  return nil
}

func (inp *InputFile) StatusDidSucceed() bool {
  if (inp.TranscodingTimeStarted == 0) || (inp.TranscodingError != "") { return false }
  md := Metadata {}
  err := dbRecordRead(&md, inp.Id)
  if err != nil { return false }
  if (md.Duration == 0) || (md.Size == 0) { return false }
  return true
}

func (inp *InputFile) StatusReset() error {
  // delete existing transcoded file (if any)
  _, _, output_path := inp.OutputNames()
  if pathExists(output_path) {
    err := os.Remove(output_path)
    if err != nil { return fmt.Errorf("error deleting transcoded file: %s", err.Error()) }
  }

  // delete existing metadata record (if any)
  md := Metadata {}
  err := dbRecordRead(&md, inp.Id)
  if err == nil {
    err = dbRecordDelete(&md)
    if err != nil { return fmt.Errorf("error deleting metadata record: %s", err.Error()) }
  }

  inp_update := inp.Copy()
  inp_update.TranscodingTimeStarted = 0
  inp_update.TranscodingTimeElapsed = 0
  inp_update.TranscodingCommand     = ""
  inp_update.TranscodingError       = ""
  err = dbRecordReplace(inp, inp_update)
  if err != nil { return ErrQueryFailed }
  return nil
}

func InputFileCreate(inp *InputFile) error {
  err := dbRecordCreate(inp)
  if err != nil { return ErrQueryFailed }
  return nil
}

func InputFileRead(id string) (*InputFile, error) {
  inp := InputFile {}
  err := dbRecordRead(&inp, id)
  if err != nil { return nil, ErrNotFound }
  return &inp, nil
}

func InputFileList() ([]InputFile, error) {
  records, err := dbRecordWhere(&InputFile{}, ``)
  if err != nil { return nil, ErrQueryFailed }
  inputs := make([]InputFile, len(records))
  for index, record := range records { inputs[index] = *(record.(*InputFile)) }
  return inputs, nil
}

func InputFileDelete(inp *InputFile) error {
  if inp.StatusDidSucceed() == false {
    // if processing did not complete, make sure to delete metadata & output file
    err := inp.StatusReset()
    if err != nil { return err }
  }
  err := dbRecordDelete(inp)
  if err != nil { return ErrQueryFailed }
  return nil
}

func InputFileExistsForSource(source_location string) bool {
  queryRow := dbHandle.QueryRow(`SELECT id FROM input_files WHERE source_location = ? LIMIT 1;`, source_location)
  err := queryRow.Scan(&source_location)
  return (err == nil)
}

func InputFileNextForTranscoding() (*InputFile, error) {
  records, err := dbRecordWhere(&InputFile{}, `(transcoding_time_started = 0) AND (stream_map <> '[]') AND (transcoding_error = '') LIMIT 1`)
  if err != nil { return nil, err }
  if len(records) == 0 { return nil, nil }
  return records[0].(*InputFile), nil
}

// ============================================================================
// dbRecord interface

func (inp *InputFile) TableName() string { return "input_files"}
func (inp *InputFile) GetId() string { return inp.Id }
func (inp *InputFile) SetId(id string) { inp.Id = id }

func (inp *InputFile) RecordCopy() (dbRecord, error) {
  return inp.Copy(), nil
}

func (inp *InputFile) RecordCreate(fields map[string]any) (instance dbRecord, err error) {
  new_instance := InputFile {}
  err = new_instance.FieldsReplace(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (inp *InputFile) FieldsRead() (fields map[string]any, err error) {
  streams_bytes, err := json.Marshal(inp.SourceStreams) ; if err != nil { return nil, err } ; streams_string := string(streams_bytes)
  map_bytes, err := json.Marshal(inp.StreamMap) ; if err != nil { return nil, err } ; map_string := string(map_bytes)

  fields = make(map[string]any)
  fields["id"]                       = inp.Id
  fields["source_location"]          = inp.SourceLocation
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

func (inp *InputFile) FieldsReplace(fields map[string]any) (err error) {
  streams_string := fields["source_streams"].(string) ; var source_streams []FileStream ; err = json.Unmarshal([]byte(streams_string), &source_streams) ; if err != nil { return err }
  map_string := fields["stream_map"].(string) ; var stream_map []int64 ; err = json.Unmarshal([]byte(map_string), &stream_map) ; if err != nil { return err }

  inp.Id                     = fields["id"].(string)
  inp.SourceLocation         = fields["source_location"].(string)
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

func (inp *InputFile) FieldsPatch(fields map[string]any) (err error) {
  if id,                       ok := fields["id"]                       ; ok { inp.Id                     = id.(string)                      }
  if source_location,          ok := fields["source_location"]          ; ok { inp.SourceLocation         = source_location.(string)         }
  if source_duration,          ok := fields["source_duration"]          ; ok { inp.SourceDuration         = source_duration.(int64)          }
  if time_scanned,             ok := fields["time_scanned"]             ; ok { inp.TimeScanned            = time_scanned.(int64)             }
  if transcoding_command,      ok := fields["transcoding_command"]      ; ok { inp.TranscodingCommand     = transcoding_command.(string)     }
  if transcoding_time_started, ok := fields["transcoding_time_started"] ; ok { inp.TranscodingTimeStarted = transcoding_time_started.(int64) }
  if transcoding_time_elapsed, ok := fields["transcoding_time_elapsed"] ; ok { inp.TranscodingTimeElapsed = transcoding_time_elapsed.(int64) }
  if transcoding_error,        ok := fields["transcoding_error"]        ; ok { inp.TranscodingError       = transcoding_error.(string)       }

  if source_streams, ok := fields["source_streams"] ; ok {
    streams_string := source_streams.(string) ; var source_streams []FileStream ; err = json.Unmarshal([]byte(streams_string), &source_streams) ; if err != nil { return err }
    inp.SourceStreams = source_streams
  }

  if stream_map, ok := fields["stream_map"] ; ok {
    map_string := stream_map.(string) ; var stream_map []int64 ; err = json.Unmarshal([]byte(map_string), &stream_map) ; if err != nil { return err }
    inp.StreamMap = stream_map
  }

  return nil
}

func (inp_a *InputFile) FieldsDifference(other dbRecord) (diff map[string]any, err error) {
  diff = make(map[string]any)
  inp_b, b_is_inp := other.(*InputFile)
  if b_is_inp == false { return diff, ErrInvalidType }

  a_streams_bytes, err := json.Marshal(inp_a.SourceStreams) ; if err != nil { return nil, err } ; a_streams_string := string(a_streams_bytes)
  b_streams_bytes, err := json.Marshal(inp_b.SourceStreams) ; if err != nil { return nil, err } ; b_streams_string := string(b_streams_bytes)
  a_map_bytes, err := json.Marshal(inp_a.StreamMap) ; if err != nil { return nil, err } ; a_map_string := string(a_map_bytes)
  b_map_bytes, err := json.Marshal(inp_b.StreamMap) ; if err != nil { return nil, err } ; b_map_string := string(b_map_bytes)

  if inp_a.Id                       != inp_b.Id                       { diff["id"]                       = inp_b.Id                       }
  if inp_a.SourceLocation           != inp_b.SourceLocation           { diff["source_location"]          = inp_b.SourceLocation           }
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
