package main

import (
  "github.com/gofiber/fiber/v2"
)

func startupAdminRoutes(server *fiber.App) {
  server.Get ("/admin/paths", adminPathList  )
  server.Post("/admin/paths", adminPathUpdate)

  server.Get   ("/admin/categories",   adminCategoryList  )
  server.Post  ("/admin/category",     adminCategoryCreate)
  server.Post  ("/admin/category/:id", adminCategoryUpdate)
  server.Delete("/admin/category/:id", adminCategoryDelete)

  server.Get   ("/admin/unprocessed",          adminUnprocessedList      )
  server.Delete("/admin/unprocessed",          adminUnprocessedEmpty     ) // empty unprocessed table
  server.Post  ("/admin/unprocessed/map/:id",  adminUnprocessedMap       ) // add stream map to unprocessed record
  server.Post  ("/admin/unprocessed/queue",    adminUnprocessedQueue     ) // add unprocessed items to transcoding queue
  server.Post  ("/admin/unprocessed/complete", adminUnprocessedComplete  ) // move completed unprocessed items (moves transcoded file to final location, updates metadata as live)
  server.Delete("/admin/unprocessed/:id",      adminUnprocessedDelete    ) // remove item from unprocessed list

  server.Get   ("/admin/transcoding",       adminTranscodingList  )
  server.Delete("/admin/transcoding",       adminTranscodingEmpty ) // empty transcoding queue
  server.Post  ("/admin/transcoding/clean", adminTranscodingClean ) // clean transcoding queue (of completed tasks (success or failure options))
  server.Delete("/admin/transcoding/:id",   adminTranscodingDelete) // remove item from transcoding queue  

  // TODO: metadata matching & record creation
}
