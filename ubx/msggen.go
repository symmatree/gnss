package ubx

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Definitions struct {
	MessageDef []*MessageDef
}

type MessageDef struct {
	Name        string
	Type        string
	Description string
	Comment     string
	Firmware    string
	Class       Number   `xml:"Structure>Class"`
	Id          Number   `xml:"Structure>Id"`
	Length      string   `xml:"Structure>Length"` // of the form A + N * B, but varying syntax
	Blocks      []*Block `xml:"Structure>Payload>Block"`

	Version int // different versions of the same MessageDef name
}

type Block struct {
	Cardinality string `xml:"type,attr"` // repeated or optional, in which case 'nested' is non nil
	LenField    string `xml:"name,attr"` // for repeated fields: name of the count field
	Name        string

	// non-repeated, non-optional fields
	Offset      string
	Type        string
	Comment     string
	Scale       string
	Unit        string
	BitfieldRef string    `xml:"Bitfield>Reference"`
	Bitfield    []*BitDef `xml:"Bitfield>Type"`
	Subtype     string    // in some MessageDefs, the first field has an additional type-switch function.  valid values are 'default' or a number

	Nested []*Block `xml:"Block"` // for repeated or optional blocks, this contains the subfields

	MessageDef *MessageDef `xml:"-"` // link back up
	LenFor     string      `xml:"-"` // for fields that are the CountField of a repeated field, name of the repeated field

}

type BitDef struct {
	Index       string
	Type        string
	Name        string
	Description string

	Block *Block `xml:"-"` // link back up
}

func (b *BitDef) Mask() string {
	parts := strings.Split(b.Index, ":")
	if len(parts) == 2 {
		hi, _ := strconv.ParseUint(parts[1], 0, 8)
		lo, _ := strconv.ParseUint(parts[0], 0, 8)
		if hi <= lo {
			log.Fatalf("hi<=lo in bit mask %q", b.Index)
		}
		return fmt.Sprintf("0x%x", ((1<<(hi+1))-1)^((1<<(lo))-1))
	}
	i, _ := strconv.ParseUint(b.Index, 0, 8)
	return fmt.Sprintf("0x%x", 1<<i)
}

func (b *BitDef) Shift() string { return strings.Split(b.Index, ":")[0] }
func (b *BitDef) OneBit() bool  { return len(strings.Split(b.Index, ":")) == 1 }

type ByNameAndLength []*MessageDef

func (v ByNameAndLength) Len() int      { return len(v) }
func (v ByNameAndLength) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v ByNameAndLength) Less(i, j int) bool {
	if v[i].Name != v[j].Name {
		return v[i].Name < v[j].Name
	}
	return v[i].MinSize() < v[j].MinSize()
}

// set Block.MessageDef and BitDef.Block pointers
func (b *Block) Link(m *MessageDef) {
	b.MessageDef = m
	for _, v := range b.Nested {
		v.Link(m)
	}
	for _, v := range b.Bitfield {
		v.Block = b
	}
}

func (m *MessageDef) ClassIDName() string {
	parts := strings.Split(strings.ToLower(m.Name), "-")
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return strings.Join(parts, "-")
}

func (m *MessageDef) TypeName() string {
	parts := strings.Split(strings.ToLower(m.Name), "-")
	for i, v := range parts {
		parts[i] = strings.Title(v)
	}
	if m.Version > 0 {
		return fmt.Sprintf("%s%d", strings.Join(parts[1:], ""), m.Version)
	}
	return strings.Join(parts[1:], "")
}

var (
	reUnit       = regexp.MustCompile(`^[a-zA-Z^/2]+$`)
	reScaleDec   = regexp.MustCompile(`^1e-\d+$`)
	reScaleLeft  = regexp.MustCompile(`^2\^-\d+$`)
	reScaleRight = regexp.MustCompile(`^2\^\d+$`)
	repl         = strings.NewReplacer("^", "", "/", "_")
)

