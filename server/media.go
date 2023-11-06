package main

import (
  "os"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/library"
)

func startupMediaRoutes(server *fiber.App) {
  server.Get("/media/:id",        mediaServeMedia)
  server.Get("/poster/:id/:size", mediaServePoster)
}

// todo: cache lookups? (https://github.com/hashicorp/golang-lru)

func mediaServeMedia(context *fiber.Ctx) error {
  id := context.Params("id")
  md, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  full_path, err := md.DiskPath(library.MetadataPathTypeMedia)
  if err != nil { return debug500(context, err) }
  if _, err := os.Stat(full_path); os.IsNotExist(err) { return context.SendStatus(404) }

  return context.SendFile(full_path, false)
}

func mediaServePoster(context *fiber.Ctx) error {
  size := context.Params("size")
  if (size != "small") && (size != "large") { return context.SendStatus(400) }

  id := context.Params("id")
  md, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  var full_path string = ""
  if size == "small" { full_path, err = md.DiskPath(library.MetadataPathTypePosterSmall) }
  if size == "large" { full_path, err = md.DiskPath(library.MetadataPathTypePosterLarge) }
  if err != nil { return debug500(context, err) }
  
  if _, err := os.Stat(full_path); os.IsNotExist(err) {
    poster_aspect := "1x1"
    switch(md.MediaType) {
      case library.MetadataMediaTypeFileVideo : fallthrough
      case library.MetadataMediaTypeSeason    : fallthrough
      case library.MetadataMediaTypeSeries    : poster_aspect = "2x3"
    }
    return context.SendFile("./static/missing." + size + "." + poster_aspect + ".png", false)
  }

  return context.SendFile(full_path, false)
}
