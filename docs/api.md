# Lua API Documentation

This is the API documentation for the Lua modules that `sqump` makes available. Please feel free to submit issues for deficiencies or clarifications!

All data types are outlined as best as possible, and all parameters are required unless otherwise noted.

## `sqump`
```
execute(request) -> data
    Parameters:
        request - string, representing the title of another script in the squmpfile collection
    Returns:
        data - table, the data returned by the specified request

export(tbl)
    Parameters:
        tbl - table, data to be made available to subsequent chained scripts

fetch(resource, options) -> response
    Parameters:
        resource - string, the HTTP URL
        options  - optional table, the options modeled after the Fetch API, with some alterations
            method  - string, representing the HTTP method to use (default GET)
            timeout - number, the timeout for the request in seconds (default 10)
            headers - table, an array of strings to use as request headers (default none)
            body    - string, the request body data (default none)
    Returns:
        response - table, holding:
            status  - number, the status code of the response
            headers - table, the headers of the response
            body    - string, the body sent in the response

drill_json(query, json) -> result
    Parameters:
        query - string, period-separated string identifying a location in the given JSON
        json  - string, the JSON being queried
    Returns:
        result - variable, either a found basic Lua value or a JSON string of a map or array if a basic type wasn't found by the query

print_response(response)
    Parameters:
        response - table, holding the result of `fetch`, to be printed to the console
```

## `sqump_kafka`
```
new_consumer(brokers, group, topic, options) -> consumer
    Parameters:
        brokers - table, an array of string addresses for the consumer to connect to
        group   - string, the group ID of the consumer
        topic   - string, the topic to consume from
        options - optional table, with the below fields:
            offset_last - string, if set to "true" will override Kafka consumer group to default to last offset on creation
    Returns:
        consumer - metatable, a custom type representing the connection handle of the consumer connection

consumer:read_message(timeout) -> message
    Parameters:
        timeout - number, an integer representing the timeout for the read in seconds
    Returns:
        message - table, containing:
            key  - string, the key of the Kafka message
            data - string, the value of the Kafka message

consumer:close()
    Note: It is the responsibility of the user to call this message when done reading from the consumer

new_producer(brokers, topic, timeout) -> producer
    Parameters:
        brokers - table, an array of string addresses for the consumer to connect to
        topic   - string, the topic to consume from
        timeout - number, an integer representing the timeout for future writes in seconds
    Returns:
        producer - metatable, a custom type representing the connection handle of the producer connection

producer:write(key, data)
    Parameters:
        key  - string, the key of the Kafka message
        data - string, the value of the Kafka message

producer:close()
    Note: It is the responsibility of the user to call this message when done writing to the producer

provision_topic(brokers, topic)
    Parameters:
        brokers - table, an array of string addresses for the consumer to connect to
        topic   - string, the topic to consume from
    Note: This function only dials the cluster to trigger topic auto-creates. It will have no effect if `allow.auto.create.topics` is not set to `true` on the cluster.
```