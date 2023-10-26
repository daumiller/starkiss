package main

import (
  "github.com/gofiber/fiber/v2"
)

func startupAdminRoutes(server *fiber.App) {
  server.Get ("/admin/paths", adminPathList  )
  server.Post("/admin/paths", adminPathUpdate)

  server.Get   ("/admin/categories",   adminCategoryList  )
  server.Put   ("/admin/category",     adminCategoryCreate)
  server.Post  ("/admin/category/:id", adminCategoryUpdate)
  server.Delete("/admin/category/:id", adminCategoryDelete)

  adminScannerStartup()
  server.Get("/admin/scanner",        adminScannerStatus)
  server.Post("/admin/scanner/start", adminScannerStart )
  server.Post("/admin/scanner/stop",  adminScannerStop  )
}
