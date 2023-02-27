# Eco

A small HTTP Go server that sends back what you specify at start up.

## Usage

By default it listens at port 8081 and sends back "Hello, World!". 
More than one port can be specified.

### Example

`eco -listen=":8081,:8082" -reponse_status=200 -response_body='Hello, world!'`

### Help

`eco -help`

## License

MIT
