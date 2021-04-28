/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

package control

/*
#include <e2sm/wrapper.h>
#cgo LDFLAGS: -le2smwrapper -lm
#cgo CFLAGS: -I/usr/local/include/e2sm
*/
import "C"

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"unsafe"
)

type E2sm struct {
}

func (c *E2sm) SetEventTriggerDefinition(buffer []byte, eventTriggerCount int, RTPeriods int64) (newBuffer []byte, err error) {
	cptr := unsafe.Pointer(&buffer[0])
	periods := RTPeriods
	size := C.e2sm_encode_ric_event_trigger_definition(cptr, C.size_t(len(buffer)), C.size_t(eventTriggerCount), (C.long)(periods))
	if size < 0 {
		return make([]byte, 0), errors.New("e2sm wrapper is unable to set EventTriggerDefinition due to wrong or invalid input")
	}
	newBuffer = C.GoBytes(cptr, (C.int(size)+7)/8)
	return
}

func (c *E2sm) SetActionDefinition(buffer []byte, ricStyleType int64) (newBuffer []byte, err error) {
	cptr := unsafe.Pointer(&buffer[0])
	size := C.e2sm_encode_ric_action_definition(cptr, C.size_t(len(buffer)), C.long(ricStyleType))
	if size < 0 {
		return make([]byte, 0), errors.New("e2sm wrapper is unable to set ActionDefinition due to wrong or invalid input")
	}
	newBuffer = C.GoBytes(cptr, (C.int(size)+7)/8)
	return
}

