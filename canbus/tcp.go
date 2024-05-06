/**
 * Copyright © 2023-2024, Staufi Tech - Switzerland
 * All rights reserved.
 *
 *   ________________________   ___ _     ________________  _  ____
 *  / _____  _  ____________/  / __|_|   /_______________  | | ___/
 * ( (____ _| |_ _____ _   _ _| |__ _      | |_____  ____| |_|_
 *  \____ (_   _|____ | | | (_   __) |     | | ___ |/ ___)  _  \
 *  _____) )| |_/ ___ | |_| | | |  | |     | | ____( (___| | | |
 * (______/  \__)_____|____/  |_|  |_|     |_|_____)\____)_| |_|
 *
 *
 *  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 *  AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 *  IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 *  ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 *  LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 *  CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 *  SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 *  INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 *  CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 *  ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 *  POSSIBILITY OF SUCH DAMAGE.
 */

package canbus

import (
	"bufio"
	"bytes"
	"net"
	"strconv"
	"sync"

	log "github.com/ChrIgiSta/go-utils/logger"

	"github.com/angelodlfrtr/go-can"
	"github.com/angelodlfrtr/go-can/transports"
)

const (
	TcpCanNetworkType = "tcp"
)

const (
	TCP_CAN_DEFAULT_PORT = 9001
)

type TcpClient struct {
	address         string
	port            uint16
	tcp             net.Conn
	useCustomParser bool
	customParser    CanFrameParser
	bus             can.Bus
}

func NewTcpClient(address string, port uint16) *TcpClient {
	return &TcpClient{
		bus: *can.NewBus(&transports.TCPCan{
			Host: address,
			Port: int(port),
		}),
		useCustomParser: false,
	}
}

func NewTcpClientCustomParser(address string, port uint16,
	parser CanFrameParser) *TcpClient {

	return &TcpClient{
		address:         address,
		port:            port,
		useCustomParser: true,
		customParser:    parser,
	}
}

func (c *TcpClient) Connect(wg *sync.WaitGroup) (<-chan *can.Frame, error) {

	var err error

	if c.useCustomParser {
		return c.connectTcpNative(wg)
	}

	err = c.bus.Open()
	if err != nil {
		return nil, err
	}

	rxCh := make(chan *can.Frame, CanbusBufferSize)

	go func() {
		defer wg.Done()
		defer close(rxCh)

		for {
			canFrame, ok := <-c.bus.ReadChan()
			if !ok {
				log.Error("tcp can", "read can tcp: %v", err)
				return
			}
			rxCh <- canFrame
		}
	}()

	return rxCh, err
}

func (c *TcpClient) connectTcpNative(wg *sync.WaitGroup) (<-chan *can.Frame, error) {
	var err error

	c.tcp, err = net.Dial(TcpCanNetworkType, c.address+":"+strconv.Itoa(int(c.port)))
	if err != nil {
		return nil, err
	}

	rxCh := make(chan *can.Frame, CanbusBufferSize)

	go func() {
		defer wg.Done()
		defer close(rxCh)

		buffer := make([]byte, CanbusBufferSize)

		packages := 0
		for {
			b, err := c.tcp.Read(buffer)
			if err != nil {
				log.Error("tcp can", "read tcp driveMode: %v", err)
			} else if b > 0 {
				// line for line
				scanner := bufio.NewScanner(bytes.NewBuffer(buffer[:b-1]))
				for scanner.Scan() {
					packages++
					canFrame := c.customParser.Unmarshal(scanner.Bytes())
					if canFrame != nil {
						select {
						// do not block, if the rx channel is full
						case rxCh <- canFrame:
						default:
							log.Warn("tcp can", "full rx channel")
						}
					}
				}
			}
		}
	}()

	return rxCh, err
}

func (c *TcpClient) Disconnect() error {

	if c.useCustomParser {
		return c.tcp.Close()
	}
	return c.bus.Close()
}

func (c *TcpClient) Send(message *can.Frame) error {

	if c.useCustomParser {
		_, err := c.tcp.Write(c.customParser.Marshal(message))
		return err
	}
	return c.bus.Write(message)
}
