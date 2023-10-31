package library

import (
  "os"
  "path/filepath"
  "github.com/daumiller/starkiss/database"
)

type MetadataPathType string
const (
  MetadataPathTypeBase        MetadataPathType = "base"
  MetadataPathTypeMedia       MetadataPathType = "media"
  MetadataPathTypePosterLarge MetadataPathType = "poster_large"
  MetadataPathTypePosterSmall MetadataPathType = "poster_small"
)

func MetadataPath(md *database.Metadata, path_type MetadataPathType) (string, error) {
  media_path, err := MediaPathGet()
  if err != nil { return "", err }

  parent_id := md.ParentId
  for parent_id != "" {
    if database.CategoryExistingId(db, parent_id) { break }
    parent, err := database.MetadataRead(db, parent_id)
    if err != nil { return "", err }
    media_path = filepath.Join(media_path, parent.NameSort)
    parent_id = parent.ParentId
  }

  media_path = filepath.Join(media_path, md.NameSort)
  switch path_type {
    case MetadataPathTypeMedia: {
      switch(md.MediaType) {
        case database.MetadataMediaTypeFileVideo: media_path += ".mp4"
        case database.MetadataMediaTypeFileAudio: media_path += ".mp3"
      }
    }
    case MetadataPathTypePosterLarge: media_path += ".large.jpg"
    case MetadataPathTypePosterSmall: media_path += ".small.jpg"
    case MetadataPathTypeBase: ; // do nothing
  }

  return media_path, nil
}

func MetadataWhere(where_string string, where_args ...any) ([]database.Metadata, error) {
  return database.MetadataWhere(db, where_string, where_args...)
}

func metadataCanMoveFilesToPath(md *database.Metadata, path string) bool {
  check_paths := []string {
    filepath.Join(path, md.NameSort) + ".mp4",
    filepath.Join(path, md.NameSort) + ".mp3",
    filepath.Join(path, md.NameSort) + ".large.jpg",
    filepath.Join(path, md.NameSort) + ".small.jpg",
  }
  for _, check_path := range check_paths {
    if _, err := os.Stat(check_path); err == nil { return false }
  }
  return true
}

func metadataMoveFilesToPath(md *database.Metadata, path string) error {
  path_after_base := filepath.Join(path, md.NameSort)
  path_before_base, err := MetadataPath(md, MetadataPathTypeBase)
  if err != nil { return err }

  paths_before := []string {
    path_before_base + ".mp4",
    path_before_base + ".mp3",
    path_before_base + ".large.jpg",
    path_before_base + ".small.jpg",
  }
  paths_after := []string {
    path_after_base + ".mp4",
    path_after_base + ".mp3",
    path_after_base + ".large.jpg",
    path_after_base + ".small.jpg",
  }

  var any_error error = nil
  for index := range paths_before {
    if _, err := os.Stat(paths_before[index]); err != nil { continue }
    err := os.Rename(paths_before[index], paths_after[index])
    if err != nil { any_error = err }
  }

  return any_error
}

/*
  MetadataTree()
  MetadataCreate(), Delete, Reparent, Rename(display, sort)
  MetadataSetPoster()
*/
