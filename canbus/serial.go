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
	"sync"

	log "github.com/ChrIgiSta/go-utils/logger"

	"github.com/angelodlfrtr/go-can"
	"github.com/angelodlfrtr/go-can/transports"
	"go.bug.st/serial"
)

const (
	SerialCanDefaultPort     = "/dev/ttyS1234"
	SerialCanDefaultBaudrate = 25000

	SerialCanParity   = serial.NoParity
	SerialCanDataBits = 8
	SerialCanStopBits = serial.OneStopBit
)

type Serial struct {
	bus             *can.Bus
	useCustomParser bool
	customParser    CanFrameParser
	port            string
	baudrate        int
	com             serial.Port
}

func NewSerial(port string, baudrate int) *Serial {
	return &Serial{
		bus: can.NewBus(&transports.USBCanAnalyzer{
			Port:     port,
			BaudRate: baudrate,
		}),
		useCustomParser: false,
		customParser:    nil,
		port:            port,
		baudrate:        baudrate,
	}
}

func NewSerialCustomParser(port string, baudrate int, parser CanFrameParser) *Serial {
	return &Serial{
		bus:             nil,
		useCustomParser: true,
		customParser:    parser,
		port:            port,
		baudrate:        baudrate,
	}
}

func (s *Serial) Connect(wg *sync.WaitGroup) (<-chan *can.Frame, error) {

	var err error

	if s.useCustomParser {
		return s.connectNative(wg)
	}

	err = s.bus.Open()
	if err != nil {
		return nil, err
	}

	rxCh := make(chan *can.Frame, CanbusBufferSize)

	go func() {
		defer wg.Done()
		defer close(rxCh)

		for {
			canFrame, ok := <-s.bus.ReadChan()
			if !ok {
				log.Error("serial can", "read can iface: %v", err)
				return
			}

			// do not block on full rx channel
			select {
			case rxCh <- canFrame:
			default:
				log.Warn("serial can", "full rx channel")
			}
		}
	}()

	return rxCh, err
}

func (s *Serial) connectNative(wg *sync.WaitGroup) (<-chan *can.Frame, error) {
	var err error

	s.com, err = serial.Open(s.port, &serial.Mode{
		BaudRate: s.baudrate,
		DataBits: SerialCanDataBits,
		Parity:   SerialCanParity,
		StopBits: SerialCanStopBits,
	})

	if err != nil {
		return nil, err
	}

	rxCh := make(chan *can.Frame, CanbusBufferSize)

	go func() {
		defer wg.Done()
		defer close(rxCh)

		buffer := make([]byte, CanbusBufferSize)
		for {
			n, err := s.com.Read(buffer)
			if err != nil {
				log.Error("serial can", "reading serial, %v", err)
				return
			}
			rxCh <- s.customParser.Unmarshal(buffer[:n])
		}
	}()

	return rxCh, err
}

func (s *Serial) Disconnect() error {
	if s.useCustomParser {
		return s.com.Close()
	}
	return s.bus.Close()
}

func (s *Serial) Send(message *can.Frame) error {
	if s.useCustomParser {
		_, err := s.com.Write(s.customParser.Marshal(message))
		return err
	}
	return s.bus.Write(message)
}
