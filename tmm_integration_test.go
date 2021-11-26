// +build integration

package tmm

import "testing"

// Contacts live 10MinuteMail service.
func TestNewSession(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Errorf("unexpected error creating session: %s", err)
	}

	if s.Address() == "" {
		t.Errorf("email address in session is empty")
	}

	if s.Expired() == true {
		t.Errorf("session is expired despite having just been created")
	}

	ok, err := s.Renew()
	if !ok {
		t.Errorf("couldn't renew fresh session")
	}
	if err != nil {
		t.Errorf("unexpected error renewing session: %s", err)
	}

	_, err = s.Messages()
	if err != nil {
		t.Errorf("unexpected error fetching messages: %s", err)
	}

	_, err = s.Latest()
	if err != nil {
		t.Errorf("unexpected error fetching latest messages: %s", err)
	}
}
