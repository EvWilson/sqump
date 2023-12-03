# sqump
[![Tests](https://github.com/EvWilson/sqump/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/EvWilson/sqump/actions/workflows/test.yml)

`sqump` is a utility for managing small, composable scripts that perform HTTP requests and operations on the results of these requests.
It is born out of a dissatisfaction with Postman, and after recent policy changes, an outright aversion.

## Quickstart
```
$ go install github.com/EvWilson/sqump@latest
$ sqump init
$ sqump exec Squmpfile.json NewReq
hello, world!
```

The above sequence should get you spun up and executing your first request! (Assuming you have Go 1.21+ installed.)
Start exploring with `sqump help` to find out what's possible.

## Design
`sqump` has two main components: Squmpfiles and its core configuration. Squmpfiles are registered with the core, which
also contains the environment that is currently active. Squmpfiles contain Lua scripts and environment mappings - these scripts
can be chained together, and the current environment can be used to insert data into the scripts.
For a practical intro, consider the following example Squmpfile:
```json
{
  "version": {
    ...
  },
  "title": "My_Example_Squmpfile",
  "requests": [
    {
      "title": "Request_One",
      "script": "..."
    },
    {
      "title": "Request_Two",
      "script": "..."
    }
  ],
  "environment": {
    "staging": {
      "query": "staging_query"
    },
    "demo": {
      "query": "demo_query"
    }
  }
}
```
The script for Request_One:
```lua
local s = require('sqump')

print('hello from request one!')

local response = s.fetch('http://localhost:5000', {
    method = 'POST',
    timeout = 3,
    headers = {
        ['Service-Name'] = 'My_Service_Name'
    },
    body = {
        query = '{{.query}}'
    }
})

s.export({
    status = response.status,
})
```
The script from Request_Two:
```lua
local s = require('sqump')

print('hello from request two!')

local one_result = s.execute('Request_One')

print('request one status:', one_result.status)
```
To narrate: we perform a request in one script, export some details from it, and use those details in another script.
We can change the details of the data we're sending between environments, like in the request body in this instance,
by providing template data to the script.

It's easy to imagine how this could be used to perform commonly chained operations, such as obtaining an access token
before performing an authorized query.

To help with networked operations, and the mentioned exporting between scripts, `sqump` provides a Lua module to tackle
these and other necessary tasks.

## `sqump` Module API
```
execute(request) -> data
    Parameters:
        request - string representing the title of another script in the squmpfile collection
    Returns:
        data - a table of the data returned by the specified request

export(tbl)
    Parameters:
        tbl - a table of data to be made available to subsequent chained scripts

fetch(resource, opts) -> response
    Parameters:
        resource - a string HTTP URL
        opts     - a table of options modeled after the Fetch API, with some alterations
            method  - a string representing the HTTP method to use
            timeout - a numeric timeout for the request in seconds
            headers - a table of strings to use as request headers
            body    - a string to use as the request body
    Returns:
        response - a table, holding:
            status  - a number, the status code of the response
            headers - a table, the headers of the response
            body    - a string, the body sent in the response

drill_json(query, json) -> result
    Parameters:
        query - a period-separated string identifying a location in the given JSON
        json  - a JSON string
    Returns:
        result - either a found basic Lua value, or a JSON string of a map or array if a basic type wasn't found by the query

print_response(response)
    Parameters:
        response - a table, holding the result of `fetch`, to be printed to the console
```

## FAQs
- Sqump?

[Sqump](https://youtu.be/MS1jJzoMUjI?si=PPH_hONo0wEKNAmx&t=414)

### Planned Work
- Finish working on a basic web view for the less CLI-inclined
- More testing: cover import cycles, improper params
- Someday, as necessary, automated config migration
