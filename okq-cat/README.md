# okq-cat

A utility for reading and writing okq events from okq servers or other sources.

Each run of okq-cat must have at least one source (src) and one destination
(dst). An example src would be a queue on an okq server, and an example dst
would be stdout.

Multiple srcs/dsts may be given. All srcs are combined to form one event stream
(where ordering of events cannot be guaranteed), and all events from that stream
are copied to each dst individually. So okq-cat acts as a fan-in/fan-out
process.

## Examples

```bash
# Read events from queue "foo" on one okq instance and write them to queue "foo"
# on another
okq-cat --src-okq 10.0.0.2:4777,foo --dst-okq 10.0.0.3:4777,foo

# Read events from "foo" and "bar", write them to stdout (json encoded, newline
# delimited)
okq-cat --src-okq 10.0.0.2:4777,foo,bar --dst-stdout

# Read events from "foo" and "bar", write them to "baz" on two different servers
okq-cat --src-okq 10.0.0.2:4777,foo,bar \
        --dst-okq 10.0.0.3:4777,baz \
        --dst-okq 10.0.0.4:4777,baz

# Read events from "foo", write them to "bar" and "baz" on a different server.
# When writing, events will be split across "bar" and "baz" using round-robin
okq-cat --src-okq 10.0.0.2:4777,foo --dst-okq 10.0.0.3:4777,bar,baz

# Write events from "foo" to a file
okq-cat --src-okq 10.0.0.2:4777,foo --dst-stdout > foo-events

# Copy events from the file to queue "bar"
cat foo-events | okq-cat --src-stdin --dst-okq 10.0.0.2:4777,bar
```
