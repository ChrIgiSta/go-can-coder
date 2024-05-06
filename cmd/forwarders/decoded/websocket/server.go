/**
 * Copyright © 2023-2024, Staufi Tech - Switzerland
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
	"math/big"
	"sync"

	"github.com/ChrIgiSta/go-can-coder/canbus"
	"github.com/ChrIgiSta/go-can-coder/cancoder"
	"github.com/ChrIgiSta/go-can-coder/server"
	ccrypt "github.com/ChrIgiSta/go-utils/crypto"
	log "github.com/ChrIgiSta/go-utils/logger"
)

// can -> websocket
// websocket -> can
func main() {
	// make a CLI
	cert, key, err := ccrypt.CreateSelfsignedX509Certificate(big.NewInt(202405060001),
		365, ccrypt.KeyLength2048Bit,
		ccrypt.CertificateSubject{
			Organisation: "StaufiTech",
			Country:      "CH",
			Province:     "Thurgau",
			Locality:     "Frauenfeld",
			OrgUnit:      "Development",
			CommonName:   "localhost",
		})
	if err != nil {
		log.Error("main", "cannot create selfsigned cert: %v", err)
	}

	can2Websocket(cancoder.OpelAstraHOpc2006, "myToken", cert, key, 19001)

	log.Info("main", "exited")
}

func can2Websocket(cancoderDef cancoder.CancoderDef, token string, cert []byte, privKey []byte, wsPort uint16) {

	var (
		err    error
		wg     sync.WaitGroup = sync.WaitGroup{}
		failed bool           = false
		// canIfs []*canbus.NetworkIf
		// canDecoders []*cancoder.Decoder
		// canRxChs    []<-chan *can.Frame
		canDecEvnts []<-chan cancoder.CanValueMap
	)

	defer wg.Wait()

	for _, def := range cancoderDef.Cancoders {
		canDev := canbus.NewIface(def.Device)
		// canIfs = append(canIfs, canDev)
		canDec := cancoder.NewCanDecoder(def.Map)
		// canDecoders = append(canDecoders, canDec)
		wg.Add(1)
		canRx, err := canDev.Connect(&wg)
		if err != nil {
			log.Error("can2ws", "connect to %s %s can: %v", cancoderDef.Name, def.Device, err)
			return
		}
		// canRxChs = append(canRxChs, canRx)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !failed {
				_, err := canDec.Decoder(<-canRx)
				if err != nil {
					log.Error("can2ws", "decode frame: %v", err)
					failed = true
				}
			}
		}()
		canDecEvnts = append(canDecEvnts, canDec.GetEventChannel())
	}

	wsDec := server.NewWsServer(
		19001,
		"decoded",
		func(msg any) {
			log.Debug("can2ws", "websocket rx: %v", msg)
			switch m := msg.(type) {
			case server.WsMsg:
				// Need to know the can channel
				for _, cancoder := range cancoderDef.Cancoders {
					if cancoder.Device == m.Device {
						// canDecoders[i].
						// msg.Msg.Value
						// err = canIfs[i].Send(&can.Frame{})
						log.Warn("can2ws", "encoder not implemented: %s %v", cancoder.Device, msg)
						break
					}
				}
			}
		},
		token,
		cert,
		privKey)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := wsDec.Serve(); err != nil {
			log.Error("can2ws", "ws server exited with err %v", err)
			failed = true
		}
	}()

	log.Info("can2ws", "started. ws listen on port %d", wsPort)

	for !failed {

		for i, rxCh := range canDecEvnts {
			select {
			case rx := <-rxCh:
				log.Debug("can2ws", "gmLan rx: %v", rx)
				err = wsDec.Send(server.WsMsg{
					Msg:    rx.CanValueDef,
					Device: cancoderDef.Cancoders[i].Device,
				})
				if err != nil {
					log.Error("can2ws", "send on %s: %v", cancoderDef.Cancoders[i].Device, err)
					failed = true
				}
			default:
			}
		}

		if failed {
			log.Error("can2ws", "something failed. exiting app")
		}
	}
}
