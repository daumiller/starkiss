package main

import (
	"os"
  "fmt"
	"math"
  "sync"
	"time"
  "strconv"
	"strings"
	"path/filepath"
  "github.com/yargevad/filepathx"
	"github.com/daumiller/starkiss/database"
	"github.com/gofiber/fiber/v2"
	"github.com/vansante/go-ffprobe"
)

func adminScannerStartup() {
  adminScannerChannel = make(chan string)
  adminScannerMutex   = sync.Mutex{}
  go adminScannerProcess()
}

func adminScannerStatus(context *fiber.Ctx) error {
  return context.JSON(adminScanner)
}

func adminScannerStart(context *fiber.Ctx) error {
  adminScannerMutex.Lock()
  defer adminScannerMutex.Unlock()

  body := struct { Path string `json:"path"` } {}
  if err := context.BodyParser(&body); err != nil { return context.SendStatus(400) }
  if body.Path == "" { return context.Status(400).JSON(map[string]string{ "error": "missing path" }) }

  if adminScanner.Status == "running" { return context.Status(400).JSON(map[string]string{ "error": "already running" }) }
  if adminScanner.Status == "idle"    { adminScannerChannel <- body.Path }
  return context.JSON(adminScanner)
}

func adminScannerStop(context *fiber.Ctx) error {
  adminScannerMutex.Lock()
  defer adminScannerMutex.Unlock()

  if adminScanner.Status == "idle" { return context.Status(400).JSON(map[string]string{ "error": "not running" }) }
  adminScanner.ShouldQuit = true
  return context.JSON(adminScanner)
}

type adminScannerData struct {
  Status         string    `json:"status"`
  ScanPath       string    `json:"scan_path"`
  LatestFile     string    `json:"latest_file"`
  StartedTime    uint64    `json:"started_time"`
  TotalCount     uint64    `json:"total_count"`
  ProcessedCount uint64    `json:"processed_count"`
  SkippedCount   uint64    `json:"skipped_count"`
  ShouldQuit     bool      `json:"should_quit"`
}
var adminScanner        adminScannerData
var adminScannerChannel chan string
var adminScannerMutex   sync.Mutex

func adminScannerProcess() {
  for {
    adminScanner.Status         = "idle"
    adminScanner.ScanPath       = ""
    adminScanner.LatestFile     = ""
    adminScanner.StartedTime    = 0
    adminScanner.TotalCount     = 0
    adminScanner.ProcessedCount = 0
    adminScanner.SkippedCount   = 0
    adminScanner.ShouldQuit     = false

    adminScanner.ScanPath    = <- adminScannerChannel
    adminScanner.Status      = "running"
    adminScanner.StartedTime = uint64(time.Now().Unix())

    paths, err := filepathx.Glob(adminScanner.ScanPath)
    if err != nil { continue }
    adminScanner.TotalCount = uint64(len(paths))

    for _, path := range paths {
      if adminScanner.ShouldQuit == true { break }
      skip_reason := adminScannerProcessFile(path)
      adminScanner.LatestFile = path
      adminScanner.ProcessedCount += 1
      if skip_reason != "" {
        adminScanner.SkippedCount += 1
        fmt.Printf("Skipping \"%s\": (%s)\n", path, skip_reason)
      }
    }
  }
}

func adminScannerGetFps(fps_string string) uint64 {
  split := strings.Split(fps_string, "/")                 ; if len(split) != 2 { return 0 }
  numerator  , err := strconv.ParseUint(split[0], 10, 64) ; if err != nil { return 0 }
  denominator, err := strconv.ParseUint(split[1], 10, 64) ; if err != nil { return 0 }
  floating := float64(numerator) / float64(denominator)
  ceiling := math.Ceil(floating)
  return uint64(ceiling)
}

