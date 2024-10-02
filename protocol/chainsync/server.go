// Copyright 2024 Blink Labs Software
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chainsync

import (
	"fmt"

	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/blinklabs-io/gouroboros/protocol"
	"github.com/blinklabs-io/gouroboros/protocol/common"
)

// Server implements the ChainSync server
type Server struct {
	*protocol.Protocol
	config          *Config
	callbackContext CallbackContext
	protoOptions    protocol.ProtocolOptions
	stateContext    any
}

// NewServer returns a new ChainSync server object
func NewServer(
	stateContext interface{},
	protoOptions protocol.ProtocolOptions,
	cfg *Config,
) *Server {
	s := &Server{
		config: cfg,
		// Save these for re-use later
		protoOptions: protoOptions,
		stateContext: stateContext,
	}
	s.callbackContext = CallbackContext{
		Server:       s,
		ConnectionId: protoOptions.ConnectionId,
	}
	s.initProtocol()
	return s
}

func (s *Server) initProtocol() {
	// Use node-to-client protocol ID
	ProtocolId := ProtocolIdNtC
	msgFromCborFunc := NewMsgFromCborNtC
	if s.protoOptions.Mode == protocol.ProtocolModeNodeToNode {
		// Use node-to-node protocol ID
		ProtocolId = ProtocolIdNtN
		msgFromCborFunc = NewMsgFromCborNtN
	}
	protoConfig := protocol.ProtocolConfig{
		Name:                ProtocolName,
		ProtocolId:          ProtocolId,
		Muxer:               s.protoOptions.Muxer,
		Logger:              s.protoOptions.Logger,
		ErrorChan:           s.protoOptions.ErrorChan,
		Mode:                s.protoOptions.Mode,
		Role:                protocol.ProtocolRoleServer,
		MessageHandlerFunc:  s.messageHandler,
		MessageFromCborFunc: msgFromCborFunc,
		StateMap:            StateMap,
		StateContext:        s.stateContext,
		InitialState:        stateIdle,
	}
	s.Protocol = protocol.New(protoConfig)
}

func (s *Server) RollBackward(point common.Point, tip Tip) error {
	s.Protocol.Logger().
		Debug(fmt.Sprintf("server called %s RollBackward(point: %+v, tip: %+v)", ProtocolName, point, tip))
	msg := NewMsgRollBackward(point, tip)
	return s.SendMessage(msg)
}

func (s *Server) AwaitReply() error {
	s.Protocol.Logger().
		Debug(fmt.Sprintf("server called %s AwaitReply()", ProtocolName))
	msg := NewMsgAwaitReply()
	return s.SendMessage(msg)
}

func (s *Server) RollForward(blockType uint, blockData []byte, tip Tip) error {
	s.Protocol.Logger().
		Debug(fmt.Sprintf("server called %s Rollforward(blockType: %+v, blockData: %x, tip: %+v)", ProtocolName, blockType, blockData, tip))
	if s.Mode() == protocol.ProtocolModeNodeToNode {
		eraId := ledger.BlockToBlockHeaderTypeMap[blockType]
		msg := NewMsgRollForwardNtN(
			eraId,
			0,
			blockData,
			tip,
		)
		return s.SendMessage(msg)
	} else {
		msg := NewMsgRollForwardNtC(
			blockType,
			blockData,
			tip,
		)
		return s.SendMessage(msg)
	}
}

func (s *Server) messageHandler(msg protocol.Message) error {
	s.Protocol.Logger().
		Debug(fmt.Sprintf("handling server message for %s", ProtocolName))
	var err error
	switch msg.Type() {
	case MessageTypeRequestNext:
		err = s.handleRequestNext()
	case MessageTypeFindIntersect:
		err = s.handleFindIntersect(msg)
	case MessageTypeDone:
		err = s.handleDone()
	default:
		err = fmt.Errorf(
			"%s: received unexpected message type %d",
			ProtocolName,
			msg.Type(),
		)
	}
	return err
}

func (s *Server) handleRequestNext() error {
	// TODO: figure out why this one log message causes a panic (and only this one)
	//   during tests
	// s.Protocol.Logger().
	// 	Debug(fmt.Sprintf("handling server request next for %s", ProtocolName))
	if s.config == nil || s.config.RequestNextFunc == nil {
		return fmt.Errorf(
			"received chain-sync RequestNext message but no callback function is defined",
		)
	}
	return s.config.RequestNextFunc(s.callbackContext)
}

func (s *Server) handleFindIntersect(msg protocol.Message) error {
	s.Protocol.Logger().
		Debug(fmt.Sprintf("handling server find intersect for %s", ProtocolName))
	if s.config == nil || s.config.FindIntersectFunc == nil {
		return fmt.Errorf(
			"received chain-sync FindIntersect message but no callback function is defined",
		)
	}
	msgFindIntersect := msg.(*MsgFindIntersect)
	point, tip, err := s.config.FindIntersectFunc(
		s.callbackContext,
		msgFindIntersect.Points,
	)
	if err != nil {
		if err == IntersectNotFoundError {
			msgResp := NewMsgIntersectNotFound(tip)
			if err := s.SendMessage(msgResp); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	msgResp := NewMsgIntersectFound(point, tip)
	if err := s.SendMessage(msgResp); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleDone() error {
	s.Protocol.Logger().
		Debug(fmt.Sprintf("handling server done for %s", ProtocolName))
	// Restart protocol
	s.Protocol.Stop()
	s.initProtocol()
	s.Protocol.Start()
	return nil
}
