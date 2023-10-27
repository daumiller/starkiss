package database

type Migration0000 struct {}

func (m *Migration0000) Up() (error) {
  println("Migration0000.Up()")
  return nil
}

func (m *Migration0000) Down() (error) {
  return nil
}
