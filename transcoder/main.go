package main

/*
Transcoding Task
  id, unprocessed_id, time_started, time_ran, status (todo, running, success, failure), error_message, command_line 

Separate Transcoding Process, launched externally to web app
  pick up (status == "todo") jobs, mark it in progress, run transcoding, read result, update task & unprocessed records
Admin app just adds tasks to be ran later
*/