func (c *E2sm) GetIndicationHeader(buffer []byte) (indHdr *IndicationHeader, err error) {
	cptr := unsafe.Pointer(&buffer[0])
	indHdr = &IndicationHeader{}
	decodedHdr := C.e2sm_decode_ric_indication_header(cptr, C.size_t(len(buffer)))
	if decodedHdr == nil {
		return indHdr, errors.New("e2sm wrapper is unable to get IndicationHeader due to wrong or invalid input")
	}
	defer C.e2sm_free_ric_indication_header(decodedHdr)

	indHdr.IndHdrType = int32(decodedHdr.indicationHeader_formats.present)
	if indHdr.IndHdrType == 1 {
		indHdrFormat1 := &IndicationHeaderFormat1{}
		indHdrFormat1_C := *(**C.E2SM_KPM_IndicationHeader_Format1_t)(unsafe.Pointer(&decodedHdr.indicationHeader_formats.choice[0]))

		if indHdrFormat1_C.kpmNodeID != nil {
			globalKPMnodeID_C := (*C.GlobalKPMnode_ID_t)(indHdrFormat1_C.kpmNodeID)

			indHdrFormat1.GlobalKPMnodeIDType = int32(globalKPMnodeID_C.present)
			if indHdrFormat1.GlobalKPMnodeIDType == 1 {
				globalgNBID := &GlobalKPMnodegNBIDType{}
				globalgNBID_C := (*C.GlobalKPMnode_gNB_ID_t)(unsafe.Pointer(&globalKPMnodeID_C.choice[0]))

				plmnID_C := globalgNBID_C.global_gNB_ID.plmn_id
				globalgNBID.GlobalgNBID.PlmnID.Buf = C.GoBytes(unsafe.Pointer(plmnID_C.buf), C.int(plmnID_C.size))
				globalgNBID.GlobalgNBID.PlmnID.Size = int(plmnID_C.size)

				globalgNBID_gNBID_C := globalgNBID_C.global_gNB_ID.gnb_id
				globalgNBID.GlobalgNBID.GnbIDType = int(globalgNBID_gNBID_C.present)
				if globalgNBID.GlobalgNBID.GnbIDType == 1 {
					gNBID := &GNBID{}
					gNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globalgNBID_gNBID_C.choice[0]))

					gNBID.Buf = C.GoBytes(unsafe.Pointer(gNBID_C.buf), C.int(gNBID_C.size))
					gNBID.Size = int(gNBID_C.size)
					gNBID.BitsUnused = int(gNBID_C.bits_unused)

					globalgNBID.GlobalgNBID.GnbID = gNBID
				}

				if globalgNBID_C.gNB_CU_UP_ID != nil {
					globalgNBID.GnbCUUPID = &Integer{}
					globalgNBID.GnbCUUPID.Buf = C.GoBytes(unsafe.Pointer(globalgNBID_C.gNB_CU_UP_ID.buf), C.int(globalgNBID_C.gNB_CU_UP_ID.size))
					globalgNBID.GnbCUUPID.Size = int(globalgNBID_C.gNB_CU_UP_ID.size)
				}

				if globalgNBID_C.gNB_DU_ID != nil {
					globalgNBID.GnbDUID = &Integer{}
					globalgNBID.GnbDUID.Buf = C.GoBytes(unsafe.Pointer(globalgNBID_C.gNB_DU_ID.buf), C.int(globalgNBID_C.gNB_DU_ID.size))
					globalgNBID.GnbDUID.Size = int(globalgNBID_C.gNB_DU_ID.size)
				}

				indHdrFormat1.GlobalKPMnodeID = globalgNBID
			} else if indHdrFormat1.GlobalKPMnodeIDType == 2 {
				globalengNBID := &GlobalKPMnodeengNBIDType{}
				globalengNBID_C := (*C.GlobalKPMnode_en_gNB_ID_t)(unsafe.Pointer(&globalKPMnodeID_C.choice[0]))

				plmnID_C := globalengNBID_C.global_gNB_ID.pLMN_Identity
				globalengNBID.PlmnID.Buf = C.GoBytes(unsafe.Pointer(plmnID_C.buf), C.int(plmnID_C.size))
				globalengNBID.PlmnID.Size = int(plmnID_C.size)

				globalengNBID_gNBID_C := globalengNBID_C.global_gNB_ID.gNB_ID
				globalengNBID.GnbIDType = int(globalengNBID_gNBID_C.present)
				if globalengNBID.GnbIDType == 1 {
					engNBID := &ENGNBID{}
					engNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globalengNBID_gNBID_C.choice[0]))

					engNBID.Buf = C.GoBytes(unsafe.Pointer(engNBID_C.buf), C.int(engNBID_C.size))
					engNBID.Size = int(engNBID_C.size)
					engNBID.BitsUnused = int(engNBID_C.bits_unused)

					globalengNBID.GnbID = engNBID
				}

				indHdrFormat1.GlobalKPMnodeID = globalengNBID
			} else if indHdrFormat1.GlobalKPMnodeIDType == 3 {
				globalngeNBID := &GlobalKPMnodengeNBIDType{}
				globalngeNBID_C := (*C.GlobalKPMnode_ng_eNB_ID_t)(unsafe.Pointer(&globalKPMnodeID_C.choice[0]))

				plmnID_C := globalngeNBID_C.global_ng_eNB_ID.plmn_id
				globalngeNBID.PlmnID.Buf = C.GoBytes(unsafe.Pointer(plmnID_C.buf), C.int(plmnID_C.size))
				globalngeNBID.PlmnID.Size = int(plmnID_C.size)

				globalngeNBID_eNBID_C := globalngeNBID_C.global_ng_eNB_ID.enb_id
				globalngeNBID.EnbIDType = int(globalngeNBID_eNBID_C.present)
				if globalngeNBID.EnbIDType == 1 {
					ngeNBID := &NGENBID_Macro{}
					ngeNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globalngeNBID_eNBID_C.choice[0]))

					ngeNBID.Buf = C.GoBytes(unsafe.Pointer(ngeNBID_C.buf), C.int(ngeNBID_C.size))
					ngeNBID.Size = int(ngeNBID_C.size)
					ngeNBID.BitsUnused = int(ngeNBID_C.bits_unused)

					globalngeNBID.EnbID = ngeNBID
				} else if globalngeNBID.EnbIDType == 2 {
					ngeNBID := &NGENBID_ShortMacro{}
					ngeNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globalngeNBID_eNBID_C.choice[0]))

					ngeNBID.Buf = C.GoBytes(unsafe.Pointer(ngeNBID_C.buf), C.int(ngeNBID_C.size))
					ngeNBID.Size = int(ngeNBID_C.size)
					ngeNBID.BitsUnused = int(ngeNBID_C.bits_unused)

					globalngeNBID.EnbID = ngeNBID
				} else if globalngeNBID.EnbIDType == 3 {
					ngeNBID := &NGENBID_LongMacro{}
					ngeNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globalngeNBID_eNBID_C.choice[0]))

					ngeNBID.Buf = C.GoBytes(unsafe.Pointer(ngeNBID_C.buf), C.int(ngeNBID_C.size))
					ngeNBID.Size = int(ngeNBID_C.size)
					ngeNBID.BitsUnused = int(ngeNBID_C.bits_unused)

					globalngeNBID.EnbID = ngeNBID
				}

				indHdrFormat1.GlobalKPMnodeID = globalngeNBID
			} else if indHdrFormat1.GlobalKPMnodeIDType == 4 {
				globaleNBID := &GlobalKPMnodeeNBIDType{}
				globaleNBID_C := (*C.GlobalKPMnode_eNB_ID_t)(unsafe.Pointer(&globalKPMnodeID_C.choice[0]))

				plmnID_C := globaleNBID_C.global_eNB_ID.pLMN_Identity
				globaleNBID.PlmnID.Buf = C.GoBytes(unsafe.Pointer(plmnID_C.buf), C.int(plmnID_C.size))
				globaleNBID.PlmnID.Size = int(plmnID_C.size)

				globaleNBID_eNBID_C := globaleNBID_C.global_eNB_ID.eNB_ID
				globaleNBID.EnbIDType = int(globaleNBID_eNBID_C.present)
				if globaleNBID.EnbIDType == 1 {
					eNBID := &ENBID_Macro{}
					eNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globaleNBID_eNBID_C.choice[0]))

					eNBID.Buf = C.GoBytes(unsafe.Pointer(eNBID_C.buf), C.int(eNBID_C.size))
					eNBID.Size = int(eNBID_C.size)
					eNBID.BitsUnused = int(eNBID_C.bits_unused)

					globaleNBID.EnbID = eNBID
				} else if globaleNBID.EnbIDType == 2 {
					eNBID := &ENBID_Home{}
					eNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globaleNBID_eNBID_C.choice[0]))

					eNBID.Buf = C.GoBytes(unsafe.Pointer(eNBID_C.buf), C.int(eNBID_C.size))
					eNBID.Size = int(eNBID_C.size)
					eNBID.BitsUnused = int(eNBID_C.bits_unused)

					globaleNBID.EnbID = eNBID
				} else if globaleNBID.EnbIDType == 3 {
					eNBID := &ENBID_ShortMacro{}
					eNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globaleNBID_eNBID_C.choice[0]))

					eNBID.Buf = C.GoBytes(unsafe.Pointer(eNBID_C.buf), C.int(eNBID_C.size))
					eNBID.Size = int(eNBID_C.size)
					eNBID.BitsUnused = int(eNBID_C.bits_unused)

					globaleNBID.EnbID = eNBID
				} else if globaleNBID.EnbIDType == 4 {
					eNBID := &ENBID_LongMacro{}
					eNBID_C := (*C.BIT_STRING_t)(unsafe.Pointer(&globaleNBID_eNBID_C.choice[0]))

					eNBID.Buf = C.GoBytes(unsafe.Pointer(eNBID_C.buf), C.int(eNBID_C.size))
					eNBID.Size = int(eNBID_C.size)
					eNBID.BitsUnused = int(eNBID_C.bits_unused)

					globaleNBID.EnbID = eNBID
				}

				indHdrFormat1.GlobalKPMnodeID = globaleNBID
			}
		} else {
			indHdrFormat1.GlobalKPMnodeIDType = 0
		}

		if indHdrFormat1_C.fileFormatversion != nil {
			indHdrFormat1.FileFormatVersion = &PrintableString{}

			indHdrFormat1.FileFormatVersion.Buf = C.GoBytes(unsafe.Pointer(indHdrFormat1_C.fileFormatversion.buf), C.int(indHdrFormat1_C.fileFormatversion.size))
			indHdrFormat1.FileFormatVersion.Size = int(indHdrFormat1_C.fileFormatversion.size)
		}

		if indHdrFormat1_C.senderName != nil {
			indHdrFormat1.SenderName = &PrintableString{}

			indHdrFormat1.SenderName.Buf = C.GoBytes(unsafe.Pointer(indHdrFormat1_C.senderName.buf), C.int(indHdrFormat1_C.senderName.size))
			indHdrFormat1.SenderName.Size = int(indHdrFormat1_C.senderName.size)
		}

		if indHdrFormat1_C.senderType != nil {
			indHdrFormat1.SenderType = &PrintableString{}

			indHdrFormat1.SenderType.Buf = C.GoBytes(unsafe.Pointer(indHdrFormat1_C.senderType.buf), C.int(indHdrFormat1_C.senderType.size))
			indHdrFormat1.SenderType.Size = int(indHdrFormat1_C.senderType.size)
		}

		if indHdrFormat1_C.vendorName != nil {
			indHdrFormat1.VendorName = &PrintableString{}

			indHdrFormat1.VendorName.Buf = C.GoBytes(unsafe.Pointer(indHdrFormat1_C.vendorName.buf), C.int(indHdrFormat1_C.vendorName.size))
			indHdrFormat1.VendorName.Size = int(indHdrFormat1_C.vendorName.size)
		}

		if indHdrFormat1_C.colletStartTime.size > 0 {
			indHdrFormat1.ColletStartTime = &OctetString{}

			indHdrFormat1.ColletStartTime.Buf = C.GoBytes(unsafe.Pointer(indHdrFormat1_C.colletStartTime.buf), C.int(indHdrFormat1_C.colletStartTime.size))
			indHdrFormat1.ColletStartTime.Size = int(indHdrFormat1_C.colletStartTime.size)
		}

		indHdr.IndHdr = indHdrFormat1
	} else {
		return indHdr, errors.New("Unknown RIC Indication Header type")
	}

	return
}

