package main

import (
	"os"
  "fmt"
	"time"
	"path/filepath"
  "github.com/yargevad/filepathx"
  "github.com/daumiller/starkiss/library"
)

func main() {
  db_path := os.Getenv("DBFILE")
  if db_path == "" { fmt.Printf("DBFILE environment variable not set.\n") ; os.Exit(-1) }
  err := library.LibraryStartup(db_path)
  if err != nil { fmt.Printf("Error starting library: %s\n", err.Error()) ; os.Exit(-1) }
  defer library.LibraryShutdown()
  if library.LibraryReady() != nil {
    fmt.Printf("Library not ready: %s\n", err.Error())
    os.Exit(-1)
  }

  if len(os.Args) < 2 {
    fmt.Printf("Usage: scanner <path>\n")
    fmt.Printf("  <path> is a glob pattern to match files against (ex: \"/media/**/*.mp4\")\n")
    fmt.Printf("  <path> can also be a single file\n")
    fmt.Printf("Use DBFILE environment variable to set alternate path to database.\n")
    fmt.Printf("\n")
    os.Exit(0)
  }

  paths, err := filepathx.Glob(os.Args[1])
  if err != nil { fmt.Printf("Error processing glob: %s\n", err.Error()) ; os.Exit(-1) }

  fmt.Printf("Scanning %d paths...\n", len(paths))
  for _, path := range paths {
    skipped_reason := processFile(path)
    if skipped_reason != "" { fmt.Printf("Skipped \"%s\" (%s)\n", path, skipped_reason) }
  }

  fmt.Printf("Done!\n")
}

func processFile(path string) (skip_reason string) {
  exists := library.InputFileExistsForSource(path)
  if exists { return "already-processed" }

  basename := filepath.Base(path)
  if basename[0] == '.' { return "hidden" }

  fileinfo, err := os.Stat(path)
  if err != nil { return "stat-error" }

  if fileinfo.IsDir() { return "directory" }

  source_streams, source_duration, err := library.FileStreamsList(path)
  if err != nil { return "probe-error" }
  if len(source_streams) < 1 { return "no-streams" }

  inp := library.InputFile{}
  inp.Id                     = ""
  inp.SourceLocation         = path
  inp.SourceStreams          = source_streams
  inp.StreamMap              = []int64 {}
  inp.SourceDuration         = source_duration
  inp.TimeScanned            = time.Now().Unix()
  inp.TranscodingCommand     = ""
  inp.TranscodingTimeStarted = 0
  inp.TranscodingTimeElapsed = 0
  inp.TranscodingError       = ""

  video_stream_count    := 0 ; video_stream_index    := int64(0)
  audio_stream_count    := 0 ; audio_stream_index    := int64(0)
  subtitle_stream_count := 0 ; subtitle_stream_index := int64(0)
  for _, stream := range source_streams {
    if stream.StreamType == library.FileStreamTypeVideo {
      video_stream_index = stream.Index
      video_stream_count += 1
    } else if stream.StreamType == library.FileStreamTypeAudio {
      audio_stream_index = stream.Index
      audio_stream_count += 1
    } else if stream.StreamType == library.FileStreamTypeSubtitle {
      subtitle_stream_index = stream.Index
      subtitle_stream_count += 1
    }
  }

  if (video_stream_count == 0) && (audio_stream_count == 0) { return "no a/v streams found" }

  // auto-map, for simple cases
  if (video_stream_count < 2) && (audio_stream_count < 2) && (subtitle_stream_count < 2) {
    stream_map := []int64 {}
    if (video_stream_count    == 1) { stream_map = append(stream_map, int64(video_stream_index   )) }
    if (audio_stream_count    == 1) { stream_map = append(stream_map, int64(audio_stream_index   )) }
    if (subtitle_stream_count == 1) { stream_map = append(stream_map, int64(subtitle_stream_index)) }
    inp.StreamMap = stream_map
  }

  err = library.InputFileCreate(&inp)
  if err != nil {
    println(err.Error())
    return "database-error"
  }
  return ""
}
