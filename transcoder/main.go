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
	"github.com/daumiller/starkiss/database"
  "github.com/daumiller/starkiss/library"
	"github.com/schollz/progressbar/v3"
)

var DB *sql.DB = nil

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
  library.SetDatabase(DB)

  if library.MediaPathValid() == false {
    fmt.Printf("Media path is not valid.\n")
    os.Exit(-1)
  }

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

func getTask() *database.InputFile {
  inp, err := database.InputFileNextForTranscoding(DB)
  if err != nil { fmt.Printf("Error getting next input file: %s\n", err.Error()) ; os.Exit(-1) }
  return inp
}

func getArguments(inp *database.InputFile, primary_type database.FileStreamType) []string {
  arguments := []string {
    "-i", inp.SourceLocation,
    "-progress", "pipe:1",
  }
  if primary_type == database.FileStreamTypeVideo {
    arguments = append(arguments,
      "-map_metadata", "-1",
    )
  }

  has_video    := false
  has_audio    := false ; audio_mp3 := false
  has_subtitle := false

  for _, stream_index := range inp.StreamMap {
    stream := inp.SourceStreams[stream_index]
    if stream.StreamType == database.FileStreamTypeVideo    { has_video    = true }
    if stream.StreamType == database.FileStreamTypeSubtitle { has_subtitle = true }
    if stream.StreamType == database.FileStreamTypeAudio {
      has_audio = true
      if stream.Codec == "mp3" { audio_mp3 = true }
    }
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
    if primary_type == database.FileStreamTypeVideo {
      arguments = append(arguments,
        "-acodec", "aac",
        "-ac"    , "2",
      )
    }
    if primary_type == database.FileStreamTypeAudio {
      if audio_mp3 {
        arguments = append(arguments,
          "-vn",  // disable video
          "-acodec", "copy",
        )
      } else {
        arguments = append(arguments,
          "-vn",  // disable video
          "-acodec", "libmp3lame",
          "-ac"    , "2",
          "-b:a"   , "320k",
        )
      }
    }
  }
  if has_subtitle {
    arguments = append(arguments,
      "-scodec", "mov_text",
    )
  }

  return arguments
}

func setReady(inp *database.InputFile, arguments []string) {
  err := library.InputFileStart(inp, time.Now().Unix(), "ffmpeg " + strings.Join(arguments, " "))
  if err != nil { fmt.Printf("Error updating input file: %s\n", err.Error()) ; os.Exit(-1) }
}

func setFailed(inp *database.InputFile, message string) {
  fmt.Printf("Transcoding task failed: %s -- %s\n", inp.Id, message)
  err := library.InputFileFail(inp, time.Now().Unix(), message)
  if err != nil { fmt.Printf("Error updating failed transcoding task: %s\n", err.Error()) ; os.Exit(-1) }
}

func setComplete(inp *database.InputFile, output_path string, name_display string, name_sort string) {
  // get ouput size
  output_stat, err := os.Stat(output_path)
  if err != nil { setFailed(inp, fmt.Sprintf("Error getting output file size: %s\n", err.Error())) ; return }
  output_size := output_stat.Size()

  // get streams from transcoded file
  output_streams, output_duration, err := library.FileStreamsList(output_path)
  if err != nil { setFailed(inp, fmt.Sprintf("Error getting streams from transcoded file: %s\n", err.Error())) ; return }

  // create metadata record
  file_type := database.MetadataMediaTypeFileAudio
  for _, stream := range output_streams {
    if stream.StreamType == database.FileStreamTypeVideo { file_type = database.MetadataMediaTypeFileVideo ; break }
  }
  md := database.Metadata {}
  md.Id          = inp.Id
  md.ParentId    = ""
  md.MediaType   = file_type
  md.NameDisplay = name_display
  md.NameSort    = name_sort
  md.Streams     = output_streams
  md.Duration    = output_duration
  md.Size        = output_size
  err = database.MetadataCreate(DB, &md)
  if err != nil { setFailed(inp, fmt.Sprintf("Error creating metadata record: %s\n", err.Error())) ; return }

  // update InputFile record
  err = library.InputFileSucceed(inp, time.Now().Unix())
  if err != nil { setFailed(inp, fmt.Sprintf("Error updating unprocessed entry: %s\n", err.Error())) ; return }
}

func runTask(inp *database.InputFile) {
  // get output file name
  output_name_display, output_name_sort, output_path := library.InputFileOutputNames(inp)
  output_primary_type := library.InputFileOutputType(inp)
  if (output_primary_type != database.FileStreamTypeVideo) && (output_primary_type != database.FileStreamTypeAudio) {
    setFailed(inp, fmt.Sprintf("Unable to determine if output is audio/video"))
    return
  }
  if output_primary_type == database.FileStreamTypeVideo { output_path = output_path + ".mp4" }
  if output_primary_type == database.FileStreamTypeAudio { output_path = output_path + ".mp3" }

  // ensure this file doesn't already exist
  _, err := os.Stat(output_path)
  if err == nil {
    setFailed(inp, fmt.Sprintf("Unable to process file, file named \"%s\" already exists.\n", output_path))
    return
  }

  // build arguments, mark task as started
  arguments := getArguments(inp, output_primary_type)
  arguments = append(arguments, "-y", output_path)
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
  if err != nil { setFailed(inp, fmt.Sprintf("Error creating ffmpeg output pipe: %s", err.Error())) ; return }
  err = ffmpeg.Start()
  if err != nil { setFailed(inp, fmt.Sprintf("Error starting ffmpeg: %s", err.Error())) ; return }

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
  if err != nil { setFailed(inp, fmt.Sprintf("Error waiting for ffmpeg to complete: %s", err.Error())) ; return }

  setComplete(inp, output_path, output_name_display, output_name_sort)
}
