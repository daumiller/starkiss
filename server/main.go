package main

import (
  "os"
  "fmt"
  "github.com/labstack/echo/v4"
  "github.com/daumiller/starkiss/library"
)

// globals & defaults
var DBFILE     string  = "starkiss.db"
var ADDRESS    string  = ":4331" // server binding address; can be overridden by environment variable
var DEBUG      bool    = false   // debug mode; can be overridden by environment variable
var JWT_KEY    []byte  = nil     // JWT key; created or read from DB in propertiesMain()

// simple debug handler for 500s
func debug500(context echo.Context, err error) error {
  if DEBUG == false { return context.JSON(500, map[string]string{"error": "internal server error"}) }
  return context.JSON(500, map[string]string{"error": err.Error()})
}
func json200(context echo.Context, data interface{}) error {
  return context.JSON(200, data)
}
func json400(context echo.Context, err error) error {
  return context.JSON(400, map[string]string{"error": err.Error()})
}
func json404(context echo.Context) error {
  return context.JSON(404, map[string]string{"error": "record not found"})
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
  // TODO: need an option (environment variable?) to skip auto-migration
  exit_code := startupMigration("latest")
  if exit_code != 0 { os.Exit(exit_code) }

  // get default properties
  startupProperties()

  // ensure Library is ready
  err = library.LibraryReady()
  if err != nil { fmt.Printf("Error starting library; not ready: %s\n", err.Error()) ; os.Exit(-1) }

  // startup server, and register routes
  server := echo.New()
  server.Static("/web-admin", "./../web-admin")
  startupMediaRoutes(server)
  startupClientRoutes(server)
  startupAdminRoutes(server)

  server.Logger.Fatal(server.Start(ADDRESS))
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
