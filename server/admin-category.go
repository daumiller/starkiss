package main

import (
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

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
  proposed := original
  if changes["name"] != "" { proposed.Name = changes["name"] }
  if changes["type"] != "" { proposed.Type = database.CategoryType(changes["type"]) }

  err = database.TableUpdate(DB, &original, &proposed)
  if err == database.ErrValidationFailed { return context.Status(400).JSON(map[string]string{"error": err.Error()}) }
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}

func adminCategoryDelete(context *fiber.Ctx) error {
  id := context.Params("id")
  category, err := database.CategoryRead(DB, id)
  if err == database.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }
  err = database.TableDelete(DB, &category)
  if err != nil { return debug500(context, err) }
  return context.SendStatus(200)
}
