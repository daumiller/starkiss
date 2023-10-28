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
  server.Post("/admin/unprocessed/map/:id", adminUnprocessedMap)   // add stream map to unprocessed record
  server.Post("/admin/unprocessed/queue",   adminUnprocessedQueue) // add unprocessed item to transcoding queue
  // remove item from unprocessed list (and corresponding provisional entry)
  // empty unprocessed table (and provisional table)
  // move completed unprocessed items
  //  - move provisional to metadata
  //  - move transcoded file, or copy no-transcoding-needed file, to media location

  // empty transcoding queue
  // clean transcoding queue (of completed tasks (success or failure options))
  // transcoder command line option to run continuously (polling DB every x seconds)
}
