package library

import (
  "os"
  "fmt"
  "image"
  "image/jpeg"
  "github.com/nfnt/resize"
  "strings"
  "path/filepath"
  "encoding/json"
)
type MetadataPathType string
const (
  MetadataPathTypeBase        MetadataPathType = "base"
  MetadataPathTypeMedia       MetadataPathType = "media"
  MetadataPathTypePosterLarge MetadataPathType = "poster_large"
  MetadataPathTypePosterSmall MetadataPathType = "poster_small"
)

type MetadataMediaType string
const (
  MetadataMediaTypeFileVideo MetadataMediaType = "file-video"
  MetadataMediaTypeFileAudio MetadataMediaType = "file-audio"
  MetadataMediaTypeSeries    MetadataMediaType = "series"
  MetadataMediaTypeSeason    MetadataMediaType = "season"
  MetadataMediaTypeArtist    MetadataMediaType = "artist"
  MetadataMediaTypeAlbum     MetadataMediaType = "album"
)

type Metadata struct {
  Id          string            `json:"id"`
  ParentId    string            `json:"parent_id"`
  MediaType   MetadataMediaType `json:"media_type"`
  NameDisplay string            `json:"name_display"`
  NameSort    string            `json:"name_sort"`
  Streams     []FileStream      `json:"streams"`
  Duration    int64             `json:"duration"`
  Size        int64             `json:"size"`
}

// ============================================================================
// Public Interface

func (md *Metadata) Copy() (*Metadata) {
  copy := Metadata {}
  copy.Id          = md.Id
  copy.ParentId    = md.ParentId
  copy.MediaType   = md.MediaType
  copy.NameDisplay = md.NameDisplay
  copy.NameSort    = md.NameSort
  copy.Streams     = make([]FileStream, len(md.Streams))
  copy.Duration    = md.Duration
  copy.Size        = md.Size

  for index, stream := range md.Streams {
    stream_copy := stream.Copy()
    copy.Streams[index] = *stream_copy
  }

  return &copy
}

func (md *Metadata) DiskPath(path_type MetadataPathType) (string, error) {
  md_path := md.NameSort

  parent_id := md.ParentId
  for parent_id != "" {
    if CategoryIdExists(parent_id) {
      cat := Category {}
      err := dbRecordRead(&cat, parent_id)
      if err != nil { return "", fmt.Errorf("category not found: %s", parent_id) }
      md_path = filepath.Join(cat.Name, md_path)
      break
    } else {
      parent := Metadata {}
      err := dbRecordRead(&parent, parent_id)
      if err != nil { return "", fmt.Errorf("metadata not found: %s", parent_id) }
      md_path = filepath.Join(parent.NameSort, md_path)
      parent_id = parent.ParentId
    }
  }
  md_path = filepath.Join(mediaPath, md_path)

  switch path_type {
    case MetadataPathTypeMedia: {
      switch(md.MediaType) {
        case MetadataMediaTypeFileVideo: md_path += ".mp4"
        case MetadataMediaTypeFileAudio: md_path += ".mp3"
      }
    }
    case MetadataPathTypePosterLarge: md_path += ".large.jpg"
    case MetadataPathTypePosterSmall: md_path += ".small.jpg"
    case MetadataPathTypeBase: ; // do nothing
  }

  return md_path, nil
}

