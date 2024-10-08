//go:build !unit

package driver_test

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/SAP/go-hdb/driver"
)

// WriterLob defines a io.Writer based data type for scanning Lobs.
type WriterLob []byte

// Write implements the io.Writer interface.
func (b *WriterLob) Write(p []byte) (n int, err error) {
	*b = append(*b, p...)
	return len(p), nil
}

// Scan implements the database.sql.Scanner interface.
func (b *WriterLob) Scan(arg any) error { return driver.ScanLobWriter(arg, b) }

// ExampleScanLobWriter demontrates how to read Lob data using a io.Writer based data type.
func ExampleScanLobWriter() {
	// Open Test database.
	db := sql.OpenDB(driver.MT.Connector())
	defer db.Close()

	table := driver.RandomIdentifier("lob_")

	if _, err := db.Exec(fmt.Sprintf("create table %s (n1 nclob, n2 nclob)", table)); err != nil {
		log.Fatalf("create table failed: %s", err)
	}

	tx, err := db.Begin() // Start Transaction to avoid database error: SQL Error 596 - LOB streaming is not permitted in auto-commit mode.
	if err != nil {
		log.Fatal(err)
	}

	// Lob content can be written using a string.
	content := "scan lob writer"
	_, err = tx.Exec(fmt.Sprintf("insert into %s values (?, ?)", table), content, content)
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	// Select.
	stmt, err := db.Prepare(fmt.Sprintf("select * from %s", table))
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Scan into WriterLob and sql.Null[WriterLob].
	var w WriterLob
	var nw sql.Null[WriterLob]
	if err := stmt.QueryRow().Scan(&w, &nw); err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(w))
	fmt.Println(string(nw.V))

	// output: scan lob writer
	// scan lob writer
}
