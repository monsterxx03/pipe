# Capture and mirror network traffic

[![Build Status](https://travis-ci.org/monsterxx03/pipe.svg?branch=master)](https://travis-ci.org/monsterxx03/pipe)



## Example

Capture tcp traffic on port 80 and decode as ascii

    pipe -p 80 -d ascii

Capture tcp traffic on port 6379 and decode as redis

    pipe -p 6379 -d redis

Decode redis traffic on port 6379 and write decoded msg to local file

    pipe -p 6379 -d redis -w result.txt

Decode http traffic on port 80 with filter(fitler value should be valid golang regexp):

    pipe -p 80 -d http -f "method: POST & url: /hello & Content-Type: application/json"

Decode in stream mode:

    pipe -p 80 -d http -stream

## About -stream

by default, stream mode is disabled, pipe will handle packet it gets one by one.

It may fail to decode some packets since tcp is a stream protocol, every packet can have multi
requests or one request maybe splitted into multi packets.

If enable stream, pipe will take use of gopacket's tcpassembly function, then it can walk through the whole stream,
but the problem is tcpassembly only track complete tcp stream, so for tcp connections opened before it starts sniffing(keep-alive),
tcpassembly will discard them, so pipe will output nothing for them.
