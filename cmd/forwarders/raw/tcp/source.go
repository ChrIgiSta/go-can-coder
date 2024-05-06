/**
 * Copyright © 2024, Staufi Tech - Switzerland
 * All rights reserved.
 *
 *   ________________________   ___ _     ________________________
 *  / _____  _  ____________/  / __|_|   /_______________   _____/
 * ( (____ _| |_ _____ _   _ _| |__ _      | |_____  ____| |__
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

package main

import (
	"flag"
	"sync"

	"github.com/ChrIgiSta/go-can-coder/canbus"
	"github.com/ChrIgiSta/go-can-coder/utils"
	"github.com/ChrIgiSta/go-utils/connection"
	"github.com/ChrIgiSta/go-utils/connection/socket"
	log "github.com/ChrIgiSta/go-utils/logger"
)

func main() {
	var canInterface = flag.String("interface", "can0", "can interface to forward (lookup with ifconfig)")
	var tcpListenerPort = flag.Uint("port", 9001, "port to bind the tcp server")

	flag.Parse()

	forwarder(*canInterface, uint16(*tcpListenerPort), utils.NewCanDriveParser())
}

func forwarder(canInterface string, tcpPort uint16, customParser canbus.CanFrameParser) {
	var wg sync.WaitGroup

	defer wg.Wait()

	tcpRx := make(chan connection.Message, 1024)
	tcpEvnt := make(chan connection.Event, 1024)

	tcpE2c := connection.NewEventsToChannel(tcpRx, tcpEvnt)

	tcpServer := socket.NewTcpServer("", tcpPort, tcpE2c)
	err := tcpServer.ListenAndServe()
	if err != nil {
		log.Error("tcp forwarder", "listen %v", err)
		return
	}
	defer tcpServer.Stop()

	canIf := canbus.NewIface(canInterface)
	wg.Add(1)
	canRx, err := canIf.Connect(&wg)
	if err != nil {
		log.Error("tcp forwarder", "can connect %v", err)
		return
	}
	defer canIf.Disconnect()

	for {
		select {

		case msg, ok := <-canRx:
			if !ok {
				log.Error("tcp forwarder", "error can rx")
				return
			}

			err = tcpServer.Broadcast(customParser.Marshal(msg))
			if err != nil {
				log.Error("tcp forwarder", "error send txp: %v", err)
				return
			}

		case msg, ok := <-tcpRx:
			if !ok {
				log.Error("tcp forwarder", "error can rx")
				return
			}

			cFrame := customParser.Unmarshal(msg.Content)

			err = canIf.Send(cFrame)
			if err != nil {
				log.Error("tcp forwarder", "error send on can interface: %v", err)
				return
			}

		default:
			continue
		}
	}
}