func (md *Metadata) Reparent(new_parent_id string) error {
  if new_parent_id == md.ParentId { return nil }

  parent_path := mediaPath
  if new_parent_id != "" {
    if CategoryIdExists(new_parent_id) {
      cat := Category {}
      err := dbRecordRead(&cat, new_parent_id)
      if err != nil { return fmt.Errorf("category not found: %s", new_parent_id) }
      parent_path = cat.DiskPath()
    } else {
      parent := Metadata {}
      err := dbRecordRead(&parent, new_parent_id)
      if err != nil { return fmt.Errorf("metadata not found: %s", new_parent_id) }
      parent_path, err = parent.DiskPath(MetadataPathTypeBase)
      if err != nil { return err }
    }
  }

  // verify name_sort, within parent, is unique
  records, err := dbRecordWhere(&Metadata{}, `(parent_id = ?) AND (name_sort = ?) LIMIT 1;`, new_parent_id, md.NameSort)
  if err != nil { return ErrQueryFailed }
  if len(records) > 0 { return fmt.Errorf("metadata named \"%s\" already exists in parent \"%s\"", md.NameSort, new_parent_id) }

  // verify name_sort, on disk, isn't taken
  if metadataCanMoveFilesToPath(md, parent_path) == false {
    return fmt.Errorf("metadata named \"%s\" already exists on disk", md.NameSort)
  }

  // move files on disk
  old_parent_id := md.ParentId
  paths_source := make([]string, 3)
  paths_dest   := make([]string, 3)
  md.ParentId = old_parent_id
  paths_source[0], _ = md.DiskPath(MetadataPathTypeMedia)
  paths_source[1], _ = md.DiskPath(MetadataPathTypePosterLarge)
  paths_source[2], _ = md.DiskPath(MetadataPathTypePosterSmall)
  md.ParentId = new_parent_id
  paths_dest[0], _ = md.DiskPath(MetadataPathTypeMedia)
  paths_dest[1], _ = md.DiskPath(MetadataPathTypePosterLarge)
  paths_dest[2], _ = md.DiskPath(MetadataPathTypePosterSmall)
  md.ParentId = old_parent_id
  for index := range paths_source {
    if pathExists(paths_source[index]) == false { continue }
    err := os.Rename(paths_source[index], paths_dest[index])
    if err != nil { return err }
  }

  // update record
  err = dbRecordPatch(md, map[string]any { "parent_id":new_parent_id })
  if err != nil { return ErrQueryFailed }
  return nil
}

func (md *Metadata) Rename(new_name_display string, new_name_sort string) error {
  if new_name_sort == "" { new_name_sort = nameGetSortForDisplay(new_name_display) }
  new_name_sort = strings.TrimSpace(new_name_sort)

  if !nameValidForDisk(new_name_sort) { return ErrInvalidName }

  if new_name_sort != md.NameSort {
    // verify name_sort, within parent, is unique
    records, err := dbRecordWhere(&Metadata{}, `(parent_id = ?) AND (name_sort = ?) LIMIT 1;`, md.ParentId, new_name_sort)
    if err != nil { return ErrQueryFailed }
    if len(records) > 0 { return fmt.Errorf("metadata named \"%s\" already exists in parent \"%s\"", new_name_sort, md.ParentId) }

    // verify name_sort, on disk, isn't taken
    dest_path, _ := md.DiskPath(MetadataPathTypeBase)
    dest_path = strings.TrimSuffix(dest_path, "/" + md.NameSort)
    old_name_sort := md.NameSort
    md.NameSort = new_name_sort
    if metadataCanMoveFilesToPath(md, dest_path) == false {
      md.NameSort = old_name_sort
      return fmt.Errorf("metadata named \"%s\" already exists on disk", new_name_sort)
    }

    // move files on disk
    paths_source := make([]string, 3)
    paths_dest   := make([]string, 3)
    md.NameSort = old_name_sort
    paths_source[0], _ = md.DiskPath(MetadataPathTypeMedia)
    paths_source[1], _ = md.DiskPath(MetadataPathTypePosterLarge)
    paths_source[2], _ = md.DiskPath(MetadataPathTypePosterSmall)
    md.NameSort = new_name_sort
    paths_dest[0], _ = md.DiskPath(MetadataPathTypeMedia)
    paths_dest[1], _ = md.DiskPath(MetadataPathTypePosterLarge)
    paths_dest[2], _ = md.DiskPath(MetadataPathTypePosterSmall)
    md.NameSort = old_name_sort
    for index := range paths_source {
      if pathExists(paths_source[index]) == false { continue }
      err := os.Rename(paths_source[index], paths_dest[index])
      if err != nil { return err }
    }
  }

  // update record
  err := dbRecordPatch(md, map[string]any { "name_display":new_name_display, "name_sort":new_name_sort })
  if err != nil { return ErrQueryFailed }
  return nil
}

