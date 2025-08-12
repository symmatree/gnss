package msgparse

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// compareVersions compares two version strings like "3.14.9" and "3.14.10"
// Returns true if v1 < v2
func compareVersions(v1, v2 string) bool {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		var err error

		if i < len(parts1) {
			n1, err = strconv.Atoi(parts1[i])
			if err != nil {
				return v1 < v2 // fallback to string comparison
			}
		}

		if i < len(parts2) {
			n2, err = strconv.Atoi(parts2[i])
			if err != nil {
				return v1 < v2 // fallback to string comparison
			}
		}

		if n1 < n2 {
			return true
		} else if n1 > n2 {
			return false
		}
		// if equal, continue to next part
	}

	return false // versions are equal
}

func TestSplitIntoMessages(t *testing.T) {
	// Open the UBX-INF section file
	file, err := os.Open("F9-HPS1.40-interface/section-3.12-UBX-INF.txt")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Parse the messages
	messages, err := SplitIntoMessages(file)
	if err != nil {
		t.Fatalf("Failed to parse messages: %v", err)
	}

	// Check that we got at least 2 messages
	if len(messages) < 2 {
		t.Fatalf("Expected at least 2 messages, got %d", len(messages))
	}

	expectedInf := []Message{
		{
			name:        "UBX-INF-DEBUG",
			section:     "3.12.1",
			description: "ASCII output with debug contents",
			msgType:     "Output",
			comment:     "This message has a variable length payload, representing an ASCII string.",
			structure:   "0xb5 0x62 0x04 0x04 [0..n] see below CK_A CK_B",
			payloadDescription: []string{
				"Byte offset Type Name Scale Unit Description",
				"Start of repeated group (N times)",
				"0 + n CH str - - ASCII Character",
				"End of repeated group (N times)",
			},
		},
		{
			name:        "UBX-INF-ERROR",
			section:     "3.12.2",
			description: "ASCII output with error contents",
			msgType:     "Output",
			comment:     "This message has a variable length payload, representing an ASCII string.",
			structure:   "0xb5 0x62 0x04 0x00 [0..n] see below CK_A CK_B",
			payloadDescription: []string{
				"Byte offset Type Name Scale Unit Description",
				"Start of repeated group (N times)",
				"0 + n CH str - - ASCII Character",
				"UBXDOC-963802114-13138 - R01 3 UBX protocol Page 75 of 296",
				"C1-Public",
				"u-blox F9 HPS 1.40 - Interface description",
				"End of repeated group (N times)",
			},
		},
	}

	// Compare first two messages
	for i := 0; i < 2; i++ {
		if !reflect.DeepEqual(messages[i], expectedInf[i]) {
			t.Errorf("Message %d does not match expected:\nGot:      %+v\nExpected: %+v", i, messages[i], expectedInf[i])

			// More detailed comparison
			if messages[i].name != expectedInf[i].name {
				t.Errorf("Name mismatch: got %q, expected %q", messages[i].name, expectedInf[i].name)
			}
			if messages[i].section != expectedInf[i].section {
				t.Errorf("Section mismatch: got %q, expected %q", messages[i].section, expectedInf[i].section)
			}
			if messages[i].description != expectedInf[i].description {
				t.Errorf("Description mismatch: got %q, expected %q", messages[i].description, expectedInf[i].description)
			}
			if messages[i].msgType != expectedInf[i].msgType {
				t.Errorf("MsgType mismatch: got %q, expected %q", messages[i].msgType, expectedInf[i].msgType)
			}
			if messages[i].comment != expectedInf[i].comment {
				t.Errorf("Comment mismatch: got %q, expected %q", messages[i].comment, expectedInf[i].comment)
			}
			if messages[i].structure != expectedInf[i].structure {
				t.Errorf("Structure mismatch: got %q, expected %q", messages[i].structure, expectedInf[i].structure)
			}
			if !reflect.DeepEqual(messages[i].payloadDescription, expectedInf[i].payloadDescription) {
				t.Errorf("PayloadDescription mismatch:\nGot:      %v\nExpected: %v", messages[i].payloadDescription, expectedInf[i].payloadDescription)
			}
		}
	}
}

func TestAllSectionFiles(t *testing.T) {
	// List of all section files with their expected section prefixes
	sectionFiles := []struct {
		filename      string
		sectionPrefix string
	}{
		{"section-3.9-UBX-ACK.txt", "3.9"},
		{"section-3.10-UBX-CFG.txt", "3.10"},
		{"section-3.11-UBX-ESF.txt", "3.11"},
		{"section-3.12-UBX-INF.txt", "3.12"},
		{"section-3.13-UBX-MGA.txt", "3.13"},
		{"section-3.14-UBX-MON.txt", "3.14"},
		{"section-3.15-UBX-NAV.txt", "3.15"},
		{"section-3.16-UBX-NAV2.txt", "3.16"},
		{"section-3.17-UBX-RXM.txt", "3.17"},
		{"section-3.18-UBX-SEC.txt", "3.18"},
		{"section-3.19-UBX-TIM.txt", "3.19"},
		{"section-3.20-UBX-UPD.txt", "3.20"},
	}

	for _, sf := range sectionFiles {
		t.Run(sf.filename, func(t *testing.T) {
			// Open the section file
			file, err := os.Open("F9-HPS1.40-interface/" + sf.filename)
			if err != nil {
				t.Fatalf("Failed to open test file %s: %v", sf.filename, err)
			}
			defer file.Close()

			// Parse the messages
			messages, err := SplitIntoMessages(file)
			if err != nil {
				t.Fatalf("Failed to parse messages from %s: %v", sf.filename, err)
			}

			// Check that we got at least one message
			if len(messages) == 0 {
				t.Fatalf("Expected at least one message from %s, got 0", sf.filename)
			}

			// Verify all messages have sections starting with the expected prefix
			for i, msg := range messages {
				if !strings.HasPrefix(msg.section, sf.sectionPrefix) {
					t.Errorf("Message %d in %s has section %q, expected to start with %q",
						i, sf.filename, msg.section, sf.sectionPrefix)
				}

				// Check that section field is not empty
				if msg.section == "" {
					t.Errorf("Message %d in %s has empty section field", i, sf.filename)
				}

				// Check that name follows UBX pattern
				if !strings.HasPrefix(msg.name, "UBX-") {
					t.Errorf("Message %d in %s has name %q, expected to start with 'UBX-'",
						i, sf.filename, msg.name)
				}
			}

			// Verify sections are sequential within each file
			prevSection := ""
			for i, msg := range messages {
				if prevSection != "" {
					// Compare section numbers to ensure they're sequential
					if !compareVersions(prevSection, msg.section) {
						t.Errorf("Message %d in %s has section %q which is not greater than previous section %q",
							i, sf.filename, msg.section, prevSection)
					}
				}
				prevSection = msg.section
			}

			t.Logf("Successfully parsed %d messages from %s", len(messages), sf.filename)
		})
	}
}
