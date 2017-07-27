# Capture and mirror network traffic

[![Build Status](https://travis-ci.org/monsterxx03/pipe.svg?branch=master)](https://travis-ci.org/monsterxx03/pipe)



## Example

Capture tcp traffic on port 80 and decode as text

    pipe -p 80 -d text

Capture tcp traffic on port 6379 and decode as redis

    pipe -p 6379 -d redis


Decode http traffic on port 80 with filter(fitler value should be valid golang regexp):

    pipe -p 80 -d http -f "method: POST & url: /hello & Content-Type: application/json"
    
    
##  TODO

- [] redis filter
- [] http response filter
- [] traffic redirect

