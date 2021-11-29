package tmm

import (
	"encoding/json"
	"testing"
	"time"
)

const ExampleMessages = `[
	{
	    "read": false,
	    "expanded": false,
	    "forwarded": false,
	    "repliedTo": false,
	    "sentDate": "2021-11-28T08:21:06.000+0000",
	    "sentDateFormatted": "Nov 28, 2021, 8:21:06 AM",
	    "sender": "example@example.com",
	    "from": "[Ljavax.mail.internet.InternetAddress;@683d4237",
	    "subject": "Testing",
	    "bodyPlainText": "hello world",
	    "bodyHtmlContent": "<div>hello world<br></div>",
	    "bodyPreview": "hello world",
	    "id": "-14532887521908171110"
	}
]`

const ExampleMessage = `{
    "read": false,
    "expanded": false,
    "forwarded": false,
    "repliedTo": false,
    "sentDate": "2021-11-28T08:21:06.000+0000",
    "sentDateFormatted": "Nov 28, 2021, 8:21:06 AM",
    "sender": "example@example.com",
    "from": "[Ljavax.mail.internet.InternetAddress;@683d4237",
    "subject": "Testing",
    "bodyPlainText": "hello world",
    "bodyHtmlContent": "<div>hello world<br></div>",
    "bodyPreview": "hello world",
    "id": "-14532887521908171110"
}`

func TestUnmarshalMessage(t *testing.T) {
	var m Message

	err := json.Unmarshal([]byte(ExampleMessage), &m)
	if err != nil {
		t.Errorf("failed to unmarshal message: %s", err)
	}

	if m.SentDate.IsZero() {
		t.Error("time object is zero, shouldn't be")
	}
}

func TestExpired(t *testing.T) {
	s := Session{
		lastreset: time.Now().Add(-10 * time.Minute),
	}

	if s.Expired() != true {
		t.Errorf("session should be expired")
	}

	s.lastreset = time.Now()

	if s.Expired() != false {
		t.Errorf("session should NOT be expired")
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name string
		b    string
		n    []string
		want string
	}{
		{"base with one extra", "https://example.com", []string{"mypath"}, "https://example.com/mypath"},
		{"base with no extra", "https://example.com", []string{}, "https://example.com"},
		{"base with three extra", "https://example.com", []string{"one", "two", "three"}, "https://example.com/one/two/three"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := join(tt.b, tt.n...)
			if s != tt.want {
				t.Errorf("Got %s, want %s", s, tt.want)
			}
		})
	}
}
