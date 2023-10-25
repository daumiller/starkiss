package database

import (
  "os"
  "testing"
)

// Read, Upser, Delete, List

func TestBasicProperties(test *testing.T) {
  Location = "./test.database"
  _ = os.Remove(Location)
  defer os.Remove(Location)

  db, err := Open()
  if err != nil { test.Fatalf("TestBasicProperties: Open failed: %s", err) }
  defer Close(db)

  // create table
  _, err = db.Exec(`CREATE TABLE properties (
    id    INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
    key   TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL
  );`)
  if err != nil { test.Fatalf("TestBasicProperties: CREATE TABLE failed: %s", err) }
  _, err = db.Exec(`CREATE UNIQUE INDEX property_key ON properties (key);`)
  if err != nil { test.Fatalf("TestBasicProperties: CREATE INDEX failed: %s", err) }

  // PropertyRead fails
  _, err = PropertyRead(db, "test-key")
  if err == nil { test.Error("TestBasicProperties: Read succeeded when value doesn't exist") }
  if err != ErrNotFound { test.Errorf("TestBasicProperties: Read returned wrong error: %s", err) }

  // PropertyUpsert & PropertyRead succeed
  err = PropertyUpsert(db, "test-key", "test-value")
  if err != nil { test.Errorf("TestBasicProperties: Upsert failed: %s", err) }
  value, err := PropertyRead(db, "test-key")
  if err != nil { test.Errorf("TestBasicProperties: Read failed: %s", err) }
  if value != "test-value" { test.Errorf("TestBasicProperties: Read returned wrong value: %s", value) }

  // PropetyDelete succeeds
  err = PropertyDelete(db, "test-key")
  if err != nil { test.Errorf("TestBasicProperties: Delete failed: %s", err) }
  value, err = PropertyRead(db, "test-key")
  if err == nil { test.Error("TestBasicProperties: Read succeeded when value doesn't exist") }
  if err != ErrNotFound { test.Errorf("TestBasicProperties: Read returned wrong error: %s", err) }

  // PropertyList empty
  properties, err := PropertyList(db)
  if err != nil { test.Errorf("TestBasicProperties: List failed: %s", err) }
  if len(properties) != 0 { test.Errorf("TestBasicProperties: List returned wrong number of properties: %d", len(properties)) }

  // PropertyList single key
  err = PropertyUpsert(db, "test-key", "test-value")
  if err != nil { test.Errorf("TestBasicProperties: Upsert failed: %s", err) }
  properties, err = PropertyList(db)
  if err != nil { test.Errorf("TestBasicProperties: List failed: %s", err) }
  if len(properties) != 1 { test.Errorf("TestBasicProperties: List returned wrong number of properties: %d", len(properties)) }
  if properties["test-key"] != "test-value" { test.Errorf("TestBasicProperties: List returned wrong value: %s", properties["test-key"]) }

  // PropetyList multiple keys
  err = PropertyUpsert(db, "second-key", "second-value")
  if err != nil { test.Errorf("TestBasicProperties: Upsert failed: %s", err) }
  properties, err = PropertyList(db)
  if err != nil { test.Errorf("TestBasicProperties: List failed: %s", err) }
  if len(properties) != 2 { test.Errorf("TestBasicProperties: List returned wrong number of properties: %d", len(properties)) }
  if properties["test-key"] != "test-value" { test.Errorf("TestBasicProperties: List returned wrong value: %s", properties["test-key"]) }
  if properties["second-key"] != "second-value" { test.Errorf("TestBasicProperties: List returned wrong value: %s", properties["second-key"]) }
}
