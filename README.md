# A TCP proxy for the 3CPlus V2 Eventbroker

Some languages do not have a complying websocket library at their disposal so in order to integrate with older systems TCP is the only choice left. I'm not going to add a public TCP proxy to our stack (sorry!) because that will require some changes in our load balancers, maintenance, updates and security concerns.
 
This is a simple proxy that exposes the application events from the 3CPlus V2 dialer from Websockets to TCP.

# Installation

Clone the repository, build and run:

```
go build .

3C_PROXY_ADDR=":9090" \
	3C_LOG_LEVEL=error \
	3C_EVENTS_SERVER_ADDR="events.3c.fluxoti.com" \
	./tcp-proxy-3c
```

# Usage

After connecting to the server send your user's token in order to authenticate:

```
echo "my_token_here" | nc -q1 localhost 9090
```

You can then consume the application events, in JSON format, with a unix line terminator ("\n") before each message.

This proxy will also periodically send a "PING" message in order to ensure the connection is still responsive, you do not need to answer to this message. In fact, you don't have to send anything to the server besides the initial token and if done otherwise the server will simply discard it.
