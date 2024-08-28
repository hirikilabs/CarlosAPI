# CarlosAPI

```
___/\/\/\/\/\______/\/\______/\/\/\/\/\____/\/\__________/\/\/\/\______/\/\/\/\/\_
_/\/\____________/\/\/\/\____/\/\____/\/\__/\/\________/\/\____/\/\__/\/\_________
_/\/\__________/\/\____/\/\__/\/\/\/\/\____/\/\________/\/\____/\/\____/\/\/\/\___
_/\/\__________/\/\/\/\/\/\__/\/\__/\/\____/\/\________/\/\____/\/\__________/\/\_
___/\/\/\/\/\__/\/\____/\/\__/\/\____/\/\__/\/\/\/\/\____/\/\/\/\____/\/\/\/\/\___

```

HTTP API code for the C.A.R.L.O.S (Cooperative Amateur Radio-telescope Listening Outer Space)

## Endpoints

* / : GET info from the app (JSON)
* /status : GET info on all the requested recordings (JSON)
* /status/id : GET info on a recording identified by "id" (JSON)
* /record : POST request a new recording (JSON)
* /download/id : GET download the data file from a recording identified by "id"

## Record data

Record parameters are passed as JSON to the /record enpoint as a POST request using the following template:

```
{
    "user": "username",
    "time": integer,
    "frequency": integer,
    "sample_rate": integer,
    "gain": integer,
    "rec_time": integer,
    "wait_time": integer,
    "az": float,
    "el": float,
    "az_range": float,
    "az_step": float,
    "el_range": float,
    "el_step": float,
}
```

where "time" is a unix timestamp in milliseconds, "frequency" is in Hz, "gain" is an integer representing SDR gain*10, "rec_time" and "wait_time" (seconds recording each coordinate and time to wait after moving the rotor) are in milliseconds, and coordinates are all floats. Azimuth and elevation ("az", "el") are in degrees and represent the center coordinates of the recording, "az_range" and "el_range" are the span of the observation, and "az_step" and "el_step" are the step in degrees between each movement.

For example:
```
{
    "user": "dpello",
    "time": 1724764842135,
    "frequency": 1420000000,
    "sample_rate": 2400000,
    "gain": 480,
    "rec_time": 5000,
    "wait_time": 5000,
    "az": 30.0,
    "el": 30.0,
    "az_range": 3.0,
    "az_step": 2.0,
    "el_range": 1.0,
    "el_step": 1.0,
}
```

The request, if valid, will give a JSON answer with the same format but also including this fields:

* "id" : integer, a unique ID for the recording, valid for downloading the data or query the record status.
* "calc_time": the estimated time the recording will take (milliseconds)
* "status": a string representing the recording status, it can be "Created", "Recording", "Recorded" or "Finished"


## Diagram
```
                                        
                 +---------+              
                 |  Main   |              
                 +---------+              
                      |                   
       +--------------+-------------+       
       |              |             |       
   +--------+   +-----------+    +------+   
   |  HTTP  |   | Scheduler |<-->|  DB  |
   +--------+   +-----------+    +------+
       |              |             ^ ^
       |              |             | |
       +----------------------------+ |
                      |               |
                +-----------+         |               
                |   Record  |---------+                   
                +-----------+                        
                      |                  
       +--------------+--------------+       
       |              |              |       
   +--------+   +------------+   +-------+   
   |  SDR   |   | Filesystem |   | Rotor |
   +--------+   +------------+   +-------+
```
                                        
                                        
                                        
                                        
