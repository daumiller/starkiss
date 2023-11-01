package library

import (
  "os"
  "testing"
)

func TestTransactions(test *testing.T) {
  testDbPath := "./test.database"
  _ = os.Remove(testDbPath)

  err := LibraryStartup(testDbPath)
  if err != nil { test.Fatalf("TestTransactions: Open failed: %s", err) }
  defer os.Remove(testDbPath)
  defer LibraryShutdown()

  _, err = dbHandle.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);`)
  if err != nil { test.Fatalf("TestTransactions: CREATE TABLE failed: %s", err) }

  // committed transaction should read back changes
  err = dbTransactionBegin()
  if err != nil { test.Fatalf("TestTransactions: TransactionBegin failed: %s", err) }
  _, err = dbHandle.Exec(`INSERT INTO test (id, name) VALUES (1, "committed");`)
  err = dbTransactionCommit()
  if err != nil { test.Fatalf("TestTransactions: TransactionCommit failed: %s", err) }
  row := dbHandle.QueryRow(`SELECT id, name FROM test WHERE name = 'committed';`)
  var id int ; var name string
  err = row.Scan(&id, &name)
  if err != nil { test.Fatalf("TestTransactions: SELECT failed for committed transaction: %s", err) }
  if id != 1 { test.Fatalf("TestTransactions: SELECT returned wrong id: %d", id) }
  if name != "committed" { test.Fatalf("TestTransactions: SELECT returned wrong name: %s", name) }

  // rolled back transaction should not read back changes
  err = dbTransactionBegin()
  if err != nil { test.Fatalf("TestTransactions: TransactionBegin failed: %s", err) }
  _, err = dbHandle.Exec(`INSERT INTO test (id, name) VALUES (2, "rollback");`)
  if err != nil { test.Fatalf("TestTransactions: INSERT failed: %s", err) }
  err = dbTransactionRollback()
  if err != nil { test.Fatalf("TestTransactions: TransactionRollback failed: %s", err) }
  row = dbHandle.QueryRow(`SELECT id, name FROM test WHERE name = 'rollback';`)
  err = row.Scan(&id, &name)
  if err == nil { test.Fatalf("TestTransactions: SELECT returned row for rolled back transaction") }
}

func TestBackups(test *testing.T) {
  testDbPath := "./test.database"
  _ = os.Remove(testDbPath)

  err := LibraryStartup(testDbPath)
  if err != nil { test.Fatalf("TestBackups: Open failed: %s", err) }
  defer os.Remove(testDbPath)
  defer os.Remove(testDbPath + ".bak")

  _, err = dbHandle.Exec(`CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);`)
  if err != nil { test.Fatalf("TestTransactions: CREATE TABLE failed: %s", err) }
  _, err = dbHandle.Exec(`INSERT INTO test (id, name) VALUES (3, "three");`)
  if err != nil { test.Fatalf("TestTransactions: INSERT failed: %s", err) }
  _, err = dbHandle.Exec(`INSERT INTO test (id, name) VALUES (32, "thirty-two");`)
  if err != nil { test.Fatalf("TestTransactions: INSERT failed: %s", err) }
  LibraryShutdown()

  err = dbBackupCreate()
  if err != nil { test.Fatalf("TestBackups: BackupCreate failed: %s", err) }

  err = LibraryStartup(testDbPath)
  if err != nil { test.Fatalf("TestBackups: Open failed: %s", err) }
  row := dbHandle.QueryRow(`SELECT name FROM test WHERE id = 3;`)
  var name string
  err = row.Scan(&name)
  if err != nil { test.Fatalf("TestBackups: SELECT failed: %s", err) }
  if name != "three" { test.Fatalf("TestBackups: SELECT returned wrong name: %s", name) }
  _, err = dbHandle.Exec(`UPDATE test SET name = "three-fifty" WHERE id = 3;`)
  if err != nil { test.Fatalf("TestBackups: UPDATE failed: %s", err) }
  row = dbHandle.QueryRow(`SELECT name FROM test WHERE id = 3;`)
  err = row.Scan(&name)
  if err != nil { test.Fatalf("TestBackups: SELECT failed: %s", err) }
  if name != "three-fifty" { test.Fatalf("TestBackups: SELECT returned wrong name: %s", name) }
  LibraryShutdown()

  err = dbBackupRestore()
  if err != nil { test.Fatalf("TestBackups: BackupRestore failed: %s", err) }
  err = LibraryStartup(testDbPath)
  if err != nil { test.Fatalf("TestBackups: Open failed: %s", err) }
  row = dbHandle.QueryRow(`SELECT name FROM test WHERE id = 3;`)
  err = row.Scan(&name)
  if err != nil { test.Fatalf("TestBackups: SELECT failed: %s", err) }
  if name != "three" { test.Fatalf("TestBackups: SELECT returned wrong name: %s", name) }
  LibraryShutdown()
}