func (md *Metadata) SetPoster(img image.Image) error {
  // get image sizes
  large_width  := uint(1)
  large_height := uint(1)
  small_width  := uint(1)
  small_height := uint(1)

  switch(md.MediaType) {
    case MetadataMediaTypeFileVideo: fallthrough
    case MetadataMediaTypeSeries   : fallthrough
    case MetadataMediaTypeSeason   : {
      large_width  = 360
      large_height = 540
      small_width  = 183
      small_height = 275
    }
    case MetadataMediaTypeFileAudio: fallthrough
    case MetadataMediaTypeArtist   : fallthrough
    case MetadataMediaTypeAlbum    : {
      large_width  = 512
      large_height = 512
      small_width  = 200
      small_height = 200
    }
  }

  // save large poster
  large_path, err := md.DiskPath(MetadataPathTypePosterLarge)
  if err != nil { return err }
  large_file, err := os.Create(large_path)
  if err != nil { return err }
  defer large_file.Close()
  large_image := resize.Resize(large_width, large_height, img, resize.NearestNeighbor)
  err = jpeg.Encode(large_file, large_image, nil)
  if err != nil { return err }

  // save small poster
  small_path, err := md.DiskPath(MetadataPathTypePosterSmall)
  if err != nil { return err }
  small_file, err := os.Create(small_path)
  if err != nil { return err }
  defer small_file.Close()
  small_image := resize.Resize(small_width, small_height, img, resize.NearestNeighbor)
  err = jpeg.Encode(small_file, small_image, nil)
  if err != nil { return err }

  return nil
}

func MetadataRead(id string) (*Metadata, error) {
  md := Metadata {}
  err := dbRecordRead(&md, id)
  if err != nil { return nil, err }
  return &md, nil
}

func MetadataForParent(parent_id string) ([]Metadata, error) {
  records, err := dbRecordWhere(&Metadata{}, `(parent_id = ?) ORDER BY name_sort ASC`, parent_id)
  if err != nil { return nil, ErrQueryFailed }
  metadata := make([]Metadata, len(records))
  for index, record := range records { metadata[index] = *(record.(*Metadata)) }
  return metadata, nil
}

func MetadataCreate(md *Metadata) error {
  // verify valid name_sort
  if md.NameSort == "" { md.NameSort = nameGetSortForDisplay(md.NameDisplay) }
  if !nameValidForDisk(md.NameSort) { return ErrInvalidName }

  // verify name_sort, within parent, is unique
  records, err := dbRecordWhere(&Metadata{}, `(parent_id = ?) AND (name_sort = ?) LIMIT 1;`, md.ParentId, md.NameSort)
  if err != nil { return ErrQueryFailed }
  if len(records) > 0 { return fmt.Errorf("metadata named \"%s\" already exists in parent \"%s\"", md.NameSort, md.ParentId) }

  // if not file type, create directory
  if (md.MediaType != MetadataMediaTypeFileVideo) && (md.MediaType != MetadataMediaTypeFileAudio) {
    dir_path, err := md.DiskPath(MetadataPathTypeBase)
    if err != nil { return err }
    err = os.Mkdir(dir_path, 0770)
    if err != nil { return fmt.Errorf("error creating directory \"%s\" on disk: %s", dir_path, err.Error()) }
  }

  // save record
  err = dbRecordCreate(md)
  if err != nil { return ErrQueryFailed }
  return nil
}

func MetadataDelete(md *Metadata, delete_children bool) error {
  if delete_children == true {
    // delete all children first (cannot bulk because of parent_id hierarchy)
    children, err := MetadataForParent(md.Id)
    if err != nil { return err }
    for _, child := range children {
      err := MetadataDelete(&child, true)
      if err != nil { return err }
    }
  } else {
    // reparent all children to "lost" (empty parent)
    children, err := MetadataForParent(md.Id)
    if err != nil { return err }
    for _, child := range children {
      err := child.Reparent("")
      if err != nil { return err }
    }
  }

  // delete files on disk
  paths := make([]string, 3)
  paths[0], _ = md.DiskPath(MetadataPathTypeBase)
  paths[1], _ = md.DiskPath(MetadataPathTypePosterLarge)
  paths[2], _ = md.DiskPath(MetadataPathTypePosterSmall)
  var anyerr error = nil
  for index := range paths {
    if pathExists(paths[index]) == false { continue }
    err := os.Remove(paths[index])
    if err != nil { anyerr = err }
  }
  if anyerr != nil { return anyerr }

  // delete record
  err := dbRecordDelete(md)
  if err != nil { return ErrQueryFailed }
  return nil
}

