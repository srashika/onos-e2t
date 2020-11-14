// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package asn1cgo

//#cgo CFLAGS: -I. -D_DEFAULT_SOURCE -DASN_DISABLE_OER_SUPPORT
//#cgo LDFLAGS: -lm
//#include <stdio.h>
//#include <stdlib.h>
//#include <assert.h>
//#include "RICindicationType.h"
import "C"
import (
	"encoding/binary"
	"fmt"
	"github.com/onosproject/onos-e2t/api/e2ap/v1beta1/e2apies"
)

func newRicIndicationType(rit *e2apies.RicindicationType) (*C.RICindicationType_t, error) {
	var ret C.RICindicationType_t
	switch *rit {
	case e2apies.RicindicationType_RICINDICATION_TYPE_REPORT:
		ret = C.RICindicationType_report
	case e2apies.RicindicationType_RICINDICATION_TYPE_INSERT:
		ret = C.RICindicationType_insert
	default:
		return nil, fmt.Errorf("unexpected RicIndicationType %v", rit)
	}
	return &ret, nil
}

func decodeRicIndicationTypeBytes(bytes []byte) e2apies.RicindicationType {
	raIDC := C.long(binary.LittleEndian.Uint64(bytes[:8]))
	return decodeRicIndicationType(&raIDC)
}

func decodeRicIndicationType(raIDC *C.RICindicationType_t) e2apies.RicindicationType {
	return e2apies.RicindicationType(*raIDC)
}