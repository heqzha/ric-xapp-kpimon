package control

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Control struct {
	ranList            []string             //nodeB list
	eventCreateExpired int32                //maximum time for the RIC Subscription Request event creation procedure in the E2 Node
	eventDeleteExpired int32                //maximum time for the RIC Subscription Request event deletion procedure in the E2 Node
	rcChan             chan *xapp.RMRParams //channel for receiving rmr message
	client             influxdb.Client      //influxdb client
	eventCreateExpiredMap map[string]bool //map for recording the RIC Subscription Request event creation procedure is expired or not
	eventDeleteExpiredMap map[string]bool //map for recording the RIC Subscription Request event deletion procedure is expired or not
	eventCreateExpiredMu  *sync.Mutex     //mutex for eventCreateExpiredMap
	eventDeleteExpiredMu  *sync.Mutex     //mutex for eventDeleteExpiredMap
}

func init() {
	file := "/opt/kpimon.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	log.SetPrefix("[qSkipTool]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	xapp.Logger.SetLevel(4)
}

func NewControl() Control {
	str := os.Getenv("ranList")
	url := os.Getenv("influxAddr")
	client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     url,
		Username: "admin",
		Password: "",
	})
	if err != nil {
		panic(err)
	}

	return Control{strings.Split(str, ","),
		5, 5,
		make(chan *xapp.RMRParams),
		client,
		make(map[string]bool),
		make(map[string]bool),
		&sync.Mutex{},
		&sync.Mutex{}}
}

func ReadyCB(i interface{}) {
	c := i.(*Control)

	c.startTimerSubReq()
	go c.controlLoop()
}

func (c *Control) Run() {
	if len(c.ranList) > 0 {
		xapp.SetReadyCB(ReadyCB, c)
		xapp.Run(c)
	} else {
		xapp.Logger.Error("gNodeB not set for subscription")
		log.Printf("gNodeB not set for subscription")
	}

}

func (c *Control) startTimerSubReq() {
	timerSR := time.NewTimer(5 * time.Second)
	count := 0

	go func(t *time.Timer) {
		defer timerSR.Stop()
		for {
			<-t.C
			count++
			xapp.Logger.Debug("send RIC_SUB_REQ to gNodeB with cnt=%d", count)
			log.Printf("send RIC_SUB_REQ to gNodeB with cnt=%d", count)
			err := c.sendRicSubRequest(1001, 1001, 0)
			if err != nil && count < MAX_SUBSCRIPTION_ATTEMPTS {
				t.Reset(5 * time.Second)
			} else {
				break
			}
		}
	}(timerSR)
}

func (c *Control) Consume(rp *xapp.RMRParams) (err error) {
	c.rcChan <- rp
	return
}

func (c *Control) rmrSend(params *xapp.RMRParams) (err error) {
	if !xapp.Rmr.Send(params, false) {
		err = errors.New("rmr.Send() failed")
		xapp.Logger.Error("Failed to rmrSend to %v", err)
		log.Printf("Failed to rmrSend to %v", err)
	}
	return
}

func (c *Control) rmrReplyToSender(params *xapp.RMRParams) (err error) {
	if !xapp.Rmr.Send(params, true) {
		err = errors.New("rmr.Send() failed")
		xapp.Logger.Error("Failed to rmrReplyToSender to %v", err)
		log.Printf("Failed to rmrReplyToSender to %v", err)
	}
	return
}

func (c *Control) controlLoop() {
	for {
		msg := <-c.rcChan
		xapp.Logger.Debug("Received message type: %d", msg.Mtype)
		log.Printf("Received message type: %d", msg.Mtype)
		switch msg.Mtype {
		case 12050:
			c.handleIndication(msg)
		case 12011:
			c.handleSubscriptionResponse(msg)
		case 12012:
			c.handleSubscriptionFailure(msg)
		case 12021:
			c.handleSubscriptionDeleteResponse(msg)
		case 12022:
			c.handleSubscriptionDeleteFailure(msg)
		default:
			err := errors.New("Message Type " + strconv.Itoa(msg.Mtype) + " is discarded")
			xapp.Logger.Error("Unknown message type: %v", err)
			log.Printf("Unknown message type: %v", err)
		}
	}
}

