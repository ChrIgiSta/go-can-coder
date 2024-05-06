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

package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func InterfaceToString(i any) string {
	return fmt.Sprintf("%v", i)
}

func ReplaceHexWithDecimal(text string) string {
	hexPattern := `0x[0-9a-fA-F]+`

	re := regexp.MustCompile(hexPattern)

	hexMatches := re.FindAllString(text, -1)

	for _, hexMatch := range hexMatches {
		decimalValue, err := strconv.ParseInt(hexMatch[2:], 16, 64)
		if err == nil {
			text = strings.Replace(text, hexMatch, strconv.FormatInt(decimalValue, 10), -1)
		}
	}

	return text
}

func ComaSeperatedDecimalsToAscii(in string) string {
	var intArray []byte = make([]byte, 0)

	split := strings.Split(in, ",")

	for _, element := range split {
		num, _ := strconv.Atoi(element)
		intArray = append(intArray, byte(num))
	}
	return string(intArray[:len(intArray)-2])
}

func CanTimeStringToTime(timeStr string) (time.Time, error) {

	var t time.Time

	split := strings.Split(timeStr, "T")
	if len(split) != 2 {
		return t, errors.New("unknown format")
	}
	dataSplit := strings.Split(split[0], "-")
	if len(dataSplit) != 3 {
		return t, errors.New("date format")
	}
	timeSplit := strings.Split(split[1], ":")
	if len(timeSplit) != 3 {
		return t, errors.New("time format")
	}

	layout := "2006-01-02T15:04:05Z"

	year, err := strconv.Atoi(dataSplit[0])
	if err != nil {
		return t, errors.New("parse year")
	}
	month, err := strconv.Atoi(dataSplit[1])
	if err != nil {
		return t, errors.New("parse month")
	}
	day, err := strconv.Atoi(dataSplit[2])
	if err != nil {
		return t, errors.New("parse day")
	}
	hours, err := strconv.Atoi(timeSplit[0])
	if err != nil {
		return t, errors.New("parse hour")
	}
	mins, err := strconv.Atoi(timeSplit[1])
	if err != nil {
		return t, errors.New("parse minutes")
	}
	secs, err := strconv.Atoi(timeSplit[2])
	if err != nil {
		return t, errors.New("parse seconds")
	}

	return time.Parse(layout, fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ", year+2000, month, day, hours, mins, secs))
}
