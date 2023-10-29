package main

import (
  "database/sql"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

func adminUnprocessedList(context *fiber.Ctx) error {
  unprocessed, err := database.UnprocessedList(DB, false, false, false)
  if err != nil { return debug500(context, err) }
  return context.JSON(unprocessed)
}

func adminUnprocessedEmpty(context *fiber.Ctx) error {
  _, err := DB.Exec(`DELETE FROM unprocessed;`)
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

func adminUnprocessedMap(context *fiber.Ctx) error {
  id := context.Params("id")
  unproc, err := database.UnprocessedRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  // read stream map
  stream_map := []int64 {}
  if err := context.BodyParser(&stream_map); err != nil { return context.Status(400).JSON(map[string]string { "error": "error parsing json" }) }

  // validate stream map
  if len(stream_map) == 0 { return context.Status(400).JSON(map[string]string { "error": "stream map is empty" }) }
  for _, stream_index := range stream_map {
    if stream_index < 0 { return context.Status(400).JSON(map[string]string { "error": "stream map contains negative values" }) }
    if stream_index >= int64(len(unproc.SourceStreams)) { return context.Status(400).JSON(map[string]string { "error": "stream map contains out-of-bounds values" }) }
  }

  // set new mapping
  unproc_update := unproc.Copy()
  unproc_update.TranscodedStreams = stream_map
  if len(stream_map) > 0 { unproc_update.NeedsStreamMap = false }
  err = unproc.Update(DB, unproc_update)
  if err != nil { return debug500(context, err) }

  return context.SendStatus(200)
}

func adminUnprocessedQueue(context *fiber.Ctx) error {
  id_list := []string {}
  if err := context.BodyParser(&id_list); err != nil { return context.Status(400).JSON(map[string]string { "error": "error parsing json" }) }
  if len(id_list) == 0 { return context.SendStatus(200) }

  // validate all items first
  errors := map[string]string {}
  for _, id := range id_list {
    unproc, err := database.UnprocessedRead(DB, id)
    if err == database.ErrNotFound { errors[id] = "item not found"; continue }
    if err != nil { return debug500(context, err) }

    // validate item needs transcoding
    if !unproc.NeedsTranscoding {
      errors[id] = "item does not need transcoding"
      continue
    }

    // validate item stream map
    if unproc.NeedsStreamMap {
      errors[id] = "item needs stream map"
      continue
    }
    if len(unproc.TranscodedStreams) == 0 {
      errors[id] = "item has empty stream map"
      continue
    }

    // validate this item not already in queue
    var count int64
    err = DB.QueryRow(`SELECT COUNT(*) FROM transcoding_tasks WHERE unprocessed_id = ?;`, id).Scan(&count)
    if err != nil && err != sql.ErrNoRows { return debug500(context, err) }
    if count > 0 {
      errors[id] = "item already in queue"
      continue
    }
  }
  if len(errors) > 0 { return context.Status(400).JSON(errors) }

  // add to queue
  for _, id := range id_list {
    unproc, err := database.UnprocessedRead(DB, id)
    if err != nil { return debug500(context, err) }

    tct := database.TranscodingTask {
      UnprocessedId: unproc.Id,
      TimeStarted:   0,
      TimeElapsed:   0,
      Status:        database.TranscodingTaskTodo,
      ErrorMessage:  "",
      CommandLine:   "",
    }
    err = tct.Create(DB)
    if err != nil { return debug500(context, err) }
  }

  return context.SendStatus(200)
}

func adminUnprocessedComplete(context *fiber.Ctx) error {
  id_list := []string {}
  if err := context.BodyParser(&id_list); err != nil { return context.Status(400).JSON(map[string]string { "error": "error parsing json" }) }
  if len(id_list) == 0 { return context.SendStatus(200) }

  // TODO:
  //  - validate that mapping, transcoding, and metadata are complete
  //  - move transcoded file to final location
  //  - mark metadata as live (by setting newly added "live" field)
  //  - if no errors, remove unprocessed record (and any associated transcoding tasks)
  return context.SendStatus(200)
}

func adminUnprocessedDelete(context *fiber.Ctx) error {
  id := context.Params("id")
  unproc, err := database.UnprocessedRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  err = unproc.Delete(DB)
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}
