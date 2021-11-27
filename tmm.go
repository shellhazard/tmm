// Package tmm provides a simple interface to the 10MinuteMail web service.
//   // Create a new session
//   s, err := tmm.New()
//   if err != nil {
// 	   log.Fatal(err)
//   }
//
//   // Check the email address
//   addr := s.Address()
//
//   // Retrieve all messages
//   mail, err := s.Messages()
//   for _, m := range mail {
// 	   fmt.Println(mail.Plaintext)
//   }
package tmm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/shellhazard/tmm/internal"
)

const (
	DefaultTimeout = 10 * time.Second

	baseURL = "https://10minutemail.com"

	endpointAddress     = "session/address"
	endpointExpired     = "session/expired"
	endpointReset       = "session/reset"
	endpointSecondsLeft = "session/secondsLeft"

	endpointMessagesAfter  = "messages/messagesAfter"
	endpointMessageReply   = "messages/reply"
	endpointMessageForward = "messages/forward"
)

var (
	ErrBuildingRequest = errors.New("failed to construct request object")
	ErrRequestFailed   = errors.New("request to 10minutemail failed")
	ErrReadBody        = errors.New("reading response body failed")
	ErrMarshalFailed   = errors.New("marshalling request body failed")
	ErrUnmarshalFailed = errors.New("unmarshalling response body failed")
	ErrMissingSession  = errors.New("missing session cookie in response")
	ErrBlockedByServer = errors.New("server is blocking requests from this host; probably rate limited")
)

// Message represents a single email message sent to a temporary mail.
type Message struct {
	ID string `json:"id"`

	Forwarded         bool   `json:"forwarded"`
	RepliedTo         bool   `json:"repliedTo"`
	SentDate          string `json:"sentDate"`
	SentDateFormatted string `json:"sentDateFormatted"`
	Sender            string `json:"sender"`
	From              string `json:"from"`
	Subject           string `json:"subject"`
	Plaintext         string `json:"bodyPlainText"`
	HTML              string `json:"bodyHtmlContent"`
	Prewview          string `json:"bodyPreview"`
}

// Session holds information required to maintain a 10MinuteMail session.
type Session struct {
	address string
	token   string

	// The last time the session was reset.
	lastreset time.Time

	// The number of the last message fetched,
	// to ensure we aren't refetching the same data.
	lastcount int64

	baseurl string
	c       *http.Client
}

// New creates a new 10MinuteMail session with a random address.
func New() (*Session, error) {
	s := &Session{
		baseurl: baseURL,
		c: &http.Client{
			Timeout: DefaultTimeout,
		},
		// It's better to assume that we have less time than more time.
		// Assume our mail will expire 10 minutes from initialisation,
		// before the request is made.
		lastreset: time.Now(),
	}

	return newSession(s)
}

// NewWithClient is identical to New but allows
// for passing a custom *http.Client object.
func NewWithClient(c *http.Client) (*Session, error) {
	s := &Session{
		baseurl:   baseURL,
		c:         c,
		lastreset: time.Now(),
	}

	return newSession(s)
}

// newSession abstracts the logic of the New function
// to enable testing.
func newSession(s *Session) (*Session, error) {
	// Initialise session
	res, err := s.c.Get(join(baseURL, endpointAddress))
	if err != nil {
		return s, fmt.Errorf("%w: %s", ErrRequestFailed, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		return s, ErrBlockedByServer
	}

	// Read body
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return s, fmt.Errorf("%w: %s", ErrReadBody, err)
	}

	// Store session cookie
	for _, cookie := range res.Cookies() {
		if cookie.Name == "JSESSIONID" {
			s.token = cookie.Value
		}
	}
	if s.token == "" {
		return s, ErrMissingSession
	}

	// Store address
	v := &internal.AddressResponse{}
	err = json.Unmarshal(b, v)
	if err != nil {
		return s, fmt.Errorf("%w: %s", ErrUnmarshalFailed, err)
	}
	s.address = v.Address

	return s, nil
}

// Address returns the email address attached to the current session.
func (s *Session) Address() string {
	return s.address
}

// Expired returns whether or not the session is due to have expired
// and is in need of renewal.
func (s *Session) Expired() bool {
	return !time.Now().Before(s.lastreset.Add(10 * time.Minute))
}

// ExpiresAt returns a time.Time object representing the instant
// in time that the session is due to expire.
func (s *Session) ExpiresAt() time.Time {
	return s.lastreset.Add(10 * time.Minute)
}

// Messages contacts the server and returns a list of all messages
// received to the email address attached to this session.
//
// Note that if any new messages are found, the same counter will
// be updated that is used when calling the session.Latest() method,
// so you won't need to call it afterwards.
func (s *Session) Messages() ([]Message, error) {
	return s.messages(0)
}

