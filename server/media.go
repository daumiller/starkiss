package main

import (
  "os"
  "github.com/gofiber/fiber/v2"
  "github.com/hashicorp/golang-lru/v2"
  "github.com/daumiller/starkiss/library"
)

var poster_cache *lru.Cache[string, string]

func startupMediaRoutes(server *fiber.App) {
  poster_cache, _ = lru.New[string, string](1024)
  server.Get("/media/:id",          mediaServeMedia)
  server.Get("/poster/:id/:size",   mediaServePoster)
  server.Get("/poster/reset-cache", mediaServePosterResetCache)
}

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

func resetMetadataPosterCache(id string) {
  poster_cache.Remove(id + "/small")
  poster_cache.Remove(id + "/large")
}

func mediaServePoster(context *fiber.Ctx) error {
  size := context.Params("size")
  if (size != "small") && (size != "large") { return context.SendStatus(400) }

  cache_path := context.Params("id") + "/" + size
  disk_path, ok := poster_cache.Get(cache_path)
  if ok { return context.SendFile(disk_path, false) }

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
    missing_location := "./static/missing." + size + "." + poster_aspect + ".png"
    poster_cache.Add(cache_path, missing_location)
    return context.SendFile(missing_location, false)
  }

  poster_cache.Add(cache_path, full_path)
  return context.SendFile(full_path, false)
}

func mediaServePosterResetCache(context *fiber.Ctx) error {
  poster_cache.Purge()
  return json200(context, map[string]string {})
}
