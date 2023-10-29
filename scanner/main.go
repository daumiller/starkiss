package main

import (
	"os"
  "fmt"
	"math"
	"time"
  "strconv"
	"strings"
	"path/filepath"
  "database/sql"
  "github.com/google/uuid"
  "github.com/yargevad/filepathx"
	"github.com/daumiller/starkiss/database"
	"github.com/vansante/go-ffprobe"
)

var DB *sql.DB = nil

func main() {
  if os.Getenv("DBFILE") != "" { database.Location = os.Getenv("DBFILE") }

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

  DB, err = database.Open()
  if err != nil { fmt.Printf("Error opening database: %s\n", err.Error()) ; os.Exit(-1) }
  defer DB.Close()

  fmt.Printf("Scanning %d paths...\n", len(paths))
  for _, path := range paths {
    skipped_reason := processFile(path)
    if skipped_reason != "" { fmt.Printf("Skipped \"%s\" (%s)\n", path, skipped_reason) }
  }

  fmt.Printf("Done!\n")
}

func convertFpsString(fps_string string) int64 {
  split := strings.Split(fps_string, "/")                ; if len(split) != 2 { return 0 }
  numerator  , err := strconv.ParseInt(split[0], 10, 64) ; if err != nil { return 0 }
  denominator, err := strconv.ParseInt(split[1], 10, 64) ; if err != nil { return 0 }
  if denominator < 1 { return 0 }
  floating := float64(numerator) / float64(denominator)
  ceiling := math.Ceil(floating)
  // we have a value, but it's often wrong, do basic checks
  if ceiling < 10.0  { return 0 }
  if ceiling > 320.0 { return 0 } // found some invalid files with things like 90000/1
  return int64(ceiling)
}

func processFile(path string) (skip_reason string) {
  lookup_row := DB.QueryRow("SELECT id FROM input_files WHERE source_location = ?", path)
  lookup_id := ""
  err := lookup_row.Scan(&lookup_id)
  if err == nil { return "already-processed" }

  basename := filepath.Base(path)
  if basename[0] == '.' { return "hidden" }

  fileinfo, err := os.Stat(path)
  if err != nil { return "stat-error" }

  if fileinfo.IsDir() { return "directory" }

  probe, err := ffprobe.GetProbeData(path, time.Second * 30)
  if err != nil { return "probe-error" }
  if len(probe.Streams) < 1 { return "no-streams" }

  inp := database.InputFile{}
  inp.Id                     = uuid.NewString()
  inp.SourceLocation         = path
  inp.TranscodedLocation     = ""
  inp.SourceStreams          = []database.FileStream {}
  inp.StreamMap              = []int64 {}
  inp.SourceDuration         = int64(probe.Format.DurationSeconds)
  inp.TimeScanned            = time.Now().Unix()
  inp.TranscodingCommand     = ""
  inp.TranscodingTimeStarted = 0
  inp.TranscodingTimeElapsed = 0
  inp.TranscodingError       = ""

  video_stream_count    := 0 ; video_stream_index    := 0
  audio_stream_count    := 0 ; audio_stream_index    := 0
  subtitle_stream_count := 0 ; subtitle_stream_index := 0
  for _, probe_stream := range probe.Streams {
    if probe_stream.CodecType == "video" {
      video_stream := database.FileStream{}
      video_stream.StreamType = database.FileStreamTypeVideo
      video_stream.Index      = int64(probe_stream.Index)
      video_stream.Codec      = probe_stream.CodecName
      video_stream.Width      = int64(probe_stream.Width)
      video_stream.Height     = int64(probe_stream.Height)
      video_stream.Fps        = 0
      video_stream.Channels   = 0
      video_stream.Language   = ""

      r_fps := convertFpsString(probe_stream.RFrameRate)
      a_fps := convertFpsString(probe_stream.AvgFrameRate)
      if (r_fps  > 0) && (a_fps == 0) { video_stream.Fps = r_fps }
      if (r_fps == 0) && (a_fps  > 0) { video_stream.Fps = a_fps }
      if (r_fps  > 0) && (a_fps  > 0) { video_stream.Fps = int64(math.Min(float64(r_fps), float64(a_fps))) }

      inp.SourceStreams = append(inp.SourceStreams, video_stream)
      video_stream_index = probe_stream.Index
      video_stream_count += 1
    } else if probe_stream.CodecType == "audio" {
      audio_stream := database.FileStream{}
      audio_stream.StreamType = database.FileStreamTypeAudio
      audio_stream.Index      = int64(probe_stream.Index)
      audio_stream.Codec      = probe_stream.CodecName
      audio_stream.Width      = 0
      audio_stream.Height     = 0
      audio_stream.Fps        = 0
      audio_stream.Channels   = int64(probe_stream.Channels)
      audio_stream.Language   = probe_stream.Tags.Language

      inp.SourceStreams = append(inp.SourceStreams, audio_stream)
      audio_stream_index = probe_stream.Index
      audio_stream_count += 1
    } else if probe_stream.CodecType == "subtitle" {
      subtitle_stream := database.FileStream{}
      subtitle_stream.StreamType = database.FileStreamTypeSubtitle
      subtitle_stream.Index      = int64(probe_stream.Index)
      subtitle_stream.Codec      = probe_stream.CodecName
      subtitle_stream.Width      = 0
      subtitle_stream.Height     = 0
      subtitle_stream.Fps        = 0
      subtitle_stream.Channels   = 0
      subtitle_stream.Language   = probe_stream.Tags.Language

      inp.SourceStreams = append(inp.SourceStreams, subtitle_stream)
      subtitle_stream_index = probe_stream.Index
      subtitle_stream_count += 1
    }
  }

  if (video_stream_count == 0) && (audio_stream_count == 0) { return "no streams found" }

  // auto-map, for simple cases
  if (video_stream_count < 2) && (audio_stream_count < 2) && (subtitle_stream_count < 2) {
    stream_map := []int64 {}
    if (video_stream_count    == 1) { stream_map = append(stream_map, int64(video_stream_index   )) }
    if (audio_stream_count    == 1) { stream_map = append(stream_map, int64(audio_stream_index   )) }
    if (subtitle_stream_count == 1) { stream_map = append(stream_map, int64(subtitle_stream_index)) }
    inp.StreamMap = stream_map
  }

  err = inp.Create(DB)
  if err != nil {
    println(err.Error())
    return "database-error"
  }
  return ""
}
