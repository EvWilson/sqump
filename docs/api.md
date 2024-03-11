# Lua API Documentation

This is the API documentation for the Lua modules that `sqump` makes available. Please feel free to submit issues for deficiencies or clarifications!

All data types are outlined as best as possible, and all parameters are required unless otherwise noted.

## `sqump`
```
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

to_json(value) -> json
    Parameters:
        value - any, a value to convert to JSON representation
    Returns:
        json - string, the given value in JSON representation

to_json_pretty(value) -> json
    Parameters:
        value - any, a value to convert to JSON representation
    Returns:
        json - string, the given value in JSON representation
    Note: Same as the above, just with the output indented.

from_json(json) -> value
    Parameters:
        json - string, a valid JSON string
    Returns:
        value - any, the given JSON string converted to a Lua value

print_response(response)
    Parameters:
        response - table, holding the result of `fetch`, to be printed to the console
```

## `sqump_kafka`
```
new_consumer(brokers, group, topic, offset) -> consumer
    Parameters:
        brokers - table, an array of string addresses for the consumer to connect to
        group   - string, the group ID of the consumer
        topic   - string, the topic to consume from
        offset  - string, either "first" or "last", specifying the offset to initialize a new consumer group to
    Returns:
        consumer - metatable, a custom type representing the connection handle of the consumer connection

consumer:read_message(timeout, fail_on_timeout) -> message
    Parameters:
        timeout         - number, an integer representing the timeout for the read in seconds
        fail_on_timeout - boolean, determining whether the the script should fail on read timeout
    Returns:
        message - table (or nil on timeout set to not fail), containing:
            key  - string, the key of the Kafka message
            data - string, the value of the Kafka message
    Notes:
        Specifying a read that does not fail on timeout can be helpful for reading to the current end of the topic and then taking action.

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

random_group_id() -> group_id
    Returns:
        group_id - string, random group ID, though not strictly guaranteed to be unique (should be in practice)

provision_topic(brokers, topic)
    Parameters:
        brokers - table, an array of string addresses for the consumer to connect to
        topic   - string, the topic to consume from
    Note: This function only dials the cluster to trigger topic auto-creates. It will have no effect if `allow.auto.create.topics` is not set to `true` on the cluster.
```
