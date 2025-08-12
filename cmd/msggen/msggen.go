// This program generates messages.go from messages.xml
// TODO generate string methods for all bitfield types

package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/symmatree/gnss/ubx"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("msggen: ")
	flag.Parse()

	if len(flag.Args()) != 3 {
		log.Fatalf("Usage: %s code.tmpl messages.xml code.go", os.Args[0])
	}

	tmpl, err := template.New(filepath.Base(flag.Arg(0))).Funcs(tmplfuncs).ParseFiles(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if flag.Arg(1) != "-" {
		os.Stdin, err = os.Open(flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}
	}

	var definitions ubx.Definitions

	if err := xml.NewDecoder(os.Stdin).Decode(&definitions); err != nil {
		log.Fatal(err)
	}

	sort.Stable(ubx.ByNameAndLength(definitions.MessageDef))

	for _, v := range definitions.MessageDef {
		for _, b := range v.Blocks {
			b.Link(v)
		}
	}

	// name repeated and annotate 'count' fields
	for _, msg := range definitions.MessageDef {
		for _, b := range msg.Blocks {
			if b.Cardinality == "repeated" {
				b.Name = "Items"
				if strings.HasPrefix(strings.ToLower(b.LenField), "num") {
					b.Name = strings.Title(b.LenField[3:])
				}

				// link back
				for _, bb := range msg.Blocks {
					if bb.Name == b.LenField {
						bb.LenFor = b.Name
						break
					}
				}

			}
		}
	}

	// sort by class/id, and length options
	msgs := map[string][]*ubx.MessageDef{}
	for _, v := range definitions.MessageDef {
		n := v.ClassIDName()
		v.Version = len(msgs[n])
		msgs[n] = append(msgs[n], v)
	}

	// figure out if we can figure out the type from just class, id and size
	for k, v := range msgs {
		m := map[int]int{}
		// first count all minimum and minimum+opt sizes
		for _, vv := range v {
			if sz := vv.VarSize(); sz == 0 {
				m[vv.MinSize()]++
				if vv.MaxFixSize() != vv.MinSize() {
					m[vv.MaxFixSize()]++
				}
			}
		}
		// now mark all variable sizes that could alias the existing ones
		for _, vv := range v {
			if sz := vv.VarSize(); sz != 0 {
				for kk, _ := range m {
					if (kk >= vv.MinSize()) && (kk-vv.MinSize())%sz == 0 {
						m[vv.MinSize()]++
					}
				}
			}
		}
		for _, cnt := range m {
			if len(v) > 1 && cnt > 1 {
				log.Println("Ambiguous type", k, m)
				isambiguous[k] = true
				break
			}
		}
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "// Generated Code -- DO NOT EDIT.\n//go:generate go run msggen.go %s %s %s\n\n", flag.Arg(0), flag.Arg(1), flag.Arg(2))
	if err := tmpl.Execute(&buf, msgs); err != nil {
		log.Fatal(err)
	}
	b := buf.Bytes()

	// fix all &#xx; xml entities
	b = regexp.MustCompile(`&#[0-9]{2};`).ReplaceAllFunc(b, func(b []byte) []byte {
		v, _ := strconv.ParseUint(string(b[2:4]), 10, 8)
		return []byte{byte(v)}
	})
	b = regexp.MustCompile(`&gt;`).ReplaceAllLiteral(b, []byte(">"))
	b = regexp.MustCompile(`&lt;`).ReplaceAllLiteral(b, []byte("<"))
	b = regexp.MustCompile(`&amp;`).ReplaceAllLiteral(b, []byte("&"))

	// try to format as valid Go
	bb, err := format.Source(b)
	if err != nil {
		log.Println(err)
		bb = b
	}

	if flag.Arg(2) != "-" {
		os.Stdout, err = os.Create(flag.Arg(2))
		if err != nil {
			log.Fatal(err)
		}
		defer os.Stdout.Close()
	}

	if _, err := os.Stdout.Write(bb); err != nil {
		log.Fatal(err)
	}

}

// Helper functions for in the template
var tmplfuncs = template.FuncMap{
	"lower":       strings.ToLower,
	"upper":       strings.ToUpper,
	"title":       strings.Title,
	"notabs":      notabs,
	"isambiguous": isAmbiguous,
}

var wstospace = strings.NewReplacer("\t", " ", "\n", " ")

func notabs(s string) string { return wstospace.Replace(s) }

var isambiguous = map[string]bool{}

func isAmbiguous(classid string) bool { return isambiguous[classid] }
