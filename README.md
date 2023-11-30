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


