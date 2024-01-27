package library

type migration0003 struct {}

func (m *migration0003) Up() (err error) {
  err = createTableUsers()        ; if err != nil { return err }
  err = createTableUserMetadata() ; if err != nil { return err }
  err = createDefaultUser()       ; if err != nil { return err }

  return nil
}

func (m *migration0003) Down() (err error) {
  _, err = dbHandle.Exec(`DROP TABLE users;`       ) ; if err != nil { return err }
  _, err = dbHandle.Exec(`DROP TABLE usermetadata;`) ; if err != nil { return err }

  return nil
}

func createTableUsers() (err error) {
  _, err = dbHandle.Exec(`CREATE TABLE users (
    id       TEXT NOT NULL PRIMARY KEY UNIQUE,
    name     TEXT NOT NULL UNIQUE,
    password TEXT
  );`)
  if err != nil { return err }

  _, err = dbHandle.Exec(`CREATE UNIQUE INDEX user_name ON users (name);`)
  return err
}

func createTableUserMetadata() (err error) {
  _, err = dbHandle.Exec(`CREATE TABLE user_metadata (
    id          TEXT NOT NULL PRIMARY KEY UNIQUE,
    user_id     TEXT NOT NULL,
    metadata_id TEXT NOT NULL,
    started     INTEGER NOT NULL,
    timestamp   INTEGER NOT NULL
  );`)
  if err != nil { return err }

  _, err = dbHandle.Exec(`CREATE INDEX user_metadata_user     ON user_metadata (user_id);     ` ) ; if err != nil { return err }
  _, err = dbHandle.Exec(`CREATE INDEX user_metadata_metadata ON user_metadata (metadata_id); ` ) ; if err != nil { return err }
  return nil
}

func createDefaultUser() (err error) {
  // this will eventually be removed for actual multi-user support
  _, err = dbHandle.Exec(`INSERT INTO users (id, name, password) VALUES ('8e5624ad-8ce5-4d86-ac5c-1d2ecf120d05', 'default', '');`)
  return err
}
