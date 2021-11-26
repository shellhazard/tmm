package main

import (
	"fmt"
	logpkg "log"
	"os"
	"strings"
	"time"

	"github.com/shellhazard/tmm"
)

var log = logpkg.New(os.Stdout, "", 0)

func main() {
	s, err := tmm.New()
	if err != nil {
		log.Fatalf("Error during session initialisation: %s", err)
	}

	log.Printf("Your address is: %s", s.Address())

	// Poll for new messages every two seconds.
	tk := time.NewTicker(2 * time.Second)

	log.Println("Waiting for new messages.")

	for range tk.C {
		// Check if we need to renew our session.
		if s.Expired() {
			log.Println("Renewing session..")
			ok, err := s.Renew()
			if !ok {
				log.Fatal("Session permanently expired, exiting.")
			}
			if err != nil {
				log.Fatalf("Error renewing session: %s", err)
			}
			log.Println("Successfully renewed session.")
		}

		mail, err := s.Latest()
		if err != nil {
			log.Fatalf("Error fetching latest mail: %s", err)
		}

		output := []string{}

		for _, m := range mail {
			var out strings.Builder
			out.WriteString(fmt.Sprintf("Sender: %s\n", m.Sender))
			out.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
			out.WriteString(fmt.Sprintf("Body: %s\n", m.Plaintext))

			output = append(output, out.String())
		}

		if len(output) > 0 {
			log.Printf("-----\n%s", strings.Join(output, "-----\n"))
		}
	}
}