type MetadataTreeNode struct {
  Id        string             `json:"id"`
  Name      string             `json:"name"`
  MediaType string             `json:"media_type"`
  Children  []MetadataTreeNode `json:"children"` // not a map because want this ordered
}
func MetadataParentTree(parent_id string) ([]MetadataTreeNode, error) {
  dbLock.RLock()
  defer dbLock.RUnlock()
  listing := []MetadataTreeNode {}
  rows, err := dbHandle.Query(`SELECT id, name_display, media_type FROM metadata WHERE parent_id = ? ORDER BY name_sort;`, parent_id)
  if err != nil { return listing, err }
  defer rows.Close()

  for rows.Next() {
    var id, name, media_type string
    err = rows.Scan(&id, &name, &media_type)
    if err != nil { return listing, err }

    if (media_type == string(MetadataMediaTypeFileAudio)) || (media_type == string(MetadataMediaTypeFileVideo)) { continue }
    entry := MetadataTreeNode { Id: id, Name: name, MediaType: media_type, Children: []MetadataTreeNode{} }
    entry.Children, err = MetadataParentTree(id)
    if err != nil { return listing, nil }

    listing = append(listing, entry)
  }

  return listing, nil
}

// ============================================================================
// public utilities

func MetadataIdExists(id string) bool {
  dbLock.RLock()
  defer dbLock.RUnlock()
  queryRow := dbHandle.QueryRow(`SELECT id FROM metadata WHERE id = ? LIMIT 1;`, id)
  err := queryRow.Scan(&id)
  return (err == nil)
}

func MetadataIsEmpty(id string) bool {
  dbLock.RLock()
  defer dbLock.RUnlock()
  row := dbHandle.QueryRow(`SELECT id FROM metadata WHERE parent_id = ? LIMIT 1;`, id)
  err := row.Scan(&id)
  found := (err == nil)
  return !found
}

// ============================================================================
// private utilities

func metadataMediaTypeValidString(media_type string) bool {
  md_media_type := MetadataMediaType(media_type)
  return metadataMediaTypeValid(md_media_type)
}
func metadataMediaTypeValid(media_type MetadataMediaType) bool {
  switch(media_type) {
    case MetadataMediaTypeFileVideo : fallthrough
    case MetadataMediaTypeFileAudio : fallthrough
    case MetadataMediaTypeSeries    : fallthrough
    case MetadataMediaTypeSeason    : fallthrough
    case MetadataMediaTypeArtist    : fallthrough
    case MetadataMediaTypeAlbum     : return true
    default: return false
  }
}

func metadataCanMoveFilesToPath(md *Metadata, path string) bool {
  check_paths := []string {
    filepath.Join(path, md.NameSort) + ".large.jpg",
    filepath.Join(path, md.NameSort) + ".small.jpg",
  }
  switch(md.MediaType) {
    case MetadataMediaTypeFileVideo: check_paths = append(check_paths, filepath.Join(path, md.NameSort) + ".mp4")
    case MetadataMediaTypeFileAudio: check_paths = append(check_paths, filepath.Join(path, md.NameSort) + ".mp3")
    default:                         check_paths = append(check_paths, filepath.Join(path, md.NameSort)         )
  }
  for _, check_path := range check_paths {
    if pathExists(check_path) { return false }
  }
  return true
}
func metadataMoveFilesToPath(md *Metadata, path string) error {
  path_base_before, err := md.DiskPath(MetadataPathTypeBase)
  path_base_after := filepath.Join(path, md.NameSort)
  if err != nil { return err }

  paths_before := []string {
    path_base_before + ".large.jpg",
    path_base_before + ".small.jpg",
  }
  paths_after := []string {
    path_base_after + ".large.jpg",
    path_base_after + ".small.jpg",
  }
  switch(md.MediaType) {
    case MetadataMediaTypeFileVideo: paths_before = append(paths_before, path_base_before + ".mp4") ; paths_after = append(paths_after, path_base_after + ".mp4")
    case MetadataMediaTypeFileAudio: paths_before = append(paths_before, path_base_before + ".mp3") ; paths_after = append(paths_after, path_base_after + ".mp3")
    default:                         paths_before = append(paths_before, path_base_before         ) ; paths_after = append(paths_after, path_base_after         )
  }

  var any_error error = nil
  for index := range paths_before {
    if !pathExists(paths_before[index]) { continue }
    err := os.Rename(paths_before[index], paths_after[index])
    if err != nil { any_error = err }
  }

  return any_error
}

// ============================================================================
// dbRecord interface

func (md *Metadata) TableName() string {
  return "metadata"
}

