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
	"encoding/hex"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/angelodlfrtr/go-can"
	"github.com/mattn/go-tty"

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

	canDev := flag.String("device", "can0",
		"CAN Network Interface to connect")
	enDecoder := flag.String("parser",
		"Opel_Astra_H_OPC_2006", "en- decoder for parsing can frames")
	port := flag.Int("port", -1,
		"If using tcp as source, define this tcp port")
	baudrate := flag.Int("baud", -1,
		"If using serial device as source, define this baudrate")
	verbose := flag.Bool("verbose", false, "print also undecodable can frames")
	raw := flag.Bool("raw", false, "if only the raw output should displayed")
	utf8 := flag.Bool("utf8", false, "show utf8 encoded package")

	flag.Parse()

	if *port > 0 {
		t = TCP
	} else if *baudrate > 0 {
		t = Serial
	}

	for _, coder := range cancoder.CancoderDefs {
		if coder.Name == *enDecoder {
			canCli(*canDev, &coder, t, *port, *baudrate, *verbose, *raw, *utf8)
		}
	}

	log.Info("main", "exited")
}

func help() {
	fmt.Println("StaufiTec- CAN Tool CLI")
	fmt.Println("")
	fmt.Println("send:")
	fmt.Println("     - raw: <arbitration id as hex>:<8 byte data as hex>: e.g. 160:022070d600000000")
}

func canCli(device string, endecoder *cancoder.CancoderDef,
	canType CanType, port int, baudrate int, verbose bool,
	raw bool, utf8 bool) {

	var (
		wg        sync.WaitGroup
		canBus    canbus.CanBus
		codec     cancoder.Decoder
		inputEvnt chan bool     = make(chan bool, 1)
		prettyOut *PrettyOutput = NewPrettyOutput()
		inReader  *InputScanner = NewInputScanner(inputEvnt)
	)

	defer wg.Wait()

	go inReader.Scan()

	codec = *cancoder.NewCanCoder(endecoder.Cancoders[0].Map) // add all available decoders? or select over flag?

	switch canType {
	case TCP:
		fmt.Println("connecting to can via tcp ", device, port)
		canBus = canbus.NewTcpClient(device, uint16(port))
	case NetIf:
		fmt.Println("connecting to can via network interface ", device)
		canBus = canbus.NewIface(device)
	case Serial:
		fmt.Println("connecting to can via serial ", device, baudrate)
		canBus = canbus.NewSerial(device, baudrate)
	}

	wg.Add(1)
	canFrameCh, err := canBus.Connect(&wg)
	if err != nil {
		log.Error("cli", "cannot open connection to can device: %v", err)
		return
	}
	defer canBus.Disconnect()

	go func() {
		for range inputEvnt {
			typed, complete := inReader.Typed()
			prettyOut.Print(typed)
			if complete {
				cmd := inReader.ReadLine()
				// to send on can
				cmdInterpreter(cmd, canBus)
			}
		}
	}()

	for canFrame := range canFrameCh {
		values, err := codec.Decoder(canFrame)
		if err != nil {
			log.Warn("cli", "decoder: %v", err)
		} else if values != nil {
			// fmt.Printf("rx[decoded]: [%s]:%v%s\r\n",
			// 	val.CanValueDef.Name,
			// 	val.CanValueDef.Value,
			// 	val.CanValueDef.Unit)
			for _, val := range values {
				if !raw {
					prettyOut.Add(val)
				} else {
					prettyOut.Add(rawOut(canFrame, utf8))
				}
			}
		} else {
			// fmt.Printf("rx[encoded]: %x:%X",
			// 	canFrame.ArbitrationID,
			// 	canFrame.Data) // only verbose
			if verbose {
				prettyOut.Add(rawOut(canFrame, utf8))
			}
		}
		userIn, _ := inReader.Typed()
		prettyOut.Print(userIn)
	}
}

func rawOut(canFrame *can.Frame, utf8 bool) *cancoder.CanValueMap {
	spaces := ""
	for i := 0; i < 8-int(canFrame.DLC); i++ {
		spaces += " "
	}

	raw := &cancoder.CanValueMap{
		ArbitrationID: canFrame.ArbitrationID,
		OriginalData:  canFrame.Data[:],
		CanValueDef: cancoder.CanValueDef{
			Name:  cancoder.CanVars(fmt.Sprintf("0x%x", canFrame.ArbitrationID)),
			Value: fmt.Sprintf("[%d] 0x%X%s", canFrame.DLC, canFrame.Data[:canFrame.DLC], spaces),
		},
	}

	if utf8 {
		raw.CanValueDef.Value =
			fmt.Sprintf("[%d] 0x%X%s\t%s", canFrame.DLC, canFrame.Data[:canFrame.DLC], spaces,
				string(canFrame.Data[:canFrame.DLC]))
	}

	return raw
}

