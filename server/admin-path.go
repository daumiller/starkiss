package main

import (
  "os"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

type AdminPath struct {
  Media  string `json:"media"`
  Poster string `json:"poster"`
}

func adminPathList(context *fiber.Ctx) error {
  paths := AdminPath{ MEDIAPATH, POSTERPATH }
  return context.JSON(paths)
}

func adminPathUpdate(context *fiber.Ctx) error {
  proposed := AdminPath {}
  if err := context.BodyParser(&proposed); err != nil { return context.Status(400).JSON(map[string]string { "error": "error parsing json" }) }

  if proposed.Media != "" {
    if _, err := os.Stat(proposed.Media); os.IsNotExist(err) { return context.Status(400).JSON(map[string]string { "error": "media path does not exist" }) }
    err := database.PropertyUpsert(DB, "mediapath", proposed.Media)
    if err != nil { return debug500(context, err) }
    MEDIAPATH = proposed.Media
  }

  if proposed.Poster != "" {
    if _, err := os.Stat(proposed.Poster); os.IsNotExist(err) { return context.Status(400).JSON(map[string]string { "error": "poster path does not exist" }) }
    err := database.PropertyUpsert(DB, "posterpath", proposed.Poster)
    if err != nil { return debug500(context, err) }
    POSTERPATH = proposed.Poster
  }

  proposed.Media = MEDIAPATH
  proposed.Poster = POSTERPATH
  return context.JSON(proposed)
}
