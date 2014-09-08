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

import (
	"net"
	"sync"
	"time"
	"testing"
)

type testTcpSocketProcessorServer struct {
	sync.Mutex
	listener net.Listener
	connection net.Conn
}

func (self *testTcpSocketProcessorServer) close() {

	if self.connection != nil { self.connection.Close() }
	if self.listener != nil { self.listener.Close() }

}

func createTestTcpSocketProcessorServer(t *testing.T) *testTcpSocketProcessorServer {

	server := &testTcpSocketProcessorServer{}

	var err error
	server.listener, err = net.Listen("tcp", "127.0.0.1:9999")

	if err != nil {
		t.Errorf("TestTcpSocketProcessor is broken - unable to start listener: %v", err)
		return nil
	}

	go func() {
		buffer := make([]byte, 10)
		server.connection, err = server.listener.Accept()

		if err != nil {
			t.Errorf("TestTcpSocketProcessor is broken - to accept: %v", err)
			return
		}

		server.connection.Read(buffer)
	}()

	return server
}

func writeTestTcpSocketProcessorMessageAndVerify(	writeChannel chan TcpSocketProcessorWrite,
													responseChannel chan TcpSocketProcessorWrite,
													t *testing.T) {

	writeChannel <- TcpSocketProcessorWrite{ Data: []byte("hello"), ResponseChannel: responseChannel }
	response := <- responseChannel
	if response.Error != nil { t.Errorf("TestTcpSocketProcessor is broken - write error : %v", response.Error) }
	if response.BytesWritten != 5 { t.Errorf("TestTcpSocketProcessor - bad byte count - receved: %d - expected: 5", response.BytesWritten) }
}

func TestTcpSocketProcessor(t *testing.T) {

	server := createTestTcpSocketProcessorServer(t)

	logger := Logger{ Prefix: "test", Appenders: []Appender{ LevelFilter(Info, StdErrAppender()) } }

	writeChannel := make(chan TcpSocketProcessorWrite)
	readChannel := make (chan TcpSocketProcessorRead)

	processor := NewTcpSocketProcessor("127.0.0.1:9999", 1000, 0, 1000, 10, writeChannel, readChannel, logger)
	processor.Start()

	responseChannel := make(chan TcpSocketProcessorWrite)

	writeTestTcpSocketProcessorMessageAndVerify(writeChannel, responseChannel, t)

	time.Sleep(100* time.Millisecond)

	server.close()

	// Close the listener and open again to confirm the reconnect
	server = createTestTcpSocketProcessorServer(t)

	writeTestTcpSocketProcessorMessageAndVerify(writeChannel, responseChannel, t)

	time.Sleep(100* time.Millisecond)

	processor.Stop()
	server.close()
}

