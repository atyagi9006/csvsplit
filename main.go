/*
Command csvsplit splits a .csv into multiple, smaller files.

Resulting files will be saved as 1.csv, 2.csv, etc.  in the currect directory, unless the -output flag is used.

Install

Requires Go to be installed first, https://golang.org/doc/install.

	$ go get github.com/JeffPaine/csvsplit

Flags

Basic usage: csvsplit -records <number of records> <file>

	-records Number of records per file
	-output Output filename / path (optional)
	-headers Number of header lines in the input file to add to each ouput file (optional, default=0)

Examples

Split file.csv into files with 300 records a piece.
	$ csvplit -records 300 file.csv

Accept csv data from stdin.
	$ cat file.csv | csvsplit -records 20

Split file.csv into files with 40 records a piece and two header lines (preserved in all files).
	$ csvplit -records 40 -headers 2 file.csv

You can use the -output flag to customize the resulting filenames.
The below will generate custom_filename-001.csv, custom_filename-002.csv, etc..
	$ csvsplit -records 20 -output custom_filename- file.csv

Split file.csv into files with 37 records a piece into the subfolder 'stuff'.
	$ csvplit -records 37 -output stuff/ file.csv
*/
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	flagRecords = flag.Int("records", 0, "The number of records per output file")
	flagOutput  = flag.String("output", "", "Filename / path of the output file (leave blank for current directory)")
	flagHeaders = flag.Int("headers", 0, "Number of header lines in the input file to preserve in each output file")
)

func main() {
	flag.Parse()

	// Sanity check command line flags.
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: csvsplit [options] -records <number of records> <file>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *flagRecords < 1 {
		fmt.Fprintln(os.Stderr, "-records must be > 1")
		flag.Usage()
	}
	if *flagHeaders < 0 {
		fmt.Fprintln(os.Stderr, "-headers must be > 0")
		flag.Usage()
	}
	if *flagHeaders >= *flagRecords {
		fmt.Fprintln(os.Stderr, "-headers must be >= -records")
		flag.Usage()
	}

	// Get input from a given file or stdin
	var r *csv.Reader
	if len(flag.Args()) == 1 {
		f, err := os.Open(flag.Args()[0])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		r = csv.NewReader(f)
	} else {
		r = csv.NewReader(os.Stdin)
	}

	// Read the input .csv file line by line. Save to a new file after reaching
	// the amount of records prescribed by the -records flag.
	var recs [][]string
	count := 1
	for {
		record, err := r.Read()
		if err == io.EOF {
			save(&recs, count)
			break
		} else if err != nil {
			log.Fatal(err)
		}

		recs = append(recs, record)
		if len(recs) == *flagRecords {
			save(&recs, count)
			// Reset records to include just the header lines (if any)
			recs = recs[:*flagHeaders]
			count++
		}
	}
}

// save() saves the given *[][]string of csv data to a .csv file. Files are named
// sequentially in the form of 1.csv, 2.csv, etc.
func save(recs *[][]string, c int) {
	name := fmt.Sprintf("%v%d%v", *flagOutput, c, ".csv")

	// Make sure we don't overwrite existing files
	if _, err := os.Stat(name); err == nil {
		log.Fatal("file exists: ", name)
	}

	// If a directory is specified, make sure that directory exists
	if filepath.Dir(*flagOutput) != "." {
		_, err := os.Stat(filepath.Dir(*flagOutput))
		if err != nil {
			log.Fatal("no such directory:", *flagOutput)
		}
	}

	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.WriteAll(*recs)
}
