package main

import (
	"time"
)

// проверка на пустую запись об использовании ТС
func VehicleRecordIsEmpty(vr *сardVehicleRecord) bool {
	var result bool

	result = vr.VehicleFirstUse == time.Unix(0, 0).UTC()
	result = result && vr.VehicleLastUse == time.Unix(0, 0).UTC()
	result = result && vr.VehicleOdometerBegin == 0
	result = result && vr.VehicleOdometerEnd == 0

	return result
}

// проверка на пустую запись о неисправности
func FaultRecordIsEmpty(fault *cardFaultRecord) bool {
	var result bool

	result = fault.FaultBeginTime == time.Unix(0, 0).UTC()
	result = result && fault.FaultEndTime == time.Unix(0, 0).UTC()
	result = result && fault.FaultTypeId == 0

	return result
}

// проверка на пустую запись о событии
func EventRecordIsEmpty(event *cardEventRecord) bool {
	var result bool

	result = event.EventBeginTime == time.Unix(0, 0).UTC()
	result = result && event.EventEndTime == time.Unix(0, 0).UTC()
	result = result && event.EventTypeId == 0

	return result
}

// проверка на пустую запись о месте
func PlaceRecordIsEmpty(place *placeRecord) bool {
	var result bool

	result = place.EntryTime == time.Unix(0, 0).UTC()
	result = result && place.DailyWorkPeriodCountry == 0
	result = result && place.DailyWorkPeriodRegion == 0
	result = result && place.TypePeriodId == 0
	result = result && place.VehicleOdometerValue == 0

	return result
}

// проверка на пустую запись о специальных условиях
func SpecificConditionIsEmpty(sc *specificConditionRecord) bool {
	var result bool

	result = sc.EntryTime == time.Unix(0, 0).UTC()
	result = result && sc.SpecificConditionTypeId == 0

	return result
}

// проверка на пустую запись о контроле
func ControlActivityDataIsEmpty(cad *cardControlActivityDataRecord) bool {
	var result bool

	result = cad.ControlTypeId == 0
	result = result && cad.ControlTime == time.Unix(0, 0).UTC()
	result = result && cad.CardTypeId == 0
	result = result && cad.CardIssuingMemberState == 0
	result = result && cad.ControlCardNumber == ""
	result = result && cad.ControlDownloadPeriodBegin == time.Unix(0, 0).UTC()
	result = result && cad.ControlDownloadPeriodEnd == time.Unix(0, 0).UTC()

	return result
}