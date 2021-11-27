## mailwatch

A tiny command line program that creates and continiously renews a temporary email address, logging out any recieved messages. Pass a forwarding address as an argument to turn it into a simple mail proxy.

### Usage

```
go install github.com/shellhazard/tmm/cmd/mailwatch
mailwatch -fwd=realemail@example.com
```