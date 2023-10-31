package library

import (
  "os"
  "fmt"
  "strings"
  "path/filepath"
  "github.com/daumiller/starkiss/database"
)

var ErrInvalidStreamIndex = fmt.Errorf("invalid stream index")
var ErrMissingVideoStream = fmt.Errorf("missing video stream")
var ErrMissingAudioStream = fmt.Errorf("missing audio stream")

func InputFileList() ([]database.InputFile, error) {
  return database.InputFileWhere(db, "")
}

func InputFileMap(inp *database.InputFile, source_stream_map []int64) error {
  has_video := false
  has_audio := false
  stream_index_map := map[int64]*(database.FileStream) {}
  for _, input_stream := range inp.SourceStreams {
    if input_stream.StreamType == database.FileStreamTypeVideo { has_video = true }
    if input_stream.StreamType == database.FileStreamTypeAudio { has_audio = true }
    stream_index_map[input_stream.Index] = &input_stream
  }
  matched_video := false
  matched_audio := false
  for _, stream_index := range source_stream_map {
    stream, ok := stream_index_map[stream_index]
    if !ok { return ErrInvalidStreamIndex }
    if stream.StreamType == database.FileStreamTypeVideo { matched_video = true }
    if stream.StreamType == database.FileStreamTypeAudio { matched_audio = true }
  }
  if has_video && !matched_video { return ErrMissingVideoStream }
  if has_audio && !matched_audio { return ErrMissingAudioStream }

  inp_update := inp.Copy()
  inp_update.StreamMap = source_stream_map
  return database.InputFileReplace(db, inp, inp_update)
}

func InputFileOutputNames(inp *database.InputFile) (name_display string, name_sort string, path string) {
  media_path, _ := MediaPathGet()

  path_base   := filepath.Base(inp.SourceLocation)
  name_display = strings.TrimSuffix(path_base, filepath.Ext(path_base))
  name_sort    = SortName(name_display)
  path = filepath.Join(media_path, name_sort)

  return name_display, name_sort, path
}
func InputFileOutputType(inp *database.InputFile) database.FileStreamType {
  has_video := false
  has_audio := false
  stream_index_map := map[int64]*(database.FileStream) {}
  for index := range inp.SourceStreams {
    if inp.SourceStreams[index].StreamType == database.FileStreamTypeVideo { has_video = true }
    if inp.SourceStreams[index].StreamType == database.FileStreamTypeAudio { has_audio = true }
    stream_index_map[inp.SourceStreams[index].Index] = &(inp.SourceStreams[index])
  }
  matched_video := false
  matched_audio := false
  for _, stream_index := range inp.StreamMap {
    stream, ok := stream_index_map[stream_index]
    if !ok { continue }
    if stream.StreamType == database.FileStreamTypeVideo { matched_video = true }
    if stream.StreamType == database.FileStreamTypeAudio { matched_audio = true }
  }
  if has_video && matched_video { return database.FileStreamTypeVideo }
  if has_audio && matched_audio { return database.FileStreamTypeAudio }
  return database.FileStreamTypeSubtitle // error state
}

func InputFileStart(inp *database.InputFile, time int64, command string) error {
  inp_update := inp.Copy()
  inp_update.TranscodingTimeStarted = time
  inp_update.TranscodingCommand     = command
  return database.InputFileReplace(db, inp, inp_update)
}

func InputFileFail(inp *database.InputFile, time int64, error string) error {
  inp_update := inp.Copy()
  inp_update.TranscodingError       = error
  inp_update.TranscodingTimeElapsed = time - inp.TranscodingTimeStarted
  return database.InputFileReplace(db, inp, inp_update)
}

func InputFileSucceed(inp *database.InputFile, time int64) error {
  inp_update := inp.Copy()
  inp_update.TranscodingError = ""
  inp_update.TranscodingTimeElapsed = time - inp.TranscodingTimeStarted
  return database.InputFileReplace(db, inp, inp_update)
}

func InputFileDidSucceed(inp *database.InputFile) bool {
  if (inp.TranscodingTimeStarted == 0) || (inp.TranscodingError != "") { return false }
  md, err := database.MetadataRead(db, inp.Id)
  if err != nil { return false }
  if (md.Duration == 0) || (md.Size == 0) { return false }
  return true
}

func InputFileReset(inp *database.InputFile) error {
  // delete existing transcoded file (if any)
  _, _, output_path := InputFileOutputNames(inp)
  _, err := os.Stat(output_path)
  if err == nil {
    err = os.Remove(output_path)
    if err != nil { return err }
  }

  // delete existing metadata record (if any)
  md, err := database.MetadataRead(db, inp.Id)
  if err == nil {
    err = database.MetadataDelete(db, md)
    if err != nil { return err }
  }

  inp_update := inp.Copy()
  inp_update.TranscodingTimeStarted = 0
  inp_update.TranscodingTimeElapsed = 0
  inp_update.TranscodingCommand     = ""
  inp_update.TranscodingError       = ""
  return database.InputFileReplace(db, inp, inp_update)
}

func InputFileDelete(inp *database.InputFile) error {
  succeeded := InputFileDidSucceed(inp)
  if succeeded == false {
    // if processing did not complete, make sure to delete metadata & output file
    err := InputFileReset(inp)
    if err != nil { return err }
  }
  return database.InputFileDelete(db, inp)
}
