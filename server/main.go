package main

import (
  "os"
  "fmt"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/library"
)

// globals & defaults
var DBFILE     string  = "starkiss.db"
var ADDRESS    string  = ":4331" // server binding address; can be overridden by environment variable
var DEBUG      bool    = false   // debug mode; can be overridden by environment variable
var JWT_KEY    []byte  = nil     // JWT key; created or read from DB in propertiesMain()

// simple debug handler for 500s
func debug500(context *fiber.Ctx, err error) error {
  if DEBUG == false { return context.Status(500).JSON(map[string]string{"error": "internal server error"}) }
  return context.Status(500).JSON(map[string]string{"error": err.Error()})
}
func json200(context *fiber.Ctx, data interface{}) error {
  return context.Status(200).JSON(data)
}
func json400(context *fiber.Ctx, err error) error {
  return context.Status(400).JSON(map[string]string{"error": err.Error()})
}
func json404(context *fiber.Ctx) error {
  return context.Status(404).JSON(map[string]string{"error": "record not found"})
}

func main() {
  // check for environment variables
  startupEnvironment()

  // startup Library
  err := library.LibraryStartup(DBFILE)
  if err != nil { fmt.Printf("Error starting library: %s\n", err.Error()) ; os.Exit(-1) }
  defer library.LibraryShutdown()

  // check for command line arguments
  startupCommands()

  // update database to latest migration (creating DB if necessary)
  exit_code := startupMigration("latest")
  if exit_code != 0 { os.Exit(exit_code) }

  // get default properties
  startupProperties()

  // ensure Library is ready
  err = library.LibraryReady()
  if err != nil { fmt.Printf("Error starting library; not ready: %s\n", err.Error()) ; os.Exit(-1) }

  // startup server, and register routes
  server := fiber.New()
  server.Static("/web-admin", "./../web-admin")
  startupMediaRoutes(server)
  startupClientRoutes(server)
  startupAdminRoutes(server)

  // 404 handler
  server.Use(func(context *fiber.Ctx) error {
    return context.SendStatus(404)
  })

  server.Listen(ADDRESS)
}

// Look for environment variables. If present, override defaults.
func startupEnvironment() {
  if os.Getenv("DEBUG")   == "true" { DEBUG   = true                 }
  if os.Getenv("ADDRESS") != ""     { ADDRESS = os.Getenv("ADDRESS") }
  if os.Getenv("DBFILE")  != ""     { DBFILE  = os.Getenv("DBFILE")  }
}

// Look for command line arguments. If present, execute them and exit.
func startupCommands() {
  if len(os.Args) < 2 { return }
  switch os.Args[1] {
    case "migrate":
      if len(os.Args) > 2 {
        os.Exit(startupMigration(os.Args[2]))
      } else {
        os.Exit(startupMigration("latest"))
      }
    default:
      fmt.Printf("Unknown command: \"%s\"\n", os.Args[1])
      os.Exit(-1)
  }
}
