package msgparse

import (
	"os"
	"reflect"
	"testing"
)

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
