package streamrproxyclient

// #include <stdlib.h>
// #include "shim.h"
// #include "streamrproxyclient.h"
import "C"
import (
	"fmt"
	"unsafe"
)

func openLibrary() (error) { 
	fileName, err := SaveLibToTempFile()
	if err != nil {
		return err
	}
	C.loadSharedLibrary(C.CString(fileName))
	return nil
}

func closeLibrary() {
	C.closeSharedLibrary()
}

// Error codes from the C library.
const (
	ERROR_INVALID_ETHEREUM_ADDRESS = "INVALID_ETHEREUM_ADDRESS"
	ERROR_INVALID_STREAM_PART_ID   = "INVALID_STREAM_PART_ID"
	ERROR_PROXY_CLIENT_NOT_FOUND   = "PROXY_CLIENT_NOT_FOUND"
	ERROR_INVALID_PROXY_URL        = "INVALID_PROXY_URL"
	ERROR_NO_PROXIES_DEFINED       = "NO_PROXIES_DEFINED"
	ERROR_PROXY_CONNECTION_FAILED  = "PROXY_CONNECTION_FAILED"
	ERROR_PROXY_BROADCAST_FAILED   = "PROXY_BROADCAST_FAILED"
)

// Proxy struct represents a proxy with a websocket URL and an Ethereum address.
type Proxy struct {
	WebsocketUrl    string
	EthereumAddress string
}

// NewProxy creates a new Proxy instance.
func NewProxy(websocketUrl, ethereumAddress string) *Proxy {
	return &Proxy{
		WebsocketUrl:    websocketUrl,
		EthereumAddress: ethereumAddress,
	}
}

// FromCProxy converts a C Proxy to a Go Proxy.
func (p *Proxy) FromCProxy(cProxy *C.Proxy) *Proxy {
	return &Proxy{
		WebsocketUrl:    C.GoString(cProxy.websocketUrl),
		EthereumAddress: C.GoString(cProxy.ethereumAddress),
	}
}

// String returns a string representation of the Proxy.
func (p *Proxy) String() string {
	return fmt.Sprintf("Proxy(websocketUrl=%s, ethereumAddress=%s)", p.WebsocketUrl, p.EthereumAddress)
}

// Equals checks if two Proxy instances are equal.
func (p *Proxy) Equals(other *Proxy) bool {
		return p.WebsocketUrl == other.WebsocketUrl && p.EthereumAddress == other.EthereumAddress
}

// ProxyClientError struct represents an error in the ProxyClient.
type ProxyClientError struct {
	Message string
	Code    string
	Proxy   *Proxy
}

// Error implements the error interface for ProxyClientError.
func (e *ProxyClientError) Error() string {
	return e.String()
}

// NewProxyClientError creates a new ProxyClientError from a C Error.
func NewProxyClientError(cError *C.Error) *ProxyClientError {
	var proxy *Proxy
	if cError.proxy != nil {
		proxy = NewProxy(C.GoString(cError.proxy.websocketUrl), C.GoString(cError.proxy.ethereumAddress))
	}
	return &ProxyClientError{
		Message: C.GoString(cError.message),
		Code:    C.GoString(cError.code),
		Proxy:   proxy,
	}
}

// String returns a string representation of the ProxyClientError.
func (e *ProxyClientError) String() string {
	return fmt.Sprintf("Error(message=%s, code=%s, proxy=%v)", e.Message, e.Code, e.Proxy)
}

// ProxyClientResult struct represents the result of a ProxyClient operation.
type ProxyClientResult struct {
	Errors     []*ProxyClientError
	Successful []*Proxy
}

// NewProxyClientResult creates a new ProxyClientResult from a C ProxyResult.
func NewProxyClientResult(proxyResultPtr *C.ProxyResult) *ProxyClientResult {
	result := &ProxyClientResult{
		Errors:     make([]*ProxyClientError, proxyResultPtr.numErrors),
		Successful: make([]*Proxy, proxyResultPtr.numSuccessful),
	}

	for i := 0; i < int(proxyResultPtr.numErrors); i++ {
		cError := (*C.Error)(unsafe.Pointer(uintptr(unsafe.Pointer(proxyResultPtr.errors)) + uintptr(i)*unsafe.Sizeof(*proxyResultPtr.errors)))
		result.Errors[i] = NewProxyClientError(cError)
	}

	for i := 0; i < int(proxyResultPtr.numSuccessful); i++ {
		cProxy := (*C.Proxy)(unsafe.Pointer(uintptr(unsafe.Pointer(proxyResultPtr.successful)) + uintptr(i)*unsafe.Sizeof(*proxyResultPtr.successful)))
		result.Successful[i] = NewProxy(C.GoString(cProxy.websocketUrl), C.GoString(cProxy.ethereumAddress))
	}

	return result
}

