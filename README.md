# Replicated Memory Cache

This is a replicated memory cache system implemented in Go. The system allows for quick access to data using multiple servers, with data being stored in memory as key/value pairs and replicated to other connected servers.

## Features

- Add, retrieve, and delete key/value pairs in the cache
- Replicate data to multiple peer servers for fault tolerance and scalability
- Monitor cache content in real-time with Server-Sent Events (SSE)

## Getting Started

### Prerequisites

- Go programming language installed
- Basic understanding of HTTP and concurrent programming in Go

### Installation

1. Clone this repository:

```bash
git clone https://github.com/your-username/replicated-memory-cache.git
```
## Retrieving Data from the Cache

To retrieve the value associated with a key from the cache, send a GET request to /get endpoint with the key query parameter:

```bash
GET /get?key=mykey
```
## Deleting Data from the Cache

To delete a key/value pair from the cache, send a GET request to /delete endpoint with the key query parameter:


```bash
GET /delete?key=mykey
```
## Monitoring Cache Content

To monitor the content of the cache in real-time, you can use the Server-Sent Events (SSE) endpoint /updates. Connect to this endpoint using an SSE-compatible client to receive updates whenever the cache content changes.
Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.
License

This project is licensed under the MIT License - see the LICENSE file for details.
 