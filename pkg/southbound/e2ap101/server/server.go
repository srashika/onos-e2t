// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package server

import (
	"context"
	"fmt"
	"github.com/onosproject/onos-e2t/pkg/broker/subscription"
	"strconv"

	e2smtypes "github.com/onosproject/onos-api/go/onos/e2t/e2sm"

	e2appducontents "github.com/onosproject/onos-e2t/api/e2ap/v1beta2/e2ap-pdu-contents"
	"github.com/onosproject/onos-e2t/pkg/modelregistry"
	e2 "github.com/onosproject/onos-e2t/pkg/protocols/e2ap101"
	"github.com/onosproject/onos-e2t/pkg/southbound/e2ap101/pdubuilder"
	"github.com/onosproject/onos-e2t/pkg/southbound/e2ap101/pdudecoder"
	"github.com/onosproject/onos-e2t/pkg/southbound/e2ap101/types"
)

// TODO: Change the RIC ID to something appropriate
var ricID = types.RicIdentifier{
	RicIdentifierValue: 0xABCDE,
	RicIdentifierLen:   20,
}

func NewE2Server(channels ChannelManager, subs subscription.Broker, modelRegistry modelregistry.ModelRegistry) *E2Server {
	return &E2Server{
		server:        e2.NewServer(),
		channels:      channels,
		subs:          subs,
		modelRegistry: modelRegistry,
	}
}

type E2Server struct {
	server        *e2.Server
	channels      ChannelManager
	subs          subscription.Broker
	modelRegistry modelregistry.ModelRegistry
}

func (s *E2Server) Serve() error {
	return s.server.Serve(func(channel e2.ServerChannel) e2.ServerInterface {
		return &E2ChannelServer{
			serverChannel: channel,
			manager:       s.channels,
			subs:          s.subs,
			modelRegistry: s.modelRegistry,
		}
	})
}

func (s *E2Server) Stop() error {
	return s.server.Stop()
}

type E2ChannelServer struct {
	manager       ChannelManager
	subs          subscription.Broker
	serverChannel e2.ServerChannel
	e2Channel     *E2Channel
	modelRegistry modelregistry.ModelRegistry
}

// uint24ToUint32 converts uint24 uint32
func uint24ToUint32(val []byte) uint32 {
	r := uint32(0)
	for i := uint32(0); i < 3; i++ {
		r |= uint32(val[i]) << (8 * i)
	}
	return r
}

func (e *E2ChannelServer) E2Setup(ctx context.Context, request *e2appducontents.E2SetupRequest) (*e2appducontents.E2SetupResponse, *e2appducontents.E2SetupFailure, error) {
	nodeID, ranFuncs, err := pdudecoder.DecodeE2SetupRequest(request)
	if err != nil {
		return nil, nil, err
	}
	channelID := ChannelID(fmt.Sprintf("%x:%d", string(nodeID.NodeIdentifier), nodeID.NodeType))
	rawPlmnid := []byte{nodeID.Plmn[0], nodeID.Plmn[1], nodeID.Plmn[2]}
	plmnID := strconv.FormatUint(uint64(uint24ToUint32(rawPlmnid)), 10)

	rfAccepted := make(types.RanFunctionRevisions)
	rfRejected := make(types.RanFunctionCauses)
	ranFuncIDs := make(map[e2smtypes.OID]types.RanFunctionID)
	plugins := e.modelRegistry.GetPlugins()
	for id, ranFunc := range *ranFuncs {
		rfAccepted[id] = ranFunc.Revision
		for smOid, sm := range plugins {
			if string(smOid) == string(ranFunc.OID) {
				log.Infof("Decoding RanFunction description for OID: %s", ranFunc.OID)
				names, triggers, reports, err := sm.DecodeRanFunctionDescription(ranFunc.Description)
				if err != nil {
					log.Warn(err)
					continue
				}

				log.Infof("RanFunctionDescription ShortName: %s, Desc: %s,"+
					"Instance: %d, Oid: %s. #Triggers: %d. #Reports: %d",
					names.RanFunctionShortName,
					names.RanFunctionDescription,
					names.RanFunctionInstance,
					names.RanFunctionE2SmOid,
					len(*triggers), len(*reports))
				oid := names.RanFunctionE2SmOid
				ranFuncIDs[oid] = id

			}
		}
	}

	e.e2Channel = NewE2Channel(channelID, plmnID, e.serverChannel, e.subs, ranFuncIDs)
	e.manager.Open(channelID, e.e2Channel)

	// Create an E2 setup response
	response, err := pdubuilder.NewE2SetupResponse(nodeID.Plmn, ricID, rfAccepted, rfRejected)
	if err != nil {
		return nil, nil, err
	}
	return response, nil, nil
}

func (e *E2ChannelServer) RICIndication(ctx context.Context, request *e2appducontents.Ricindication) error {
	return e.e2Channel.ricIndication(ctx, request)
}
