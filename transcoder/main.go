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
	"github.com/schollz/progressbar/v3"
)

var DB *sql.DB = nil
var MEDIAPATH string = ""

func main() {
  // command line options
  continuous := false
  if len(os.Args) == 2 {
    switch(os.Args[1]) {
      case "--continuous": continuous = true
      case "-c":           continuous = true
      default:
        fmt.Printf("Usage: transcoder (-c|--continuous)\n")
        fmt.Printf("  continuous: run continuously, polling database for new tasks\n")
        fmt.Printf("              otherwise, run until queue is empty, and exit\n")
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
  MEDIAPATH, err = database.PropertyRead(DB, "mediapath")
  if err != nil { fmt.Printf("Error reading mediapath: %s\n", err.Error()) ; os.Exit(-1) }
  if MEDIAPATH == "" { fmt.Printf("Error: media path not set in database.\n") ; os.Exit(-1) }

  // main loop
  for {
    for {
      tct, unproc := transcoderTaskGet()
      if (tct == nil) || (unproc == nil) { fmt.Printf("Transcoder queue empty...\n") ; break }
      transcoderTaskRun(tct, unproc)
    }
    if continuous == false { break }
    time.Sleep(5 * time.Second)
  }
}

func transcoderTaskGet() (tct *database.TranscodingTask, unproc *database.Unprocessed) {
  var err error
  // get next task that is "todo", fetching this task will automatically mark it as "running"
  tct, err = database.TranscodingTaskNext(DB)
  if err == database.ErrNotFound { return nil, nil }
  if err != nil { fmt.Printf("Error getting next transcoding task: %s\n", err.Error()) ; os.Exit(-1) }

  // get corresponding unprocessed entry, if it doesn't exist, mark task as "failure"
  unproc_item, err := database.UnprocessedRead(DB, tct.UnprocessedId)
  if err == database.ErrNotFound {
    fmt.Printf("Task references missing Unprocessed entry: %s\n", tct.UnprocessedId)
    tct_update := tct.Copy()
    tct_update.Status = database.TranscodingTaskFailure
    tct_update.ErrorMessage = "Task references missing Unprocessed entry"
    err = tct.Update(DB, tct_update)
    if err != nil { fmt.Printf("Error updating transcoding task: %s\n", err.Error()) ; os.Exit(-1) }
    return nil, nil
  } else if err != nil {
    fmt.Printf("Error reading Unprocessed entry: %s\n", err.Error()) ; os.Exit(-1)
  }

  return tct, &unproc_item
}

func transcoderTaskArguments(tct *database.TranscodingTask, unproc *database.Unprocessed) (arguments []string) {
  /*
    ffmpeg
    -i "/Volumes/Starkiss/import/Video/Documentary/BBS - The Documentary/Season 1/01 Baud.avi"
    -progress pipe:1
    -vcodec h264
    -acodec aac
    -map 0:0
    -map 0:1
    -y "/Volumes/Starkiss/.trans/a2919edf-11b3-41db-83bd-a6a849f8ddbb.mp4"
  */
  arguments = []string { "-i", unproc.SourceLocation, "-progress", "pipe:1" }

  has_video    := false ; copy_video    := true
  has_audio    := false ; copy_audio    := true
  has_subtitle := false ; copy_subtitle := true

  for index := range unproc.TranscodedStreams {
    stream := unproc.SourceStreams[unproc.TranscodedStreams[index]]
    if stream.Type == "video"    { has_video    = true ; if stream.Codec != "h264"     { copy_video    = false } }
    if stream.Type == "audio"    { has_audio    = true ; if stream.Codec != "aac"      { copy_audio    = false } ; if stream.Channels != 2 { copy_audio = false } }
    if stream.Type == "subtitle" { has_subtitle = true ; if stream.Codec != "mov_text" { copy_subtitle = false } }
    arguments = append(arguments, "-map", "0:" + strconv.Itoa(int(stream.Index)))
  }
  if has_video {
    arguments = append(arguments, "-vcodec")
    if copy_video { arguments = append(arguments, "copy") } else { arguments = append(arguments, "h264") }
  }
  if has_audio {
    arguments = append(arguments, "-acodec")
    if copy_audio { arguments = append(arguments, "copy") } else { arguments = append(arguments, "aac") }
    arguments = append(arguments, "-ac", "2")
  }
  if has_subtitle {
    arguments = append(arguments, "-scodec")
    if copy_subtitle { arguments = append(arguments, "copy") } else { arguments = append(arguments, "mov_text") }
  }

  return arguments
}

func transcoderTaskFail(tct *database.TranscodingTask, unproc *database.Unprocessed, message string) {
  fmt.Printf("Transcoding task failed: %s -- %s\n", tct.Id, message)
  tct_update := tct.Copy()
  tct_update.Status = database.TranscodingTaskFailure
  tct_update.ErrorMessage = message
  err := tct.Update(DB, tct_update)
  if err != nil { fmt.Printf("Error updating failed transcoding task: %s\n", err.Error()) ; os.Exit(-1) }
}

func transcoderTaskRun(tct *database.TranscodingTask, unproc *database.Unprocessed) {
  // prep output
  fmt.Printf("Processing \"%s\"...\n", unproc.SourceLocation)
  progress_bar := progressbar.NewOptions64(
    unproc.Duration * 1000,
    progressbar.OptionSetWidth(64),
    progressbar.OptionSetDescription("Transcoding"),
  )

  // get all arguments for ffmpeg
  arguments := transcoderTaskArguments(tct, unproc)
  transcoded_location := MEDIAPATH + "/.trans/" + unproc.Id + ".mp4"
  arguments = append(arguments, "-y", transcoded_location)
  
  // update record for starting process
  tct_update := tct.Copy()
  tct_update.TimeStarted = time.Now().Unix()
  tct_update.CommandLine = "ffmpeg " + strings.Join(arguments, " ")
  err := tct.Update(DB, tct_update)
  if err != nil { transcoderTaskFail(tct, unproc, fmt.Sprintf("Error updating transcoding task: %s\n", err.Error())) ; return }
  tct = tct_update.Copy()

  // run ffmpeg
  ffmpeg := exec.Command("ffmpeg", arguments...)
  ffmpeg_output, err := ffmpeg.StdoutPipe()
  if err != nil { transcoderTaskFail(tct, unproc, fmt.Sprintf("Error creating ffmpeg output pipe: %s\n", err.Error())) ; return }
  err = ffmpeg.Start()
  if err != nil { transcoderTaskFail(tct, unproc, fmt.Sprintf("Error starting ffmpeg: %s\n", err.Error())) ; return }

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
  if err != nil { transcoderTaskFail(tct, unproc, fmt.Sprintf("Error waiting for ffmpeg to complete: %s\n", err.Error())) ; return }

  // update unproc record
  unproc_update := unproc.Copy()
  unproc_update.NeedsTranscoding = false
  unproc_update.TranscodedLocation = transcoded_location
  err = unproc.Update(DB, unproc_update)
  if err != nil { transcoderTaskFail(tct, unproc, fmt.Sprintf("Error updating unprocessed entry: %s\n", err.Error())) ; return }

  // update tct record
  tct_update.Status = database.TranscodingTaskSuccess
  tct_update.TimeElapsed = time.Now().Unix() - tct_update.TimeStarted
  tct_update.ErrorMessage = ""
  err = tct.Update(DB, tct_update)
  if err != nil { transcoderTaskFail(tct, unproc, fmt.Sprintf("Error updating transcoding task to success: %s\n", err.Error())) ; return }
}