func (c *E2sm) GetIndicationMessage(buffer []byte) (indMsg *IndicationMessage, err error) {
	cptr := unsafe.Pointer(&buffer[0])
	indMsg = &IndicationMessage{}
	decodedMsg := C.e2sm_decode_ric_indication_message(cptr, C.size_t(len(buffer)))
	if decodedMsg == nil {
		return indMsg, errors.New("e2sm wrapper is unable to get IndicationMessage due to wrong or invalid input")
	}
	defer C.e2sm_free_ric_indication_message(decodedMsg)

	indMsg.IndMsgType = int32(decodedMsg.indicationMessage_formats.present)

	if indMsg.IndMsgType == 1 {
		indMsgFormat1 := &IndicationMessageFormat1{}
		indMsgFormat1_C := *(**C.E2SM_KPM_IndicationMessage_Format1_t)(unsafe.Pointer(&decodedMsg.indicationMessage_formats.choice[0]))

		if indMsgFormat1_C.granulPeriod != nil {
			indMsgFormat1.GranulPeriod = int64(*indMsgFormat1_C.granulPeriod)
		} else {
			indMsgFormat1.GranulPeriod = -1
		}

		if indMsgFormat1_C.subscriptID.size > 0 {
			indMsgFormat1.SubscriptID = &Integer{}

			indMsgFormat1.SubscriptID.Buf = C.GoBytes(unsafe.Pointer(indMsgFormat1_C.subscriptID.buf), C.int(indMsgFormat1_C.subscriptID.size))
			indMsgFormat1.SubscriptID.Size = int(indMsgFormat1_C.subscriptID.size)
		}

		if indMsgFormat1_C.cellObjID != nil {
			indMsgFormat1.CellObjID = &PrintableString{}

			indMsgFormat1.CellObjID.Buf = C.GoBytes(unsafe.Pointer(indMsgFormat1_C.cellObjID.buf), C.int(indMsgFormat1_C.cellObjID.size))
			indMsgFormat1.CellObjID.Size = int(indMsgFormat1_C.cellObjID.size)
		}

		if indMsgFormat1_C.measInfoList != nil {
			indMsgFormat1.MeasInfoCount = int(indMsgFormat1_C.measInfoList.list.count)
			MeasInfoList := []MeasInfoItem{}

			for i := 0; i < indMsgFormat1.MeasInfoCount; i++ {
				var sizeof_MeasurementInfoItem_t *C.MeasurementInfoItem_t
				MeasInfoItem_C := *(**C.MeasurementInfoItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat1_C.measInfoList.list.array)) + (uintptr)(i)*unsafe.Sizeof(sizeof_MeasurementInfoItem_t)))
				MeasInfoItem := MeasInfoList[i]

				if MeasInfoItem_C.measType.present > 0 {
					if int32(MeasInfoItem_C.measType.present) == 1 {
						MeasInfoItem.MeasType = 1
						MeasName := &PrintableString{}
						MeasName_C := (*C.MeasurementTypeName_t)(unsafe.Pointer(&MeasInfoItem_C.measType.choice[0]))
						MeasName.Size = int(MeasName_C.size)
						MeasName.Buf = C.GoBytes(unsafe.Pointer(MeasName_C.buf), C.int(MeasName_C.size))
						MeasInfoItem.Measurement = MeasName
					} else if int32(MeasInfoItem_C.measType.present) == 2 {
						MeasInfoItem.MeasType = 2
						MeasInfoItem.Measurement = int64(MeasInfoItem_C.measType.choice[0])
					}
				}

				if MeasInfoItem_C.labelInfoList.list.count > 0 {
					MeasInfoItem.LabelInfoCount = int(MeasInfoItem_C.labelInfoList.list.count)
					LabelInfoList := []MeasLabelInfo{}

					for j := 0; j < MeasInfoItem.LabelInfoCount; j++ {
						var sizeof_MeasurementLabel_t *C.MeasurementLabel_t
						LabelInfo_C := *(**C.MeasurementLabel_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasInfoItem_C.labelInfoList.list.array)) + (uintptr)(j)*unsafe.Sizeof(sizeof_MeasurementLabel_t)))
						LabelInfo := LabelInfoList[j]

						if LabelInfo_C.plmnID != nil {
							LabelInfo.PLMNID = &OctetString{}
							LabelInfo.PLMNID.Size = int(LabelInfo_C.plmnID.size)
							LabelInfo.PLMNID.Buf = C.GoBytes(unsafe.Pointer(LabelInfo_C.plmnID.buf), C.int(LabelInfo_C.plmnID.size))
						}

						if LabelInfo_C.sliceID != nil {
							LabelInfo.SliceID = &SliceIDType{}

							if LabelInfo_C.sliceID.sST.size > 0 {
								LabelInfo.SliceID.SST.Size = int(LabelInfo_C.sliceID.sST.size)
								LabelInfo.SliceID.SST.Buf = C.GoBytes(unsafe.Pointer(LabelInfo_C.sliceID.sST.buf), C.int(LabelInfo_C.sliceID.sST.size))
							}

							if LabelInfo_C.sliceID.sD.size > 0 {
								LabelInfo.SliceID.SD = &OctetString{}
								LabelInfo.SliceID.SD.Size = int(LabelInfo_C.sliceID.sD.size)
								LabelInfo.SliceID.SD.Buf = C.GoBytes(unsafe.Pointer(LabelInfo_C.sliceID.sD.buf), C.int(LabelInfo_C.sliceID.sD.size))
							}
						}

						if LabelInfo_C.fiveQI != nil {
							LabelInfo.FiveQI = int64(*LabelInfo_C.fiveQI)
						}

						if LabelInfo_C.qCI != nil {
							LabelInfo.QCI = int64(*LabelInfo_C.qCI)
						}

						if LabelInfo_C.qCImax != nil {
							LabelInfo.QCImax = int64(*LabelInfo_C.qCImax)
						}

						if LabelInfo_C.qCImin != nil {
							LabelInfo.QCImin = int64(*LabelInfo_C.qCImin)
						}

						if LabelInfo_C.aRPmax != nil {
							LabelInfo.ARPmax = int64(*LabelInfo_C.aRPmax)
						}

						if LabelInfo_C.aRPmin != nil {
							LabelInfo.ARPmin = int64(*LabelInfo_C.aRPmin)
						}

						if LabelInfo_C.bitrateRange != nil {
							LabelInfo.BitrateRange = int64(*LabelInfo_C.bitrateRange)
						}

						if LabelInfo_C.layerMU_MIMO != nil {
							LabelInfo.LayerMU_MIMO = int64(*LabelInfo_C.layerMU_MIMO)
						}

						if LabelInfo_C.sUM != nil {
							LabelInfo.SUM = int64(*LabelInfo_C.sUM)
						}

						if LabelInfo_C.distBinX != nil {
							LabelInfo.DistBinX = int64(*LabelInfo_C.distBinX)
						}

						if LabelInfo_C.distBinY != nil {
							LabelInfo.DistBinY = int64(*LabelInfo_C.distBinY)
						}

						if LabelInfo_C.distBinZ != nil {
							LabelInfo.DistBinZ = int64(*LabelInfo_C.distBinZ)
						}

						if LabelInfo_C.preLabelOverride != nil {
							LabelInfo.PreLabelOverride = int64(*LabelInfo_C.preLabelOverride)
						}

						if LabelInfo_C.startEndInd != nil {
							LabelInfo.StartEndInd = int64(*LabelInfo_C.startEndInd)
						}
					}
				}
			}

			indMsgFormat1.MeasInfoList = MeasInfoList
		}

		if indMsgFormat1_C.measData.list.count > 0 {
			indMsgFormat1.MeasDataCount = int(indMsgFormat1_C.measData.list.count)
			MeasDataList := []MeasurementRecord{}

			for i := 0; i < indMsgFormat1.MeasDataCount; i++ {
				var sizeof_MeasurementRecord_t *C.MeasurementRecord_t
				MeasRecord_C := *(**C.MeasurementRecord_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat1_C.measData.list.array)) + (uintptr)(i)*unsafe.Sizeof(sizeof_MeasurementRecord_t)))
				MeasDataList[i].MeasRecordCount = int(MeasRecord_C.list.count)

				for j := 0; j < MeasDataList[i].MeasRecordCount; j++ {
					var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
					MeasRecordItem_C := *(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat1_C.measData.list.array)) + (uintptr)(j)*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))
					MeasDataList[i].MeasRecord[j].MeasRecordType = int32(MeasRecordItem_C.present)

					if MeasDataList[i].MeasRecord[j].MeasRecordType == 1 {
						MeasDataList[i].MeasRecord[j].MeasRecordValue = int64(MeasRecordItem_C.choice[0])
					} else if MeasDataList[i].MeasRecord[j].MeasRecordType == 2 {
						MeasDataList[i].MeasRecord[j].MeasRecordValue = float64(MeasRecordItem_C.choice[0])
					} else if MeasDataList[i].MeasRecord[j].MeasRecordType == 3 {
						MeasDataList[i].MeasRecord[j].MeasRecordValue = int32(MeasRecordItem_C.choice[0])
					}
				}
			}

			indMsgFormat1.MeasData = MeasDataList
		}

		indMsg.IndMsg = indMsgFormat1
	} else if indMsg.IndMsgType == 2 {
		indMsgFormat2 := &IndicationMessageFormat2{}
		indMsgFormat2_C := *(**C.E2SM_KPM_IndicationMessage_Format2_t)(unsafe.Pointer(&decodedMsg.indicationMessage_formats.choice[0]))

		if indMsgFormat2_C.granulPeriod != nil {
			indMsgFormat2.GranulPeriod = int64(*indMsgFormat2_C.granulPeriod)
		} else {
			indMsgFormat2.GranulPeriod = -1
		}

		if indMsgFormat2_C.subscriptID.size > 0 {
			indMsgFormat2.SubscriptID = &Integer{}

			indMsgFormat2.SubscriptID.Buf = C.GoBytes(unsafe.Pointer(indMsgFormat2_C.subscriptID.buf), C.int(indMsgFormat2_C.subscriptID.size))
			indMsgFormat2.SubscriptID.Size = int(indMsgFormat2_C.subscriptID.size)
		}

		if indMsgFormat2_C.cellObjID != nil {
			indMsgFormat2.CellObjID = &PrintableString{}

			indMsgFormat2.CellObjID.Buf = C.GoBytes(unsafe.Pointer(indMsgFormat2_C.cellObjID.buf), C.int(indMsgFormat2_C.cellObjID.size))
			indMsgFormat2.CellObjID.Size = int(indMsgFormat2_C.cellObjID.size)
		}

		indMsgFormat2.MeasInfoUeidCount = int(indMsgFormat2_C.measCondUEidList.list.count)
		MeasInfoUeidList := []MeasInfoUeidItem{}

		for i := 0; i < indMsgFormat2.MeasInfoUeidCount; i++ {
			var sizeof_MeasurementCondUEidItem_t *C.MeasurementCondUEidItem_t
			MeasInfoUeidItem_C := *(**C.MeasurementCondUEidItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measCondUEidList.list.array)) + (uintptr)(i)*unsafe.Sizeof(sizeof_MeasurementCondUEidItem_t)))
			MeasInfoUeidItem := MeasInfoUeidList[i]

			if MeasInfoUeidItem_C.measType.present > 0 {
				if int32(MeasInfoUeidItem_C.measType.present) == 1 {
					MeasInfoUeidItem.MeasType = 1
					MeasName := &PrintableString{}
					MeasName_C := (*C.MeasurementTypeName_t)(unsafe.Pointer(&MeasInfoUeidItem_C.measType.choice[0]))
					MeasName.Size = int(MeasName_C.size)
					MeasName.Buf = C.GoBytes(unsafe.Pointer(MeasName_C.buf), C.int(MeasName_C.size))
					MeasInfoUeidItem.Measurement = MeasName
				} else if int32(MeasInfoUeidItem_C.measType.present) == 2 {
					MeasInfoUeidItem.MeasType = 2
					MeasInfoUeidItem.Measurement = int64(MeasInfoUeidItem_C.measType.choice[0])
				}
			}

			MeasInfoUeidItem.MatchingCondCount = int(MeasInfoUeidItem_C.matchingCond.list.count)
			MatchingCondList := []MatchingCond{}

			for j := 0; j < MeasInfoUeidItem.MatchingCondCount; j++ {
				var sizeof_MatchingCondItem_t *C.MatchingCondItem_t
				MatchingCondItem_C := *(**C.MatchingCondItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(MeasInfoUeidItem_C.matchingCond.list.array)) + (uintptr)(j)*unsafe.Sizeof(sizeof_MatchingCondItem_t)))
				MatchingCond := MatchingCondList[j]

				MatchingCond.ConditionType = int32(MatchingCondItem_C.present)
				if MatchingCond.ConditionType == 1 {
					LabelInfo_C := *(**C.MeasurementLabel_t)(unsafe.Pointer(&MatchingCondItem_C.choice[0]))
					LabelInfo := &MeasLabelInfo{}

					if LabelInfo_C.plmnID != nil {
						LabelInfo.PLMNID = &OctetString{}
						LabelInfo.PLMNID.Size = int(LabelInfo_C.plmnID.size)
						LabelInfo.PLMNID.Buf = C.GoBytes(unsafe.Pointer(LabelInfo_C.plmnID.buf), C.int(LabelInfo_C.plmnID.size))
					}

					if LabelInfo_C.sliceID != nil {
						LabelInfo.SliceID = &SliceIDType{}

						if LabelInfo_C.sliceID.sST.size > 0 {
							LabelInfo.SliceID.SST.Size = int(LabelInfo_C.sliceID.sST.size)
							LabelInfo.SliceID.SST.Buf = C.GoBytes(unsafe.Pointer(LabelInfo_C.sliceID.sST.buf), C.int(LabelInfo_C.sliceID.sST.size))
						}

						if LabelInfo_C.sliceID.sD.size > 0 {
							LabelInfo.SliceID.SD = &OctetString{}
							LabelInfo.SliceID.SD.Size = int(LabelInfo_C.sliceID.sD.size)
							LabelInfo.SliceID.SD.Buf = C.GoBytes(unsafe.Pointer(LabelInfo_C.sliceID.sD.buf), C.int(LabelInfo_C.sliceID.sD.size))
						}
					}

					if LabelInfo_C.fiveQI != nil {
						LabelInfo.FiveQI = int64(*LabelInfo_C.fiveQI)
					}

					if LabelInfo_C.qCI != nil {
						LabelInfo.QCI = int64(*LabelInfo_C.qCI)
					}

					if LabelInfo_C.qCImax != nil {
						LabelInfo.QCImax = int64(*LabelInfo_C.qCImax)
					}

					if LabelInfo_C.qCImin != nil {
						LabelInfo.QCImin = int64(*LabelInfo_C.qCImin)
					}

					if LabelInfo_C.aRPmax != nil {
						LabelInfo.ARPmax = int64(*LabelInfo_C.aRPmax)
					}

					if LabelInfo_C.aRPmin != nil {
						LabelInfo.ARPmin = int64(*LabelInfo_C.aRPmin)
					}

					if LabelInfo_C.bitrateRange != nil {
						LabelInfo.BitrateRange = int64(*LabelInfo_C.bitrateRange)
					}

					if LabelInfo_C.layerMU_MIMO != nil {
						LabelInfo.LayerMU_MIMO = int64(*LabelInfo_C.layerMU_MIMO)
					}

					if LabelInfo_C.sUM != nil {
						LabelInfo.SUM = int64(*LabelInfo_C.sUM)
					}

					if LabelInfo_C.distBinX != nil {
						LabelInfo.DistBinX = int64(*LabelInfo_C.distBinX)
					}

					if LabelInfo_C.distBinY != nil {
						LabelInfo.DistBinY = int64(*LabelInfo_C.distBinY)
					}

					if LabelInfo_C.distBinZ != nil {
						LabelInfo.DistBinZ = int64(*LabelInfo_C.distBinZ)
					}

					if LabelInfo_C.preLabelOverride != nil {
						LabelInfo.PreLabelOverride = int64(*LabelInfo_C.preLabelOverride)
					}

					if LabelInfo_C.startEndInd != nil {
						LabelInfo.StartEndInd = int64(*LabelInfo_C.startEndInd)
					}

					MatchingCond.Condition = LabelInfo
				} else if MatchingCond.ConditionType == 2 {
					TestInfo_C := *(**C.TestCondInfo_t)(unsafe.Pointer(&MatchingCondItem_C.choice[0]))
					TestInfo := &TestConditionInfo{}

					TestInfo.TestConditionType = int32(TestInfo_C.testType.present)
					TestInfo.Expression = int32(TestInfo_C.testExpr)
					ValueType := int32(TestInfo_C.testValue.present)
					switch ValueType {
					case 1, 2, 3:
						TestInfo.Value = int64(TestInfo_C.testValue.choice[0])
					case 4:
						Value := &BitString{}
						TestValue := **(**C.BIT_STRING_t)(unsafe.Pointer(&TestInfo_C.testValue.choice[0]))
						Value.Size = int(TestValue.size)
						Value.Buf = C.GoBytes(unsafe.Pointer(TestValue.buf), C.int(TestValue.size))
						Value.BitsUnused = int(TestValue.bits_unused)
						TestInfo.Value = Value
					case 5, 6:
						Value := &OctetString{}
						TestValue := **(**C.OCTET_STRING_t)(unsafe.Pointer(&TestInfo_C.testValue.choice[0]))
						Value.Size = int(TestValue.size)
						Value.Buf = C.GoBytes(unsafe.Pointer(TestValue.buf), C.int(TestValue.size))
						TestInfo.Value = Value
					default:
						return indMsg, errors.New("Unknown Test Condition Value type")
					}

					MatchingCond.Condition = TestInfo
				} else {
					return indMsg, errors.New("Unknown Test Condition type")
				}
			}

		}
		indMsgFormat2.MeasInfoUeidList = MeasInfoUeidList

		if indMsgFormat2_C.measData.list.count > 0 {
			indMsgFormat2.MeasDataCount = int(indMsgFormat2_C.measData.list.count)
			MeasDataList := []MeasurementRecord{}

			for i := 0; i < indMsgFormat2.MeasDataCount; i++ {
				var sizeof_MeasurementRecord_t *C.MeasurementRecord_t
				MeasRecord_C := *(**C.MeasurementRecord_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(i)*unsafe.Sizeof(sizeof_MeasurementRecord_t)))
				MeasDataList[i].MeasRecordCount = int(MeasRecord_C.list.count)

				for j := 0; j < MeasDataList[i].MeasRecordCount; j++ {
					var sizeof_MeasurementRecordItem_t *C.MeasurementRecordItem_t
					MeasRecordItem_C := *(**C.MeasurementRecordItem_t)(unsafe.Pointer(uintptr(unsafe.Pointer(indMsgFormat2_C.measData.list.array)) + (uintptr)(j)*unsafe.Sizeof(sizeof_MeasurementRecordItem_t)))
					MeasDataList[i].MeasRecord[j].MeasRecordType = int32(MeasRecordItem_C.present)

					if MeasDataList[i].MeasRecord[j].MeasRecordType == 1 {
						MeasDataList[i].MeasRecord[j].MeasRecordValue = int64(MeasRecordItem_C.choice[0])
					} else if MeasDataList[i].MeasRecord[j].MeasRecordType == 2 {
						MeasDataList[i].MeasRecord[j].MeasRecordValue = float64(MeasRecordItem_C.choice[0])
					} else if MeasDataList[i].MeasRecord[j].MeasRecordType == 3 {
						MeasDataList[i].MeasRecord[j].MeasRecordValue = int32(MeasRecordItem_C.choice[0])
					}
				}
			}

			indMsgFormat2.MeasData = MeasDataList
		}

		indMsg.IndMsg = indMsgFormat2
	} else {
		return indMsg, errors.New("Unknown RIC Indication Message Format")
	}

	return
}

