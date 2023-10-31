package main

import (
  "os"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

/* Workflow:
  1a. admin sets media_path, creates categories
  1b. scanner creates input files
  2.  admin views input files, creates missing stream maps
  3.  transcoder gets media ready, stores in temporary location
  4.  admin creates series/season, or artist/album metadata as needed
  5.  admin creates file metadata (edit metadata from inputfile, which will edit or create records)
  6.  admin completes input file, which moves transcode to final destination, marks metadata as not hidden, and deletes input file
*/

/* Admin web ui components:
  - media path editor
  - category editor
  - input_file browser/editor
    - can edit/assign stream_map
    - can reset transcoding state (deleting any existing transcoded file)
    - can delete input file (and transcoded file, if exists)
    - can complete input file (validate transcoding & metadata, move file to final destination, remove input-file record)
    - can launch metadata editor for corresponding metadata entry (creating as needed)
  - metadata browser
    - can find orphaned or hidden metadata
    - can browser by category, through hierarchy, just like client would
  - metadata editor, can edit series/season/artist/album/file
    - can create new metadata records for input files
    - can assign parent metadata (which also sets category)
    - given file path, can copy, resize, and store posters
    - LATER: given url, can download, resize, and store posters
    - LATER: can search at tmdb/tvdb/audiodb, and import metadata
*/

func startupAdminRoutes(server *fiber.App) {
  server.Get ("/admin/properties", adminPropertiesRead)
  server.Post("/admin/properties", adminPropertiesUpdate)

  server.Get   ("/admin/categories",   adminCategoryList  )
  server.Post  ("/admin/category",     adminCategoryCreate)
  server.Post  ("/admin/category/:id", adminCategoryUpdate)
  server.Delete("/admin/category/:id", adminCategoryDelete)

  server.Get   ("/admin/metadata/tree",                 adminMetadataTree        ) // list category > metadata > metadata hierarchy
  server.Get   ("/admin/metadata/by-parent/:parent_id", adminMetadataByParentList) // list all metadata for a parent
  server.Post  ("/admin/metadata",                      adminMetadataCreate      ) // create new metadata (defaults to hidden) (requires a parent id, optionally specify a input-file for read-or-create)
  server.Delete("/admin/metadata",                      adminMetadataDelete      ) // delete metadata record(s)
  server.Post  ("/admin/metadata/:id",                  adminMetadataUpdate      ) // update metadata

  server.Get   ("/admin/input-files",             adminInputFileList    ) // list all input files
  server.Delete("/admin/input-files",             adminInputFileDelete  ) // delete input file(s) (and transcoded file(s), if they exist)
  server.Post  ("/admin/input-file/:id/map",      adminInputFileMap     ) // update stream_map
  server.Post  ("/admin/input-file/:id/reset",    adminInputFileReset   ) // reset transcoding values (and delete transcoded file, if exists)
  server.Post  ("/admin/input-file/:id/complete", adminInputFileComplete) // validate transcoding & metadata, move file to final destination, remove input-file record
}

// ============================================================================
// Properties

var PROPERTIES_HIDDEN []string = []string { "jwtkey", "migration_level" }

func adminPropertiesRead(context *fiber.Ctx) error {
  properties, err := database.PropertyList(DB)
  if err != nil { return debug500(context, err) }

  for _, key := range PROPERTIES_HIDDEN { delete(properties, key) }
  return context.JSON(properties)
}
func adminPropertiesUpdate(context *fiber.Ctx) error {
  updates := map[string]string {}
  if err := context.BodyParser(&updates); err != nil { return context.SendStatus(400) }

  for _, key := range PROPERTIES_HIDDEN { delete(updates, key) }
  for key, value := range updates {
    err := database.PropertyUpsert(DB, key, value)
    if err != nil { return debug500(context, err) }
    if key == "media_path" { MEDIA_PATH = value }
  }
  return context.SendStatus(200)
}

// ============================================================================
// Category

func adminCategoryList(context *fiber.Ctx) error {
  categories, err := database.CategoryList(DB)
  if err != nil { return debug500(context, err) }
  return context.JSON(categories)
}

func adminCategoryCreate(context *fiber.Ctx) error {
  category := database.Category{}
  if err := context.BodyParser(&category); err != nil { return context.SendStatus(400) }
  category.Id = ""
  err := database.TableCreate(DB, &category)
  if err == database.ErrValidationFailed { return context.SendStatus(400) }
  if err != nil { return debug500(context, err) }
  return context.Status(201).JSON(category)
}

func adminCategoryUpdate(context *fiber.Ctx) error {
  id := context.Params("id")
  original, err := database.CategoryRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  changes := map[string]string {}
  if err = context.BodyParser(&changes); err != nil { return context.SendStatus(400) }
  new_name, ok := changes["name"]       ; if !ok { new_name = original.Name              }
  new_type, ok := changes["media_type"] ; if !ok { new_type = string(original.MediaType) }

  err = original.Update(DB, new_name, database.CategoryMediaType(new_type))
  if err == database.ErrValidationFailed { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

func adminCategoryDelete(context *fiber.Ctx) error {
  id := context.Params("id")
  category, err := database.CategoryRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }
  err = category.Delete(DB)
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

// ============================================================================
// Metadata

func isCategory(id string) bool {
  result := DB.QueryRow(`SELECT id FROM categories WHERE id = ? LIMIT 1;`, id)
  var found_id string = ""
  err := result.Scan(&found_id)
  if err == database.ErrNotFound { return false }
  if err != nil { return false }
  return true
}

type metadataTreeNode struct {
  Id        string           `json:"id"`
  Name      string           `json:"name"`
  MediaType string           `json:"media_type"`
  Children  []metadataTreeNode `json:"children"` // not a map because want this ordered
}
func adminMetadataTreeRecurse(root *metadataTreeNode) error {
  rows, err := DB.Query(`SELECT id, name_display, media_type FROM metadata WHERE parent_id = ? ORDER BY name_sort;`, root.Id)
  if err != nil { return err }
  defer rows.Close()

  for rows.Next() {
    var id, name, media_type string
    err = rows.Scan(&id, &name, &media_type)
    if err != nil { return err }

    if (media_type == "file-video") || (media_type == "file-audio") { continue }
    child := metadataTreeNode { Id: id, Name: name, MediaType: media_type, Children: []metadataTreeNode{} }
    err = adminMetadataTreeRecurse(&child)
    if err != nil { return err }

    root.Children = append(root.Children, child)
  }

  return nil
}
func adminMetadataTree(context *fiber.Ctx) error {
  lost_items := metadataTreeNode { Id:"lost", Name:"Lost Items", MediaType:"", Children:[]metadataTreeNode{} }
  err := adminMetadataTreeRecurse(&lost_items)
  if err != nil { return debug500(context, err) }

  categories, err := database.CategoryList(DB)
  if err != nil { return debug500(context, err) }

  root_items := make([]*metadataTreeNode, len(categories) + 1)
  root_items[0] = &lost_items
  for index, cat := range categories {
    cat_tree := metadataTreeNode { Id:cat.Id, Name:cat.Name, MediaType:string(cat.MediaType), Children: []metadataTreeNode{} }
    err = adminMetadataTreeRecurse(&cat_tree)
    if err != nil { return debug500(context, err) }
    root_items[index + 1] = &cat_tree
  }

  return context.JSON(root_items)
}

func adminMetadataByParentList(context *fiber.Ctx) error {
  parent_id := context.Params("parent_id")
  if parent_id == "lost" { parent_id = "" }
  metadata_list, err := database.MetadataWhere(DB, `parent_id = ?`, parent_id)
  if err != nil { return debug500(context, err) }
  return context.JSON(metadata_list)
}

func adminMetadataCreate(context *fiber.Ctx) error {
  metadata := database.Metadata{}
  if err := context.BodyParser(&metadata); err != nil { return context.SendStatus(400) }
  metadata.Id = ""
  err := database.TableCreate(DB, &metadata)
  if err == database.ErrValidationFailed { return context.SendStatus(400) }
  if err != nil { return debug500(context, err) }
  return context.Status(201).JSON(metadata)
}

func adminMetadataUpdate(context *fiber.Ctx) error {
  id := context.Params("id")
  original, err := database.MetadataRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  changes := map[string]any {}
  if err = context.BodyParser(&changes); err != nil { return context.SendStatus(400) }

  err = original.Patch(DB, changes)
  if err == database.ErrValidationFailed { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  if err != nil { return debug500(context, err) }
  return context.Status(200).JSON(original)
}

func adminMetadataDelete(context *fiber.Ctx) error {
  id_list := []string{}
  if err := context.BodyParser(&id_list); err != nil { return context.SendStatus(400) }
  if len(id_list) == 0 { return context.SendStatus(200) }

  lookup := map[string]*(database.Metadata) {}
  errors := map[string]string {}
  for _, id := range id_list {
    md, err := database.MetadataRead(DB, id)
    if err == database.ErrNotFound { errors[id] = "not found" ; continue }
    if err != nil { return debug500(context, err) }
    lookup[id] = md
    okay_to_delete, err := md.ValidDelete(DB)
    if err != nil { return debug500(context, err) }
    if okay_to_delete == false { errors[id] = "validation failed" ; continue }
  }
  if len(errors) > 0 { return context.Status(400).JSON(errors) }

  for _, id := range id_list {
    md := lookup[id]
    err := md.Delete(DB)
    if err != nil { errors[id] = "delete failed" ; continue }
  }
  if len(errors) > 0 { return context.Status(500).JSON(errors) }

  return context.SendStatus(200)
}

// ============================================================================
// InputFile

func adminInputFileList(context *fiber.Ctx) error {
  input_files, err := database.InputFileWhere(DB, "")
  if err != nil { return debug500(context, err) }
  return context.JSON(input_files)
}

func adminInputFileDelete(context *fiber.Ctx) error {
  id_list := []string{}
  if err := context.BodyParser(&id_list); err != nil { return context.SendStatus(400) }
  if len(id_list) == 0 { return context.SendStatus(200) }

  lookup := map[string]*(database.InputFile) {}
  errors := map[string]string {}
  for _, id := range id_list {
    inp, err := database.InputFileRead(DB, id)
    if err == database.ErrNotFound { errors[id] = "not found" ; continue }
    if err != nil { return debug500(context, err) }
    lookup[id] = inp
    okay_to_delete, err := inp.ValidDelete(DB)
    if err != nil { return debug500(context, err) }
    if okay_to_delete == false { errors[id] = "validation failed" ; continue }
  }
  if len(errors) > 0 { return context.Status(400).JSON(errors) }

  for _, id := range id_list {
    inp := lookup[id]
    err := inp.DeleteTranscodedFile()
    if err != nil { errors[id] = "delete transcoded file failed" ; continue }
    err = inp.Delete(DB)
    if err != nil { errors[id] = "delete failed" ; continue }
  }
  if len(errors) > 0 { return context.Status(500).JSON(errors) }

  return context.SendStatus(200)
}

func adminInputFileMap(context *fiber.Ctx) error {
  id := context.Params("id")
  inp, err := database.InputFileRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  proposed := inp.Copy()
  if err = context.BodyParser(&(proposed.StreamMap)); err != nil { return context.SendStatus(400) }

  err = inp.Replace(DB, proposed)
  if err == database.ErrValidationFailed { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

func adminInputFileReset(context *fiber.Ctx) error {
  id := context.Params("id")
  inp, err := database.InputFileRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  err = inp.DeleteTranscodedFile()
  if err != nil { return debug500(context, err) }

  proposed := inp.Copy()
  proposed.TranscodingCommand     = ""
  proposed.TranscodingError       = ""
  proposed.TranscodingTimeStarted = 0
  proposed.TranscodingTimeElapsed = 0
  proposed.TranscodedLocation     = ""

  err = inp.Replace(DB, proposed)
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

func adminInputFileComplete(context *fiber.Ctx) error {
  // validate transcoding & metadata, move file to final destination, remove input-file record
  id := context.Params("id")
  inp, err := database.InputFileRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  // verify transcoding complete
  if inp.TranscodingError != "" { return context.Status(400).JSON(map[string]string{"error": "transcoding error: " + inp.TranscodingError}) }
  if inp.TranscodedLocation == "" { return context.Status(400).JSON(map[string]string{"error": "transcoded location not set"}) }
  _, err = os.Stat(inp.TranscodedLocation)
  if err != nil { return context.Status(400).JSON(map[string]string{"error": "transcoded file not found"}) }

  // verify metadata created
  _, err = database.MetadataRead(DB, inp.Id)
  if err == database.ErrNotFound { return context.Status(400).JSON(map[string]string{"error": "metadata not found"}) }
  // TODO: verify all metadata fields
  // TODO: populate metadata record with new ffprobe of transcoded file (and verify compatible)

  // TODO:
  // move transcoded file to final location
  // mark metadata as not hidden
  // remove input-file record

  return context.SendStatus(200)
}
