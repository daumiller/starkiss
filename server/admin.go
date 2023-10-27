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

  server.Get("/admin/unprocessed",          adminUnprocessedList)
  server.Post("/admin/unprocessed/map/:id", adminUnprocessedMap)
  server.Post("/admin/unprocessed/queue",   adminUnprocessedQueue)
}
