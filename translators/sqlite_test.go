package translators_test

import (
	"fmt"

	"github.com/gobuffalo/fizz"
	"github.com/gobuffalo/fizz/translators"
)

var IntIDCol = fizz.Column{
	Name:    "id",
	Primary: true,
	ColType: "integer",
	Options: fizz.Options{},
}

var _ fizz.Translator = (*translators.SQLite)(nil)
var schema = &fauxSchema{schema: map[string]*fizz.Table{}}
var sqt = &translators.SQLite{Schema: schema}

type fauxSchema struct {
	schema map[string]*fizz.Table
}

func (s *fauxSchema) Build() error {
	return nil
}

func (s *fauxSchema) IndexInfo(table string, column string) (*fizz.Index, error) {
	return nil, fmt.Errorf("IndexInfo is not implemented for this translator!")
}

func (s *fauxSchema) ReplaceSchema(schema map[string]*fizz.Table) {
	s.schema = schema
}

func (s *fauxSchema) DeleteColumn(table string, column string) {
	return
}

func (s *fauxSchema) ReplaceColumn(table string, column string, newColumn fizz.Column) error {
	return fmt.Errorf("ReplaceColumn is not implemented for this translator!")
}

func (s *fauxSchema) ColumnInfo(table string, column string) (*fizz.Column, error) {
	return nil, fmt.Errorf("ColumnInfo is not implemented for this translator!")
}

func (p *fauxSchema) Delete(table string) {
	delete(p.schema, table)
}

func (s *fauxSchema) SetTable(table *fizz.Table) {
	s.schema[table.Name] = table
}

func (p *fauxSchema) TableInfo(table string) (*fizz.Table, error) {
	if ti, ok := p.schema[table]; ok {
		return ti, nil
	}
	return nil, fmt.Errorf("Could not find table data for %s!", table)
}

