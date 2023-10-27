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

  server.Get("/admin/unprocessed",          adminUnprocessedList)
  server.Post("/admin/unprocessed/map/:id", adminUnprocessedMap)
  server.Post("/admin/unprocessed/queue",   adminUnprocessedQueue)
}

// DUH!!!! Make scanning a separate process as well. No need to stick it in the admin part.
/*
Transcoding Task
  id, unprocessed_id, time_started, time_ran, status (todo, running, success, failure), error_message, command_line 

Separate Transcoding Process, launched externally to web app
  pick up (status == "todo") jobs, mark it in progress, run transcoding, read result, update task & unprocessed records
Admin app just adds tasks to be ran later
*/
