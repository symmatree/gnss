package ubx

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// TextMessage represents the structure from msggen.go (subset needed for conversion)
type TextMessage struct {
	Name        string
	Type        string
	Description string
	Comment     string
	Class       uint64
	Id          uint64
	Length      string
	Blocks      []*TextBlock
}

// TextBlock represents a field structure from msggen.go (subset needed for conversion)
type TextBlock struct {
	Name    string
	Offset  string
	Type    string
	Comment string
	Scale   string
	Unit    string
}

// XMLMessage represents the XML structure for a UBX message for output
type XMLMessage struct {
	XMLName     xml.Name     `xml:"Message"`
	Name        string       `xml:"Name"`
	Type        string       `xml:"Type"`
	Description string       `xml:"Description"`
	Comment     string       `xml:"Comment"`
	Structure   XMLStructure `xml:"Structure"`
}

// XMLStructure represents the message structure for XML output
type XMLStructure struct {
	Header   string     `xml:"Header"`
	Class    string     `xml:"Class"`
	Id       string     `xml:"Id"`
	Length   string     `xml:"Length"`
	Payload  XMLPayload `xml:"Payload"`
	Checksum string     `xml:"Checksum"`
}

// XMLPayload contains the payload blocks for XML output
type XMLPayload struct {
	Blocks []XMLBlock `xml:"Block"`
}

// XMLBlock represents a field in the payload for XML output
type XMLBlock struct {
	Offset  string `xml:"Offset"`
	Name    string `xml:"Name"`
	Type    string `xml:"Type"`
	Comment string `xml:"Comment"`
	Scale   string `xml:"Scale"`
	Unit    string `xml:"Unit"`
}

// ParseTextToXML parses a UBX text format and returns XML representation
func ParseTextToXML(reader io.Reader) (*XMLMessage, error) {
	scanner := bufio.NewScanner(reader)
	message := &XMLMessage{}

	// Regular expressions for parsing different parts
	titleRegex := regexp.MustCompile(`^[\d\.]+\s+(UBX-\w+-\w+)\s+\(0x([0-9a-fA-F]+)\s+0x([0-9a-fA-F]+)\)`)
	descRegex := regexp.MustCompile(`^[\d\.]+\s+(.+)$`)
	typeRegex := regexp.MustCompile(`^Type\s+(.+)$`)
	commentRegex := regexp.MustCompile(`^Comment\s+(.+)$`)
	structureRegex := regexp.MustCompile(`^structure\s+0xb5\s+0x62\s+0x([0-9a-fA-F]+)\s+0x([0-9a-fA-F]+)\s+(\d+)\s+.*\s+(CK_A\s+CK_B)`)
	payloadRegex := regexp.MustCompile(`^(\d+)\s+(\w+)\s+(\w+)\s+([^\s]+)\s+([^\s]+)\s+(.+)$`)

	var inPayloadSection bool
	var commentLines []string
	var descriptionSet bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse title line with message name and class/id
		if matches := titleRegex.FindStringSubmatch(line); matches != nil {
			message.Name = matches[1]
			message.Structure.Class = "0x" + strings.ToLower(matches[2])
			message.Structure.Id = "0x" + strings.ToLower(matches[3])
			message.Structure.Header = "0xB5 0x62"
			continue
		}

		// Parse description (second numbered line)
		if matches := descRegex.FindStringSubmatch(line); matches != nil && !descriptionSet {
			message.Description = matches[1]
			descriptionSet = true
			continue
		}

		// Parse type
		if matches := typeRegex.FindStringSubmatch(line); matches != nil {
			message.Type = matches[1]
			continue
		}

		// Parse comment (can be multi-line)
		if matches := commentRegex.FindStringSubmatch(line); matches != nil {
			commentLines = append(commentLines, matches[1])
			continue
		}

		// Continue collecting comment lines until we hit something else
		if len(commentLines) > 0 && !strings.HasPrefix(line, "Header") && !strings.HasPrefix(line, "structure") && !strings.HasPrefix(line, "Payload") {
			commentLines = append(commentLines, line)
			continue
		}

		// Parse structure line
		if matches := structureRegex.FindStringSubmatch(line); matches != nil {
			message.Structure.Length = matches[3]
			message.Structure.Checksum = matches[4]

			// Join all comment lines
			if len(commentLines) > 0 {
				message.Comment = strings.Join(commentLines, " ")
			}
			continue
		}

		// Start of payload section
		if strings.HasPrefix(line, "Payload description:") {
			inPayloadSection = true
			continue
		}

		// Skip header line in payload section
		if inPayloadSection && strings.HasPrefix(line, "Byte offset") {
			continue
		}

		// Parse payload fields
		if inPayloadSection {
			if matches := payloadRegex.FindStringSubmatch(line); matches != nil {
				block := XMLBlock{
					Offset:  matches[1],
					Type:    matches[2],
					Name:    matches[3],
					Scale:   matches[4],
					Unit:    matches[5],
					Comment: matches[6],
				}
				message.Structure.Payload.Blocks = append(message.Structure.Payload.Blocks, block)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return message, nil
}

// ToXML converts the message to XML format
func (m *XMLMessage) ToXML() ([]byte, error) {
	output, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling to XML: %w", err)
	}
	return output, nil
}

// ConvertToXMLMessage converts a TextMessage to XMLMessage format
func ConvertToXMLMessage(msg *TextMessage) *XMLMessage {
	xmlMsg := &XMLMessage{
		Name:        msg.Name,
		Type:        msg.Type,
		Description: msg.Description,
		Comment:     msg.Comment,
		Structure: XMLStructure{
			Header:   "0xB5 0x62",
			Class:    fmt.Sprintf("0x%02x", msg.Class),
			Id:       fmt.Sprintf("0x%02x", msg.Id),
			Length:   msg.Length,
			Checksum: "CK_A CK_B",
		},
	}

	// Convert blocks
	for _, block := range msg.Blocks {
		xmlBlock := XMLBlock{
			Offset:  block.Offset,
			Name:    block.Name,
			Type:    block.Type,
			Comment: block.Comment,
			Scale:   block.Scale,
			Unit:    block.Unit,
		}
		xmlMsg.Structure.Payload.Blocks = append(xmlMsg.Structure.Payload.Blocks, xmlBlock)
	}

	return xmlMsg
}
