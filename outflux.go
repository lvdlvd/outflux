/*
Outflux extracts rows and columns from a influxdb line protocol formatted file

	https://docs.influxdata.com/influxdb/v1/write_protocols/line_protocol_tutorial/

Usage

	outflux [-csv|-tsv] [-m measurement] [tag=[val]]... [tag_or_fld [tag_or_fld]...]  < influx.db > outfile.csv

		tag=val   only process records that contain this tag with this value.
		tag=      only process records that contain this tag, irrespective of value
		tag~pfx   only process records that contain this tag with a value that has this prefix

		tag_or_fld  output this tag or field value as a column.

		only records that have all the specified columns will be output.

		if no tag_or_fld is specified, the program will output a summary over
		all records that match the filter

		TODO the influxdb timestamps are ignored because they're not really meaningful in our datasets.
*/
package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/influxdata/line-protocol/v2/lineprotocol"
)

var (
	fMeasurement = flag.String("m", "", "only output records from this measurement (the prefix of each lineprotocol record)")
	fCSV         = flag.Bool("csv", false, "output comma separated values (default space)")
	fTSV         = flag.Bool("tsv", false, "output tab separated values (default space)")
	fSkip        = flag.Uint("skip", 0, "discard this many records from the begining of the output")
	fLim         = flag.Uint("lim", math.MaxUint, "limit output to this many records ")
)

func newWriter(f io.Writer) *csv.Writer {
	w := csv.NewWriter(f)
	switch {
	case *fTSV:
		w.Comma = '\t'
	case !*fCSV:
		w.Comma = ' '
	}
	return w
}

func Select(tags map[string]string, pfixes map[string]bool, fields []string, r io.Reader) <-chan []string {
	ch := make(chan []string)
	go func() {
		defer close(ch)
		var (
			from     []byte
			fieldidx = map[string]int{}
			dec      = lineprotocol.NewDecoder(r)
			n        = 0
			recno    = 0
		)

		if *fMeasurement != "" {
			from = []byte(*fMeasurement)
		}

		for i, v := range fields {
			fieldidx[v] = i
		}

		for dec.Next() {
			recno++
			m, err := dec.Measurement()
			if err != nil {
				log.Printf("skipping record %d:%v", recno, err)
				continue
			}
			if from != nil && !bytes.Equal(from, m) {
				continue
			}

			values := make([]string, len(fields))

			// since the filter tags are unique in the record and in the filter,
			// we only have to count if they're all satisfied
			sat, flds := 0, 0

			for {
				key, val, err := dec.NextTag()
				if err != nil {
					log.Printf("skipping record %d:tag %v", recno, err)
					continue
				}
				if key == nil {
					break
				}
				k := string(key)
				if v, found := tags[k]; found {
					vv := []byte(v)
					if pfixes[k] && !bytes.HasPrefix(val, vv) {
						continue
					}
					if !pfixes[k] && !bytes.Equal(val, vv) {
						continue
					}
					sat++
				}
				if idx, found := fieldidx[k]; found {
					values[idx] = string(val)
					flds++
				}
			}

			for {
				key, _, val, err := dec.NextFieldBytes()
				if err != nil {
					log.Printf("skipping record %d:tag %v", recno, err)
					continue
				}
				if key == nil {
					break
				}
				k := string(key)
				if idx, found := fieldidx[k]; found {
					values[idx] = string(val)
					flds++
				}
			}

			if sat != len(tags) || flds != len(fields) {
				continue
			}

			ch <- values
			n++

		}
		if err := dec.Err(); err != nil {
			log.Println(err)
		}
		log.Printf("read %d of %d records.", n, recno-1)
	}()
	return ch
}

