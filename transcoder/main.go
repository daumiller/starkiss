package main

import (
	"os"
  "fmt"
  "time"
  "bufio"
	"os/exec"
	"strconv"
	"strings"
  "database/sql"
  "path/filepath"
	"github.com/daumiller/starkiss/database"
	"github.com/schollz/progressbar/v3"
)

var DB *sql.DB = nil
var MEDIA_PATH string = ""
var TRANS_PATH string = ""

func main() {
  // command line options
  continuous := false
  stop       := false
  if len(os.Args) == 2 {
    switch(os.Args[1]) {
      case "--continuous": continuous = true
      case "-c":           continuous = true
      case "--stop":       stop = true
      case "-s":           stop = true
      default:
        fmt.Printf("Usage: transcoder (-c|--continuous) (-s|--stop)\n")
        fmt.Printf("  continuous: run continuously, polling database for new tasks\n")
        fmt.Printf("              otherwise, run until queue is empty, and exit\n")
        fmt.Printf("  stop:       set transcoder stop value in database, and exit\n")
        fmt.Printf("              this will stop a running transcoder, once its current task is completed\n")
        fmt.Printf("\n")
        os.Exit(0)
    }
  }
  // transcoding destination path
  if os.Getenv("TRANS_PATH") != "" { TRANS_PATH = os.Getenv("TRANS_PATH") }
  if TRANS_PATH == "" {
    fmt.Printf("Error: TRANS_PATH environment variable not set.\n")
    os.Exit(-1)
  }

  // database
  if os.Getenv("DBFILE") != "" { database.Location = os.Getenv("DBFILE") }
  _, err := os.Stat(database.Location)
  if os.IsNotExist(err) {
    fmt.Printf("Database file \"%s\" does not exist.\n", database.Location)
    os.Exit(-1)
  }
  DB, err = database.Open()
  if err != nil { fmt.Printf("Error opening database: %s\n", err.Error()) ; os.Exit(-1) }
  defer DB.Close()
  MEDIA_PATH, err = database.PropertyRead(DB, "media_path")
  if err != nil { fmt.Printf("Error reading media_path: %s\n", err.Error()) ; os.Exit(-1) }
  if MEDIA_PATH == "" { fmt.Printf("Error: media_path not set in database.\n") ; os.Exit(-1) }

  if stop {
    setStopValue()
    os.Exit(0)
  }

  // get current transcoder termination value
  stop_value := getStopValue()

  // main loop
  for {
    for {
      inp := getTask()
      if inp == nil { fmt.Printf("Transcoder queue empty...\n") ; break }
      runTask(inp)
      if getStopValue() != stop_value { break }
    }
    if continuous == false { break }
    time.Sleep(5 * time.Second)
    if getStopValue() != stop_value { break }
  }
}

func setStopValue() {
  err := database.PropertyUpsert(DB, "transcoder_stop", strconv.FormatInt(time.Now().Unix(), 10))
  if err != nil { fmt.Printf("Error setting transcoder stop value: %s\n", err.Error()) ; os.Exit(-1) }
}

func getStopValue() string {
  stop_string, err := database.PropertyRead(DB, "transcoder_stop")
  if err != nil { return "" }
  return stop_string
}

func getTask() (inp *database.InputFile) {
  var err error

  // find InputFile where (TranscodedLocation = "") && (TranscodingTimeStarted = 0); if none, return nil
  inp_row := DB.QueryRow(`
    SELECT id FROM input_files WHERE
      (transcoded_location = '') AND
      (transcoding_time_started = 0) AND
      (stream_map <> '[]')
    LIMIT 1;`)
  var inp_id string = ""
  err = inp_row.Scan(&inp_id)
  if err == sql.ErrNoRows { return nil }
  if err != nil { fmt.Printf("Error getting next input file: %s\n", err.Error()) ; os.Exit(-1) }

  // return InputFile
  inp, err = database.InputFileRead(DB, inp_id)
  if err != nil { fmt.Printf("Error reading input file: %s\n", err.Error()) ; os.Exit(-1) }
  return inp
}