func cmdInterpreter(cmd string, can canbus.CanBus) {
	if cmd == "help" {
		help()
		return
	}

	sp := strings.Split(cmd, ":")
	if len(sp) == 2 {
		arbId, err := strconv.ParseInt(sp[0], 16, 32)
		if err != nil {
			log.Error("cmd", "cannot get valid arbitration id: %v", err)
			return
		}
		data, err := hex.DecodeString(sp[1])
		if err != nil {
			log.Error("cmd", "cannot convert data: %v", err)
			return
		}
		sendRaw(can, uint32(arbId), data)
	} else if len(sp) == 1 {
		switch sp[0] {
		case "dooropen":
			sendRaw(can, 0x160, []byte{0x02, 0x20, 0x70, 0xD6})
		case "doorclose":
			sendRaw(can, 0x160, []byte{0x02, 0x80, 0x70, 0xD6})
		case "winopen":
			// for i := 0; i < 3; i++ {
			sendRaw(can, 0x160, []byte{0x02, 0x30, 0x70, 0xD6})
			// 	time.Sleep(1000 * time.Millisecond)
			// }
		case "winclose":
			// for i := 0; i < 3; i++ {
			sendRaw(can, 0x160, []byte{0x02, 0xC0, 0x70, 0xD6})
			// 	time.Sleep(1000 * time.Millisecond)
			// }
		}
	}
}

func sendRaw(canbus canbus.CanBus, arbId uint32, data []byte) {
	var data8 [8]byte

	for i, b := range data {
		if i < 8 {
			data8[i] = b
		}
	}

	frame := can.Frame{
		ArbitrationID: arbId,
		DLC:           uint8(len(data)),
		Data:          data8,
	}
	// fmt.Println(frame)

	err := canbus.Send(&frame)
	if err != nil {
		log.Error("send raw", "couldn't send frame: %v", err)
	}
}

type PrettyOutput struct {
	values    []*cancoder.CanValueMap
	lastLines int
	mutex     sync.Mutex
}

func NewPrettyOutput() *PrettyOutput {
	return &PrettyOutput{
		values:    make([]*cancoder.CanValueMap, 0),
		lastLines: 0,
		mutex:     sync.Mutex{},
	}
}

func (o *PrettyOutput) Add(value *cancoder.CanValueMap) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	for i, v := range o.values {
		if v.CanValueDef.Name == value.CanValueDef.Name {
			o.values[i] = value
			return
		}
	}
	o.values = append(o.values, value)
}

func (o *PrettyOutput) Print(appendix string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.clearLines(o.lastLines)
	o.lastLines = 0
	for _, v := range o.values {
		o.lastLines++
		fmt.Printf("* %s:\t %v%s\r\n",
			v.CanValueDef.Name, v.CanValueDef.Value, v.CanValueDef.Unit)
	}
	if appendix != "" {
		o.lastLines++
		fmt.Print("\033[2K\r") // rm line
		fmt.Println("\r > ", appendix)
	}
}

func (o *PrettyOutput) clearLines(num int) {
	for i := 0; i < num; i++ {
		fmt.Print("\033[2K\r") // rm line
		fmt.Print("\033[A")    // curser up
	}
}

type InputScanner struct {
	buffer string
	event  chan<- bool
}

func NewInputScanner(event chan<- bool) *InputScanner {
	return &InputScanner{
		buffer: "",
		event:  event,
	}
}

func (s *InputScanner) Scan() {
	tty, err := tty.Open()

	if err != nil {
		log.Error("input scaner", "open: %v", err)
		return
	}
	defer tty.Close()

	for err == nil {
		var in rune

		in, err = tty.ReadRune()

		if in == 13 {
			in = '\n'
		}

		if in != 127 {
			s.buffer += string(in)
		} else {
			if len(s.buffer) > 0 {
				s.buffer = s.buffer[:len(s.buffer)-1]
			}
		}

		if s.event != nil {
			s.event <- true
		}
	}
}

func (s *InputScanner) Typed() (input string, complete bool) {
	return s.get(false)
}

func (s *InputScanner) ReadLine() (input string) {
	text, completed := s.get(true)
	if completed {
		return text
	}
	return ""
}

func (s *InputScanner) get(rmLine bool) (text string, complete bool) {
	if strings.ContainsAny(s.buffer, "\n") {
		complete = true
		indexNewLine := strings.Index(s.buffer, "\n")
		text = s.buffer[:indexNewLine]
		if rmLine {
			s.buffer = s.buffer[indexNewLine+1:]
		}
		return
	}

	text = s.buffer
	return
}
