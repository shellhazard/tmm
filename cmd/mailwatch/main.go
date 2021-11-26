package main

import (
	"flag"
	"fmt"
	logpkg "log"
	"os"
	"strings"
	"time"

	"github.com/shellhazard/tmm"
)

var (
	log = logpkg.New(os.Stdout, "", 0)

	fwdaddr = flag.String("fwd", "", "the email address to forward to")
)

func main() {
	flag.Parse()

	// Create new session
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

		// Check if there are any messages we haven't already received
		mail, err := s.Latest()
		if err != nil {
			log.Fatalf("Error fetching latest mail: %s", err)
		}

		output := []string{}

		for _, m := range mail {
			// Prepare console output
			var out strings.Builder
			out.WriteString(fmt.Sprintf("Sender: %s\n", m.Sender))
			out.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
			out.WriteString(fmt.Sprintf("Body: %s\n", m.Plaintext))

			output = append(output, out.String())

			// If an address was provided, forward the message on
			if *fwdaddr != "" {
				ok, err := s.Forward(m.ID, *fwdaddr)
				if !ok {
					log.Println("Server rejected message forward request.")
				}
				if err != nil {
					log.Printf("Error forwarding message: %s", err)
				}
			}
		}

		if len(output) > 0 {
			log.Printf("-----\n%s", strings.Join(output, "-----\n"))
		}
	}
}
