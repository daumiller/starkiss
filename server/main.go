package main

import (
  "os"
  "fmt"
  "database/sql"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

// globals & defaults
var DB         *sql.DB = nil     // global handle to database; initialized in main()
var ADDRESS    string  = ":4331" // server binding address; can be overridden by environment variable
var DEBUG      bool    = false   // debug mode; can be overridden by environment variable
var JWTKEY     []byte  = nil     // JWT key; created or read from DB in propertiesMain()
var MEDIAPATH  string = ""       // media path; read from DB in propertiesMain()

// simple debug handler for 500s
func debug500(context *fiber.Ctx, err error) error {
  if DEBUG == false { return context.SendStatus(500) }
  return context.Status(500).SendString(err.Error())
}

func main() {
  // check for environment variables and command-line arguments
  startupEnvironment()
  startupCommands()

  // update database to latest migration (creating DB if necessary)
  exit_code := startupMigration("latest")
  if exit_code != 0 { os.Exit(exit_code) }

  // get database handle, get default properties
  var err error
  DB, err = database.Open()
  if err != nil { fmt.Printf("Error opening database: \"%s\"\n", err.Error()); os.Exit(-1) }
  defer database.Close(DB)
  startupProperties()

  // startup server, and register routes
  server := fiber.New()
  server.Static("/web", "./../web")
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
  if os.Getenv("DEBUG")   == "true" { DEBUG = true }
  if os.Getenv("ADDRESS") != ""     { ADDRESS = os.Getenv("ADDRESS") }
  if os.Getenv("DBFILE")  != ""     { database.Location = os.Getenv("DBFILE") }
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
