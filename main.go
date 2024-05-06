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
	"flag"
	"fmt"
	"sync"

	"github.com/ChrIgiSta/go-can-coder/canbus"
	"github.com/ChrIgiSta/go-can-coder/cancoder"
	log "github.com/ChrIgiSta/go-utils/logger"
)

type CanType int

const (
	NetIf  CanType = 0
	Serial CanType = 1
	TCP    CanType = 2
)

// CLI to read encoded data
func main() {
	t := NetIf

	log.Debug("main", "stared")

	fmt.Println("available de-encoder:")
	for _, coder := range cancoder.CancoderDefs {
		fmt.Println("\t-" + coder.Name)
	}

	canDev := flag.String("device", "can0", "CAN Network Interface to connect")
	enDecoder := flag.String("parser", "Opel_Astra_H_OPC_2006", "en- decoder for parsing can frames")
	port := flag.Int("port", -1, "If using tcp as source, define this tcp port")
	baudrate := flag.Int("baud", -1, "If using serial device as source, define this baudrate")

	flag.Parse()

	if *port > 0 {
		t = TCP
	} else if *baudrate > 0 {
		t = Serial
	}

	for _, coder := range cancoder.CancoderDefs {
		if coder.Name == *enDecoder {
			canCli(*canDev, &coder, t, *port, *baudrate)
		}
	}

	log.Info("main", "exited")
}

func canCli(device string, endecoder *cancoder.CancoderDef, canType CanType, port int, baudrate int) {
	var (
		wg     sync.WaitGroup
		canBus canbus.CanBus
		codec  cancoder.Decoder
	)

	defer wg.Wait()

	codec = *cancoder.NewCanDecoder(endecoder.Cancoders[0].Map) // add all available decoders? or select over flag?

	switch canType {
	case TCP:
		canBus = canbus.NewTcpClient(device, uint16(port))
	case NetIf:
		canBus = canbus.NewIface(device)
	case Serial:
		canBus = canbus.NewSerial(device, baudrate)
	}

	wg.Add(1)
	canFrameCh, err := canBus.Connect(&wg)
	if err != nil {
		log.Error("cli", "cannot open connection to can device: %v", err)
		return
	}
	defer canBus.Disconnect()

	for canFrame := range canFrameCh {
		val, err := codec.Decoder(canFrame)
		if err != nil {
			log.Warn("cli", "decoder: %v", err)
		} else if val != nil {
			fmt.Printf("rx[decoded]: [%s]:%v%s\r\n", val.CanValueDef.Name, val.CanValueDef.Value, val.CanValueDef.Unit)
		} else {
			fmt.Printf("rx[encoded]: %x:%X", canFrame.ArbitrationID, canFrame.Data) // only verbose
		}
	}

}
