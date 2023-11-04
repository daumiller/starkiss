package main

import (
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/library"
)

/* Workflow:
  1a. admin sets media_path, creates categories
  1b. scanner creates input files
  2.  admin views input files, creates missing stream maps
  3.  transcoder gets media ready, stores parentless metadata ("lost" category)
  4.  admin creates series/season, or artist/album metadata as needed
  5.  admin creates file metadata (edit metadata from inputfile, which will edit or create records)
  6.  admin deletes completed input file
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
  server.Post  ("/admin/metadata",                      adminMetadataCreate      ) // create metadata record
  server.Delete("/admin/metadata/:id",                  adminMetadataDelete      ) // delete metadata record (and all files & children, if any)
  server.Post  ("/admin/metadata/:id",                  adminMetadataUpdate      ) // update metadata

  server.Get   ("/admin/input-files",             adminInputFileList    ) // list all input files
  server.Delete("/admin/input-file/:id",          adminInputFileDelete  ) // delete input file (and transcoded file(s), if they exist)
  server.Post  ("/admin/input-file/:id/map",      adminInputFileMap     ) // update stream_map
  server.Post  ("/admin/input-file/:id/reset",    adminInputFileReset   ) // reset transcoding values (and delete transcoded file, if exists)
}

// ============================================================================
// Properties

func adminPropertiesRead(context *fiber.Ctx) error {
  properties, err := library.PropertyList()
  if err != nil { return debug500(context, err) }
  return context.JSON(properties)
}
func adminPropertiesUpdate(context *fiber.Ctx) error {
  updates := map[string]string {}
  if err := context.BodyParser(&updates); err != nil { return context.SendStatus(400) }

  for key, value := range updates {
    err := library.PropertySet(key, value)
    if err != nil { return debug500(context, err) }
  }
  return context.SendStatus(200)
}

// ============================================================================
// Category

func adminCategoryList(context *fiber.Ctx) error {
  categories, err := library.CategoryList()
  if err != nil { return debug500(context, err) }
  return context.JSON(categories)
}

func adminCategoryCreate(context *fiber.Ctx) error {
  category := library.Category{}
  if err := context.BodyParser(&category); err != nil { return context.SendStatus(400) }
  category.Id = ""
  result, err := library.CategoryCreate(category.Name, category.MediaType)
  if err == library.ErrQueryFailed { return context.SendStatus(500) }
  if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  return context.Status(201).JSON(result)
}

func adminCategoryUpdate(context *fiber.Ctx) error {
  id := context.Params("id")
  original, err := library.CategoryRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  changes := map[string]string {}
  if err = context.BodyParser(&changes); err != nil { return context.SendStatus(400) }
  new_name, ok := changes["name"]       ; if !ok { new_name = original.Name              }
  new_type, ok := changes["media_type"] ; if !ok { new_type = string(original.MediaType) }

  err = library.CategoryUpdate(original, new_name, new_type)
  if err == library.ErrQueryFailed { return context.SendStatus(500) }
  if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  return context.SendStatus(200)
}

func adminCategoryDelete(context *fiber.Ctx) error {
  id := context.Params("id")
  category, err := library.CategoryRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  err = library.CategoryDelete(category)
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

// ============================================================================
// Metadata

func adminMetadataTree(context *fiber.Ctx) error {
  var err error
  lost_items := library.MetadataTreeNode { Id:"lost", Name:"Lost Items", MediaType:"" }
  lost_items.Children, err = library.MetadataParentTree("")
  if err != nil { return debug500(context, err) }

  categories, err := library.CategoryList()
  if err != nil { return debug500(context, err) }

  root_items := make([]*library.MetadataTreeNode, len(categories) + 1)
  root_items[0] = &lost_items
  for index, cat := range categories {
    cat_tree := library.MetadataTreeNode { Id:cat.Id, Name:cat.Name, MediaType:string(cat.MediaType) }
    cat_tree.Children, err = library.MetadataParentTree(cat.Id)
    if err != nil { return debug500(context, err) }
    root_items[index + 1] = &cat_tree
  }

  return context.JSON(root_items)
}

func adminMetadataByParentList(context *fiber.Ctx) error {
  parent_id := context.Params("parent_id")
  if parent_id == "lost" { parent_id = "" }
  metadata_list, err := library.MetadataForParent(parent_id)
  if err != nil { return debug500(context, err) }
  return context.JSON(metadata_list)
}

func adminMetadataCreate(context *fiber.Ctx) error {
  metadata := library.Metadata{}
  if err := context.BodyParser(&metadata); err != nil { return context.SendStatus(400) }
  err := library.MetadataCreate(&metadata)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  return context.Status(201).JSON(metadata)
}

func adminMetadataUpdate(context *fiber.Ctx) error {
  id := context.Params("id")
  original, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  changes := map[string]string {}
  if err = context.BodyParser(&changes); err != nil { return context.SendStatus(400) }

  var ok bool
  var new_parent       string = "" ; new_parent,       ok = changes["parent_id"]    ; if !ok { new_parent       = original.ParentId    }
  var new_name_display string = "" ; new_name_display, ok = changes["name_display"] ; if !ok { new_name_display = original.NameDisplay }
  var new_name_sort    string = "" ; new_name_sort,    ok = changes["name_sort"]    ; if !ok { new_name_sort    = original.NameSort    }

  if new_parent != original.ParentId {
    err = original.Reparent(new_parent)
    if err == library.ErrQueryFailed { return debug500(context, err) }
    if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  }
  if (new_name_display != original.NameDisplay) || (new_name_sort != original.NameSort) {
    err = original.Rename(new_name_display, new_name_sort)
    if err == library.ErrQueryFailed { return debug500(context, err) }
    if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  }

  return context.Status(200).JSON(original)
}

type MetadataDeleteRequest struct {
  DeleteChildren bool `json:"delete_children"`
}
func adminMetadataDelete(context *fiber.Ctx) error {
  id := context.Params("id")
  md, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }
  request := MetadataDeleteRequest{}
  if err := context.BodyParser(&request); err != nil { return context.SendStatus(400) }

  err = library.MetadataDelete(md, request.DeleteChildren)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  return context.SendStatus(200)
}

// ============================================================================
// InputFile

func adminInputFileList(context *fiber.Ctx) error {
  input_files, err := library.InputFileList()
  if err != nil { return debug500(context, err) }
  return context.JSON(input_files)
}

func adminInputFileDelete(context *fiber.Ctx) error {
  id := context.Params("id")
  inp, err := library.InputFileRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  err = library.InputFileDelete(inp)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  return context.SendStatus(200)
}

func adminInputFileMap(context *fiber.Ctx) error {
  id := context.Params("id")
  inp, err := library.InputFileRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  stream_map := []int64 {}
  if err = context.BodyParser(&stream_map); err != nil { return context.SendStatus(400) }

  err = inp.Remap(stream_map)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  return context.SendStatus(200)
}

func adminInputFileReset(context *fiber.Ctx) error {
  id := context.Params("id")
  inp, err := library.InputFileRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  err = inp.StatusReset()
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  return context.SendStatus(200)
}