func (c *Control) handleIndication(params *xapp.RMRParams) (err error) {
	var e2ap *E2ap
	var e2sm *E2sm

	indicationMsg, err := e2ap.GetIndicationMessage(params.Payload)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication message: %v", err)
		log.Printf("Failed to decode RIC Indication message: %v", err)
		return
	}

	log.Printf("RIC Indication message from {%s} received", params.Meid.RanName)
	log.Printf("RequestID: %d", indicationMsg.RequestID)
	log.Printf("RequestSequenceNumber: %d", indicationMsg.RequestSequenceNumber)
	log.Printf("FunctionID: %d", indicationMsg.FuncID)
	log.Printf("ActionID: %d", indicationMsg.ActionID)
	log.Printf("IndicationSN: %d", indicationMsg.IndSN)
	log.Printf("IndicationType: %d", indicationMsg.IndType)
	log.Printf("IndicationHeader: %x", indicationMsg.IndHeader)
	log.Printf("IndicationMessage: %x", indicationMsg.IndMessage)
	log.Printf("CallProcessID: %x", indicationMsg.CallProcessID)

	indicationHdr, err := e2sm.GetIndicationHeader(indicationMsg.IndHeader)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication Header: %v", err)
		log.Printf("Failed to decode RIC Indication Header: %v", err)
		return
	}

	log.Printf("-----------RIC Indication Header-----------")
	if indicationHdr.IndHdrType == 1 {
		log.Printf("RIC Indication Header Format: %d", indicationHdr.IndHdrType)
		indHdrFormat1 := indicationHdr.IndHdr.(*IndicationHeaderFormat1)

		log.Printf("GlobalKPMnodeIDType: %d", indHdrFormat1.GlobalKPMnodeIDType)

		if indHdrFormat1.GlobalKPMnodeIDType == 1 {
			globalKPMnodegNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodegNBIDType)

			globalgNBID := globalKPMnodegNBID.GlobalgNBID

			log.Printf("PlmnID: %x", globalgNBID.PlmnID.Buf)
			log.Printf("gNB ID Type: %d", globalgNBID.GnbIDType)
			if globalgNBID.GnbIDType == 1 {
				gNBID := globalgNBID.GnbID.(*GNBID)
				log.Printf("gNB ID ID: %x, Unused: %d", gNBID.Buf, gNBID.BitsUnused)
			}

			if globalKPMnodegNBID.GnbCUUPID != nil {
				log.Printf("gNB-CU-UP ID: %x", globalKPMnodegNBID.GnbCUUPID.Buf)
			}

			if globalKPMnodegNBID.GnbDUID != nil {
				log.Printf("gNB-DU ID: %x", globalKPMnodegNBID.GnbDUID.Buf)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 2 {
			globalKPMnodeengNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodeengNBIDType)

			log.Printf("PlmnID: %x", globalKPMnodeengNBID.PlmnID.Buf)
			log.Printf("en-gNB ID Type: %d", globalKPMnodeengNBID.GnbIDType)
			if globalKPMnodeengNBID.GnbIDType == 1 {
				engNBID := globalKPMnodeengNBID.GnbID.(*ENGNBID)
				log.Printf("en-gNB ID ID: %x, Unused: %d", engNBID.Buf, engNBID.BitsUnused)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 3 {
			globalKPMnodengeNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodengeNBIDType)

			log.Printf("PlmnID: %x", globalKPMnodengeNBID.PlmnID.Buf)
			log.Printf("ng-eNB ID Type: %d", globalKPMnodengeNBID.EnbIDType)
			if globalKPMnodengeNBID.EnbIDType == 1 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_Macro)
				log.Printf("ng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			} else if globalKPMnodengeNBID.EnbIDType == 2 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_ShortMacro)
				log.Printf("ng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			} else if globalKPMnodengeNBID.EnbIDType == 3 {
				ngeNBID := globalKPMnodengeNBID.EnbID.(*NGENBID_LongMacro)
				log.Printf("ng-eNB ID ID: %x, Unused: %d", ngeNBID.Buf, ngeNBID.BitsUnused)
			}
		} else if indHdrFormat1.GlobalKPMnodeIDType == 4 {
			globalKPMnodeeNBID := indHdrFormat1.GlobalKPMnodeID.(*GlobalKPMnodeeNBIDType)

			log.Printf("PlmnID: %x", globalKPMnodeeNBID.PlmnID.Buf)
			log.Printf("eNB ID Type: %d", globalKPMnodeeNBID.EnbIDType)
			if globalKPMnodeeNBID.EnbIDType == 1 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_Macro)
				log.Printf("eNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 2 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_Home)
				log.Printf("eNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 3 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_ShortMacro)
				log.Printf("eNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			} else if globalKPMnodeeNBID.EnbIDType == 4 {
				eNBID := globalKPMnodeeNBID.EnbID.(*ENBID_LongMacro)
				log.Printf("eNB ID ID: %x, Unused: %d", eNBID.Buf, eNBID.BitsUnused)
			}

		}

		if indHdrFormat1.ColletStartTime != nil {
			log.Printf("ColletStartTime: %x", indHdrFormat1.ColletStartTime.Buf)
		}

		if indHdrFormat1.FileFormatVersion != nil {
			log.Printf("FileFormatVersion: %x", indHdrFormat1.FileFormatVersion.Buf)
		}

		if indHdrFormat1.SenderName != nil {
			log.Printf("SenderName: %x", indHdrFormat1.SenderName.Buf)
		}

		if indHdrFormat1.SenderType != nil {
			log.Printf("SenderType: %x", indHdrFormat1.SenderType.Buf)
		}

		if indHdrFormat1.VendorName != nil {
			log.Printf("VendorName: %x", indHdrFormat1.VendorName.Buf)
		}
	} else {
		xapp.Logger.Error("Unknown RIC Indication Header Format: %d", indicationHdr.IndHdrType)
		log.Printf("Unknown RIC Indication Header Format: %d", indicationHdr.IndHdrType)
		return
	}

	indMsg, err := e2sm.GetIndicationMessage(indicationMsg.IndMessage)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Indication Message: %v", err)
		log.Printf("Failed to decode RIC Indication Message: %v", err)
		return
	}

	log.Printf("-----------RIC Indication Message-----------")
	log.Printf("StyleType: %d", indMsg.IndMsgType)
	if indMsg.IndMsgType == 1 {
		log.Printf("RIC Indication Message Format: %d", indMsg.IndMsgType)

		indMsgFormat1 := indMsg.IndMsg.(*IndicationMessageFormat1)

		log.Printf("GranulPeriod: %d", indMsgFormat1.GranulPeriod)

		if indMsgFormat1.SubscriptID != nil {
			log.Printf("SubscriptID: %x", indMsgFormat1.SubscriptID.Buf)
		}

		if indMsgFormat1.CellObjID != nil {
			log.Printf("CellObjID: %x", indMsgFormat1.CellObjID.Buf)
		}

		if indMsgFormat1.MeasInfoList != nil {
			log.Printf("MeasInfoCount: %d", indMsgFormat1.MeasInfoCount)
			for i := 0; i < indMsgFormat1.MeasInfoCount; i++ {
				log.Printf("MeasInfoList[%d]: ", i)
				MeasInfo := indMsgFormat1.MeasInfoList[i]

				if MeasInfo.MeasType == 1 {
					ID := MeasInfo.Measurement.(MeasID)
					log.Printf("MeasID: %d", ID)
				} else if MeasInfo.MeasType == 2 {
					Name := MeasInfo.Measurement.(MeasName)
					log.Printf("MeasName: %x", Name.Buf)
				} else {
					xapp.Logger.Error("Unknown Measurement Type: %d", MeasInfo.MeasType)
					log.Printf("Unknown Measurement Type: %d", MeasInfo.MeasType)
				}

				if MeasInfo.LabelInfoList != nil {
					log.Printf("LabelInfoCount: %d", MeasInfo.LabelInfoCount)
					for j := 0; j < MeasInfo.LabelInfoCount; j++ {
						log.Printf("LabelInfoList[%d]: ", j)
						LabelInfo := MeasInfo.LabelInfoList[j]

						if LabelInfo.PLMNID != nil {
							log.Printf("PLMNID: %x", LabelInfo.PLMNID.Buf)
						}
						if LabelInfo.SliceID != nil {
							log.Printf("SliceID.SST: %x", LabelInfo.SliceID.SST.Buf)
							if LabelInfo.SliceID.SD != nil {
								log.Printf("SliceID.SD: %x", LabelInfo.SliceID.SD.Buf)
							}
						}

						log.Printf("FiveQI: %d", LabelInfo.FiveQI)
						log.Printf("QCI: %d", LabelInfo.QCI)
						log.Printf("QCImax: %d", LabelInfo.QCImax)
						log.Printf("QCImin: %d", LabelInfo.QCImin)
						log.Printf("ARPmax: %d", LabelInfo.ARPmax)
						log.Printf("ARPmin: %d", LabelInfo.ARPmin)
						log.Printf("BitrateRange: %d", LabelInfo.BitrateRange)
						log.Printf("LayerMU_MIMO: %d", LabelInfo.LayerMU_MIMO)
						log.Printf("SUM: %d", LabelInfo.SUM)
						log.Printf("DistBinX: %d", LabelInfo.DistBinX)
						log.Printf("DistBinY: %d", LabelInfo.DistBinY)
						log.Printf("DistBinZ: %d", LabelInfo.DistBinZ)
						log.Printf("PreLabelOverride: %d", LabelInfo.PreLabelOverride)
						log.Printf("StartEndInd: %d", LabelInfo.StartEndInd)

						database := os.Getenv("influxDatabase")
						precision := os.Getenv("influxPrecision")
						bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
							Database:  database,
							Precision: precision,
						})
						if err != nil {
							xapp.Logger.Error("Failed to create batch points for influx: %v", err)
							log.Printf("Failed to create batch points for influx: %v", err)
						}

						tags := make(map[string]string)
						if LabelInfo.PLMNID != nil {
							tags["PLMNID"] = string(LabelInfo.PLMNID.Buf)
						}
						if LabelInfo.SliceID != nil {
							tags["SliceID.SST"] = string(LabelInfo.SliceID.SST.Buf)
							if LabelInfo.SliceID.SD != nil {
								tags["SliceID.SD"] = string(LabelInfo.SliceID.SD.Buf)
							}
						}

						fields := make(map[string]interface{})
						fields["FiveQI"] = LabelInfo.FiveQI
						fields["QCI"] = LabelInfo.QCI
						fields["QCImax"] = LabelInfo.QCImax
						fields["QCImin"] = LabelInfo.QCImin
						fields["ARPmax"] = LabelInfo.ARPmax
						fields["ARPmin"] = LabelInfo.ARPmin
						fields["BitrateRange"] = LabelInfo.BitrateRange
						fields["LayerMU_MIMO"] = LabelInfo.LayerMU_MIMO
						fields["SUM"] = LabelInfo.SUM
						fields["DistBinX"] = LabelInfo.DistBinX
						fields["DistBinY"] = LabelInfo.DistBinY
						fields["DistBinZ"] = LabelInfo.DistBinZ
						fields["PreLabelOverride"] = LabelInfo.PreLabelOverride
						fields["StartEndInd"] = LabelInfo.StartEndInd

						pt, err := influxdb.NewPoint("metrics", tags, fields, time.Now())
						if err != nil {
							xapp.Logger.Error("Failed to create point for influx: %v", err)
							log.Printf("Failed to create point for influx: %v", err)
						}

						bp.AddPoint(pt)
						err = c.client.Write(bp)
						if err != nil {
							xapp.Logger.Error("Failed to write influxdb: %v", err)
							log.Printf("Failed to write influxdb: %v", err)
						}
					}
				}
			}
		}

		if indMsgFormat1.MeasData != nil {
			log.Printf("MeasDataCount: %d", indMsgFormat1.MeasDataCount)

			for i := 0; i < indMsgFormat1.MeasDataCount; i++ {
				log.Printf("MeasDataList[%d]: ", i)
				MeasData := indMsgFormat1.MeasData[i]

				if MeasData.MeasRecord != nil {
					log.Printf("MeasRecordCount: %d", MeasData.MeasRecordCount)

					for j := 0; j < MeasData.MeasRecordCount; j++ {
						log.Printf("MeasRecordList[%d]: ", j)
						MeasRecord := MeasData.MeasRecord[j]

						if MeasRecord.MeasRecordType == 1 {
							Value := MeasRecord.MeasRecordValue.(Integer)
							log.Printf("Integer: %d", Value)
						} else if MeasRecord.MeasRecordType == 2 {
							Value := MeasRecord.MeasRecordValue.(Real)
							log.Printf("Real: %f", Value)
						} else if MeasRecord.MeasRecordType == 3 {
							Value := MeasRecord.MeasRecordValue.(Null)
							log.Printf("NoValue: %d", Value)
						} else {
							xapp.Logger.Error("Unknown Measured Value Type: %d", MeasRecord.MeasRecordType)
							log.Printf("Unknown Measured Value Type: %d", MeasRecord.MeasRecordType)
						}
					}
				}
			}
		}
	} else if indMsg.IndMsgType == 2 {
		log.Printf("RIC Indication Message Format: %d", indMsg.IndMsgType)

		indMsgFormat2 := indMsg.IndMsg.(*IndicationMessageFormat2)

		log.Printf("GranulPeriod: %d", indMsgFormat2.GranulPeriod)

		if indMsgFormat2.SubscriptID != nil {
			log.Printf("SubscriptID: %x", indMsgFormat2.SubscriptID.Buf)
		}

		if indMsgFormat2.CellObjID != nil {
			log.Printf("CellObjID: %x", indMsgFormat2.CellObjID.Buf)
		}

		if indMsgFormat2.MeasInfoUeidList != nil {
			log.Printf("MeasInfoUeidCount: %d", indMsgFormat2.MeasInfoUeidCount)
			for i := 0; i < indMsgFormat2.MeasInfoUeidCount; i++ {
				log.Printf("MeasInfoUeidList[%d]: ", i)
				MeasInfoUeid := indMsgFormat2.MeasInfoUeidList[i]

				if MeasInfoUeid.MeasType == 1 {
					ID := MeasInfoUeid.Measurement.(MeasID)
					log.Printf("MeasID: %d", ID)
				} else if MeasInfoUeid.MeasType == 2 {
					Name := MeasInfoUeid.Measurement.(MeasName)
					log.Printf("MeasName: %x", Name.Buf)
				} else {
					xapp.Logger.Error("Unknown Measurement Type: %d", MeasInfoUeid.MeasType)
					log.Printf("Unknown Measurement Type: %d", MeasInfoUeid.MeasType)
				}

				log.Printf("MatchingCondCount: %d", MeasInfoUeid.MatchingCondCount)
				for j := 0; j < MeasInfoUeid.MatchingCondCount; j++ {
					log.Printf("MatchingCondList[%d]: ", j)
					MatchingCondition := MeasInfoUeid.MatchingCondList[j]
					if MatchingCondition.ConditionType == 1 {
						LabelInfo := MatchingCondition.Condition.(MeasLabelInfo)
						if LabelInfo.PLMNID != nil {
							log.Printf("PLMNID: %x", LabelInfo.PLMNID.Buf)
						}
						if LabelInfo.SliceID != nil {
							log.Printf("SliceID.SST: %x", LabelInfo.SliceID.SST.Buf)
							if LabelInfo.SliceID.SD != nil {
								log.Printf("SliceID.SD: %x", LabelInfo.SliceID.SD.Buf)
							}
						}

						log.Printf("FiveQI: %d", LabelInfo.FiveQI)
						log.Printf("QCI: %d", LabelInfo.QCI)
						log.Printf("QCImax: %d", LabelInfo.QCImax)
						log.Printf("QCImin: %d", LabelInfo.QCImin)
						log.Printf("ARPmax: %d", LabelInfo.ARPmax)
						log.Printf("ARPmin: %d", LabelInfo.ARPmin)
						log.Printf("BitrateRange: %d", LabelInfo.BitrateRange)
						log.Printf("LayerMU_MIMO: %d", LabelInfo.LayerMU_MIMO)
						log.Printf("SUM: %d", LabelInfo.SUM)
						log.Printf("DistBinX: %d", LabelInfo.DistBinX)
						log.Printf("DistBinY: %d", LabelInfo.DistBinY)
						log.Printf("DistBinZ: %d", LabelInfo.DistBinZ)
						log.Printf("PreLabelOverride: %d", LabelInfo.PreLabelOverride)
						log.Printf("StartEndInd: %d", LabelInfo.StartEndInd)

						database := os.Getenv("influxDatabase")
						precision := os.Getenv("influxPrecision")
						bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
							Database:  database,
							Precision: precision,
						})
						if err != nil {
							xapp.Logger.Error("Failed to create batch points for influx: %v", err)
							log.Printf("Failed to create batch points for influx: %v", err)
						}

						tags := make(map[string]string)
						if LabelInfo.PLMNID != nil {
							tags["PLMNID"] = string(LabelInfo.PLMNID.Buf)
						}
						if LabelInfo.SliceID != nil {
							tags["SliceID.SST"] = string(LabelInfo.SliceID.SST.Buf)
							if LabelInfo.SliceID.SD != nil {
								tags["SliceID.SD"] = string(LabelInfo.SliceID.SD.Buf)
							}
						}

						fields := make(map[string]interface{})
						fields["FiveQI"] = LabelInfo.FiveQI
						fields["QCI"] = LabelInfo.QCI
						fields["QCImax"] = LabelInfo.QCImax
						fields["QCImin"] = LabelInfo.QCImin
						fields["ARPmax"] = LabelInfo.ARPmax
						fields["ARPmin"] = LabelInfo.ARPmin
						fields["BitrateRange"] = LabelInfo.BitrateRange
						fields["LayerMU_MIMO"] = LabelInfo.LayerMU_MIMO
						fields["SUM"] = LabelInfo.SUM
						fields["DistBinX"] = LabelInfo.DistBinX
						fields["DistBinY"] = LabelInfo.DistBinY
						fields["DistBinZ"] = LabelInfo.DistBinZ
						fields["PreLabelOverride"] = LabelInfo.PreLabelOverride
						fields["StartEndInd"] = LabelInfo.StartEndInd

						pt, err := influxdb.NewPoint("metrics", tags, fields, time.Now())
						if err != nil {
							xapp.Logger.Error("Failed to create point for influx: %v", err)
							log.Printf("Failed to create point for influx: %v", err)
						}

						bp.AddPoint(pt)
						err = c.client.Write(bp)
						if err != nil {
							xapp.Logger.Error("Failed to write influxdb: %v", err)
							log.Printf("Failed to write influxdb: %v", err)
						}
					} else if MatchingCondition.ConditionType == 2 {
						TestCondInfo := MatchingCondition.Condition.(TestConditionInfo)
						log.Printf("TestConditionType: %d", TestCondInfo.TestConditionType)
						log.Printf("Expression: %d", TestCondInfo.Expression)
						if TestCondInfo.ValueType == 1 {
							TestCondValue := TestCondInfo.Value.(int64)
							log.Printf("Integer: %d", TestCondValue)
						} else if TestCondInfo.ValueType == 2 {
							TestCondValue := TestCondInfo.Value.(int64)
							log.Printf("Enumerated: %d", TestCondValue)
						} else if TestCondInfo.ValueType == 3 {
							TestCondValue := TestCondInfo.Value.(int32)
							log.Printf("Boolean: %d", TestCondValue)
						} else if TestCondInfo.ValueType == 4 {
							TestCondValue := TestCondInfo.Value.(BitString)
							log.Printf("Bit string: %x, Unused: %d", TestCondValue.Buf, TestCondValue.BitsUnused)
						} else if TestCondInfo.ValueType == 5 {
							TestCondValue := TestCondInfo.Value.(OctetString)
							log.Printf("Octet string: %x", TestCondValue.Buf)
						} else if TestCondInfo.ValueType == 6 {
							TestCondValue := TestCondInfo.Value.(PrintableString)
							log.Printf("Printable string: %x", TestCondValue.Buf)
						} else {
							xapp.Logger.Error("Unknown Test Condition Value Type: %d", TestCondInfo.ValueType)
							log.Printf("Unknown Test Condition Value Type: %d", TestCondInfo.ValueType)
						}
					} else {
						xapp.Logger.Error("Unknown Matching Condition Type: %d", MatchingCondition.ConditionType)
						log.Printf("Unknown Matching Condition Type: %d", MatchingCondition.ConditionType)
					}
				}

				log.Printf("MatchedUeidCount: %d", MeasInfoUeid.MatchedUeidCount)
				for j := 0; j < MeasInfoUeid.MatchedUeidCount; j++ {
					log.Printf("MatchedUeid[%d]: %x", j, MeasInfoUeid.MatchedUeidList[j].Buf)
				}
			}
		}

		if indMsgFormat2.MeasData != nil {
			log.Printf("MeasDataCount: %d", indMsgFormat2.MeasDataCount)

			for i := 0; i < indMsgFormat2.MeasDataCount; i++ {
				log.Printf("MeasDataList[%d]: ", i)
				MeasData := indMsgFormat2.MeasData[i]

				if MeasData.MeasRecord != nil {
					log.Printf("MeasRecordCount: %d", MeasData.MeasRecordCount)

					for j := 0; j < MeasData.MeasRecordCount; j++ {
						log.Printf("MeasRecordList[%d]: ", j)
						MeasRecord := MeasData.MeasRecord[j]

						if MeasRecord.MeasRecordType == 1 {
							Value := MeasRecord.MeasRecordValue.(Integer)
							log.Printf("Integer: %d", Value)
						} else if MeasRecord.MeasRecordType == 2 {
							Value := MeasRecord.MeasRecordValue.(Real)
							log.Printf("Real: %f", Value)
						} else if MeasRecord.MeasRecordType == 3 {
							Value := MeasRecord.MeasRecordValue.(Null)
							log.Printf("NoValue: %d", Value)
						} else {
							xapp.Logger.Error("Unknown Measured Value Type: %d", MeasRecord.MeasRecordType)
							log.Printf("Unknown Measured Value Type: %d", MeasRecord.MeasRecordType)
						}
					}
				}
			}
		}
	} else {
		xapp.Logger.Error("Unknown RIC Indication Message Format: %d", indMsg.IndMsgType)
		log.Printf("Unkonw RIC Indication Message Format: %d", indMsg.IndMsgType)
		return
	}

	return nil
}

