package format

import (
	"bytes"
	"testing"
)

func TestSectionOrder_MetaBeforeTypesBeforeTables(t *testing.T) {
	input := []byte(`[tables.posts]
comment = "Blog posts"
pk = ["id"]

[tables.posts.columns.id]
type = "id"

[types.status]
kind = "enum"
values = ["active", "inactive"]

[meta]
version = 1
schema = "test"
`)
	got, err := Format(input, nil)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// meta must come before types, types before tables.
	metaIdx := bytes.Index(got, []byte("[meta]"))
	typesIdx := bytes.Index(got, []byte("[types.status]"))
	tablesIdx := bytes.Index(got, []byte("[tables.posts]"))

	if metaIdx < 0 || typesIdx < 0 || tablesIdx < 0 {
		t.Fatalf("missing sections in output:\n%s", got)
	}
	if metaIdx >= typesIdx {
		t.Errorf("[meta] (pos %d) should come before [types.*] (pos %d)", metaIdx, typesIdx)
	}
	if typesIdx >= tablesIdx {
		t.Errorf("[types.*] (pos %d) should come before [tables.*] (pos %d)", typesIdx, tablesIdx)
	}
}

func TestTableAlphabeticalOrder(t *testing.T) {
	input := []byte(`[meta]
version = 1
schema = "test"

[tables.zebra]
comment = "Z table"

[tables.zebra.columns.id]
type = "id"

[tables.alpha]
comment = "A table"

[tables.alpha.columns.id]
type = "id"

[tables.middle]
comment = "M table"

[tables.middle.columns.id]
type = "id"
`)
	config := &Config{
		TableOrder:  "alphabetical",
		ColumnOrder: "pk_fk_alpha",
	}
	got, err := Format(input, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	alphaIdx := bytes.Index(got, []byte("[tables.alpha]"))
	middleIdx := bytes.Index(got, []byte("[tables.middle]"))
	zebraIdx := bytes.Index(got, []byte("[tables.zebra]"))

	if alphaIdx < 0 || middleIdx < 0 || zebraIdx < 0 {
		t.Fatalf("missing table sections in output:\n%s", got)
	}
	if alphaIdx >= middleIdx {
		t.Errorf("alpha (pos %d) should come before middle (pos %d)", alphaIdx, middleIdx)
	}
	if middleIdx >= zebraIdx {
		t.Errorf("middle (pos %d) should come before zebra (pos %d)", middleIdx, zebraIdx)
	}
}

func TestTableDependencyOrder(t *testing.T) {
	input := []byte(`[meta]
version = 1
schema = "test"

[tables.posts]
pk = ["id"]

[tables.posts.columns.id]
type = "id"

[tables.posts.columns.author_id]
type = "ref"

[tables.posts.fks.author_fk]
columns = ["author_id"]
ref_table = "users"
ref_columns = ["id"]
on_delete = "CASCADE"

[tables.users]
pk = ["id"]

[tables.users.columns.id]
type = "id"
`)
	config := &Config{
		TableOrder:  "dependency",
		ColumnOrder: "pk_fk_alpha",
	}
	got, err := Format(input, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	usersIdx := bytes.Index(got, []byte("[tables.users]"))
	postsIdx := bytes.Index(got, []byte("[tables.posts]"))

	if usersIdx < 0 || postsIdx < 0 {
		t.Fatalf("missing table sections in output:\n%s", got)
	}
	// users must come before posts (posts depends on users via FK)
	if usersIdx >= postsIdx {
		t.Errorf("users (pos %d) should come before posts (pos %d) in dependency order", usersIdx, postsIdx)
	}
}

func TestColumnPKFKAlphaOrder(t *testing.T) {
	input := []byte(`[meta]
version = 1
schema = "test"

[tables.posts]
pk = ["id"]

[tables.posts.columns.title]
type = "text"

[tables.posts.columns.author_id]
type = "ref"

[tables.posts.columns.id]
type = "id"

[tables.posts.columns.body]
type = "text"

[tables.posts.fks.author_fk]
columns = ["author_id"]
ref_table = "users"
ref_columns = ["id"]
on_delete = "CASCADE"
`)
	config := &Config{
		TableOrder:  "dependency",
		ColumnOrder: "pk_fk_alpha",
	}
	got, err := Format(input, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Expected order: id (PK), author_id (FK), body (alpha), title (alpha)
	idIdx := bytes.Index(got, []byte("[tables.posts.columns.id]"))
	authorIdx := bytes.Index(got, []byte("[tables.posts.columns.author_id]"))
	bodyIdx := bytes.Index(got, []byte("[tables.posts.columns.body]"))
	titleIdx := bytes.Index(got, []byte("[tables.posts.columns.title]"))

	if idIdx < 0 || authorIdx < 0 || bodyIdx < 0 || titleIdx < 0 {
		t.Fatalf("missing column sections in output:\n%s", got)
	}

	if idIdx >= authorIdx {
		t.Errorf("id (PK, pos %d) should come before author_id (FK, pos %d)", idIdx, authorIdx)
	}
	if authorIdx >= bodyIdx {
		t.Errorf("author_id (FK, pos %d) should come before body (alpha, pos %d)", authorIdx, bodyIdx)
	}
	if bodyIdx >= titleIdx {
		t.Errorf("body (pos %d) should come before title (pos %d)", bodyIdx, titleIdx)
	}
}

func TestIdempotence(t *testing.T) {
	// Format an already-canonical schema; result should be identical.
	canonical := []byte(`[meta]
version = 1
schema = "test"
extensions = ["pgcrypto"]

[types.status]
kind = "enum"
values = ["active", "inactive"]

[tables.users]
pk = ["id"]

[tables.users.columns.id]
type = "id"

[tables.users.columns.email]
type = "email"

[tables.users.columns.name]
type = "text"

[tables.posts]
pk = ["id"]

[tables.posts.columns.id]
type = "id"

[tables.posts.columns.author_id]
type = "ref"

[tables.posts.columns.title]
type = "text"

[tables.posts.fks.author_fk]
columns = ["author_id"]
ref_table = "users"
ref_columns = ["id"]
on_delete = "CASCADE"
`)
	config := DefaultConfig()
	first, err := Format(canonical, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	second, err := Format(first, config)
	if err != nil {
		t.Fatalf("Format (2nd pass) error: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Errorf("formatting is not idempotent.\nFirst pass:\n%s\nSecond pass:\n%s", first, second)
	}
}

func TestCheckMode(t *testing.T) {
	// Simulate what the CLI --check flag does: format and compare to input.
	unformatted := []byte(`[tables.posts]
pk = ["id"]

[tables.posts.columns.id]
type = "id"

[meta]
version = 1
schema = "test"
`)
	config := DefaultConfig()
	formatted, err := Format(unformatted, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Unformatted input should differ from formatted output.
	if bytes.Equal(unformatted, formatted) {
		t.Error("expected unformatted input to differ from formatted output")
	}

	// Format the formatted output -- should be identical (check would pass).
	reformatted, err := Format(formatted, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if !bytes.Equal(formatted, reformatted) {
		t.Errorf("formatted output is not stable for --check:\nFormatted:\n%s\nReformatted:\n%s", formatted, reformatted)
	}
}

func TestTypesAlphabeticalOrder(t *testing.T) {
	input := []byte(`[meta]
version = 1
schema = "test"

[types.status]
kind = "enum"
values = ["active", "inactive"]

[types.email]
kind = "domain"
base_type = "text"

[types.amount]
kind = "domain"
base_type = "numeric"
`)
	got, err := Format(input, nil)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	amountIdx := bytes.Index(got, []byte("[types.amount]"))
	emailIdx := bytes.Index(got, []byte("[types.email]"))
	statusIdx := bytes.Index(got, []byte("[types.status]"))

	if amountIdx < 0 || emailIdx < 0 || statusIdx < 0 {
		t.Fatalf("missing type sections in output:\n%s", got)
	}
	if amountIdx >= emailIdx {
		t.Errorf("amount (pos %d) should come before email (pos %d)", amountIdx, emailIdx)
	}
	if emailIdx >= statusIdx {
		t.Errorf("email (pos %d) should come before status (pos %d)", emailIdx, statusIdx)
	}
}

func TestFKsAlphabeticalOrder(t *testing.T) {
	input := []byte(`[meta]
version = 1
schema = "test"

[tables.posts]
pk = ["id"]

[tables.posts.columns.id]
type = "id"

[tables.posts.columns.author_id]
type = "ref"

[tables.posts.columns.category_id]
type = "ref"

[tables.posts.fks.category_fk]
columns = ["category_id"]
ref_table = "categories"
ref_columns = ["id"]
on_delete = "SET NULL"

[tables.posts.fks.author_fk]
columns = ["author_id"]
ref_table = "users"
ref_columns = ["id"]
on_delete = "CASCADE"
`)
	got, err := Format(input, nil)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	authorFKIdx := bytes.Index(got, []byte("[tables.posts.fks.author_fk]"))
	categoryFKIdx := bytes.Index(got, []byte("[tables.posts.fks.category_fk]"))

	if authorFKIdx < 0 || categoryFKIdx < 0 {
		t.Fatalf("missing FK sections in output:\n%s", got)
	}
	if authorFKIdx >= categoryFKIdx {
		t.Errorf("author_fk (pos %d) should come before category_fk (pos %d)", authorFKIdx, categoryFKIdx)
	}
}

func TestWithinTableSectionOrder(t *testing.T) {
	// Verify that within a table, subsections follow the canonical order:
	// comment, pk, columns, fks, indexes, unique, checks
	input := []byte(`[meta]
version = 1
schema = "test"

[tables.posts]
pk = ["id"]
comment = "Blog posts"

[tables.posts.fks.author_fk]
columns = ["author_id"]
ref_table = "users"
ref_columns = ["id"]
on_delete = "CASCADE"

[tables.posts.indexes.idx_title]
columns = ["title"]

[tables.posts.columns.id]
type = "id"

[tables.posts.columns.author_id]
type = "ref"

[tables.posts.columns.title]
type = "text"

[tables.posts.checks.title_len]
expr = "length(title) > 0"

[tables.posts.unique.unique_title]
columns = ["title"]
`)
	got, err := Format(input, nil)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	commentIdx := bytes.Index(got, []byte(`comment = "Blog posts"`))
	pkIdx := bytes.Index(got, []byte(`pk = ["id"]`))
	columnsIdx := bytes.Index(got, []byte("[tables.posts.columns."))
	fksIdx := bytes.Index(got, []byte("[tables.posts.fks."))
	indexesIdx := bytes.Index(got, []byte("[tables.posts.indexes."))
	uniqueIdx := bytes.Index(got, []byte("[tables.posts.unique."))
	checksIdx := bytes.Index(got, []byte("[tables.posts.checks."))

	if commentIdx < 0 || pkIdx < 0 || columnsIdx < 0 || fksIdx < 0 ||
		indexesIdx < 0 || uniqueIdx < 0 || checksIdx < 0 {
		t.Fatalf("missing subsections in output:\n%s", got)
	}

	checks := []struct {
		name string
		pos  int
	}{
		{"comment", commentIdx},
		{"pk", pkIdx},
		{"columns", columnsIdx},
		{"fks", fksIdx},
		{"indexes", indexesIdx},
		{"unique", uniqueIdx},
		{"checks", checksIdx},
	}
	for i := 0; i < len(checks)-1; i++ {
		if checks[i].pos >= checks[i+1].pos {
			t.Errorf("%s (pos %d) should come before %s (pos %d)",
				checks[i].name, checks[i].pos, checks[i+1].name, checks[i+1].pos)
		}
	}
}

func TestColumnAlphabeticalOrder(t *testing.T) {
	input := []byte(`[meta]
version = 1
schema = "test"

[tables.users]

[tables.users.columns.name]
type = "text"

[tables.users.columns.age]
type = "integer"

[tables.users.columns.email]
type = "email"
`)
	config := &Config{
		TableOrder:  "alphabetical",
		ColumnOrder: "alphabetical",
	}
	got, err := Format(input, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	ageIdx := bytes.Index(got, []byte("[tables.users.columns.age]"))
	emailIdx := bytes.Index(got, []byte("[tables.users.columns.email]"))
	nameIdx := bytes.Index(got, []byte("[tables.users.columns.name]"))

	if ageIdx < 0 || emailIdx < 0 || nameIdx < 0 {
		t.Fatalf("missing column sections in output:\n%s", got)
	}
	if ageIdx >= emailIdx {
		t.Errorf("age (pos %d) should come before email (pos %d)", ageIdx, emailIdx)
	}
	if emailIdx >= nameIdx {
		t.Errorf("email (pos %d) should come before name (pos %d)", emailIdx, nameIdx)
	}
}

func TestColumnPreserveOrder(t *testing.T) {
	input := []byte(`[meta]
version = 1
schema = "test"

[tables.users]

[tables.users.columns.name]
type = "text"

[tables.users.columns.age]
type = "integer"

[tables.users.columns.email]
type = "email"
`)
	config := &Config{
		TableOrder:  "alphabetical",
		ColumnOrder: "preserve",
	}
	got, err := Format(input, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	nameIdx := bytes.Index(got, []byte("[tables.users.columns.name]"))
	ageIdx := bytes.Index(got, []byte("[tables.users.columns.age]"))
	emailIdx := bytes.Index(got, []byte("[tables.users.columns.email]"))

	if nameIdx < 0 || ageIdx < 0 || emailIdx < 0 {
		t.Fatalf("missing column sections in output:\n%s", got)
	}
	// Preserve order: name, age, email (as in input)
	if nameIdx >= ageIdx {
		t.Errorf("name (pos %d) should come before age (pos %d) in preserve mode", nameIdx, ageIdx)
	}
	if ageIdx >= emailIdx {
		t.Errorf("age (pos %d) should come before email (pos %d) in preserve mode", ageIdx, emailIdx)
	}
}

func TestColumnFKLastOrder(t *testing.T) {
	input := []byte(`[meta]
version = 1
schema = "test"

[tables.posts]
pk = ["id"]

[tables.posts.columns.author_id]
type = "ref"

[tables.posts.columns.title]
type = "text"

[tables.posts.columns.id]
type = "id"

[tables.posts.columns.body]
type = "text"

[tables.posts.fks.author_fk]
columns = ["author_id"]
ref_table = "users"
ref_columns = ["id"]
on_delete = "CASCADE"
`)
	config := &Config{
		TableOrder:  "dependency",
		ColumnOrder: "fk_last",
	}
	got, err := Format(input, config)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Expected order: id (PK), body (alpha, non-FK), title (alpha, non-FK), author_id (FK last)
	idIdx := bytes.Index(got, []byte("[tables.posts.columns.id]"))
	bodyIdx := bytes.Index(got, []byte("[tables.posts.columns.body]"))
	titleIdx := bytes.Index(got, []byte("[tables.posts.columns.title]"))
	authorIdx := bytes.Index(got, []byte("[tables.posts.columns.author_id]"))

	if idIdx < 0 || bodyIdx < 0 || titleIdx < 0 || authorIdx < 0 {
		t.Fatalf("missing column sections in output:\n%s", got)
	}

	if idIdx >= bodyIdx {
		t.Errorf("id (PK, pos %d) should come before body (pos %d)", idIdx, bodyIdx)
	}
	if bodyIdx >= titleIdx {
		t.Errorf("body (pos %d) should come before title (pos %d)", bodyIdx, titleIdx)
	}
	if titleIdx >= authorIdx {
		t.Errorf("title (pos %d) should come before author_id (FK, pos %d)", titleIdx, authorIdx)
	}
}
