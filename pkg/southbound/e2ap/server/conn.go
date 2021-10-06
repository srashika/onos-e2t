// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/onosproject/onos-lib-go/pkg/uri"

	"sync"
	"time"

	e2smtypes "github.com/onosproject/onos-api/go/onos/e2t/e2sm"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	subscriptionv1beta1 "github.com/onosproject/onos-e2t/pkg/broker/subscription/v1beta1"
	"github.com/onosproject/onos-e2t/pkg/southbound/e2ap/types"

	"github.com/onosproject/onos-e2t/pkg/broker/subscription"

	e2appducontents "github.com/onosproject/onos-e2t/api/e2ap/v2/e2ap-pdu-contents"
	e2 "github.com/onosproject/onos-e2t/pkg/protocols/e2ap"
)

func NewE2NodeConn(nodeID topoapi.ID, plmnID string, nodeIdentity *types.E2NodeIdentity, conn e2.ServerConn,
	streams subscription.Broker, streamsv1beta1 subscriptionv1beta1.Broker,
	serviceModels map[string]*topoapi.ServiceModelInfo, ranFunctions map[e2smtypes.OID]RANFunction, e2Cells []*topoapi.E2Cell, now time.Time) *E2APConn {

	connID := ConnID(uri.NewURI(
		uri.WithScheme("uuid"),
		uri.WithOpaque(uuid.New().String())).String())

	return &E2APConn{
		ServerConn:     conn,
		ID:             connID,
		E2NodeID:       nodeID,
		PlmnID:         plmnID,
		NodeID:         string(nodeID),
		NodeType:       nodeIdentity.NodeType,
		TimeAlive:      now,
		streams:        streams,
		streamsv1beta1: streamsv1beta1,
		ServiceModels:  serviceModels,
		RANFunctions:   ranFunctions,
		E2Cells:        e2Cells,
	}
}

// RANFunction RAN function information
type RANFunction struct {
	OID e2smtypes.OID
	ID  types.RanFunctionID
	// protobuf encoded description
	Description []byte
}

type ConnID string

type E2APConn struct {
	e2.ServerConn
	ID             ConnID
	E2NodeID       topoapi.ID
	NodeID         string
	NodeType       types.E2NodeType
	PlmnID         string
	TimeAlive      time.Time
	streams        subscription.Broker
	streamsv1beta1 subscriptionv1beta1.Broker
	ServiceModels  map[string]*topoapi.ServiceModelInfo
	RANFunctions   map[e2smtypes.OID]RANFunction
	E2Cells        []*topoapi.E2Cell
	mu             sync.RWMutex
}

func (c *E2APConn) GetRANFunctions() []RANFunction {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ranFunctions := make([]RANFunction, 0, len(c.RANFunctions))
	for _, ranFunction := range c.RANFunctions {
		ranFunctions = append(ranFunctions, ranFunction)
	}
	return ranFunctions
}

func (c *E2APConn) GetRANFunction(oid e2smtypes.OID) (RANFunction, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ranFunction, ok := c.RANFunctions[oid]
	return ranFunction, ok
}

func (c *E2APConn) ricIndication(ctx context.Context, request *e2appducontents.Ricindication) error {
	log.Debugf("Received RICIndication %+v", request)
	streamID := subscriptionv1beta1.StreamID(request.ProtocolIes.E2ApProtocolIes29.Value.RicRequestorId)
	stream, ok := c.streamsv1beta1.GetWriter(streamID)
	if !ok {
		deprecatedStreamID := subscription.StreamID(request.ProtocolIes.E2ApProtocolIes29.Value.RicRequestorId)
		deprecatedStream, err := c.streams.GetStream(deprecatedStreamID)
		if err != nil {
			return err
		}
		return deprecatedStream.Send(request)
	}
	return stream.Send(request)
}