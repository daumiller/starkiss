package main

import (
  "strings"
  "image"
  _ "image/jpeg"
  _ "image/png"
  _ "golang.org/x/image/webp"
  "github.com/labstack/echo/v4"
  "github.com/daumiller/starkiss/library"
  "net/http"
)

func startupAdminRoutes(server *echo.Echo) {
  server.GET   ("/admin/properties",   adminPropertiesRead)
  server.POST  ("/admin/properties",   adminPropertiesUpdate)

  server.GET   ("/admin/categories",   adminCategoryList  )
  server.POST  ("/admin/category",     adminCategoryCreate)
  server.POST  ("/admin/category/:id", adminCategoryUpdate)
  server.DELETE("/admin/category/:id", adminCategoryDelete)

  server.GET   ("/admin/metadata/tree",                 adminMetadataTree        )
  server.GET   ("/admin/metadata/by-parent/:parent_id", adminMetadataByParentList)
  server.POST  ("/admin/metadata",                      adminMetadataCreate      )
  server.DELETE("/admin/metadata/:id",                  adminMetadataDelete      )
  server.POST  ("/admin/metadata/:id",                  adminMetadataUpdate      )
  server.POST  ("/admin/metadata/:id/poster",           adminMetadataPoster      )

  server.GET   ("/admin/input-files",             adminInputFileList    )
  server.DELETE("/admin/input-file/:id",          adminInputFileDelete  )
  server.POST  ("/admin/input-file/:id/map",      adminInputFileMap     )
  server.POST  ("/admin/input-file/:id/reset",    adminInputFileReset   )
}

// ============================================================================
// Properties

func adminPropertiesRead(context echo.Context) error {
  properties, err := library.PropertyList()
  if err != nil { return debug500(context, err) }
  return json200(context, properties)
}
func adminPropertiesUpdate(context echo.Context) error {
  updates := map[string]string{}
  if err := context.Bind(&updates); err != nil { return json400(context, err) }

  for key, value := range updates {
    err := library.PropertySet(key, value)
    if err == library.ErrQueryFailed { return debug500(context, err) }
    if err != nil { return json400(context, err) }
  }
  return json200(context, map[string]string{})
}

// ============================================================================
// Category

func adminCategoryList(context echo.Context) error {
  categories, err := library.CategoryList()
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }
  return json200(context, categories)
}

func adminCategoryCreate(context echo.Context) error {
  category := library.Category{}
  if err := context.Bind(&category); err != nil { return json400(context, err) }
  category.Id = ""
  result, err := library.CategoryCreate(category.Name, category.MediaType)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }
  return context.JSON(http.StatusCreated, result)
}

type CategoryUpdateRequest struct {
  Name       string `json:"name"`
  MediaType  string `json:"media_type"`
  SortIndex  int64  `json:"sort_index"`
}
func adminCategoryUpdate(context echo.Context) error {
  id := context.Param("id")
  original, err := library.CategoryRead(id)
  if err == library.ErrNotFound { return json404(context) }
  if err != nil { return debug500(context, err) }

  changes := CategoryUpdateRequest{}
  if err = context.Bind(&changes); err != nil { return json400(context, err) }
  changes.Name      = strings.TrimSpace(changes.Name)
  if changes.Name       == "" { changes.Name      = original.Name              }
  if changes.MediaType  == "" { changes.MediaType = string(original.MediaType) }

  if (changes.Name != original.Name) || (changes.MediaType != string(original.MediaType)) {
    err = library.CategoryUpdate(original, changes.Name, changes.MediaType)
    if err == library.ErrQueryFailed { return debug500(context, err) }
    if err != nil { return json400(context, err) }
  }

  if changes.SortIndex != original.SortIndex {
    err = library.CategoryReindex(original, changes.SortIndex)
    if err == library.ErrQueryFailed { return debug500(context, err) }
    if err != nil { return json400(context, err) }
  }
  return json200(context, map[string]string{})
}

func adminCategoryDelete(context echo.Context) error {
  id := context.Param("id")
  category, err := library.CategoryRead(id)
  if err == library.ErrNotFound { return json404(context) }
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }

  err = library.CategoryDelete(category)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }
  return json200(context, map[string]string{})
}

// ============================================================================
// Metadata

func adminMetadataTree(context echo.Context) error {
  var err error
  lost_items := library.MetadataTreeNode{Id: "lost", Name: "Lost Items", MediaType: ""}
  lost_items.Children, err = library.MetadataParentTree("")
  if err != nil { return debug500(context, err) }

  categories, err := library.CategoryList()
  if err != nil { return debug500(context, err) }

  root_items := make([]*library.MetadataTreeNode, len(categories)+1)
  root_items[0] = &lost_items
  for index, cat := range categories {
    cat_tree := library.MetadataTreeNode{Id: cat.Id, Name: cat.Name, MediaType: string(cat.MediaType)}
    cat_tree.Children, err = library.MetadataParentTree(cat.Id)
    if err != nil { return debug500(context, err) }
    root_items[index+1] = &cat_tree
  }

  return json200(context, root_items)
}

