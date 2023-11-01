package library

type migration0000 struct {}

func (m *migration0000) Up() (error) {
  println("Migration0000.Up()")
  return nil
}

func (m *migration0000) Down() (error) {
  return nil
}
