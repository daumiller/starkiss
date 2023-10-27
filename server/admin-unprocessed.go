package main

import (
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

func adminUnprocessedList(context *fiber.Ctx) error {
  unprocessed, err := database.UnprocessedList(DB, false, false, false)
  if err != nil { return debug500(context, err) }
  return context.JSON(unprocessed)
}

func adminUnprocessedMap(context *fiber.Ctx) error {
  id := context.Params("id")
  _, err := database.UnprocessedRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  // set new mapping
  return context.SendStatus(200)
}

func adminUnprocessedQueue(context *fiber.Ctx) error {
  id := context.Params("id")
  _, err := database.UnprocessedRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  // create new transcoding task record
  return context.SendStatus(200)
}

// TODO: endpoint to list transcoder results
