# Overview

capture -> decode -> stdout 
        -> remote:host


## Example

Capture tcp traffic on port 80 and decode as ascii

    pipe -p 80 -d ascii

Capture tcp traffic on port 6379 and decode as redis

    pipe -p 6379 -d redis

Capture udp traffic 
    
    pipe -p 53 -u

Mirror traffic on port 80 to remote address 

    pipe -p 80 -t example.com:8000
