package msgparse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Message struct {
	name               string
	section            string
	description        string
	msgType            string
	comment            string
	structure          string
	payloadDescription []string
}

func SplitIntoMessages(r io.Reader) ([]Message, error) {
	scanner := bufio.NewScanner(r)
	var messages []Message
	var currentMessage *Message
	var inPayloadDescription bool

	// Regular expressions for parsing
	sectionRegex := regexp.MustCompile(`^(\d+\.\d+\.\d+)\s+(UBX-\w+-\w+)\s+\(0x[0-9a-fA-F]+\s+0x[0-9a-fA-F]+\)`)
	messageLineRegex := regexp.MustCompile(`^Message\s+(UBX-\w+-\w+)`)
	typeRegex := regexp.MustCompile(`^Type\s+(.+)`)
	commentRegex := regexp.MustCompile(`^Comment\s+(.+)`)
	structureRegex := regexp.MustCompile(`^structure\s+(.+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check for new message section
		if matches := sectionRegex.FindStringSubmatch(line); matches != nil {
			// Save previous message if exists
			if currentMessage != nil {
				messages = append(messages, *currentMessage)
			}

			// Start new message
			currentMessage = &Message{
				name:    matches[2],
				section: matches[1],
			}
			inPayloadDescription = false
			continue
		}

		if currentMessage == nil {
			continue
		}

		// Parse description (line after section number but before "Message")
		// Look for lines that start with section number like "3.12.1.1 description text"
		descRegex := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+\s+(.+)$`)
		if currentMessage.description == "" && !strings.HasPrefix(line, "Message") &&
			!strings.HasPrefix(line, "Type") && !strings.HasPrefix(line, "Comment") {
			if matches := descRegex.FindStringSubmatch(line); matches != nil {
				currentMessage.description = matches[1]
			} else {
				currentMessage.description = line
			}
			continue
		}

		// Parse message line (redundant but for consistency)
		if matches := messageLineRegex.FindStringSubmatch(line); matches != nil {
			// Already have the name from section header
			continue
		}

		// Parse type
		if matches := typeRegex.FindStringSubmatch(line); matches != nil {
			currentMessage.msgType = matches[1]
			continue
		}

		// Parse comment
		if matches := commentRegex.FindStringSubmatch(line); matches != nil {
			currentMessage.comment = matches[1]
			continue
		}

		// Parse structure
		if matches := structureRegex.FindStringSubmatch(line); matches != nil {
			currentMessage.structure = matches[1]
			continue
		}

		// Start of payload description
		if line == "Payload description:" {
			inPayloadDescription = true
			continue
		}

		// Collect payload description lines
		if inPayloadDescription {
			// Stop when we hit the next section
			if strings.HasPrefix(line, "3.") && strings.Contains(line, "UBX-") {
				// This is the start of a new section, stop collecting payload
				inPayloadDescription = false
				// Parse this line as a new section
				if matches := sectionRegex.FindStringSubmatch(line); matches != nil {
					// Save current message
					if currentMessage != nil {
						messages = append(messages, *currentMessage)
					}

					// Start new message
					currentMessage = &Message{
						name:    matches[2],
						section: matches[1],
					}
				}
				continue
			}

			currentMessage.payloadDescription = append(currentMessage.payloadDescription, line)
		}
	}

	// Don't forget the last message
	if currentMessage != nil {
		messages = append(messages, *currentMessage)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	return messages, nil
}
