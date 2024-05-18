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

type CanVars string

const (
	ACTemperature      CanVars = "AC Temperature"  // tested
	ACMode             CanVars = "AC Mode"         // tested
	ACFanSpeed         CanVars = "AC Fan Speed"    // problem with Auto
	BatteryVoltage     CanVars = "Battery Voltage" // tested
	BusWakeup          CanVars = "CAN-Bus Wakeup"  // tested
	BreakState         CanVars = "Break State"     // tested
	DateTime           CanVars = "Date"            // tested
	EngineSpeedRPM     CanVars = "Engine RPM"      // tested
	FullInjection      CanVars = "Full Injection"  // partialy tested
	FullInjectionMid   CanVars = "Full Injection Mid"
	FullLevel          CanVars = "Full Level" // tested
	FullLevelMid       CanVars = "Full Level Mid"
	LedBrightness      CanVars = "Led Brightness"     // tested
	Milage             CanVars = "Milage"             // tested
	OutdoorTemperature CanVars = "Output Temperature" // tested
	VehicleSpeed       CanVars = "Speed"              // tested
	VehicleSpeedMid    CanVars = "Speed Mid"
	WeelKey            CanVars = "Weel Remote Key"        // tested
	DisplayR1C1        CanVars = "Display Row 1 Column 1" // tested
	DisplayR1C2        CanVars = "Display Row 1 Column 2" // tested
	DisplayR1C3        CanVars = "Display Row 1 Column 3" // tested
	DisplayR1C4        CanVars = "Display Row 1 Column 4" // tested
	LightSwitch        CanVars = "Light Switch"           // tested
	LightLeveler       CanVars = "Light Leveler"          // tested
	LightBack          CanVars = "Light Back"             // tested
	DoorState          CanVars = "Door State"             // tested
	EngineRunningState CanVars = "Engine State"           // tested
	TurnLights         CanVars = "Turn Lights"            // dont work
	CoolantTemperature CanVars = "Coolant Temperature"    // tested
	TPMS               CanVars = "Tire Pressure Monitoring System"
	CruseControl       CanVars = "Cruse Control" // tested
	SystemTime         CanVars = "System Time"
	DisplayTemperature CanVars = "Display Temperature"
	SensorTemperature  CanVars = "Outdoor Sensor Temperature" // tested
	LeftTravelRange    CanVars = "Range"                      // tested
	RangeWarning       CanVars = "Range Warning"              // tested
	TraveledDistance   CanVars = "Traveled Distance"          // tested
	Distance           CanVars = "Distance"
	DoorLook           CanVars = "Door Lock"
)

type CanValueDef struct {
	Calculation      string
	FormatSeperators []string // if using ; in calc
	Condition        string
	Unit             string      `json:"unit"`
	Name             CanVars     `json:"name"`
	Value            interface{} `json:"value"`
}

type CanValueMap struct {
	CanValueDef   CanValueDef
	ArbitrationID uint32
	TriggerEvent  bool
	OriginalData  []byte
}

type Cancoder struct {
	Map    []CanValueMap
	Device string
}

type Cancoders []Cancoder
