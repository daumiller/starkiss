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
