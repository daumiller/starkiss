package main

import (
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
    err := database.PropertyUpsert(DB, "mediapath", proposed.Media)
    if err != nil { return debug500(context, err) }
    MEDIAPATH = proposed.Media
  }

  if proposed.Poster != "" {
    err := database.PropertyUpsert(DB, "posterpath", proposed.Poster)
    if err != nil { return debug500(context, err) }
    POSTERPATH = proposed.Poster
  }

  proposed.Media = MEDIAPATH
  proposed.Poster = POSTERPATH
  return context.JSON(proposed)
}