func Summary(tags map[string]string, pfixes map[string]bool, r io.Reader) <-chan map[string]string {
	ch := make(chan map[string]string)
	go func() {
		defer close(ch)
		var (
			from  []byte
			dec   = lineprotocol.NewDecoder(r)
			n     = 0
			recno = 0
		)

		if *fMeasurement != "" {
			from = []byte(*fMeasurement)
		}

		for dec.Next() {
			recno++
			m, err := dec.Measurement()
			if err != nil {
				log.Printf("skipping record %d:%v", recno, err)
				continue
			}
			if from != nil && !bytes.Equal(from, m) {
				continue
			}

			values := make(map[string]string)
			values[""] = string(m) // treat measurement as the empty tag's value

			// since the filter tags are unique in the record and in the filter,
			// we only have to count if they're all satisfied
			sat := 0

			for {
				key, val, err := dec.NextTag()
				if err != nil {
					log.Printf("skipping record %d:tag %v", recno, err)
					continue
				}
				if key == nil {
					break
				}
				k := string(key)
				if v, found := tags[k]; found {
					vv := []byte(v)
					if pfixes[k] && !bytes.HasPrefix(val, vv) {
						continue
					}
					if !pfixes[k] && !bytes.Equal(val, vv) {
						continue
					}
					sat++
				}
				values[string(key)] = string(val)
			}

			if sat != len(tags) {
				continue
			}

			var fields []string

			for {
				key, _, _, err := dec.NextFieldBytes()
				if err != nil {
					log.Printf("skipping record %d:tag %v", recno, err)
					continue
				}
				if key == nil {
					break
				}
				fields = append(fields, string(key))
			}
			sort.Strings(fields)
			values[" _fields"] = strings.Join(fields, ",")

			ch <- values
			n++
		}
		if err := dec.Err(); err != nil {
			log.Println(err)
		}
		log.Printf("read %d of %d records.", n, recno-1)
	}()
	return ch
}

func main() {

	log.SetPrefix("outflux:")
	log.SetFlags(log.Lmsgprefix)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] [-m measurement] [tag=[val]]... [tag_or_fld [tag_or_fld]...]  < influxdb.log\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	var (
		start  = time.Now()
		args   = flag.Args()
		filter = map[string]string{}
		pfxes  = map[string]bool{}
	)

	for ; len(args) > 0; args = args[1:] {
		if strings.ContainsRune(args[0], '=') {
			parts := strings.SplitN(args[0], "=", 2)
			if _, found := filter[parts[0]]; found {
				log.Fatalf("Duplicate tag %s, have %q", parts[0], filter)
			}
			filter[parts[0]] = parts[1] // we did contain a =,
			if parts[1] == "" {
				pfxes[parts[0]] = true
			}
			continue
		}
		if strings.ContainsRune(args[0], '~') {
			parts := strings.SplitN(args[0], "~", 2)
			if len(parts) != 2 {
				log.Fatalf("Prefix filter can not be the empty string %q", args[0])
			}
			if _, found := filter[parts[0]]; found {
				log.Fatalf("Duplicate tag %s, have %q", parts[0], filter)
			}
			filter[parts[0]] = parts[1]
			pfxes[parts[0]] = true
			continue
		}
		// no = or ~: we are in the fields list
		break
	}

	// remaining args are field names

	N := 0

	if len(args) == 0 {
		// no fields listed, produce summary
		tagvals := map[string]map[string]int{}

		for tvs := range Summary(filter, pfxes, os.Stdin) {
			for tag, value := range tvs {
				N++
				if m, found := tagvals[tag]; found {
					_, found := m[value]
					if found || len(m) < 20 {
						m[value]++
					}
				} else {
					tagvals[tag] = map[string]int{value: 1}
				}
			}
		}

		var tags []string
		for k, _ := range tagvals {
			tags = append(tags, k)
		}
		sort.Strings(tags) // empty will be first
		for _, tag := range tags {
			switch tag {
			case "":
				fmt.Printf("MEASUREMENTS:\n")
			case " _fields":
				fmt.Printf("RECORD TYPES:\n")
			default:
				fmt.Printf("%s:\n", tag)
			}
			var values []string
			for k, _ := range tagvals[tag] {
				values = append(values, k)
			}
			sort.Strings(values)
			long := len(values) > 19
			if long {
				values = values[:19]
			}
			for _, k := range values {
				v := tagvals[tag][k]
				fmt.Printf("\t%7d %s\n", v, k)
			}
			if long {
				fmt.Printf("\t...\n")
			}
		}

	} else {

		w := newWriter(os.Stdout)
		os.Stdout.WriteString("#")
		w.Write(args)
		for values := range Select(filter, pfxes, args, os.Stdin) {
			if *fSkip > 0 {
				*fSkip--
				continue
			}

			w.Write(values)
			N++
			if *fLim--; *fLim == 0 {
				break
			}

		}
		w.Flush()
		if err := w.Error(); err != nil {
			log.Fatal(err)
		}

	}

	log.Printf("wrote %d records in %v", N, time.Now().Sub(start))

}
