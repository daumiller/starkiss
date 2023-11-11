package library

import (
  "os"
  "time"
  "sync"
  "strconv"
  "math/rand"
  "testing"
)

func testConcurrency_Reader(test *testing.T, wait_group *sync.WaitGroup) {
  defer wait_group.Done()
  time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
  _, err := InputFileList()
  if err != nil { test.Fatalf("testConcurrency_Reader: InputFileList failed: %s", err) }
}

func testConcurrency_Writer(test *testing.T, wait_group *sync.WaitGroup, inp *InputFile) {
  defer wait_group.Done()
  time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
  err := InputFileDelete(inp)
  if err != nil { test.Fatalf("testConcurrency_Write: InputFileDelete failed: %s", err) }
}

func TestConcurrency(test *testing.T) {
  testDbPath := "./test.database"
  _ = os.Remove(testDbPath)

  err := LibraryStartup(testDbPath)
  if err != nil { test.Fatalf("TestConcurrency: Open failed: %s", err) }
  defer os.Remove(testDbPath)
  defer LibraryShutdown()

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
  if err != nil { test.Fatalf("TestConcurrency: CREATE TABLE failed: %s", err) }

  // create a bunch of input files
  inps := make([]InputFile, 1000)
  for index := 0; index < len(inps); index += 1 {
    inps[index] = InputFile {
      Id:                       "id" + strconv.Itoa(index),
      SourceLocation:           "source_location" + strconv.Itoa(index),
      SourceStreams:            make([]FileStream, 0),
      StreamMap:                make([]int64, 0),
      SourceDuration:           1,
      TimeScanned:              time.Now().Unix(),
      TranscodingCommand:       "ffmpeg -i input -c:v libx264 -preset slow -crf 22 -c:a copy output",
      TranscodingTimeStarted:   time.Now().Unix(),
      TranscodingTimeElapsed:   int64(index + 1),
      TranscodingError:         "error",
    }
    err := InputFileCreate(&inps[index])
    if err != nil { test.Fatalf("TestConcurrency: InputFileCreate failed: %s", err) }
  }

  // concurrently do reads/writes
  var wait_group sync.WaitGroup
  for index := 0; index < len(inps); index += 1 {
    wait_group.Add(2)
    go testConcurrency_Reader(test, &wait_group)
    go testConcurrency_Writer(test, &wait_group, &(inps[index]))
  }
  wait_group.Wait()
}

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
