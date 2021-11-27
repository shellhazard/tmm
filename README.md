# tmm

[![Go Reference](https://pkg.go.dev/badge/github.com/shellhazard/tmm.svg)](https://pkg.go.dev/github.com/shellhazard/tmm)

Tiny package that uses [10MinuteMail](https://10minutemail.com) to generate temporary email addresses. Zero dependancies. Supports receiving, forwarding and replying to messages.

## Install

Try out the example program:
```
go install github.com/shellhazard/tmm/cmd/mailwatch
mailwatch
```

Or get stuck in:
```
go get github.com/shellhazard/tmm
```

## Usage

Short example for a common use case:

```go
// Create a new session
s, err := tmm.New()
if err != nil {
	panic(err)
}

// Get the email string
addr := s.Address()

// <register a new account with a service etc. here>

// Wait for verification email to come through
var received tmm.Message

tk := time.NewTicker(2 * time.Second)
for range tk.C {
	mail, err := s.Latest()
	if err != nil {
		panic(err)
	}

	for _, m := range mail {
		if m.Sender == "noreply@example.org" {
			// We've got our email!
			received = m
			break
		}
	}
}

body := received.Plaintext

// <parse body, click any links etc.>
```