// Latest contacts the server and returns a list of any messages
// that haven't already been received by this session.
func (s *Session) Latest() ([]Message, error) {
	return s.messages(s.lastcount)
}

func (s *Session) messages(i int64) ([]Message, error) {
	var m []Message

	// Prepare request
	u := join(s.baseurl, endpointMessagesAfter, strconv.FormatInt(i, 10))
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return m, fmt.Errorf("%w: %s", ErrBuildingRequest, err)
	}

	// Attach token
	req.AddCookie(&http.Cookie{
		Name:   "JSESSIONID",
		Value:  s.token,
		MaxAge: 300,
	})

	// Make request
	res, err := s.c.Do(req)
	if err != nil {
		return m, fmt.Errorf("%w: %s", ErrRequestFailed, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		return m, ErrBlockedByServer
	}

	// Read body
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return m, fmt.Errorf("%w: %s", ErrReadBody, err)
	}

	// Unmarshal response
	err = json.Unmarshal(b, &m)
	if err != nil {
		return m, fmt.Errorf("%w: %s", ErrUnmarshalFailed, err)
	}

	// Update last received counter
	s.lastcount = i + int64(len(m))

	return m, nil
}

// Renew attempts to extend the session by an additional 10 minutes.
//
// Returns a bool indicating whether the server indicated that the
// reset was successful or not and an error if issues were encountered
// while making the request.
func (s *Session) Renew() (bool, error) {
	// If our reset was successful, assume that we have
	// 10 minutes from when this routine began, to be safe.
	resetAt := time.Now()

	// Prepare request
	u := join(s.baseurl, endpointReset)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrBuildingRequest, err)
	}

	// Attach token
	req.AddCookie(&http.Cookie{
		Name:   "JSESSIONID",
		Value:  s.token,
		MaxAge: 300,
	})

	// Make request
	res, err := s.c.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrRequestFailed, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		return false, ErrBlockedByServer
	}

	// Read body
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrReadBody, err)
	}

	// Unmarshal response
	v := &internal.ResetResponse{}
	err = json.Unmarshal(b, v)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrUnmarshalFailed, err)
	}

	// As far as I know, this string indicates success
	if v.Response != "reset" {
		return false, nil
	}

	// Update reset time
	s.lastreset = resetAt

	return true, nil
}

// Reply asks 10MinuteMail to send a reply to the email that sent
// the message with the provided ID, with the provided body.
//
// Returns a bool indicating whether or not the reply was issued
// successfully - failure generally means the message is too old -
// and an error if issues were encountered while making the request.
func (s *Session) Reply(messageid, body string) (bool, error) {
	// Prepare body
	reqbody := &internal.ReplyRequest{}
	reqbody.Reply.MessageID = messageid
	reqbody.Reply.ReplyBody = body

	reqbytes, err := json.Marshal(reqbody)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrMarshalFailed, err)
	}

	// Prepare request
	u := join(s.baseurl, endpointMessageReply)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(reqbytes))
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrBuildingRequest, err)
	}

	// Attach token
	req.AddCookie(&http.Cookie{
		Name:   "JSESSIONID",
		Value:  s.token,
		MaxAge: 300,
	})

	// Make request
	res, err := s.c.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrRequestFailed, err)
	}
	defer res.Body.Close()

	// Check status code to determine result
	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusForbidden:
		return false, ErrBlockedByServer
	default:
		return false, nil
	}
}

// Forward asks 10MinuteMail to forward the message with the
// provided ID to the provided recipient address.
//
// Returns a bool indicating whether or not the forward request was
// issued successfully and an error if issues were encountered while
// making the request.
//
// Note that the server will claim to be successful even if the recipient
// address is invalid or the mail gets rejected after sending.
func (s *Session) Forward(messageid, recipient string) (bool, error) {
	// Prepare body
	reqbody := &internal.ForwardRequest{}
	reqbody.Forward.MessageID = messageid
	reqbody.Forward.ForwardAddress = recipient

	reqbytes, err := json.Marshal(reqbody)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrMarshalFailed, err)
	}

	// Prepare request
	u := join(s.baseurl, endpointMessageForward)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(reqbytes))
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrBuildingRequest, err)
	}

	// Set headers
	req.Header.Add("Content-Type", "application/json")

	// Attach token
	req.AddCookie(&http.Cookie{
		Name:   "JSESSIONID",
		Value:  s.token,
		MaxAge: 300,
	})

	// Make request
	res, err := s.c.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %s", ErrRequestFailed, err)
	}
	defer res.Body.Close()

	// Check status code to determine result
	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusForbidden:
		return false, ErrBlockedByServer
	default:
		return false, nil
	}
}

// join concatinates URL components.
func join(b string, n ...string) string {
	u, err := url.Parse(b)
	if err != nil {
		// should never happen..
		panic(err)
	}
	for _, str := range n {
		u.Path = path.Join(u.Path, str)
	}

	return u.String()
}
