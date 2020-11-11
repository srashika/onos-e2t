// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package pdudecoder

import (
	"github.com/onosproject/onos-e2t/pkg/southbound/e2ap/asn1cgo"
	"gotest.tools/assert"
	"io/ioutil"
	"testing"
)

func Test_DecodeRicSubscriptionResponsePdu(t *testing.T) {
	ricSubscriptionResponseXer, err := ioutil.ReadFile("../test/RICsubscriptionResponse.xml")
	assert.NilError(t, err, "Unexpected error when loading file")
	e2apPdu, err := asn1cgo.XerDecodeE2apPdu(ricSubscriptionResponseXer)
	assert.NilError(t, err)

	rfID, rrID, ricActionIDs, _, err := DecodeRicSubscriptionResponsePdu(e2apPdu)
	assert.NilError(t, err)
	assert.Assert(t, rfID != nil)
	assert.Equal(t, 20, int(*rfID))

	assert.Assert(t, rrID != nil)
	assert.Equal(t, 22, int(rrID.RequestorID))
	assert.Equal(t, 6, int(rrID.InstanceID))

	assert.Equal(t, 1, len(ricActionIDs))
	assert.Equal(t, 5, int(ricActionIDs[0]))
}
