package database

import (
  "os"
  "testing"
)

func TestOpenClose(test *testing.T) {
  Location = "./test.database"
  _ = os.Remove(Location)
  defer os.Remove(Location)

  // Open() creates file if it doesn't exist
  db, err := Open()
  if err != nil {
    test.Errorf("TestOpenClose: Open failed: %s", err)
  } else {
    err = Close(db)
    if err != nil { test.Errorf("TestOpenClose: Close failed (after Open): %s", err) }
  }

  // Open() should fail for an invalid file
  Location = "./test.dir"
  err = os.Mkdir(Location, 0755)
  if err != nil { test.Fatalf("TestOpenClose: couldn't create test directory: %s", err) }
  defer os.Remove(Location)
  db, err = Open()
  if err == nil {
    test.Error("TestOpenClose: Open succeeded when pointed at a directory")
    _ = Close(db)
  }
}

func TestTransactions(test *testing.T) {
  Location = "./test.database"
  _ = os.Remove(Location)
  defer os.Remove(Location)

  db, err := Open()
  if err != nil { test.Fatalf("TestTransactions: Open failed: %s", err) }
  defer Close(db)

  _, err = db.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);`)
  if err != nil { test.Fatalf("TestTransactions: CREATE TABLE failed: %s", err) }

  // committed transaction should read back changes
  err = TransactionBegin(db)
  if err != nil { test.Fatalf("TestTransactions: TransactionBegin failed: %s", err) }
  _, err = db.Exec(`INSERT INTO test (id, name) VALUES (1, "committed");`)
  err = TransactionCommit(db)
  if err != nil { test.Fatalf("TestTransactions: TransactionCommit failed: %s", err) }
  row := db.QueryRow(`SELECT id, name FROM test WHERE name = 'committed';`)
  var id int ; var name string
  err = row.Scan(&id, &name)
  if err != nil { test.Fatalf("TestTransactions: SELECT failed for committed transaction: %s", err) }
  if id != 1 { test.Fatalf("TestTransactions: SELECT returned wrong id: %d", id) }
  if name != "committed" { test.Fatalf("TestTransactions: SELECT returned wrong name: %s", name) }

  // rolled back transaction should not read back changes
  err = TransactionBegin(db)
  if err != nil { test.Fatalf("TestTransactions: TransactionBegin failed: %s", err) }
  _, err = db.Exec(`INSERT INTO test (id, name) VALUES (2, "rollback");`)
  if err != nil { test.Fatalf("TestTransactions: INSERT failed: %s", err) }
  err = TransactionRollback(db)
  if err != nil { test.Fatalf("TestTransactions: TransactionRollback failed: %s", err) }
  row = db.QueryRow(`SELECT id, name FROM test WHERE name = 'rollback';`)
  err = row.Scan(&id, &name)
  if err == nil { test.Fatalf("TestTransactions: SELECT returned row for rolled back transaction") }
}

func TestBackups(test *testing.T) {
  Location = "./test.database"
  _ = os.Remove(Location)
  defer os.Remove(Location)
  defer os.Remove(Location + ".bak")

  db, err := Open()
  if err != nil { test.Fatalf("TestBackups: Open failed: %s", err) }
  _, err = db.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);`)
  if err != nil { test.Fatalf("TestTransactions: CREATE TABLE failed: %s", err) }
  _, err = db.Exec(`INSERT INTO test (id, name) VALUES (3, "three");`)
  if err != nil { test.Fatalf("TestTransactions: INSERT failed: %s", err) }
  _, err = db.Exec(`INSERT INTO test (id, name) VALUES (32, "thirty-two");`)
  if err != nil { test.Fatalf("TestTransactions: INSERT failed: %s", err) }
  err = Close(db)
  if err != nil { test.Fatalf("TestBackups: Close failed: %s", err) }

  err = BackupCreate()
  if err != nil { test.Fatalf("TestBackups: BackupCreate failed: %s", err) }

  db, err = Open()
  if err != nil { test.Fatalf("TestBackups: Open failed: %s", err) }
  row := db.QueryRow(`SELECT name FROM test WHERE id = 3;`)
  var name string
  err = row.Scan(&name)
  if err != nil { test.Fatalf("TestBackups: SELECT failed: %s", err) }
  if name != "three" { test.Fatalf("TestBackups: SELECT returned wrong name: %s", name) }
  _, err = db.Exec(`UPDATE test SET name = "three-fifty" WHERE id = 3;`)
  if err != nil { test.Fatalf("TestBackups: UPDATE failed: %s", err) }
  row = db.QueryRow(`SELECT name FROM test WHERE id = 3;`)
  err = row.Scan(&name)
  if err != nil { test.Fatalf("TestBackups: SELECT failed: %s", err) }
  if name != "three-fifty" { test.Fatalf("TestBackups: SELECT returned wrong name: %s", name) }
  Close(db)

  err = BackupRestore()
  if err != nil { test.Fatalf("TestBackups: BackupRestore failed: %s", err) }
  db, err = Open()
  if err != nil { test.Fatalf("TestBackups: Open failed: %s", err) }
  row = db.QueryRow(`SELECT name FROM test WHERE id = 3;`)
  err = row.Scan(&name)
  if err != nil { test.Fatalf("TestBackups: SELECT failed: %s", err) }
  if name != "three" { test.Fatalf("TestBackups: SELECT returned wrong name: %s", name) }
  Close(db)
}
