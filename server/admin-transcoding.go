package main

import (
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

func adminTranscodingList(context *fiber.Ctx) error {
  trans_list, err := database.TranscodingTaskList(DB)
  if err != nil { return debug500(context, err) }
  return context.JSON(trans_list)
}

func adminTranscodingEmpty(context *fiber.Ctx) error {
  _, err := DB.Exec(`DELETE FROM transcoding_tasks;`)
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

func adminTranscodingClean(context *fiber.Ctx) error {
  // TODO: allow passing options for (success|failure|either)
  _, err := DB.Exec(`DELETE FROM transcoding_tasks WHERE status = 'success';`)
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

func adminTranscodingDelete(context *fiber.Ctx) error {
  id := context.Params("id")
  transcoding, err := database.TranscodingTaskRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  // delete item
  err = transcoding.Delete(DB)
  if err != nil { return debug500(context, err) }

  return context.SendStatus(200)
}
