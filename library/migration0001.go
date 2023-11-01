package library

type migration0001 struct {}

func (m *migration0001) Up() (err error) {
  err = createTableProperties() ; if err != nil { return err }
  err = createTableCategories() ; if err != nil { return err }
  err = createTableMetadata()   ; if err != nil { return err }
  err = createTableInputFiles() ; if err != nil { return err }

  return nil
}

func (m *migration0001) Down() (err error) {
  _, err = dbHandle.Exec(`DROP TABLE input_files;` ) ; if err != nil { return err }
  _, err = dbHandle.Exec(`DROP TABLE metadata;`    ) ; if err != nil { return err }
  _, err = dbHandle.Exec(`DROP TABLE categories;`  ) ; if err != nil { return err }
  _, err = dbHandle.Exec(`DROP TABLE properties;`  ) ; if err != nil { return err }

  return nil
}

func createTableProperties() (err error) {
  _, err = dbHandle.Exec(`CREATE TABLE properties (
    id    INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
    key   TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL
  );`)
  if err != nil { return err }

  _, err = dbHandle.Exec(`CREATE UNIQUE INDEX property_key ON properties (key);`)
  return err
}

func createTableCategories() (err error) {
  _, err = dbHandle.Exec(`CREATE TABLE categories (
    id         TEXT NOT NULL PRIMARY KEY UNIQUE,
    media_type TEXT NOT NULL,
    name       TEXT NOT NULL UNIQUE
  );`)
  if err != nil { return err }

  _, err = dbHandle.Exec(`CREATE UNIQUE INDEX categories_name ON categories (name);     `) ; if err != nil { return err }
  _, err = dbHandle.Exec(`CREATE INDEX categories_media_type ON categories (media_type);`) ; if err != nil { return err }
  return nil
}

func createTableMetadata() (err error) {
  _, err = dbHandle.Exec(`CREATE TABLE metadata (
    id                TEXT NOT NULL PRIMARY KEY UNIQUE,
    parent_id         TEXT NOT NULL,
    media_type        TEXT NOT NULL,
    name_display      TEXT NOT NULL,
    name_sort         TEXT NOT NULL,
    streams           TEXT NOT NULL,
    duration          INTEGER NOT NULL,
    size              INTEGER NOT NULL
  );`)
  if err != nil { return err }

  _, err = dbHandle.Exec(`CREATE INDEX metadata_parent       ON metadata (parent_id);    ` ) ; if err != nil { return err }
  _, err = dbHandle.Exec(`CREATE INDEX metadata_media_type   ON metadata (media_type);   ` ) ; if err != nil { return err }
  _, err = dbHandle.Exec(`CREATE INDEX metadata_name_display ON metadata (name_display); ` ) ; if err != nil { return err }
  return nil
}

func createTableInputFiles() (err error) {
  // transcoded_location should be unique, but only once populated, so... not unique
  _, err = dbHandle.Exec(`CREATE TABLE input_files (
    id                       TEXT NOT NULL PRIMARY KEY UNIQUE,
    source_location          TEXT NOT NULL UNIQUE,
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