// LibStreamrProxyClient struct represents the Streamr Proxy Client library.
type LibStreamrProxyClient struct {
}

// NewLibStreamrProxyClient creates a new LibStreamrProxyClient instance.
func NewLibStreamrProxyClient() *LibStreamrProxyClient {
	err := openLibrary()
	if err != nil {
		panic(err)
	}
	lib := &LibStreamrProxyClient{}
	C.proxyClientInitLibraryWrapper()
	return lib
}

// Close cleans up the Streamr Proxy Client library.
func (l *LibStreamrProxyClient) Close() {
	C.proxyClientCleanupLibraryWrapper()
	closeLibrary()
}

// ProxyClient struct represents a client that connects to proxies.
type ProxyClient struct {
	ownEthereumAddress string
	streamPartId       string
	clientHandle       C.uint64_t
}

// NewProxyClient creates a new ProxyClient instance.
func NewProxyClient(ownEthereumAddress, streamPartId string) (*ProxyClient, *ProxyClientError) {
	client := &ProxyClient{
		ownEthereumAddress: ownEthereumAddress,
		streamPartId:       streamPartId,
	}
	var result *C.ProxyResult
	client.clientHandle = C.proxyClientNewWrapper(&result, C.CString(client.ownEthereumAddress), C.CString(client.streamPartId))
	if result.numErrors > 0 {
		firstError := (*C.Error)(unsafe.Pointer(uintptr(unsafe.Pointer(result.errors)) + uintptr(0)*unsafe.Sizeof(*result.errors)))
		return nil, NewProxyClientError(firstError)
	}
	C.proxyClientResultDeleteWrapper(result)
	return client, nil
}

// Close deletes the ProxyClient instance.
func (p *ProxyClient) Close() *ProxyClientError {
	var result *C.ProxyResult
	C.proxyClientDeleteWrapper(&result, p.clientHandle)
	if result.numErrors > 0 {
		firstError := (*C.Error)(unsafe.Pointer(uintptr(unsafe.Pointer(result.errors)) + uintptr(0)*unsafe.Sizeof(*result.errors)))
		fmt.Printf("Error: %s\n", C.GoString(firstError.message))
		return NewProxyClientError(firstError)
	}
	C.proxyClientResultDeleteWrapper(result)
	return nil
}

// Connect connects the ProxyClient to the given proxies.
func (p *ProxyClient) Connect(proxies []Proxy) *ProxyClientResult {
	numProxies := C.uint64_t(len(proxies))
	var proxyArray []C.Proxy
	if numProxies > 0 {
		proxyArray = make([]C.Proxy, numProxies)
		for i, proxy := range proxies {
			proxyArray[i] = C.Proxy{
				websocketUrl:    C.CString(proxy.WebsocketUrl),
				ethereumAddress: C.CString(proxy.EthereumAddress),
			}
		}
	}
	var result *C.ProxyResult

	if numProxies > 0 {
		C.proxyClientConnectWrapper(&result, p.clientHandle, &proxyArray[0], numProxies)
	} else {
		C.proxyClientConnectWrapper(&result, p.clientHandle, nil, numProxies)
	}
	res := NewProxyClientResult(result)
	C.proxyClientResultDeleteWrapper(result)
	return res
}

// Publish publishes data using the ProxyClient.
func (p *ProxyClient) Publish(data []byte, ethereumPrivateKey string) *ProxyClientResult {
	var result *C.ProxyResult
	if ethereumPrivateKey != "" {
		C.proxyClientPublishWrapper(&result, p.clientHandle, (*C.char)(unsafe.Pointer(&data[0])), C.uint64_t(len(data)), C.CString(ethereumPrivateKey))
	} else {
		C.proxyClientPublishWrapper(&result, p.clientHandle, (*C.char)(unsafe.Pointer(&data[0])), C.uint64_t(len(data)), nil)
	}
	res := NewProxyClientResult(result)
	C.proxyClientResultDeleteWrapper(result)
	return res
}
