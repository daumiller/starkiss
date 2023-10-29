package main

import (
  "github.com/gofiber/fiber/v2"
)

func startupAdminRoutes(server *fiber.App) {
  server.Get ("/admin/media-path", adminMediaPathList  )
  server.Post("/admin/media-path", adminMediaPathUpdate)

  server.Get   ("/admin/categories",   adminCategoryList  )
  server.Post  ("/admin/category",     adminCategoryCreate)
  server.Post  ("/admin/category/:id", adminCategoryUpdate)
  server.Delete("/admin/category/:id", adminCategoryDelete)

  server.Get   ("/admin/metadata/by-parent/:parent_id", adminMetadataByParentList) // list all metadata for a parent
  server.Post  ("/admin/metadata",                      adminMetadataCreate      ) // create new metadata (defaults to hidden) (requires a parent id, optionally specify a input-file for read-or-create)
  server.Post  ("/admin/metadata/:id",                  adminMetadataUpdate      ) // update metadata
  server.Delete("/admin/metadata/:id",                  adminMetadataDelete      ) // delete metadata (options for what to do with any children records)

  server.Get   ("/admin/input-files",             adminInputFileList    ) // list all input files
  server.Delete("/admin/input-files",             adminInputFileDelete  ) // delete input file(s) (and transcoded file(s), if they exist)
  server.Post  ("/admin/input-file/:id/map",      adminInputFileMap     ) // update stream_map
  server.Post  ("/admin/input-file/:id/reset",    adminInputFileReset   ) // reset transcoding values (and delete transcoded file, if exists)
  server.Post  ("/admin/input-file/:id/complete", adminInputFileComplete) // validate transcoding & metadata, move file to final destination, remove input-file record
}
