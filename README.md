# Go wrapper for the StreamrProxyClient

This is a Go wrapper for the StreamrProxyClient C++ library. It is used to publish data to the Streamr network.

## Installation

```bash
go get github.com/streamr-dev/goproxyclient
```
The package is distributed with a precompiled shared library and is available for MacOS (arm64 and x86_64) and Linux (arm64 and x86_64).

## Usage

```go
package main

import (
	"log"

	streamrproxyclient "github.com/streamr-dev/goproxyclient"
)

func main() {
	// This is a widely-used test account
	ownEthereumAddress :=
	 "0xa5374e3c19f15e1847881979dd0c6c9ffe846bd5";
	ethereumPrivateKey :=
	 "23bead9b499af21c4c16e4511b3b6b08c3e22e76e0591f5ab5ba8d4c3a5b1820";

	// A Proxy server run by Streamr
	proxyUrl := "ws://95.216.15.80:44211"
	proxyEthereumAddress := "0xd0d14b38d1f6b59d3772a63d84ece0a79e6e1c1f"
	
	// The stream to publish to
	streamPartId := "0xd2078dc2d780029473a39ce873fc182587be69db/low-level-client#0"
	
	// Create a new library instance
	lib := streamrproxyclient.NewLibStreamrProxyClient()
	defer lib.Close()

	
	// Create a new ProxyClient instance
	client, err := streamrproxyclient.NewProxyClient(ownEthereumAddress, streamPartId)
	if err != nil {
			log.Fatalf("Error creating ProxyClient: %v", err)
	}
	defer client.Close()

	// Create definition of the proxy server to connect to
	proxies := []streamrproxyclient.Proxy{
		*streamrproxyclient.NewProxy(proxyUrl, proxyEthereumAddress),
	}

	// Connect to the proxy server
	result := client.Connect(proxies)
	if len(result.Errors) != 0 {
		log.Fatalf("Errors during connection: %v", result.Errors)
	}
	if len(result.Successful) != 1 {
		log.Fatalf("Unexpected number of successful connections: %d", len(result.Successful))
	}

	// Publish a message to the test stream
	// You should see the result in the Streamr HUB web UI at
	// https://streamr.network/hub/streams/0xd2078dc2d780029473a39ce873fc182587be69db%2Flow-level-client/live-data
	data := []byte("Hello from Go!")
	result = client.Publish(data, ethereumPrivateKey)
	if len(result.Errors) != 0 {
		log.Fatalf("Errors during publish: %v", result.Errors)
	}
	if len(result.Successful) != 1 {
		log.Fatalf("Unexpected number of successful publishes: %d", len(result.Successful))
	}
}
```

## API documentation

### LibStreamrProxyClient

The main library class that initializes the native library.

#### Methods

- `NewLibStreamrProxyClient() *LibStreamrProxyClient` - Creates a new library instance and initializes the native library
- `Close()` - Cleans up and closes the native library

### ProxyClient 

Client for connecting to proxy servers and publishing messages.

#### Methods

- `NewProxyClient(ownEthereumAddress string, streamPartId string) (*ProxyClient, *ProxyClientError)` - Creates a new proxy client instance
  - `ownEthereumAddress` - Ethereum address of the publisher
  - `streamPartId` - ID of the stream partition to publish to
  - Returns the client instance and any error that occurred

- `Close() *ProxyClientError` - Closes and cleans up the proxy client
  - Returns any error that occurred during cleanup

- `Connect(proxies []Proxy) *ProxyClientResult` - Connects to the specified proxy servers
  - `proxies` - Array of proxy servers to connect to
  - Returns result containing successful connections and any errors

- `Publish(data []byte, ethereumPrivateKey string) *ProxyClientResult` - Publishes data to the stream
  - `data` - Byte array of data to publish
  - `ethereumPrivateKey` - Optional private key for signing messages
  - Returns result containing successful publishes and any errors

### Proxy

Contains information about a proxy server.

#### Methods

- `NewProxy(websocketUrl string, ethereumAddress string) *Proxy` - Creates a new proxy definition
  - `websocketUrl` - WebSocket URL of the proxy server
  - `ethereumAddress` - Ethereum address of the proxy server

### ProxyClientResult

Result of proxy client operations containing successes and failures.

#### Fields

- `Errors []*ProxyClientError` - Array of errors that occurred
- `Successful []*Proxy` - Array of successful proxy operations

### ProxyClientError 

Error type for proxy client operations.

#### Fields

- `Message string` - Error message
- `Code string` - Error code
- `Proxy *Proxy` - Associated proxy if applicable

#### Error Codes

- `ERROR_INVALID_ETHEREUM_ADDRESS` - Invalid Ethereum address provided
- `ERROR_INVALID_STREAM_PART_ID` - Invalid stream partition ID
- `ERROR_PROXY_CLIENT_NOT_FOUND` - Proxy client instance not found
- `ERROR_INVALID_PROXY_URL` - Invalid proxy WebSocket URL
- `ERROR_NO_PROXIES_DEFINED` - No proxy servers defined
- `ERROR_PROXY_CONNECTION_FAILED` - Failed to connect to proxy
- `ERROR_PROXY_BROADCAST_FAILED` - Failed to broadcast message