func adminScannerProcessFile(path string) (skip_reason string) {
  basename := filepath.Base(path)
  if basename[0] == '.' { return "hidden" }
  fileinfo, err := os.Stat(path)
  if err != nil { return "stat-error" }
  if fileinfo.IsDir() { return "directory" }

  probe, err := ffprobe.GetProbeData(path, time.Second * 5)
  if err != nil { return "probe-error" }
  if len(probe.Streams) < 1 { return "no-streams" }

  unproc := database.Unprocessed{}
  unproc.NeedsStreamMap     = false
  unproc.NeedsTranscoding   = false
  unproc.NeedsMetadata      = true
  unproc.SourceLocation     = path
  unproc.SourceStreams      = []database.Stream{}
  unproc.SourceContainer    = probe.Format.FormatLongName
  unproc.TranscodedLocation = ""
  unproc.TranscodedStreams  = []database.Stream{}
  unproc.MatchData          = ""
  unproc.ProvisionalId      = ""
  unproc.CreatedAt          = uint64(time.Now().Unix())

  video_usable    := false ; video_stream_count    := 0
  audio_usable    := false ; audio_stream_count    := 0
  subtitle_usable := false ; subtitle_stream_count := 0
  for _, probe_stream := range probe.Streams {
    if probe_stream.CodecType == "video" {
      video_stream := database.Stream{}
      video_stream.Type     = database.StreamTypeVideo
      video_stream.Index    = uint64(probe_stream.Index)
      video_stream.Codec    = probe_stream.CodecName
      video_stream.Width    = uint64(probe_stream.Width)
      video_stream.Height   = uint64(probe_stream.Height)
      video_stream.Fps      = adminScannerGetFps(probe_stream.AvgFrameRate)
      video_stream.Channels = 0
      video_stream.Language = ""
      unproc.SourceStreams = append(unproc.SourceStreams, video_stream)

      video_stream_count += 1
      if (video_usable == false) && (probe_stream.CodecName == "h264") { video_usable = true }
    } else if probe_stream.CodecType == "audio" {
      audio_stream := database.Stream{}
      audio_stream.Type     = database.StreamTypeAudio
      audio_stream.Index    = uint64(probe_stream.Index)
      audio_stream.Codec    = probe_stream.CodecName
      audio_stream.Width    = 0
      audio_stream.Height   = 0
      audio_stream.Fps      = 0
      audio_stream.Channels = uint64(probe_stream.Channels)
      audio_stream.Language = probe_stream.Tags.Language
      unproc.SourceStreams = append(unproc.SourceStreams, audio_stream)

      audio_stream_count += 1
      if (audio_usable == false) && (probe_stream.CodecName == "aac") { audio_usable = true }
    } else if probe_stream.CodecType == "subtitle" {
      subtitle_stream := database.Stream{}
      subtitle_stream.Type     = database.StreamTypeSubtitle
      subtitle_stream.Index    = uint64(probe_stream.Index)
      subtitle_stream.Codec    = probe_stream.CodecName
      subtitle_stream.Width    = 0
      subtitle_stream.Height   = 0
      subtitle_stream.Fps      = 0
      subtitle_stream.Channels = 0
      subtitle_stream.Language = probe_stream.Tags.Language
      unproc.SourceStreams = append(unproc.SourceStreams, subtitle_stream)

      subtitle_stream_count += 1
      if (subtitle_usable == false) && (probe_stream.CodecName == "mov_text") { subtitle_usable = true }
    }
  }

  if (video_stream_count == 0) && (audio_stream_count == 0) { fmt.Printf("No streams found: %s\n", path); return } // TODO: report skips
  if (video_stream_count    > 1) || (audio_stream_count > 1  ) { unproc.NeedsStreamMap   = true }
  if (video_stream_count    > 0) && (video_usable    == false) { unproc.NeedsTranscoding = true }
  if (audio_stream_count    > 0) && (audio_usable    == false) { unproc.NeedsTranscoding = true }
  if (subtitle_stream_count > 0) && (subtitle_usable == false) { unproc.NeedsTranscoding = true }
  if (unproc.SourceContainer != "QuickTime / MOV") { unproc.NeedsTranscoding = true }

  err = unproc.Create(DB)
  if err != nil {
    println(err.Error())
    return "database-error"
  } // TODO: report skips
  return ""
}
