/*
 * Generated by asn1c-0.9.29 (http://lionet.info/asn1c)
 * From ASN.1 module "E2SM-KPM-IEs"
 * 	found in "E2SM-KPM-v02.00.03.asn"
 * 	`asn1c -pdu=auto -fno-include-deps -fcompound-names -findirect-choice -gen-PER -gen-OER -no-gen-example -D E2SM-KPM-v02.00.03`
 */

#include "MeasurementCondUEidList.h"

#include "MeasurementCondUEidItem.h"
static asn_oer_constraints_t asn_OER_type_MeasurementCondUEidList_constr_1 CC_NOTUSED = {
	{ 0, 0 },
	-1	/* (SIZE(1..65535)) */};
asn_per_constraints_t asn_PER_type_MeasurementCondUEidList_constr_1 CC_NOTUSED = {
	{ APC_UNCONSTRAINED,	-1, -1,  0,  0 },
	{ APC_CONSTRAINED,	 16,  16,  1,  65535 }	/* (SIZE(1..65535)) */,
	0, 0	/* No PER value map */
};
asn_TYPE_member_t asn_MBR_MeasurementCondUEidList_1[] = {
	{ ATF_POINTER, 0, 0,
		(ASN_TAG_CLASS_UNIVERSAL | (16 << 2)),
		0,
		&asn_DEF_MeasurementCondUEidItem,
		0,
		{ 0, 0, 0 },
		0, 0, /* No default value */
		""
		},
};
static const ber_tlv_tag_t asn_DEF_MeasurementCondUEidList_tags_1[] = {
	(ASN_TAG_CLASS_UNIVERSAL | (16 << 2))
};
asn_SET_OF_specifics_t asn_SPC_MeasurementCondUEidList_specs_1 = {
	sizeof(struct MeasurementCondUEidList),
	offsetof(struct MeasurementCondUEidList, _asn_ctx),
	0,	/* XER encoding is XMLDelimitedItemList */
};
asn_TYPE_descriptor_t asn_DEF_MeasurementCondUEidList = {
	"MeasurementCondUEidList",
	"MeasurementCondUEidList",
	&asn_OP_SEQUENCE_OF,
	asn_DEF_MeasurementCondUEidList_tags_1,
	sizeof(asn_DEF_MeasurementCondUEidList_tags_1)
		/sizeof(asn_DEF_MeasurementCondUEidList_tags_1[0]), /* 1 */
	asn_DEF_MeasurementCondUEidList_tags_1,	/* Same as above */
	sizeof(asn_DEF_MeasurementCondUEidList_tags_1)
		/sizeof(asn_DEF_MeasurementCondUEidList_tags_1[0]), /* 1 */
	{ &asn_OER_type_MeasurementCondUEidList_constr_1, &asn_PER_type_MeasurementCondUEidList_constr_1, SEQUENCE_OF_constraint },
	asn_MBR_MeasurementCondUEidList_1,
	1,	/* Single element */
	&asn_SPC_MeasurementCondUEidList_specs_1	/* Additional specs */
};
