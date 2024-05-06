/**
 * Copyright © 2024, Staufi Tech - Switzerland
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

package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ChrIgiSta/go-can-coder/cancoder"
	"github.com/ChrIgiSta/go-easy-websockets/websocket"
	log "github.com/ChrIgiSta/go-utils/logger"
)

type WsMsg struct {
	Device string               `json:"device"`
	Msg    cancoder.CanValueDef `json:"message"`
}

type CanFrame struct {
	ArbitrationID uint32 `json:"arbitrationID"`
	DLC           uint8  `json:"DLC"`
	Data          []byte `json:"data"` // base64
}

func NewCanFrame(arbitrationID uint32, DLC uint8, data []byte) *CanFrame {
	return &CanFrame{
		ArbitrationID: arbitrationID,
		DLC:           DLC,
		Data:          []byte(base64.RawStdEncoding.EncodeToString(data)),
	}
}

func (f *CanFrame) GetData() (raw []byte, err error) {
	raw, err = base64.RawStdEncoding.DecodeString(string(f.Data))
	return
}

type ReceiveCallback func(msg any)

type WsServer struct {
	server *websocket.Server
	rxClbk ReceiveCallback
}

func NewWsServer(port uint16, path string, receiveCallback ReceiveCallback, apiKey string, certificate []byte, privateKey []byte) *WsServer {
	ws := WsServer{
		rxClbk: receiveCallback,
	}

	portS := fmt.Sprintf(":%d/%s", port, path)

	ws.server = websocket.NewServer(":"+portS, &ws)
	ws.server.SetupTls(certificate, privateKey)
	ws.server.SetAuthHeader(websocket.NewAuthHeader("X-Api-Key", apiKey, websocket.HashAlgoSHA256))

	return &ws
}

func (s *WsServer) Serve() error {
	return s.server.ListenAndServe()
}

func (s *WsServer) SendEncoded(msg WsMsg) error {

	return s.Send(msg)
}

func (s *WsServer) SendCanFrame(frame CanFrame) error {

	return s.Send(frame)
}

func (s *WsServer) Send(msg any) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	go s.server.Broadcast(&websocket.Message{
		MessageType: 1,
		Data:        payload,
	})

	return nil
}

func (t *WsServer) OnReceive(msg websocket.Message) {
	_ = log.Fine("Websocket", "onReceive: %v", msg)

	var wsMsg any

	if strings.ContainsAny(string(msg.Data), "message") {
		wsMsg = WsMsg{}
	} else if strings.ContainsAny(string(msg.Data), "arbitrationID") {
		wsMsg = CanFrame{}
	}

	err := json.Unmarshal(msg.Data, &wsMsg)
	if err != nil {
		log.Warn("websocket", "rx msg error: %v", err)
	} else if t.rxClbk != nil {
		t.rxClbk(wsMsg)
	}
}

func (t *WsServer) OnDisconnect(id int) {
	_ = log.Fine("Websocket", "onDisconnect: %v", id)
}

func (t *WsServer) OnConnect(id int) {
	_ = log.Fine("Websocket", "onConnect: %v", id)
}

func (t *WsServer) OnFailure(exited bool, err error) {
	_ = log.Fine("Websocket", "onFailure: %v", err)

}
