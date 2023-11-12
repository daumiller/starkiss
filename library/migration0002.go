package library

type migration0002 struct {}

func (m *migration0002) Up() (err error) {
  _, err = dbHandle.Exec(`ALTER TABLE categories ADD COLUMN sort_index INTEGER NOT NULL DEFAULT 9999;`)
  return err
}

func (m *migration0002) Down() (err error) {
  _, err = dbHandle.Exec(`ALTER TABLE categories DROP COLUMN sort_index;`)
  return err
}