func getArguments(inp *database.InputFile) []string {
  arguments := []string {
    "-i", inp.SourceLocation,
    "-progress", "pipe:1",
    "-map_metadata", "-1",
  }

  has_video    := false
  has_audio    := false
  has_subtitle := false

  for _, stream_index := range inp.StreamMap {
    stream := inp.SourceStreams[stream_index]
    if stream.StreamType == database.FileStreamTypeVideo    { has_video    = true }
    if stream.StreamType == database.FileStreamTypeAudio    { has_audio    = true }
    if stream.StreamType == database.FileStreamTypeSubtitle { has_subtitle = true }
    arguments = append(arguments, "-map", "0:" + strconv.Itoa(int(stream_index)))
  }

  if has_video {
    arguments = append(arguments,
      "-vcodec"   , "libx264",
      "-preset"   , "slower",
      "-crf"      , "21",
      "-pix_fmt"  , "yuv420p",
      "-profile:v", "high",
      "-level"    , "4.0",
      "-movflags" , "+faststart",
    )
  }
  if has_audio {
    arguments = append(arguments,
      "-acodec", "aac",
      "-ac"    , "2",
    )
  }
  if has_subtitle {
    arguments = append(arguments,
      "-scodec", "mov_text",
    )
  }

  arguments = append(arguments,
    "-y", filepath.Join(TRANS_PATH, inp.Id + ".mp4"),
  )

  return arguments
}

func setReady(inp *database.InputFile, arguments []string) {
  inp_update := inp.Copy()
  inp_update.TranscodingTimeStarted = time.Now().Unix()
  inp_update.TranscodedLocation     = filepath.Join(TRANS_PATH, inp.Id + ".mp4")
  inp_update.TranscodingCommand     = "ffmpeg " + strings.Join(arguments, " ")
  err := inp.Replace(DB, inp_update)
  if err != nil { fmt.Printf("Error updating input file: %s\n", err.Error()) ; os.Exit(-1) }
}

func setFailed(inp *database.InputFile, message string) {
  fmt.Printf("Transcoding task failed: %s -- %s\n", inp.Id, message)
  inp_update := inp.Copy()
  inp_update.TranscodingError = message
  inp_update.TranscodingTimeElapsed = time.Now().Unix() - inp.TranscodingTimeStarted
  err := inp.Replace(DB, inp_update)
  if err != nil { fmt.Printf("Error updating failed transcoding task: %s\n", err.Error()) ; os.Exit(-1) }
}

func runTask(inp *database.InputFile) {
  // build arguments, mark task as started
  arguments := getArguments(inp)
  setReady(inp, arguments)

  // prep output display
  fmt.Printf("Processing \"%s\"...\n", inp.SourceLocation)
  progress_bar := progressbar.NewOptions64(
    inp.SourceDuration * 1000,
    progressbar.OptionSetWidth(64),
    progressbar.OptionSetDescription("Transcoding"),
  )

  // run ffmpeg
  ffmpeg := exec.Command("ffmpeg", arguments...)
  ffmpeg_output, err := ffmpeg.StdoutPipe()
  if err != nil { setFailed(inp, fmt.Sprintf("Error creating ffmpeg output pipe: %s\n", err.Error())) ; return }
  err = ffmpeg.Start()
  if err != nil { setFailed(inp, fmt.Sprintf("Error starting ffmpeg: %s\n", err.Error())) ; return }

  // monitor progress
  ffmpeg_reader := bufio.NewReader(ffmpeg_output)
  for {
    time.Sleep(500 * time.Millisecond) // default stats_period for ffmpeg is 0.5 seconds
    ffmpeg_completed := false
    for {
      line, err := ffmpeg_reader.ReadString('\n')
      if err != nil { time.Sleep(100 * time.Millisecond); continue }

      // ideally, we'd use "out_time_ms=", but ffmpeg is broken: https://trac.ffmpeg.org/ticket/7345
      if strings.HasPrefix(line, "out_time_us=") {
        timestamp_string := strings.TrimSuffix(strings.TrimPrefix(line, "out_time_us="), "\n")
        timestamp_us, _ := strconv.ParseInt(timestamp_string, 10, 64)
        progress_bar.Set64(timestamp_us / 1000)
      }
      if line == "progress=end\n" { ffmpeg_completed=true ; break }
    }
    if ffmpeg_completed { break }
  }
  progress_bar.Finish()
  fmt.Printf("\n")
  err = ffmpeg.Wait()
  if err != nil { setFailed(inp, fmt.Sprintf("Error waiting for ffmpeg to complete: %s\n", err.Error())) ; return }

  // update inp record
  inp_update := inp.Copy()
  inp_update.TranscodingError = ""
  inp_update.TranscodingTimeElapsed = time.Now().Unix() - inp.TranscodingTimeStarted
  err = inp.Replace(DB, inp_update)
  if err != nil { setFailed(inp, fmt.Sprintf("Error updating unprocessed entry: %s\n", err.Error())) ; return }
}