func (c *Control) handleSubscriptionResponse(params *xapp.RMRParams) (err error) {
	xapp.Logger.Debug("The SubId in RIC_SUB_RESP is %d", params.SubId)
	log.Printf("The SubId in RIC_SUB_RESP is %d", params.SubId)

	ranName := params.Meid.RanName
	c.eventCreateExpiredMu.Lock()
	_, ok := c.eventCreateExpiredMap[ranName]
	if !ok {
		c.eventCreateExpiredMu.Unlock()
		xapp.Logger.Debug("RIC_SUB_REQ has been deleted!")
		log.Printf("RIC_SUB_REQ has been deleted!")
		return nil
	} else {
		c.eventCreateExpiredMap[ranName] = true
		c.eventCreateExpiredMu.Unlock()
	}

	var cep *E2ap
	subscriptionResp, err := cep.GetSubscriptionResponseMessage(params.Payload)
	if err != nil {
		xapp.Logger.Error("Failed to decode RIC Subscription Response message: %v", err)
		log.Printf("Failed to decode RIC Subscription Response message: %v", err)
		return
	}

	log.Printf("RIC Subscription Response message from {%s} received", params.Meid.RanName)
	log.Printf("SubscriptionID: %d", params.SubId)
	log.Printf("RequestID: %d", subscriptionResp.RequestID)
	log.Printf("RequestSequenceNumber: %d", subscriptionResp.RequestSequenceNumber)
	log.Printf("FunctionID: %d", subscriptionResp.FuncID)

	log.Printf("ActionAdmittedList:")
	for index := 0; index < subscriptionResp.ActionAdmittedList.Count; index++ {
		log.Printf("[%d]ActionID: %d", index, subscriptionResp.ActionAdmittedList.ActionID[index])
	}

	log.Printf("ActionNotAdmittedList:")
	for index := 0; index < subscriptionResp.ActionNotAdmittedList.Count; index++ {
		log.Printf("[%d]ActionID: %d", index, subscriptionResp.ActionNotAdmittedList.ActionID[index])
		log.Printf("[%d]CauseType: %d    CauseID: %d", index, subscriptionResp.ActionNotAdmittedList.Cause[index].CauseType, subscriptionResp.ActionNotAdmittedList.Cause[index].CauseID)
	}

	return nil
}

