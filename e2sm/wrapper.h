#ifndef	_WRAPPER_H_
#define	_WRAPPER_H_

#include "ARP.h"
#include "BOOLEAN.h"
#include "E2SM-KPM-ActionDefinition-Format1.h"
#include "FiveQI.h"
#include "GranularityPeriod.h"
#include "LabelInfoItem.h"
#include "LabelInfoList.h"
#include "MeasurementData.h"
#include "MeasurementInfo-Action-Item.h"
#include "MeasurementInfoItem.h"
#include "MeasurementInfoList.h"
#include "MeasurementLabel.h"
#include "MeasurementRecord.h"
#include "MeasurementRecordItem.h"
#include "MeasurementType.h"
#include "MeasurementTypeID.h"
#include "MeasurementTypeName.h"
#include "NativeReal.h"
#include "NULL.h"
#include "QCI.h"
#include "REAL.h"
#include "SubscriptionID.h"
#include "TimeStamp.h"
#include "MeasurementInfoList.h"
#include "MeasurementTypeName.h"
#include "E2SM-KPM-EventTriggerDefinition.h"
#include "E2SM-KPM-EventTriggerDefinition-Format1.h"
#include "E2SM-KPM-ActionDefinition.h"
#include "E2SM-KPM-IndicationHeader.h"
#include "E2SM-KPM-IndicationHeader-Format1.h"
#include "GlobalKPMnode-ID.h"
#include "GlobalKPMnode-gNB-ID.h"
#include "GlobalKPMnode-en-gNB-ID.h"
#include "GlobalKPMnode-ng-eNB-ID.h"
#include "GlobalKPMnode-eNB-ID.h"
#include "PLMN-Identity.h"
#include "GNB-ID-Choice.h"
#include "GNB-CU-UP-ID.h"
#include "GNB-DU-ID.h"
#include "ENGNB-ID.h"
#include "ENB-ID-Choice.h"
#include "ENB-ID.h"
#include "NRCGI.h"
#include "SNSSAI.h"
#include "E2SM-KPM-IndicationMessage.h"
#include "E2SM-KPM-IndicationMessage-Format1.h"
#include "E2SM-KPM-IndicationMessage-Format2.h"
#include "MatchingCondItem.h"
#include "MeasurementCondUEidItem.h"
#include "TestCondInfo.h"


extern ssize_t e2sm_encode_ric_event_trigger_definition(void *buffer, size_t buf_size, size_t event_trigger_count, long RT_periods);
extern ssize_t e2sm_encode_ric_action_definition(void *buffer, size_t buf_size, long ric_style_type);
extern E2SM_KPM_IndicationHeader_t* e2sm_decode_ric_indication_header(void *buffer, size_t buf_size);
extern void e2sm_free_ric_indication_header(E2SM_KPM_IndicationHeader_t* indHdr);
extern E2SM_KPM_IndicationMessage_t* e2sm_decode_ric_indication_message(void *buffer, size_t buf_size);
extern void e2sm_free_ric_indication_message(E2SM_KPM_IndicationMessage_t* indMsg);

#endif /* _WRAPPER_H_ */
