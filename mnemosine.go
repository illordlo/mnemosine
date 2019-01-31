package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Check for errors
func check(e error) {
	if e != nil {
		fmt.Printf("[+] Error: %s\n", e)
		os.Exit(1)
	}
}

func import_records(in_file string, out_file string, skip_file string) {

	var count_inserted int // The counter of the inserted records
	var count_read int     // The counter of the read records
	var src string         // The source of the leaked credential (e.g.: Dropbox)
	var f *os.File         // The input file object
	var sf *os.File        // The file object used to keeping track of the skipped records
	var db *sql.DB         // The db instance
	var err error          // Generic error

	// Open the db
	db, err = sql.Open("sqlite3", out_file)
	check(err)
	defer db.Close()

	// Create the db schema
	t := "create table leak (source, username text, domain text, password text);"
	_, err = db.Exec(t)
	check(err)

	// Prepare the INSERT SQL transaction
	tx, err := db.Begin()
	check(err)

	stmt, err := tx.Prepare("insert into leak(source, username, domain, password) values(?, ?, ?, ?)")
	check(err)
	defer stmt.Close()

	// Create the file that will contain the skipped records
	if skip_file != "" {
		sf, err = os.Create(skip_file)
		check(err)
		defer sf.Close()
	}

	src = strings.TrimSuffix(in_file, filepath.Ext(in_file))

	f, err = os.Open(in_file)
	check(err)
	defer f.Close()

	/*
	* Increasing the scanner buffer to 1GB, just to be sure that the Scan()
	* call never fails due to "Token too long" error.
	 */
	s := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 1024*1024*1000)

	count_read = 0

	for {

		count_read++

		if !s.Scan() {
			if s.Err() != nil {
				fmt.Printf("[+] Error while reading: %s\n", s.Err())
				continue
			}
			break
		}

		t = s.Text()

		tmp1 := strings.Split(t, ":")
		if len(tmp1) != 2 {
			tmp1 = strings.Split(t, ";")
			if len(tmp1) != 2 {
				_, err = sf.WriteString(t + "\n")
				check(err)
				continue
			}
		}

		tmp2 := strings.Split(tmp1[0], "@")
		if len(tmp2) != 2 {
			_, err = sf.WriteString(t + "\n")
			check(err)
			continue
		}

		u := tmp2[0]
		d := tmp2[1]
		p := tmp1[1]

		_, err = stmt.Exec(src, u, d, p)
		check(err)

		count_inserted++

		if count_inserted%1000000 == 0 {
			fmt.Printf("[+] Imported %d records ...\n", count_inserted)
		}
	}
	tx.Commit()

	// Create indexes
	fmt.Println()
	fmt.Println("[+] Creating source index ...")
	t = "create index ids_source on leak(source);"
	_, err = db.Exec(t)
	check(err)

	fmt.Println()
	fmt.Println("[+] Creating usernames index ...")
	t = "create index idx_username on leak(username);"
	_, err = db.Exec(t)
	check(err)

	fmt.Println()
	fmt.Println("[+] Creating domains index ...")
	t = "create index idx_domain on leak(domain);"
	_, err = db.Exec(t)
	check(err)

	// Show import statistics
	fmt.Println()
	fmt.Printf("Read records: %d\n", count_read)
	fmt.Printf("Imported records: %d (%f %%)\n", count_inserted, ((float64(count_inserted) / float64(count_read)) * float64(100)))
}

func main() {

	var in_file string    // The path of the file to import
	var in_dir string     // The folder where the input files reside (alternative to in_file)
	var in_ext string     // The extension of the files that have to be imported
	var in_files []string // The list of the files in folder in_dir
	var out_file string   // The path of the SQLite db file
	var skip_file string  // The path of the file containing skipped records

	flag.StringVar(&in_file, "in-file", "", "The path of the file to import (ALTERNATIVE to \"in-dir\")")
	flag.StringVar(&in_dir, "in-dir", "", "The path of the folder where the files to import reside (ALTERNATIVE to \"in-file\")")
	flag.StringVar(&in_ext, "in-ext", "", "The extension of the file to be imported (MANDATORY with \"in-dir\")")
	flag.StringVar(&out_file, "out-file", "", "The path of the SQLite file (used only for single file as input, otherwise it is equal to the input file + \".db\")")
	flag.StringVar(&skip_file, "skip-file", "", "The path of the file containing skipped records")

	flag.Parse()

	flag.Usage = func() {
		fmt.Printf("Usage: mnemosine [options]\n")
		flag.PrintDefaults()
	}

	if in_dir != "" {
		if in_ext == "" || in_file != "" {
			flag.Usage()
			os.Exit(1)
		}
	} else if in_file != "" {
		if in_dir != "" || in_ext != "" || out_file == "" {
			flag.Usage()
			os.Exit(1)
		}
	}

	in_ext = "." + in_ext

	if in_dir != "" {
		err := filepath.Walk(in_dir, func(path string, info os.FileInfo, err error) error {
			if filepath.Ext(path) == in_ext {
				in_files = append(in_files, path)
			}
			return nil
		})
		check(err)
		for _, in_file := range in_files {
			out_file = strings.TrimSuffix(in_file, filepath.Ext(in_file))
			out_file = out_file + ".db"
			import_records(in_file, out_file, skip_file)
		}
	} else {
		import_records(in_file, out_file, skip_file)
	}

}
