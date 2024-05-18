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
	"encoding/hex"
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	log "github.com/ChrIgiSta/go-utils/logger"

	"github.com/ChrIgiSta/go-can-coder/utils"

	"github.com/Knetic/govaluate"
	"github.com/angelodlfrtr/go-can"
)

const EventChannelBufferSize = 100

type Decoder struct {
	frameBuffer []can.Frame
	valueMaps   []CanValueMap

	eventChannels []chan<- CanValueMap
}

func NewCanCoder(valueMaps []CanValueMap) *Decoder {
	return &Decoder{
		frameBuffer: []can.Frame{},

		valueMaps: valueMaps,
	}
}

func (d *Decoder) GetEventChannel() <-chan CanValueMap {
	event := make(chan CanValueMap, EventChannelBufferSize)
	d.eventChannels = append(d.eventChannels, event)

	return event
}

func (d *Decoder) GetValue(name CanVars) *CanValueMap {
	for _, val := range d.valueMaps {
		if val.CanValueDef.Name == name {
			return &val
		}
	}
	return nil
}

func (d *Decoder) Decoder(frame *can.Frame) (values []*CanValueMap, err error) {

	for i, mapping := range d.valueMaps {
		if mapping.ArbitrationID == frame.ArbitrationID {
			val, err := d.processFrame(&d.valueMaps[i], frame)
			if err != nil {
				return values, err
			} else if val != nil {
				values = append(values, val)
				continue
			}
		}
	}

	return values, nil
}

func (c *Decoder) Encode(value *CanValueMap) (frame *can.Frame, err error) {

	frame = &can.Frame{
		ArbitrationID: value.ArbitrationID,
	}

	split := strings.Split(value.CanValueDef.Calculation, ";")

	for i, hVal := range split {
		b, err := hex.DecodeString(hVal)
		if err != nil {
			return nil, err
		}
		if len(b) != 1 {
			return nil, fmt.Errorf("unexpected len in calc to encode: %v", b)
		}
		frame.Data[i] = b[0]
		frame.DLC = uint8(i + 1)
	}

	return
}

func (d *Decoder) PushFrame(frame *can.Frame) error {
	if frame == nil {
		return errors.New("frame <nil>")
	}

	found := d.findFrameByArbitrationId(frame.ArbitrationID)
	if found != nil {
		found.ArbitrationID = frame.ArbitrationID
		found.DLC = frame.DLC
		found.Data = frame.Data
	} else {
		d.frameBuffer = append(d.frameBuffer, *frame)
	}

	_, err := d.Decoder(frame)
	return err
}

func (d *Decoder) findFrameByArbitrationId(arbitrationID uint32) *can.Frame {
	for i, f := range d.frameBuffer {
		if f.ArbitrationID == arbitrationID {
			return &d.frameBuffer[i]
		}
	}
	return nil
}

func (d *Decoder) processFrame(mapping *CanValueMap, frame *can.Frame) (*CanValueMap, error) {
	condition, err := d.substituteVars(mapping.CanValueDef.Condition, frame)
	if err != nil {
		return nil, err
	}

	fileSet := token.NewFileSet()
	tav, err := types.Eval(fileSet, nil, token.NoPos, condition)
	if err != nil {
		return nil, err
	}
	if tav.Value.String() == "true" {
		equation, err := d.substituteVars(mapping.CanValueDef.Calculation, frame)
		if err != nil {
			return nil, err
		}

		mapping.OriginalData = frame.Data[0:frame.DLC]

		formatedString := false
		splittedEquation := strings.Split(equation, ";")
		if len(splittedEquation) != 1 {
			formatedString = true
		}
		if !formatedString {
			expression, err := govaluate.NewEvaluableExpression(equation)
			if err != nil {
				return nil, err
			}
			result, err := expression.Evaluate(nil)
			if err != nil {
				return nil, err
			}
			mapping.CanValueDef.Value = result
			if mapping.TriggerEvent {
				d.processEvent(mapping)
			}
			return mapping, nil
		} else {
			output := ""
			for sIndx, split := range splittedEquation {
				expression, err := govaluate.NewEvaluableExpression(split)
				if err != nil {
					return nil, err
				}
				result, err := expression.Evaluate(nil)
				if err != nil {
					return nil, err
				}
				output += utils.InterfaceToString(result)
				if len(mapping.CanValueDef.FormatSeperators) > sIndx {
					output += mapping.CanValueDef.FormatSeperators[sIndx]
				}
			}
			mapping.CanValueDef.Value = output
			if mapping.TriggerEvent {
				d.processEvent(mapping)
			}
			return mapping, nil
		}
	}

	return nil, nil
}

func (d *Decoder) processEvent(canVal *CanValueMap) {
	if canVal != nil {
		for _, evtCh := range d.eventChannels {
			select {
			case evtCh <- *canVal:
			default:
				log.Warn("decoder", "event channel full. you need to process faster ;)")
			}
		}
	} else {
		log.Error("decoder", "event: value <nil>")
	}
}

func (d *Decoder) substituteVars(query string, frame *can.Frame) (string, error) {
	subst := query
	for i := 0; i < 8; i++ {
		subst = strings.ReplaceAll(subst, "${"+strconv.Itoa(i)+"}", strconv.Itoa(int(frame.Data[i])))
	}
	return utils.ReplaceHexWithDecimal(subst), nil
}