func (c *E2sm) ParseNRCGI(nRCGI NRCGIType) (CellID string, err error) {
	var plmnID OctetString
	var nrCellID BitString

	plmnID = nRCGI.PlmnID
	CellID, _ = c.ParsePLMNIdentity(plmnID.Buf, plmnID.Size)

	nrCellID = nRCGI.NRCellID

	if plmnID.Size != 3 || nrCellID.Size != 5 {
		return "", errors.New("Invalid input: illegal length of NRCGI")
	}

	var former []uint8 = make([]uint8, 3)
	var latter []uint8 = make([]uint8, 6)

	former[0] = nrCellID.Buf[0] >> 4
	former[1] = nrCellID.Buf[0] & 0xf
	former[2] = nrCellID.Buf[1] >> 4
	latter[0] = nrCellID.Buf[1] & 0xf
	latter[1] = nrCellID.Buf[2] >> 4
	latter[2] = nrCellID.Buf[2] & 0xf
	latter[3] = nrCellID.Buf[3] >> 4
	latter[4] = nrCellID.Buf[3] & 0xf
	latter[5] = nrCellID.Buf[4] >> uint(nrCellID.BitsUnused)

	CellID = CellID + strconv.Itoa(int(former[0])) + strconv.Itoa(int(former[1])) + strconv.Itoa(int(former[2])) + strconv.Itoa(int(latter[0])) + strconv.Itoa(int(latter[1])) + strconv.Itoa(int(latter[2])) + strconv.Itoa(int(latter[3])) + strconv.Itoa(int(latter[4])) + strconv.Itoa(int(latter[5]))

	return
}