func (c *Control) handleSubscriptionFailure(params *xapp.RMRParams) (err error) {
	xapp.Logger.Debug("The SubId in RIC_SUB_FAILURE is %d", params.SubId)
	log.Printf("The SubId in RIC_SUB_FAILURE is %d", params.SubId)

	ranName := params.Meid.RanName
	c.eventCreateExpiredMu.Lock()
	_, ok := c.eventCreateExpiredMap[ranName]
	if !ok {
		c.eventCreateExpiredMu.Unlock()
		xapp.Logger.Debug("RIC_SUB_REQ has been deleted!")
		log.Printf("RIC_SUB_REQ has been deleted!")
		return nil
	} else {
		c.eventCreateExpiredMap[ranName] = true
		c.eventCreateExpiredMu.Unlock()
	}

	return nil
}

func (c *Control) handleSubscriptionDeleteResponse(params *xapp.RMRParams) (err error) {
	xapp.Logger.Debug("The SubId in RIC_SUB_DEL_RESP is %d", params.SubId)
	log.Printf("The SubId in RIC_SUB_DEL_RESP is %d", params.SubId)

	ranName := params.Meid.RanName
	c.eventDeleteExpiredMu.Lock()
	_, ok := c.eventDeleteExpiredMap[ranName]
	if !ok {
		c.eventDeleteExpiredMu.Unlock()
		xapp.Logger.Debug("RIC_SUB_DEL_REQ has been deleted!")
		log.Printf("RIC_SUB_DEL_REQ has been deleted!")
		return nil
	} else {
		c.eventDeleteExpiredMap[ranName] = true
		c.eventDeleteExpiredMu.Unlock()
	}

	return nil
}