func (p *SQLiteSuite) Test_SQLite_CreateTable() {
	r := p.Require()
	ddl := `CREATE TABLE "users" (
"id" INTEGER PRIMARY KEY AUTOINCREMENT,
"first_name" TEXT NOT NULL,
"last_name" TEXT NOT NULL,
"email" TEXT NOT NULL,
"permissions" TEXT,
"age" INTEGER DEFAULT '40',
"raw" BLOB NOT NULL,
"into" INTEGER NOT NULL,
"flotante" REAL NOT NULL,
"json" TEXT NOT NULL,
"bytes" BLOB NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);`

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("id", "integer", {"primary": true})
		t.Column("first_name", "string", {})
		t.Column("last_name", "string", {})
		t.Column("email", "string", {"size":20})
		t.Column("permissions", "text", {"null": true})
		t.Column("age", "integer", {"null": true, "default": 40})
		t.Column("raw", "blob", {})
		t.Column("into", "int", {})
		t.Column("flotante", "float", {})
		t.Column("json", "json", {})
		t.Column("bytes", "[]byte", {})
	}
	`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_CreateTable_UUID() {
	r := p.Require()
	ddl := `CREATE TABLE "users" (
"first_name" TEXT NOT NULL,
"last_name" TEXT NOT NULL,
"email" TEXT NOT NULL,
"permissions" TEXT,
"age" INTEGER DEFAULT '40',
"company_id" char(36) NOT NULL DEFAULT lower(hex(randomblob(16))),
"uuid" TEXT PRIMARY KEY,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);`

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("first_name", "string", {})
		t.Column("last_name", "string", {})
		t.Column("email", "string", {"size":20})
		t.Column("permissions", "text", {"null": true})
		t.Column("age", "integer", {"null": true, "default": 40})
		t.Column("company_id", "uuid", {"default_raw": "lower(hex(randomblob(16)))"})
		t.Column("uuid", "uuid", {"primary": true})
	}
	`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_CreateTables_WithCompositePrimaryKey() {
	r := p.Require()
	ddl := `CREATE TABLE "user_profiles" (
"user_id" INTEGER NOT NULL,
"profile_id" INTEGER NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
PRIMARY KEY("user_id", "profile_id")
);`

	res, _ := fizz.AString(`
	create_table("user_profiles") {
		t.Column("user_id", "INT")
		t.Column("profile_id", "INT")
		t.PrimaryKey("user_id", "profile_id")
	}
	`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_DropTable() {
	r := p.Require()

	ddl := `DROP TABLE "users";`

	res, _ := fizz.AString(`drop_table("users")`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_RenameTable() {
	r := p.Require()

	ddl := `ALTER TABLE "users" RENAME TO "people";`
	schema.schema["users"] = &fizz.Table{}

	res, _ := fizz.AString(`rename_table("users", "people")`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_RenameTable_NotEnoughValues() {
	r := p.Require()

	_, err := sqt.RenameTable([]fizz.Table{})
	r.Error(err)
}

func (p *SQLiteSuite) Test_SQLite_ChangeColumn() {
	r := p.Require()

	ddl := `ALTER TABLE "users" RENAME TO "_users_tmp";
CREATE TABLE "users" (
"id" INTEGER PRIMARY KEY AUTOINCREMENT,
"created_at" TEXT NOT NULL DEFAULT 'foo',
"updated_at" DATETIME NOT NULL
);
INSERT INTO "users" (id, created_at, updated_at) SELECT id, created_at, updated_at FROM "_users_tmp";
DROP TABLE "_users_tmp";`

	schema.schema["users"] = &fizz.Table{
		Name: "users",
		Columns: []fizz.Column{
			IntIDCol,
			fizz.CREATED_COL,
			fizz.UPDATED_COL,
		},
	}

	res, _ := fizz.AString(`change_column("users", "created_at", "string", {"default": "foo", "size": 50})`, sqt)

	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_AddColumn() {
	r := p.Require()

	ddl := `ALTER TABLE "users" ADD COLUMN "mycolumn" TEXT NOT NULL DEFAULT 'foo';`
	schema.schema["users"] = &fizz.Table{}

	res, _ := fizz.AString(`add_column("users", "mycolumn", "string", {"default": "foo", "size": 50})`, sqt)

	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_DropColumn() {
	r := p.Require()
	ddl := `ALTER TABLE "users" RENAME TO "_users_tmp";
CREATE TABLE "users" (
"id" INTEGER PRIMARY KEY AUTOINCREMENT,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "users" (id, updated_at) SELECT id, updated_at FROM "_users_tmp";
DROP TABLE "_users_tmp";`

	schema.schema["users"] = &fizz.Table{
		Name: "users",
		Columns: []fizz.Column{
			IntIDCol,
			fizz.CREATED_COL,
			fizz.UPDATED_COL,
		},
	}
	res, _ := fizz.AString(`drop_column("users", "created_at")`, sqt)

	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_RenameColumn() {
	r := p.Require()
	ddl := `ALTER TABLE "users" RENAME TO "_users_tmp";
CREATE TABLE "users" (
"id" INTEGER PRIMARY KEY AUTOINCREMENT,
"created_when" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
INSERT INTO "users" (id, created_when, updated_at) SELECT id, created_at, updated_at FROM "_users_tmp";
DROP TABLE "_users_tmp";`

	schema.schema["users"] = &fizz.Table{
		Name: "users",
		Columns: []fizz.Column{
			IntIDCol,
			fizz.CREATED_COL,
			fizz.UPDATED_COL,
		},
	}
	res, _ := fizz.AString(`rename_column("users", "created_at", "created_when")`, sqt)

	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_AddIndex() {
	r := p.Require()

	schema.schema["table_name"] = &fizz.Table{
		Name: "table_name",
		Columns: []fizz.Column{
			{
				Name: "column_name",
			},
		},
	}

	ddl := `CREATE INDEX "table_name_column_name_idx" ON "table_name" (column_name);`

	res, _ := fizz.AString(`add_index("table_name", "column_name", {})`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_AddIndex_Unique() {
	r := p.Require()
	ddl := `CREATE UNIQUE INDEX "table_name_column_name_idx" ON "table_name" (column_name);`

	res, _ := fizz.AString(`add_index("table_name", "column_name", {"unique": true})`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_AddIndex_MultiColumn() {
	r := p.Require()
	ddl := `CREATE INDEX "table_name_col1_col2_col3_idx" ON "table_name" (col1, col2, col3);`

	res, _ := fizz.AString(`add_index("table_name", ["col1", "col2", "col3"], {})`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_AddIndex_CustomName() {
	r := p.Require()
	ddl := `CREATE INDEX "custom_name" ON "table_name" (column_name);`

	res, _ := fizz.AString(`add_index("table_name", "column_name", {"name": "custom_name"})`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_DropIndex() {
	r := p.Require()

	schema.schema["my_table"] = &fizz.Table{
		Name: "my_table",
		Indexes: []fizz.Index{
			{
				Name: "my_idx",
			},
		},
	}

	ddl := `DROP INDEX IF EXISTS "my_idx";`

	res, _ := fizz.AString(`drop_index("my_table", "my_idx")`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_RenameIndex() {
	r := p.Require()

	ddl := `DROP INDEX IF EXISTS "old_ix";
CREATE UNIQUE INDEX "new_ix" ON "users" (id, created_at);`

	schema.schema["users"] = &fizz.Table{
		Name: "users",
		Columns: []fizz.Column{
			IntIDCol,
			fizz.CREATED_COL,
			fizz.UPDATED_COL,
		},
		Indexes: []fizz.Index{
			{
				Name:    "old_ix",
				Columns: []string{"id", "created_at"},
				Unique:  true,
			},
		},
	}

	res, _ := fizz.AString(`rename_index("users", "old_ix", "new_ix")`, sqt)
	r.Equal(ddl, res)
}

func (p *SQLiteSuite) Test_SQLite_DropColumnWithForeignKey() {
	r := p.Require()

	res, _ := fizz.AString(`
	create_table("users") {
		t.Column("uuid", "uuid", {"primary": true})
	}

	create_table("user_notes") {
		t.Column("uuid", "uuid", {"primary": true})
		t.Column("user_id", "uuid")
		t.Column("notes", "string")
    	t.ForeignKey("user_id", {"users": ["uuid"]}, {"on_delete": "cascade"})
	}`, sqt)
	r.Equal(`CREATE TABLE "users" (
"uuid" TEXT PRIMARY KEY,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL
);
CREATE TABLE "user_notes" (
"uuid" TEXT PRIMARY KEY,
"user_id" char(36) NOT NULL,
"notes" TEXT NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
FOREIGN KEY (user_id) REFERENCES users (uuid) ON DELETE cascade
);`, res)

	res, _ = fizz.AString(`drop_column("user_notes","notes")`, sqt)
	r.Equal(`ALTER TABLE "user_notes" RENAME TO "_user_notes_tmp";
CREATE TABLE "user_notes" (
"uuid" TEXT PRIMARY KEY,
"user_id" char(36) NOT NULL,
"created_at" DATETIME NOT NULL,
"updated_at" DATETIME NOT NULL,
FOREIGN KEY (user_id) REFERENCES users (uuid) ON DELETE cascade
);
INSERT INTO "user_notes" (uuid, user_id, created_at, updated_at) SELECT uuid, user_id, created_at, updated_at FROM "_user_notes_tmp";
DROP TABLE "_user_notes_tmp";`, res)

	res, _ = fizz.AString(`rename_table("users","user_accounts")`, sqt)
	r.Equal(`ALTER TABLE "users" RENAME TO "user_accounts";`, res)

	res, _ = fizz.AString(`add_column("user_notes","notes","string")`, sqt)
	r.Equal(`ALTER TABLE "user_notes" ADD COLUMN "notes" TEXT NOT NULL;`, res)
}
