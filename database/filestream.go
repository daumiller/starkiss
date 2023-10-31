package database

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