func (c *E2sm) ParsePLMNIdentity(buffer []byte, size int) (PlmnID string, err error) {
	if size != 3 {
		return "", errors.New("Invalid input: illegal length of PlmnID")
	}

	var mcc []uint8 = make([]uint8, 3)
	var mnc []uint8 = make([]uint8, 3)

	mcc[0] = buffer[0] >> 4
	mcc[1] = buffer[0] & 0xf
	mcc[2] = buffer[1] >> 4
	mnc[0] = buffer[1] & 0xf
	mnc[1] = buffer[2] >> 4
	mnc[2] = buffer[2] & 0xf

	if mnc[0] == 0xf {
		PlmnID = strconv.Itoa(int(mcc[0])) + strconv.Itoa(int(mcc[1])) + strconv.Itoa(int(mcc[2])) + strconv.Itoa(int(mnc[1])) + strconv.Itoa(int(mnc[2]))
	} else {
		PlmnID = strconv.Itoa(int(mcc[0])) + strconv.Itoa(int(mcc[1])) + strconv.Itoa(int(mcc[2])) + strconv.Itoa(int(mnc[0])) + strconv.Itoa(int(mnc[1])) + strconv.Itoa(int(mnc[2]))
	}

	return
}

func (c *E2sm) ParseSliceID(sliceID SliceIDType) (combined int32, err error) {
	if sliceID.SST.Size != 1 || (sliceID.SD != nil && sliceID.SD.Size != 3) {
		return 0, errors.New("Invalid input: illegal length of sliceID")
	}

	var temp uint8
	var sst int32
	var sd int32

	byteBuffer := bytes.NewBuffer(sliceID.SST.Buf)
	binary.Read(byteBuffer, binary.BigEndian, &temp)
	sst = int32(temp)

	if sliceID.SD == nil {
		combined = sst << 24
	} else {
		for i := 0; i < sliceID.SD.Size; i++ {
			byteBuffer = bytes.NewBuffer(sliceID.SD.Buf[i : i+1])
			binary.Read(byteBuffer, binary.BigEndian, &temp)
			sd = sd*256 + int32(temp)
		}
		combined = sst<<24 + sd
	}

	return
}

func (c *E2sm) ParseInteger(buffer []byte, size int) (value int64, err error) {
	var temp uint8
	var byteBuffer *bytes.Buffer

	for i := 0; i < size; i++ {
		byteBuffer = bytes.NewBuffer(buffer[i : i+1])
		binary.Read(byteBuffer, binary.BigEndian, &temp)
		value = value*256 + int64(temp)
	}

	return
}

func (c *E2sm) ParseTimestamp(buffer []byte, size int) (timestamp *Timestamp, err error) {
	var temp uint8
	var byteBuffer *bytes.Buffer
	var index int
	var sec int64
	var nsec int64

	for index := 0; index < size-8; index++ {
		byteBuffer = bytes.NewBuffer(buffer[index : index+1])
		binary.Read(byteBuffer, binary.BigEndian, &temp)
		sec = sec*256 + int64(temp)
	}

	for index = size - 8; index < size; index++ {
		byteBuffer = bytes.NewBuffer(buffer[index : index+1])
		binary.Read(byteBuffer, binary.BigEndian, &temp)
		nsec = nsec*256 + int64(temp)
	}

	timestamp = &Timestamp{TVsec: sec, TVnsec: nsec}
	return
}
