package main

import (
	"github.com/burntcarrot/heaputil"
	"github.com/burntcarrot/heaputil/record"
)

type RecordInfo struct {
	RecordType    record.RecordType
	RecordTypeStr string
}

func GetUniqueRecordTypes(records []heaputil.RecordData) []RecordInfo {
	recordTypesMap := map[record.RecordType]bool{}
	for _, recordInfo := range records {
		recordTypesMap[recordInfo.RecordType] = true
	}

	recordTypes := []RecordInfo{}
	for rType := range recordTypesMap {
		recordTypes = append(recordTypes, RecordInfo{RecordType: rType, RecordTypeStr: record.GetStrFromRecordType(rType)})
	}

	return recordTypes
}
