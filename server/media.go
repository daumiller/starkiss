package main

import (
  "os"
  "path"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

func startupMediaRoutes(server *fiber.App) {
  server.Get("/media/:id",        mediaServeMedia)
  server.Get("/poster/:id/:size", mediaServePoster)
}

func getMetadataPath(md *database.Metadata) (string, error) {
  // todo: cache these lookups for non-file media types (series/season/artist/album) (https://github.com/hashicorp/golang-lru)
  components := []string { md.NameSort }
  parent_id := md.ParentId
  for parent_id != "" {
    parent, err := database.MetadataRead(DB, parent_id)
    if err != nil { return "", err }
    components = append([]string { parent.NameSort }, components...)
    parent_id = parent.ParentId
  }
  components = append([]string { MEDIA_PATH }, components...)
  return path.Join(components...), nil
}

func mediaServeMedia(context *fiber.Ctx) error {
  id := context.Params("id")
  md, err := database.MetadataRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  full_path, err := getMetadataPath(md)
  if err != nil { return debug500(context, err) }
  full_path += ".mp4"
  if _, err := os.Stat(full_path); os.IsNotExist(err) { return context.SendStatus(404) }

  return context.SendFile(full_path, false)
}

func mediaServePoster(context *fiber.Ctx) error {
  size := context.Params("size")
  if (size != "small") && (size != "large") { return context.SendStatus(400) }

  id := context.Params("id")
  md, err := database.MetadataRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  full_path, err := getMetadataPath(md)
  if err != nil { return debug500(context, err) }
  full_path += "." + size + ".jpg"
  // TODO: maybe, serve a default image if poster doesn't exist?
  if _, err := os.Stat(full_path); os.IsNotExist(err) { return context.SendStatus(404) }

  return context.SendFile(full_path, false)
}