func (c *Control) handleSubscriptionDeleteFailure(params *xapp.RMRParams) (err error) {
	xapp.Logger.Debug("The SubId in RIC_SUB_DEL_FAILURE is %d", params.SubId)
	log.Printf("The SubId in RIC_SUB_DEL_FAILURE is %d", params.SubId)

	ranName := params.Meid.RanName
	c.eventDeleteExpiredMu.Lock()
	_, ok := c.eventDeleteExpiredMap[ranName]
	if !ok {
		c.eventDeleteExpiredMu.Unlock()
		xapp.Logger.Debug("RIC_SUB_DEL_REQ has been deleted!")
		log.Printf("RIC_SUB_DEL_REQ has been deleted!")
		return nil
	} else {
		c.eventDeleteExpiredMap[ranName] = true
		c.eventDeleteExpiredMu.Unlock()
	}

	return nil
}

func (c *Control) setEventCreateExpiredTimer(ranName string) {
	c.eventCreateExpiredMu.Lock()
	c.eventCreateExpiredMap[ranName] = false
	c.eventCreateExpiredMu.Unlock()

	timer := time.NewTimer(time.Duration(c.eventCreateExpired) * time.Second)
	go func(t *time.Timer) {
		defer t.Stop()
		xapp.Logger.Debug("RIC_SUB_REQ[%s]: Waiting for RIC_SUB_RESP...", ranName)
		log.Printf("RIC_SUB_REQ[%s]: Waiting for RIC_SUB_RESP...", ranName)
		for {
			select {
			case <-t.C:
				c.eventCreateExpiredMu.Lock()
				isResponsed := c.eventCreateExpiredMap[ranName]
				delete(c.eventCreateExpiredMap, ranName)
				c.eventCreateExpiredMu.Unlock()
				if !isResponsed {
					xapp.Logger.Debug("RIC_SUB_REQ[%s]: RIC Event Create Timer experied!", ranName)
					log.Printf("RIC_SUB_REQ[%s]: RIC Event Create Timer experied!", ranName)
					// c.sendRicSubDelRequest(subID, requestSN, funcID)
					return
				}
			default:
				c.eventCreateExpiredMu.Lock()
				flag := c.eventCreateExpiredMap[ranName]
				if flag {
					delete(c.eventCreateExpiredMap, ranName)
					c.eventCreateExpiredMu.Unlock()
					xapp.Logger.Debug("RIC_SUB_REQ[%s]: RIC Event Create Timer canceled!", ranName)
					log.Printf("RIC_SUB_REQ[%s]: RIC Event Create Timer canceled!", ranName)
					return
				} else {
					c.eventCreateExpiredMu.Unlock()
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(timer)
}

func (c *Control) setEventDeleteExpiredTimer(ranName string) {
	c.eventDeleteExpiredMu.Lock()
	c.eventDeleteExpiredMap[ranName] = false
	c.eventDeleteExpiredMu.Unlock()

	timer := time.NewTimer(time.Duration(c.eventDeleteExpired) * time.Second)
	go func(t *time.Timer) {
		defer t.Stop()
		xapp.Logger.Debug("RIC_SUB_DEL_REQ[%s]: Waiting for RIC_SUB_DEL_RESP...", ranName)
		log.Printf("RIC_SUB_DEL_REQ[%s]: Waiting for RIC_SUB_DEL_RESP...", ranName)
		for {
			select {
			case <-t.C:
				c.eventDeleteExpiredMu.Lock()
				isResponsed := c.eventDeleteExpiredMap[ranName]
				delete(c.eventDeleteExpiredMap, ranName)
				c.eventDeleteExpiredMu.Unlock()
				if !isResponsed {
					xapp.Logger.Debug("RIC_SUB_DEL_REQ[%s]: RIC Event Delete Timer experied!", ranName)
					log.Printf("RIC_SUB_DEL_REQ[%s]: RIC Event Delete Timer experied!", ranName)
					return
				}
			default:
				c.eventDeleteExpiredMu.Lock()
				flag := c.eventDeleteExpiredMap[ranName]
				if flag {
					delete(c.eventDeleteExpiredMap, ranName)
					c.eventDeleteExpiredMu.Unlock()
					xapp.Logger.Debug("RIC_SUB_DEL_REQ[%s]: RIC Event Delete Timer canceled!", ranName)
					log.Printf("RIC_SUB_DEL_REQ[%s]: RIC Event Delete Timer canceled!", ranName)
					return
				} else {
					c.eventDeleteExpiredMu.Unlock()
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(timer)
}

func (c *Control) sendRicSubRequest(subID int, requestSN int, funcID int) (err error) {
	var e2ap *E2ap
	var e2sm *E2sm

	var eventTriggerCount int = 1
	var periods int64 = 1
	var eventTriggerDefinition []byte = make([]byte, 8)
	_, err = e2sm.SetEventTriggerDefinition(eventTriggerDefinition, eventTriggerCount, periods)
	if err != nil {
		xapp.Logger.Error("Failed to send RIC_SUB_REQ: %v", err)
		log.Printf("Failed to send RIC_SUB_REQ: %v", err)
		return err
	}
	log.Printf("Set EventTriggerDefinition: %x", eventTriggerDefinition)

	var actionCount int = 1
	var ricStyleType []int64 = []int64{0}
	var actionIds []int64 = []int64{0}
	var actionTypes []int64 = []int64{0}
	var actionDefinitions []ActionDefinition = make([]ActionDefinition, actionCount)
	var subsequentActions []SubsequentAction = []SubsequentAction{SubsequentAction{0, 0, 0}}

	for index := 0; index < actionCount; index++ {
		if ricStyleType[index] == 0 {
			actionDefinitions[index].Buf = nil
			actionDefinitions[index].Size = 0
		} else {
			actionDefinitions[index].Buf = make([]byte, 8)
			_, err = e2sm.SetActionDefinition(actionDefinitions[index].Buf, ricStyleType[index])
			if err != nil {
				xapp.Logger.Error("Failed to send RIC_SUB_REQ: %v", err)
				log.Printf("Failed to send RIC_SUB_REQ: %v", err)
				return err
			}
			actionDefinitions[index].Size = len(actionDefinitions[index].Buf)

			log.Printf("Set ActionDefinition[%d]: %x", index, actionDefinitions[index].Buf)
		}
	}

	for index := 0; index < 1; index++ { //len(c.ranList)
		params := &xapp.RMRParams{}
		params.Mtype = 12010
		params.SubId = subID

		//xapp.Logger.Debug("Send RIC_SUB_REQ to {%s}", c.ranList[index])
		//log.Printf("Send RIC_SUB_REQ to {%s}", c.ranList[index])

		params.Payload = make([]byte, 1024)
		params.Payload, err = e2ap.SetSubscriptionRequestPayload(params.Payload, 1001, uint16(requestSN), uint16(funcID), eventTriggerDefinition, len(eventTriggerDefinition), actionCount, actionIds, actionTypes, actionDefinitions, subsequentActions)
		if err != nil {
			xapp.Logger.Error("Failed to send RIC_SUB_REQ: %v", err)
			log.Printf("Failed to send RIC_SUB_REQ: %v", err)
			return err
		}

		log.Printf("Set Payload: %x", params.Payload)

		//params.Meid = &xapp.RMRMeid{RanName: c.ranList[index]}
		params.Meid = &xapp.RMRMeid{PlmnID: "373437", EnbID: "10110101110001100111011110001", RanName: "gnb_734_733_b5c67788"}
		xapp.Logger.Debug("The RMR message to be sent is %d with SubId=%d", params.Mtype, params.SubId)
		log.Printf("The RMR message to be sent is %d with SubId=%d", params.Mtype, params.SubId)

		err = c.rmrSend(params)
		if err != nil {
			xapp.Logger.Error("Failed to send RIC_SUB_REQ: %v", err)
			log.Printf("Failed to send RIC_SUB_REQ: %v", err)
			return err
		}

		c.setEventCreateExpiredTimer(params.Meid.RanName)
		//c.ranList = append(c.ranList[:index], c.ranList[index+1:]...)
		//index--
	}

	return nil
}

func (c *Control) sendRicSubDelRequest(subID int, requestSN int, funcID int) (err error) {
	params := &xapp.RMRParams{}
	params.Mtype = 12020
	params.SubId = subID
	var e2ap *E2ap

	params.Payload = make([]byte, 1024)
	params.Payload, err = e2ap.SetSubscriptionDeleteRequestPayload(params.Payload, 100, uint16(requestSN), uint16(funcID))
	if err != nil {
		xapp.Logger.Error("Failed to send RIC_SUB_DEL_REQ: %v", err)
		return err
	}

	log.Printf("Set Payload: %x", params.Payload)

	if funcID == 0 {
		//params.Meid = &xapp.RMRMeid{PlmnID: "::", EnbID: "::", RanName: "0"}
		params.Meid = &xapp.RMRMeid{PlmnID: "373437", EnbID: "10110101110001100111011110001", RanName: "gnb_734_733_b5c67788"}
	} else {
		//params.Meid = &xapp.RMRMeid{PlmnID: "::", EnbID: "::", RanName: "3"}
		params.Meid = &xapp.RMRMeid{PlmnID: "373437", EnbID: "10110101110001100111011110001", RanName: "gnb_734_733_b5c67788"}
	}

	xapp.Logger.Debug("The RMR message to be sent is %d with SubId=%d", params.Mtype, params.SubId)
	log.Printf("The RMR message to be sent is %d with SubId=%d", params.Mtype, params.SubId)

	err = c.rmrSend(params)
	if err != nil {
		xapp.Logger.Error("Failed to send RIC_SUB_DEL_REQ: %v", err)
		log.Printf("Failed to send RIC_SUB_DEL_REQ: %v", err)
		return err
	}

	c.setEventDeleteExpiredTimer(params.Meid.RanName)

	return nil
}
