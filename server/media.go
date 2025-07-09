package main

import (
  "os"
  "github.com/labstack/echo/v4"
  "github.com/hashicorp/golang-lru/v2"
  "github.com/daumiller/starkiss/library"
)

var poster_cache *lru.Cache[string, string]

func startupMediaRoutes(server *echo.Echo) {
  poster_cache, _ = lru.New[string, string](1024)
  server.GET("/media/:id", mediaServeMedia)
  server.GET("/poster/:id/:size", mediaServePoster)
  server.GET("/poster/reset-cache", mediaServePosterResetCache)
}

func mediaServeMedia(context echo.Context) error {
  id := context.Param("id")
  md, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return context.NoContent(404) }
  if err != nil { return debug500(context, err) }

  full_path, err := md.DiskPath(library.MetadataPathTypeMedia)
  if err != nil { return debug500(context, err) }
  if _, err := os.Stat(full_path); os.IsNotExist(err) { return context.NoContent(404) }

  return context.File(full_path)
}

func resetMetadataPosterCache(id string) {
  poster_cache.Remove(id + "/small")
  poster_cache.Remove(id + "/large")
}

func mediaServePoster(context echo.Context) error {
  size := context.Param("size")
  if (size != "small") && (size != "large") { return context.NoContent(400) }

  cache_path := context.Param("id") + "/" + size
  disk_path, ok := poster_cache.Get(cache_path)
  if ok { return context.File(disk_path) }

  id := context.Param("id")
  md, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return context.NoContent(404) }
  if err != nil { return debug500(context, err) }

  var full_path string = ""
  if size == "small" { full_path, err = md.DiskPath(library.MetadataPathTypePosterSmall) }
  if size == "large" { full_path, err = md.DiskPath(library.MetadataPathTypePosterLarge) }
  if err != nil { return debug500(context, err) }

  if _, err := os.Stat(full_path); os.IsNotExist(err) {
    poster_aspect := "1x1"
    switch md.MediaType {
      case library.MetadataMediaTypeFileVideo : fallthrough
      case library.MetadataMediaTypeSeason    : fallthrough
      case library.MetadataMediaTypeSeries    : poster_aspect = "2x3"
    }
    missing_location := "./static/missing." + size + "." + poster_aspect + ".png"
    poster_cache.Add(cache_path, missing_location)
    return context.File(missing_location)
  }

  poster_cache.Add(cache_path, full_path)
  return context.File(full_path)
}

func mediaServePosterResetCache(context echo.Context) error {
  poster_cache.Purge()
  return json200(context, map[string]string{})
}
