package main

import (
  "os"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

func startupAdminRoutes(server *fiber.App) {
  server.Get ("/admin/media-path", adminMediaPathRead  )
  server.Post("/admin/media-path", adminMediaPathUpdate)

  server.Get   ("/admin/categories",   adminCategoryList  )
  server.Post  ("/admin/category",     adminCategoryCreate)
  server.Post  ("/admin/category/:id", adminCategoryUpdate)
  server.Delete("/admin/category/:id", adminCategoryDelete)

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
// MediaPath

func adminMediaPathRead(context *fiber.Ctx) error {
  return context.JSON(map[string]string { "media_path": MEDIAPATH })
}
func adminMediaPathUpdate(context *fiber.Ctx) error {
  body_obj := map[string]string {}
  err := context.BodyParser(&body_obj)
  if err != nil { return debug500(context, err) }
  if _, ok := body_obj["media_path"]; ok == false { return context.Status(400).JSON(map[string]string { "error": "missing media_path" }) }

  result := database.PropertyUpsert(DB, "mediapath", body_obj["media_path"])
  if result != nil { return debug500(context, result) }
  MEDIAPATH = body_obj["media_path"]
  return context.JSON(map[string]string { "media_path": MEDIAPATH })
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

func adminMetadataByParentList(context *fiber.Ctx) error {
  parent_id := context.Params("parent_id")
  where_string := `WHERE parent_id = ?`
  if isCategory(parent_id) { where_string = `WHERE (parent_id = '') AND (category_id = ?)` }

  metadata_list, err := database.MetadataWhere(DB, where_string, parent_id)
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
  // TODO: populate metadata record with new ffprobe of transcoded file

  // TODO:
  // move transcoded file to final location
  // mark metadata as not hidden
  // remove input-file record

  return context.SendStatus(200)
}