func (md *Metadata) RecordCopy() (dbRecord, error) {
  return md.Copy(), nil
}

func (md *Metadata) RecordCreate(fields map[string]any) (instance dbRecord, err error) {
  new_instance := Metadata {}
  err = new_instance.FieldsReplace(fields)
  if err != nil { return nil, err }
  return &new_instance, nil
}

func (md *Metadata) GetId() string {
  return md.Id
}
func (md *Metadata) SetId(id string) {
  md.Id = id
}

func (md *Metadata) FieldsRead() (fields map[string]any, err error) {
  fields = make(map[string]any)
  streams_bytes, err := json.Marshal(md.Streams) ; if err != nil { return nil, err } ; streams_string := string(streams_bytes)

  fields["id"           ] = md.Id
  fields["parent_id"    ] = md.ParentId
  fields["media_type"   ] = string(md.MediaType)
  fields["name_display" ] = md.NameDisplay
  fields["name_sort"    ] = md.NameSort
  fields["streams"      ] = streams_string
  fields["duration"     ] = md.Duration
  fields["size"         ] = md.Size

  return fields, nil
}

func (md *Metadata) FieldsReplace(fields map[string]any) (err error) {
  streams_string := fields["streams"].(string) ; var streams []FileStream ; err = json.Unmarshal([]byte(streams_string), &streams) ; if err != nil { return err }
  media_type :=  MetadataMediaType(fields["media_type"].(string))

  md.Id               = fields["id"               ].(string)
  md.ParentId         = fields["parent_id"        ].(string)
  md.MediaType        = media_type
  md.NameDisplay      = fields["name_display"     ].(string)
  md.NameSort         = fields["name_sort"        ].(string)
  md.Streams          = streams
  md.Duration         = fields["duration"         ].(int64)
  md.Size             = fields["size"             ].(int64)

  return nil
}

func (md *Metadata) FieldsPatch(fields map[string]any) (err error) {
  if id,           ok := fields["id"]           ; ok { md.Id          = id.(string)                            }
  if parent_id,    ok := fields["parent_id"]    ; ok { md.ParentId    = parent_id.(string)                     }
  if media_type,   ok := fields["media_type"]   ; ok { md.MediaType   = MetadataMediaType(media_type.(string)) }
  if name_display, ok := fields["name_display"] ; ok { md.NameDisplay = name_display.(string)                  }
  if name_sort,    ok := fields["name_sort"]    ; ok { md.NameSort    = name_sort.(string)                     }
  if duration,     ok := fields["duration"]     ; ok { md.Duration    = duration.(int64)                       }
  if size,         ok := fields["size"]         ; ok { md.Size        = size.(int64)                           }

  if streams, ok := fields["streams"] ; ok {
    streams_string := streams.(string)
    var streams []FileStream
    err = json.Unmarshal([]byte(streams_string), &streams)
    if err != nil { return err }
    md.Streams = streams
  }

  return nil
}

func (md_a *Metadata) FieldsDifference(other dbRecord) (diff map[string]any, err error) {
  diff = make(map[string]any)
  md_b, b_is_md := other.(*Metadata)
  if b_is_md == false { return diff, ErrInvalidType }

  a_streams_bytes, err := json.Marshal(md_a.Streams) ; if err != nil { return nil, err } ; a_streams_string := string(a_streams_bytes)
  b_streams_bytes, err := json.Marshal(md_b.Streams) ; if err != nil { return nil, err } ; b_streams_string := string(b_streams_bytes)

  if md_a.Id          != md_b.Id          { diff["id"           ] = md_b.Id                }
  if md_a.ParentId    != md_b.ParentId    { diff["parent_id"    ] = md_b.ParentId          }
  if md_a.MediaType   != md_b.MediaType   { diff["media_type"   ] = string(md_b.MediaType) }
  if md_a.NameDisplay != md_b.NameDisplay { diff["name_display" ] = md_b.NameDisplay       }
  if md_a.NameSort    != md_b.NameSort    { diff["name_sort"    ] = md_b.NameSort          }
  if a_streams_string != b_streams_string { diff["streams"      ] = b_streams_string       }
  if md_a.Duration    != md_b.Duration    { diff["duration"     ] = md_b.Duration          }
  if md_a.Size        != md_b.Size        { diff["size"         ] = md_b.Size              }

  return diff, nil
}
