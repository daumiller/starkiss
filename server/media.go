package main

import (
  "os"
  "fmt"
  "path"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

func startupMediaRoutes(server *fiber.App) {
  server.Get("/media/:id",        mediaServeMedia)
  server.Get("/poster/:id/:size", mediaServePoster)
}

func mediaServeMedia(context *fiber.Ctx) error {
  id := context.Params("id")
  md, err := database.MetadataRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  full_path := path.Join(MEDIAPATH, md.Location)
  if _, err := os.Stat(full_path); os.IsNotExist(err) { return context.SendStatus(404) }

  return context.SendFile(full_path, false)
}

func mediaServePoster(context *fiber.Ctx) error {
  id := context.Params("id")
  md, err := database.MetadataRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  size := context.Params("size")
  if (size != "small") && (size != "large") { return context.SendStatus(400) }

  if md.HasPoster == false { return context.SendStatus(404) }
  poster_name := fmt.Sprintf("%s.%s.jpg", id, size)
  full_path := path.Join(POSTERPATH, poster_name)
  if _, err := os.Stat(full_path); os.IsNotExist(err) { return context.SendStatus(404) }

  return context.SendFile(full_path, false)
}
