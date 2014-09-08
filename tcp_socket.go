/**
 * (C) Copyright 2014, Deft Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dlshared

import(
	"io"
	"net"
	"time"
	"strings"
	"crypto/tls"
)

const(
	tcpNetwork = "tcp"
	tcpNetworkPortSep = ":"
)

// The TCP socket processor is a wrapper around a TCP socket that reconnects
// if there is a failure. It also spawns a reading and a writing goroutine to
// handle data in/out. After the struct is created, you must call Start. When
// you are done using, call the Stop method. Make sure you check the error returned
// when calling Start. If this is a tls connection, the component will not run
// if either either the certificate or key file is not accessible. The component
// will also not run if it the address is not set.
type TcpSocketProcessor struct {
	Logger

	shutdownChannel chan bool

	address string
	certificateFile string
	keyFile string

	connectTimeout time.Duration
	readTimeout time.Duration
	writeTimeout time.Duration

	readBufferSize int

	readChannel chan TcpSocketProcessorRead
	writeChannel chan TcpSocketProcessorWrite

	tlsConf *tls.Config
}

type TcpSocketProcessorRead struct {
	Data []byte
	BytesRead int
	Error error
}


type TcpSocketProcessorWrite struct {
	Data []byte
	ResponseChannel chan TcpSocketProcessorWrite
	BytesWritten int
	Error error
}

// Create a new tcp socket processor.
func NewTcpSocketProcessor(	address string,
							connectTimeoutInMs int64,
							readTimeoutInMs int64,
							writeTimeoutInMs int64,
							readBufferSize int,
							writeChannel chan TcpSocketProcessorWrite,
							readChannel chan TcpSocketProcessorRead,
							logger Logger) *TcpSocketProcessor {

	return &TcpSocketProcessor{
		Logger: logger,
		address: address,
		connectTimeout: time.Duration(connectTimeoutInMs) * time.Millisecond,
		readTimeout: time.Duration(readTimeoutInMs) * time.Millisecond,
		writeTimeout: time.Duration(writeTimeoutInMs) * time.Millisecond,
		readBufferSize: readBufferSize,
		readChannel: readChannel,
		writeChannel: writeChannel,
		shutdownChannel: make(chan bool),
	}
}

// Create a new tcp socket processor.
func NewTlsTcpSocketProcessor(	address string,
								connectTimeoutInMs int64,
								readTimeoutInMs int64,
								writeTimeoutInMs int64,
								readBufferSize int,
								writeChannel chan TcpSocketProcessorWrite,
								readChannel chan TcpSocketProcessorRead,
								logger Logger,
								certificateFile,
								keyFile string) *TcpSocketProcessor {

	processor := NewTcpSocketProcessor(	address,
										connectTimeoutInMs,
										readTimeoutInMs,
										writeTimeoutInMs,
										readBufferSize,
										writeChannel,
										readChannel,
										logger)

	processor.certificateFile = certificateFile
	processor.keyFile = keyFile
	return processor
}

// This method should block and then return when the socket is closed or if the remote server
// directs it to do so.
func (self *TcpSocketProcessor) networkReader(connection net.Conn, readerLostConnectionChannel chan bool) {

	read := &TcpSocketProcessorRead { Data: make([]byte, self.readBufferSize, self.readBufferSize) }

	for {
		if self.readTimeout > 0 { connection.SetReadDeadline(time.Now().Add(self.readTimeout)) }

		read.BytesRead, read.Error = connection.Read(read.Data)

		// The socket was closed.
		if read.BytesRead == 0 || read.Error == io.EOF { readerLostConnectionChannel <- true; return }

		self.readChannel <- *read
	}
}

// This returns true if the connection was lost through normal reasons. If false is returned it received
// a shutdown message on the writer shutdown channel. This method blocks until it is shutdown or it
// loses the connection.
func (self *TcpSocketProcessor) networkWriter(connection net.Conn, readerLostConnectionChannel, writerShutdownChannel chan bool) bool {
	for {
		select {
			case write := <- self.writeChannel:

				if self.writeTimeout > 0 { connection.SetWriteDeadline(time.Now().Add(self.writeTimeout)) }

				write.BytesWritten, write.Error = connection.Write(write.Data)

				write.ResponseChannel <- write

				// We are going to close the socket if there is a write error.
				if write.Error != nil { return true }

			case <- readerLostConnectionChannel: return true

			case <- writerShutdownChannel: return false
		}
	}

	return true
}

func (self *TcpSocketProcessor) connectAndProcess(connectionLostChannel, writerShutdownChannel chan bool) {

	readerLostConnectionChannel := make(chan bool)

	connection, err := self.connect()

	if err != nil {
		self.Logf(Error, "Unable to connect to socket - err: %v", err)
		time.Sleep(2 * time.Second)
		connectionLostChannel <- true
		return
	}

	defer connection.Close()

	go self.networkReader(connection, readerLostConnectionChannel)

	// If we lost the channel because of an error, we are going to trigger
	// another connection to be opened.
	if self.networkWriter(connection, readerLostConnectionChannel, writerShutdownChannel) { connectionLostChannel <- true }
}

// Manages the socket and the send/receive goroutines. If a socket connection
// is lost, a new one is opened.
func (self *TcpSocketProcessor) process() {

	writerShutdownChannel := make(chan bool)
	connectionLostChannel := make(chan bool)

	// Trigger the connection to opened.
	go func() { connectionLostChannel <- true }()

	for {
		select {
			case <- connectionLostChannel: go self.connectAndProcess(connectionLostChannel, writerShutdownChannel)
			case <- self.shutdownChannel: writerShutdownChannel <- true; return
		}
	}
}

// Open a connection. You must close this connection when you are done ;)
func (self *TcpSocketProcessor) connect() (net.Conn, error) {

	connection, err := net.DialTimeout(tcpNetwork, self.address, self.connectTimeout)
	if err != nil { return nil, NewStackError("Unable to dial address: %s - err: %v", self.address, err) }

	if self.tlsConf == nil { return connection, err }

	// We have a tls connection.
	tlsConnection := tls.Client(connection, self.tlsConf)
	if err = tlsConnection.Handshake(); err != nil { return nil, NewStackError("Handshake failed - address: %s - err: %v", self.address, err) }
	return tlsConnection, nil
}

func (self *TcpSocketProcessor) Start() error {

	if self.readBufferSize <= 0 { return NewStackError("The read buffer size must be at least one - received : %d", self.readBufferSize) }

	if len(self.address) == 0 { return NewStackError("Unable to run - no address") }

	// Check to see if this a TLS connection - if so, setup the config.
	if len(self.certificateFile) > 0 {
		cert, err := tls.LoadX509KeyPair(self.certificateFile, self.keyFile)
		if err != nil { return NewStackError("Unable to init tls cert - cert: %s - key: %s", self.certificateFile, self.keyFile) }

		parts := strings.Split(self.address, tcpNetworkPortSep)

		if len(parts) != 1 {  return NewStackError("Invalid address: %s", self.address) }

		self.tlsConf = &tls.Config{
			Certificates: []tls.Certificate{ cert },
			ServerName: parts[0],
		}
	}

	go self.process()

	return nil
}

func (self *TcpSocketProcessor) Stop() error {
	close(self.readChannel)
	self.shutdownChannel <- true
	return nil
}

