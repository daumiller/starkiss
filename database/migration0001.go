package database

import (
  "database/sql"
)

type Migration0001 struct {}

func (m *Migration0001) Up() (err error) {
  db, err := Open()
  if err != nil { return err }
  defer Close(db)

  err = createTableProperties(db)              ; if err != nil { return err }
  err = createTableCategories(db)              ; if err != nil { return err }
  err = createTableMetadata(db, "metadata")    ; if err != nil { return err }
  err = createTableMetadata(db, "provisional") ; if err != nil { return err }
  err = createTableUnprocessed(db)             ; if err != nil { return err }
  err = createTableTranscodingTasks(db)        ; if err != nil { return err }

  return nil
}

func (m *Migration0001) Down() (error) {
  db, err := Open()
  if err != nil { return err }
  defer Close(db)

  _, err = db.Exec(`DROP TABLE transcoding_tasks;`) ; if err != nil { return err }
  _, err = db.Exec(`DROP TABLE unprocessed;`      ) ; if err != nil { return err }
  _, err = db.Exec(`DROP TABLE provisional;`      ) ; if err != nil { return err }
  _, err = db.Exec(`DROP TABLE metadata;`         ) ; if err != nil { return err }
  _, err = db.Exec(`DROP TABLE categories;`       ) ; if err != nil { return err }
  _, err = db.Exec(`DROP TABLE properties;`       ) ; if err != nil { return err }

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
    id   TEXT NOT NULL PRIMARY KEY UNIQUE,
    type TEXT NOT NULL,
    name TEXT NOT NULL UNIQUE
  );`)
  if err != nil { return err }

  _, err = db.Exec(`CREATE UNIQUE INDEX categories_name ON categories (name);`) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX categories_type ON categories (type);       `) ; if err != nil { return err }
  return nil
}

func createTableUnprocessed(db *sql.DB) (err error) {
  // transcoded_location and provisional_id should be unique, but only once populated, so... not unique
  _, err = db.Exec(`CREATE TABLE unprocessed (
    id                  TEXT NOT NULL PRIMARY KEY UNIQUE,
    needs_stream_map    INTEGER NOT NULL,
    needs_transcoding   INTEGER NOT NULL,
    needs_metadata      INTEGER NOT NULL,
    source_location     TEXT NOT NULL UNIQUE,
    source_streams      TEXT NOT NULL,
    source_container    TEXT NOT NULL,
    transcoded_location TEXT NOT NULL,
    transcoded_streams  TEXT NOT NULL,
    match_data          TEXT NOT NULL,
    provisional_id      TEXT NOT NULL,
    created_at          INTEGER NOT NULL
  );`)
  if err != nil { return err }

  _, err = db.Exec(`CREATE INDEX unprocessed_nsm ON unprocessed (needs_stream_map);` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX unprocessed_ntc ON unprocessed (needs_transcoding);`) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX unprocessed_nmd ON unprocessed (needs_metadata);`   ) ; if err != nil { return err }
  return nil
}

func createTableTranscodingTasks(db *sql.DB) (err error) {
  _, err = db.Exec(`CREATE TABLE transcoding_tasks (
    id              TEXT NOT NULL PRIMARY KEY UNIQUE,
    unprocessed_id  TEXT NOT NULL UNIQUE,
    time_started    INTEGER NOT NULL,
    time_elapsed    INTEGER NOT NULL,
    status          TEXT NOT NULL,
    error_message   TEXT NOT NULL,
    command_line    TEXT NOT NULL
  );`)
  if err != nil { return err }

  _, err = db.Exec(`CREATE INDEX transcoding_tasks_unprocessed_id ON transcoding_tasks (unprocessed_id);`)
  return err
}

func createTableMetadata(db *sql.DB, table_name string) (err error) {
  _, err = db.Exec(`CREATE TABLE ` + table_name + ` (
    id                TEXT NOT NULL PRIMARY KEY UNIQUE,
    category_id       TEXT NOT NULL,
    category_type     TEXT NOT NULL,
    parent_id         TEXT NOT NULL,
    grandparent_id    TEXT NOT NULL,
    type              TEXT NOT NULL,
    title_user        TEXT NOT NULL,
    title_sort        TEXT NOT NULL,
    match_data        TEXT NOT NULL,
    description_short TEXT NOT NULL,
    description_long  TEXT NOT NULL,
    genre             TEXT NOT NULL,
    release_year      INTEGER NOT NULL,
    release_month     INTEGER NOT NULL,
    release_day       INTEGER NOT NULL,
    sibling_index     INTEGER NOT NULL,
    has_poster        INTEGER NOT NULL,
    location          TEXT NOT NULL UNIQUE,
    size              INTEGER NOT NULL,
    duration          INTEGER NOT NULL,
    streams           TEXT NOT NULL
  );`)
  if err != nil { return err }

  _, err = db.Exec(`CREATE INDEX ` + table_name + `_parent   ON ` + table_name + ` (parent_id);    ` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX ` + table_name + `_cat_type ON ` + table_name + ` (category_type);` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX ` + table_name + `_type     ON ` + table_name + ` (type);         ` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX ` + table_name + `_title    ON ` + table_name + ` (title_user);   ` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX ` + table_name + `_genre    ON ` + table_name + ` (genre);        ` ) ; if err != nil { return err }
  _, err = db.Exec(`CREATE INDEX ` + table_name + `_year     ON ` + table_name + ` (release_year); ` ) ; if err != nil { return err }
  return nil
}
