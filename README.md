# Capture and mirror network traffic

[![Build Status](https://travis-ci.org/monsterxx03/pipe.svg?branch=master)](https://travis-ci.org/monsterxx03/pipe)



## Example

Capture tcp traffic on port 80 and decode as ascii

    pipe -p 80 -d ascii

Capture tcp traffic on port 6379 and decode as redis

    pipe -p 6379 -d redis

Capture udp traffic 
    
    pipe -p 53 -u

Mirror traffic on port 80 to remote address 

    pipe -p 80 -t example.com:8000

Decode redis traffic on port 6379 and write decoded msg to local file

    pipe -p 6379 -d redis -w result.txt

Decode http traffic on port 80 with filter

    pipe -p 80 -d http -f "method: GET & url /hello"