// Return the fieldname to use for a Go struct.
// if the units are not too wild they are suffixed as lowercase, with '/' converted to _
// if the scaling is 1e-d, or 2^d, or 2^-d, we suffix ed, ld or rd resp. (Daedalean convention on variable naming)
func (b *Block) FieldName() string {
	n := strings.Title(b.Name)
	if u := reUnit.FindString(b.Unit); u != "" {
		n = n + "_" + repl.Replace(strings.ToLower(u))
		if s := reScaleDec.FindString(b.Scale); s != "" {
			n = n + "e" + s[3:]
		} else if s := reScaleLeft.FindString(b.Scale); s != "" {
			n = n + "l" + s[3:]
		} else if s := reScaleRight.FindString(b.Scale); s != "" {
			n = n + "r" + s[2:]
		}
	}
	return n
}

func (b *Block) FieldType() string {
	tp := b.Type
	if i := strings.Index(tp, "["); i >= 0 {
		tp = tp[:i]
	}
	switch tp {
	case "RU1_3":
		return "Float8" // defined in spec but not found in any MessageDef
	case "R4":
		return "float32"
	case "R8":
		return "float64"
	case "I1":
		return "int8"
	case "U1", "CH", "X1":
		return "byte"
	case "I2":
		return "int16"
	case "U2", "X2":
		return "uint16"
	case "I4":
		return "int32"
	case "U4", "X4":
		return "uint32"
	case "I8":
		return "int64"
	case "U8":
		return "uint64"
	}
	return tp // probably invalid
}

func (b *Block) ArraySpec() string {
	if i := strings.Index(b.Type, "["); i >= 0 {
		return b.Type[i:]
	}
	return ""
}

func (b *Block) FieldSize() int {
	parts := strings.Split(b.Type, "[")
	sz := 1
	if len(parts) == 2 {
		v, err := strconv.ParseUint(parts[1][:len(parts[1])-1], 10, 8)
		if err != nil {
			log.Fatalf("invalid array size %#v", b.Type)
		}
		sz = int(v)
	}
	switch parts[0] {
	case "RU1_3", "I1", "U1", "CH", "X1":
		return sz
	case "I2", "U2", "X2":
		return sz * 2
	case "R4", "I4", "U4", "X4":
		return sz * 4
	case "R8", "I8", "U8":
		return sz * 8
	}
	return 0 // probably invalid
}

func (b *Block) IsSubtypeField() bool   { return b.Subtype != "" }
func (b *Block) IsSubtypeDefault() bool { return b.Subtype == "default" }
func (b *Block) SubtypeValue() int      { v, _ := strconv.ParseUint(b.Subtype, 0, 8); return int(v) }

// the non optional non repeated fields at the begining
func (m *MessageDef) MinSize() int {
	sz := 0
	for _, v := range m.Blocks {
		if v.Cardinality != "" {
			break
		}
		sz += v.FieldSize()
	}
	return sz
}

// minsize + the optional bit
func (m *MessageDef) MaxFixSize() int {
	sz := 0
	for _, v := range m.Blocks {
		switch v.Cardinality {
		case "":
			sz += v.FieldSize()
		case "optional":
			for _, vv := range v.Nested {
				sz += vv.FieldSize()
			}
		case "repeated":
		}
	}
	return sz
}

// size of the (hopefulluy single) variable block
func (m *MessageDef) VarSize() int {
	sz := 0
	for _, v := range m.Blocks {
		if v.Cardinality == "repeated" {
			for _, vv := range v.Nested {
				sz += vv.FieldSize()
			}
		}
	}
	return sz
}

// the UBX-INF-xxx MessageDefs are really just strings
func (m *MessageDef) IsString() bool {
	if len(m.Blocks) != 1 {
		return false
	}
	return m.Blocks[0].IsString()
}

func (b *Block) IsString() bool {
	if b.Cardinality != "repeated" {
		return false
	}
	if len(b.Nested) != 1 {
		return false
	}
	if b.Nested[0].Type != "CH" && b.Nested[0].Type != "U1" {
		return false
	}
	return true
}

// numeric fields in any base, not just decimal
type Number uint64

func (v *Number) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var f string
	if err := d.DecodeElement(&f, &start); err != nil {
		return err
	}
	vv, err := strconv.ParseUint(f, 0, 64)
	if err != nil {
		return err
	}
	*v = Number(vv)
	return nil
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