func adminMetadataByParentList(context echo.Context) error {
  parent_id := context.Param("parent_id")
  if parent_id == "lost" { parent_id = "" }
  metadata_list, err := library.MetadataForParent(parent_id)
  if err != nil { return debug500(context, err) }
  return json200(context, metadata_list)
}

func adminMetadataCreate(context echo.Context) error {
  metadata := library.Metadata{}
  if err := context.Bind(&metadata); err != nil { return json400(context, err) }
  err := library.MetadataCreate(&metadata)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }
  return context.JSON(http.StatusCreated, metadata)
}

func adminMetadataUpdate(context echo.Context) error {
  id := context.Param("id")
  original, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return json404(context) }
  if err != nil { return debug500(context, err) }

  changes := map[string]string{}
  if err = context.Bind(&changes); err != nil { return json400(context, err) }

  var ok bool
  var new_parent       string = "" ; new_parent,       ok = changes["parent_id"]    ; if !ok { new_parent       = original.ParentId    }
  var new_name_display string = "" ; new_name_display, ok = changes["name_display"] ; if !ok { new_name_display = original.NameDisplay }
  var new_name_sort    string = "" ; new_name_sort,    ok = changes["name_sort"]    ; if !ok { new_name_sort    = original.NameSort    }

  if new_parent != original.ParentId {
    err = original.Reparent(new_parent)
    if err == library.ErrQueryFailed { return debug500(context, err) }
    if err != nil { return json400(context, err) }
  }
  if (new_name_display != original.NameDisplay) || (new_name_sort != original.NameSort) {
    err = original.Rename(new_name_display, new_name_sort)
    if err == library.ErrQueryFailed { return debug500(context, err) }
    if err != nil { return json400(context, err) }
  }

  resetMetadataPosterCache(id)
  return json200(context, map[string]string{})
}

func adminMetadataPoster(context echo.Context) error {
  id := context.Param("id")
  md, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return json404(context) }
  if err != nil { return debug500(context, err) }

  file_header, err := context.FormFile("poster")
  if err != nil { return json400(context, err) }

  file_handle, err := file_header.Open()
  if err != nil { return debug500(context, err) }
  defer file_handle.Close()

  img, _, err := image.Decode(file_handle)
  if err != nil { return json400(context, err) }

  err = md.SetPoster(img)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }

  resetMetadataPosterCache(id)
  return json200(context, map[string]string{})
}

type MetadataDeleteRequest struct {
  DeleteChildren bool `json:"delete_children"`
}
func adminMetadataDelete(context echo.Context) error {
  id := context.Param("id")
  md, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return json404(context) }
  if err != nil { return debug500(context, err) }
  request := MetadataDeleteRequest{}
  if err := context.Bind(&request); err != nil { return json400(context, err) }

  err = library.MetadataDelete(md, request.DeleteChildren)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }
  resetMetadataPosterCache(id)
  return json200(context, map[string]string{})
}

// ============================================================================
// InputFile

func adminInputFileList(context echo.Context) error {
  input_files, err := library.InputFileList()
  if err != nil { return debug500(context, err) }
  return json200(context, input_files)
}

func adminInputFileDelete(context echo.Context) error {
  id := context.Param("id")
  inp, err := library.InputFileRead(id)
  if err == library.ErrNotFound { return json404(context) }
  if err != nil { return debug500(context, err) }

  err = library.InputFileDelete(inp)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }
  return json200(context, map[string]string{})
}

func adminInputFileMap(context echo.Context) error {
  id := context.Param("id")
  inp, err := library.InputFileRead(id)
  if err == library.ErrNotFound { return json404(context) }
  if err != nil { return debug500(context, err) }

  stream_map := []int64{}
  if err = context.Bind(&stream_map); err != nil { return json400(context, err) }

  err = inp.Remap(stream_map)
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }
  return json200(context, map[string]string{})
}

func adminInputFileReset(context echo.Context) error {
  id := context.Param("id")
  inp, err := library.InputFileRead(id)
  if err == library.ErrNotFound { return json404(context) }
  if err != nil { return debug500(context, err) }

  err = inp.StatusReset()
  if err == library.ErrQueryFailed { return debug500(context, err) }
  if err != nil { return json400(context, err) }
  return json200(context, map[string]string{})
}
