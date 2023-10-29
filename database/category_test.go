package database

import (
  "os"
  "testing"
  "database/sql"
)

func prepTest(test *testing.T) (db *sql.DB, err error) {
  Location = "./test.database"
  _ = os.Remove(Location)

  db, err = Open()
  if err != nil { return nil, err }

  _, err = db.Exec(`CREATE TABLE categories (
    id   TEXT NOT NULL PRIMARY KEY UNIQUE,
    type TEXT NOT NULL,
    name TEXT NOT NULL UNIQUE
  );`)
  if err != nil { return nil, err }
  _, err = db.Exec(`CREATE UNIQUE INDEX categories_name ON categories (name);`) ; if err != nil { return nil, err }
  _, err = db.Exec(`CREATE INDEX categories_type ON categories (type);       `) ; if err != nil { return nil, err }

  _, err = db.Exec(`CREATE TABLE metadata ( id TEXT NOT NULL PRIMARY KEY UNIQUE, category_id TEXT NOT NULL );`) ; if err != nil { return nil, err }

  return db, nil
}

func TestCreateList(test *testing.T) {
  db, err := prepTest(test)
  defer os.Remove(Location)
  defer db.Close()
  if err != nil { test.Fatalf("TestCreateDeleteList: prepTest failed: %s", err) }

  testcat := Category{ Name: "test-category", Type: CategoryType("test-type") }
  err = testcat.Create(db)
  if err != nil { test.Fatalf("TestCreateDeleteList: Create failed: %s", err) }
  testcat_id := testcat.Id
  if testcat_id == "" { test.Fatalf("TestCreateDeleteList: Create returned empty id") }

  categories, err := CategoryList(db)
  if err != nil { test.Fatalf("TestCreateDeleteList: List failed: %s", err) }
  if len(categories) != 1 { test.Fatalf("TestCreateDeleteList: List returned wrong number of categories: %d", len(categories)) }
  if categories[0].Id != testcat_id { test.Fatalf("TestCreateDeleteList: List returned wrong id: %s != %s", categories[0].Id, testcat_id) }
  if categories[0].Name != "test-category" { test.Fatalf("TestCreateDeleteList: List returned wrong name: %s", categories[0].Name) }
  if categories[0].Type != "test-type" { test.Fatalf("TestCreateDeleteList: List returned wrong type: %s", categories[0].Type) }

  err = testcat.Delete(db)
  if err != nil { test.Fatalf("TestCreateDeleteList: Delete failed: %s", err) }

  categories, err = CategoryList(db)
  if err != nil { test.Fatalf("TestCreateDeleteList: List failed: %s", err) }
  if len(categories) != 0 { test.Fatalf("TestCreateDeleteList: List returned wrong number of categories: %d", len(categories)) }
}

func TestRename(test *testing.T) {
  db, err := prepTest(test)
  defer os.Remove(Location)
  defer db.Close()
  if err != nil { test.Fatalf("TestCreateDeleteList: prepTest failed: %s", err) }

  testcat := Category{ Name: "test-category", Type: CategoryType("test-type") }
  err = testcat.Create(db)
  if err != nil { test.Fatalf("TestCreateDeleteList: Create failed: %s", err) }

  err = testcat.Rename(db, "new-name")
  if err != nil { test.Fatalf("TestCreateDeleteList: Rename failed: %s", err) }

  categories, err := CategoryList(db)
  if err != nil { test.Fatalf("TestCreateDeleteList: List failed: %s", err) }
  if len(categories) != 1 { test.Fatalf("TestCreateDeleteList: List returned wrong number of categories: %d", len(categories)) }
  if categories[0].Name != "new-name" { test.Fatalf("TestCreateDeleteList: List returned wrong name: %s", categories[0].Name) }
}

func TestSetType(test *testing.T) {
  db, err := prepTest(test)
  defer os.Remove(Location)
  defer db.Close()
  if err != nil { test.Fatalf("TestSetType: prepTest failed: %s", err) }

  cat_with_metadata    := Category{ Name: "cat-with-metadata",    Type: CategoryType("test-type") }
  cat_with_none        := Category{ Name: "cat-with-none",        Type: CategoryType("test-type") }
  err = cat_with_metadata.Create(db)    ; if err != nil { test.Fatalf("TestSetType: Create failed: %s", err) }
  err = cat_with_none.Create(db)        ; if err != nil { test.Fatalf("TestSetType: Create failed: %s", err) }

  _, err = db.Exec(`INSERT INTO metadata (id, category_id) VALUES ("test-metadata", ?);`, cat_with_metadata.Id)
  if err != nil { test.Fatalf("TestSetType: INSERT INTO metadata failed: %s", err) }

  err = cat_with_metadata.SetType(db, CategoryTypeMusic)
  if err == nil { test.Fatalf("TestSetType: SetType succeeded when metadata exists") }
  if err != ErrCategoryNotEmpty { test.Fatalf("TestSetType: SetType returned wrong error: %s", err) }

  err = cat_with_none.SetType(db, CategoryTypeMusic)
  if err != nil { test.Fatalf("TestSetType: SetType failed: %s", err) }

  _, err = db.Exec(`DELETE FROM metadata WHERE id = "test-metadata";`)
  if err != nil { test.Fatalf("TestSetType: DELETE FROM metadata failed: %s", err) }
  err = cat_with_metadata.SetType(db, CategoryTypeMusic)
  if err != nil { test.Fatalf("TestSetType: SetType failed: %s", err) }
}

func TestDelete(test *testing.T) {
  db, err := prepTest(test)
  defer os.Remove(Location)
  defer db.Close()
  if err != nil { test.Fatalf("TestDelete: prepTest failed: %s", err) }

  cat_with_metadata    := Category{ Name: "cat-with-metadata",    Type: CategoryType("test-type") }
  cat_with_none        := Category{ Name: "cat-with-none",        Type: CategoryType("test-type") }
  err = cat_with_metadata.Create(db)    ; if err != nil { test.Fatalf("TestDelete: Create failed: %s", err) }
  err = cat_with_none.Create(db)        ; if err != nil { test.Fatalf("TestDelete: Create failed: %s", err) }

  _, err = db.Exec(`INSERT INTO metadata (id, category_id) VALUES ("test-metadata", ?);`, cat_with_metadata.Id)
  if err != nil { test.Fatalf("TestDelete: INSERT INTO metadata failed: %s", err) }

  err = cat_with_metadata.Delete(db)
  if err == nil { test.Fatalf("TestDelete: Delete succeeded when metadata exists") }
  if err != ErrCategoryNotEmpty { test.Fatalf("TestDelete: Delete returned wrong error: %s", err) }

  err = cat_with_none.Delete(db)
  if err != nil { test.Fatalf("TestDelete: Delete failed: %s", err) }

  _, err = db.Exec(`DELETE FROM metadata WHERE id = "test-metadata";`)
  if err != nil { test.Fatalf("TestDelete: DELETE FROM metadata failed: %s", err) }
  err = cat_with_metadata.Delete(db)
  if err != nil { test.Fatalf("TestDelete: Delete failed: %s", err) }
}
