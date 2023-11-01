package library

import (
  "fmt"
  "time"
  "math"
  "strconv"
  "strings"
  "github.com/vansante/go-ffprobe"
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

func FileStreamsList(path string) (file_streams []FileStream, duration int64, err error) {
  streams := []FileStream {}

  probe, err := ffprobe.GetProbeData(path, time.Second * 30)
  if err != nil             { return nil, 0, fmt.Errorf("error getting probe data: %s", err.Error()) }
  if len(probe.Streams) < 1 { return nil, 0, fmt.Errorf("no streams found in file \"%s\"", path) }

  for _, probe_stream := range probe.Streams {
    if probe_stream.CodecType == "video" {
      video_stream := FileStream{}
      video_stream.StreamType = FileStreamTypeVideo
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

      // skip these streams abused into storing thumbnails
      if (probe_stream.CodecName == "mjpeg") && (video_stream.Fps == 0) { continue }

      streams = append(streams, video_stream)
    } else if probe_stream.CodecType == "audio" {
      audio_stream := FileStream{}
      audio_stream.StreamType = FileStreamTypeAudio
      audio_stream.Index      = int64(probe_stream.Index)
      audio_stream.Codec      = probe_stream.CodecName
      audio_stream.Width      = 0
      audio_stream.Height     = 0
      audio_stream.Fps        = 0
      audio_stream.Channels   = int64(probe_stream.Channels)
      audio_stream.Language   = probe_stream.Tags.Language

      streams = append(streams, audio_stream)
    } else if probe_stream.CodecType == "subtitle" {
      subtitle_stream := FileStream{}
      subtitle_stream.StreamType = FileStreamTypeSubtitle
      subtitle_stream.Index      = int64(probe_stream.Index)
      subtitle_stream.Codec      = probe_stream.CodecName
      subtitle_stream.Width      = 0
      subtitle_stream.Height     = 0
      subtitle_stream.Fps        = 0
      subtitle_stream.Channels   = 0
      subtitle_stream.Language   = probe_stream.Tags.Language

      streams = append(streams, subtitle_stream)
    }
  }

  return streams, int64(probe.Format.DurationSeconds), nil
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
