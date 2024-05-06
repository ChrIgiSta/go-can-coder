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

package cancoder

import (
	"github.com/ChrIgiSta/go-can-coder/utils"

	"fmt"
	"testing"
	"time"

	"github.com/angelodlfrtr/go-can"
)

func TestGMLanDecoder(t *testing.T) {

	var err error

	gmLan := NewCanDecoder(OpelAstraHOpc2006GMLan)

	fr := can.Frame{}

	// Speeds (RPM / km/h)
	fr.ArbitrationID = uint32(GMLanEngineSpeedRPM)
	fr.Data = [8]uint8{0x13, 0x0c, 0xf3, 0x00, 0x04, 0xe5, 0x00, 0x00}
	_, err = gmLan.Decoder(&fr)
	if err != nil {
		t.Error(err)
	}
	engineSpeed := gmLan.GetValue(EngineSpeedRPM)
	vehicleSpeed := gmLan.GetValue(VehicleSpeed)
	engineState := gmLan.GetValue(EngineRunningState)

	printCanValueInfos(t, engineSpeed)
	printCanValueInfos(t, vehicleSpeed)
	printCanValueInfos(t, engineState)
}

func TestMidspeedDecoder(t *testing.T) {
	displayData := [][8]byte{
		{0x10, 0x42, 0x40, 0x00, 0x3F, 0x03, 0x10, 0x13},
		{0x21, 0x00, 0x1B, 0x00, 0x5B, 0x00, 0x66, 0x00},
		{0x22, 0x53, 0x00, 0x5F, 0x00, 0x67, 0x00, 0x6D},
		{0x23, 0x00, 0x4E, 0x00, 0x6F, 0x00, 0x20, 0x00},
		{0x24, 0x53, 0x00, 0x6F, 0x00, 0x75, 0x00, 0x72},
		{0x25, 0x00, 0x63, 0x00, 0x65, 0x00, 0x20, 0x00},
		{0x26, 0x20, 0x00, 0x20, 0x00, 0x21, 0x00, 0x21},
		{0x27, 0x00, 0x20, 0x00, 0x20, 0x00, 0x20, 0x00},
		{0x28, 0x64, 0x00, 0x6D, 0x00, 0x20, 0x00, 0x4D},
		{0x29, 0x00, 0x41, 0x00, 0x4E, 0x00, 0x21, 0x00},
	}

	entertainmentBus := NewCanDecoder(OpelAstraHOpc2006EntertainmentCAN)

	// Display
	fr := can.Frame{}
	fr.ArbitrationID = uint32(EntertainmentCANDisplayData)

	for _, data := range displayData {

		fr.Data = data
		start := time.Now()
		_, err := entertainmentBus.Decoder(&fr)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(time.Since(start))
	}
	start := time.Now()
	r1c1 := entertainmentBus.GetValue(DisplayR1C1)
	fmt.Println(time.Since(start))
	start = time.Now()
	r1c2 := entertainmentBus.GetValue(DisplayR1C2)
	fmt.Println(time.Since(start))
	start = time.Now()
	r1c3 := entertainmentBus.GetValue(DisplayR1C3)
	fmt.Println(time.Since(start))
	start = time.Now()
	r1c4 := entertainmentBus.GetValue(DisplayR1C4)
	fmt.Println(time.Since(start))

	start = time.Now()
	row1Display := utils.ComaSeperatedDecimalsToAscii(
		r1c1.CanValueDef.Value.(string) + "," +
			r1c2.CanValueDef.Value.(string) + "," +
			r1c3.CanValueDef.Value.(string) + "," +
			r1c4.CanValueDef.Value.(string))
	fmt.Println(time.Since(start))

	if row1Display != "No Source   " {
		t.Error("row 1 display")
	}
	fmt.Println(row1Display)

	// DateTime
	dateData := [8]byte{0x46, 0x01, 0x17, 0x0a, 0x5d, 0x12, 0x27, 0xff}
	_, err := entertainmentBus.Decoder(&can.Frame{
		ArbitrationID: uint32(EntertainmentCANDate),
		DLC:           8,
		Data:          dateData,
	})
	if err != nil {
		t.Error(err)
	}
	dt := entertainmentBus.GetValue(DateTime)
	if dt.CanValueDef.Value != "23-10-11T20:18:39" {
		t.Error("Midspeed DateTime")
	}
	fmt.Println(dt.CanValueDef.Value)

	// // ClimaData
	// acData := []byte{0x23, 0xe0, 0x50, 0x00, 0x37, 0x20, 0x26, 0x02}
}

func printCanValueInfos(t *testing.T, canValue *CanValueMap) {
	t.Logf("%s is %v%s", canValue.CanValueDef.Name, canValue.CanValueDef.Value, canValue.CanValueDef.Unit)
}
