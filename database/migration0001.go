package database

import (
  "database/sql"
)

type Migration0001 struct {}

func (m *Migration0001) Up() (err error) {
  db, err := Open()
  if err != nil { return err }
  defer Close(db)

  err = createTableProperties(db) ; if err != nil { return err }
  err = createTableCategories(db) ; if err != nil { return err }
  err = createTableMetadata(db)   ; if err != nil { return err }
  err = createTableInputFiles(db) ; if err != nil { return err }

  return nil
}

func (m *Migration0001) Down() (error) {
  db, err := Open()
  if err != nil { return err }
  defer Close(db)

  _, err = db.Exec(`DROP TABLE input_files;` ) ; if err != nil { return err }
  _, err = db.Exec(`DROP TABLE metadata;`    ) ; if err != nil { return err }
  _, err = db.Exec(`DROP TABLE categories;`  ) ; if err != nil { return err }
  _, err = db.Exec(`DROP TABLE properties;`  ) ; if err != nil { return err }

  return nil
}

func createTableProperties(db *sql.DB) (err error) {
  _, err = db.Exec(`CREATE TABLE properties (
    id    INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
    key   TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL
  );`)
  if err != nil { return err }

  _, err = db.Exec(`CREATE UNIQUE INDEX property_key ON properties (key);`)
  return err
}

func createTableCategories(db *sql.DB) (err error) {
  _, err = db.Exec(`CREATE TABLE categories (
    id         TEXT NOT NULL PRIMARY KEY UNIQUE,
    media_type TEXT NOT NULL,
    name       TEXT NOT NULL UNIQUE
  );`)
  if err != nil { return err }

  _, err = db.Exec(`CREATE UNIQUE INDEX categories_name ON categories (name);     `) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX categories_media_type ON categories (media_type);`) ; if err != nil { return err }
  return nil
}

func createTableMetadata(db *sql.DB) (err error) {
  _, err = db.Exec(`CREATE TABLE metadata (
    id                TEXT NOT NULL PRIMARY KEY UNIQUE,
    parent_id         TEXT NOT NULL,
    media_type        TEXT NOT NULL,
    name_display      TEXT NOT NULL,
    name_sort         TEXT NOT NULL,
    streams           TEXT NOT NULL,
    duration          INTEGER NOT NULL,
    size              INTEGER NOT NULL,
    hidden            INTEGER NOT NULL
  );`)
  if err != nil { return err }

  _, err = db.Exec(`CREATE INDEX metadata_parent       ON metadata (parent_id);    ` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX metadata_media_type   ON metadata (media_type);   ` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX metadata_name_display ON metadata (name_display); ` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX metadata_hidden       ON metadata (hidden);       ` ) ; if err != nil { return err }
  return nil
}

func createTableInputFiles(db *sql.DB) (err error) {
  // transcoded_location should be unique, but only once populated, so... not unique
  _, err = db.Exec(`CREATE TABLE input_files (
    id                       TEXT NOT NULL PRIMARY KEY UNIQUE,
    source_location          TEXT NOT NULL UNIQUE,
    transcoded_location      TEXT NOT NULL,
    source_streams           TEXT NOT NULL,
    stream_map               TEXT NOT NULL,
    source_duration          INTEGER NOT NULL,
    time_scanned             INTEGER NOT NULL,
    transcoding_command      TEXT NOT NULL,
    transcoding_time_started INTEGER NOT NULL,
    transcoding_time_elapsed INTEGER NOT NULL,
    transcoding_error        TEXT NOT NULL
  );`)
  if err != nil { return err }
  return nil
}
