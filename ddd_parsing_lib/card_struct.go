package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type сardVehicleRecord struct {
	VehicleOdometerBegin      int       `tlv:"0505 3 0 int" json:"vehicle_odometer_begin"`
	VehicleOdometerEnd        int       `tlv:"0505 3 3 int" json:"vehicle_odometer_end"`
	VehicleFirstUse           time.Time `tlv:"0505 4 6 date" json:"vehicle_first_use"`
	VehicleLastUse            time.Time `tlv:"0505 4 10 date" json:"vehicle_last_use"`
	VehicleRegistrationNation int       `tlv:"0505 1 14 int" json:"vehicle_registration_nation"`
	VehicleRegistrationNumber string    `tlv:"0505 14 15 string" json:"vehicle_registration_number"`
}

type сardVehicleRecords []сardVehicleRecord

type activityChangeInfo struct {
	TachographCardReaderId int       `json:"tachograph_card_reader_id"`
	StateDrivingId         int       `json:"state_driving_id"`
	CardPositionId         int       `json:"card_position_id"`
	ActivityKindId         int       `json:"activity_kind_id"`
	ActivityChangeInfoT    int       `json:"activity_change_info_t"`
	CalculatedTime         time.Time `json:"calculated_time"`
}

type activityDailyRecord struct {
	ActivityRecordDate           time.Time            `tlv:"0504 4 0 date" json:"activity_record_date"`
	ActivityDailyPresenceCounter string               `tlv:"0504 2 4 daylicounter" json:"activity_daily_presence_counter"`
	ActivityDayDistance          int                  `tlv:"0504 2 6 int" json:"activity_day_distance"`
	ActivitiesS                  string               `tlv:"0504 -1 8 activites" json:"activities_s"`
	ActivityChangeInfos          []activityChangeInfo `json:"activity_change_infos"`
}

func (adr *activityDailyRecord) ParseChangeInfo() error {
	beginAciIdx := 0
	nextAciIdx := beginAciIdx + 4 // т.к. 2 байта записанные строкой это 4 символа
	aci := activityChangeInfo{}
	adr.ActivityChangeInfos = []activityChangeInfo{}
	for beginAciIdx < len(adr.ActivitiesS) {
		currentACI := adr.ActivitiesS[beginAciIdx:nextAciIdx]
		if err := parseChangeInfoData(adr.ActivityRecordDate, currentACI, &aci); err != nil {
			log.Println("Can't upload change info for activity daily record")
			return err
		}

		beginAciIdx = nextAciIdx
		nextAciIdx = nextAciIdx + 4

		adr.ActivityChangeInfos = append(adr.ActivityChangeInfos, aci)
	}

	return nil
}

type activityDailyRecords []activityDailyRecord

type placeRecord struct {
	EntryTime              time.Time `tlv:"0506 4 0 date" json:"entry_time"`
	TypePeriodId           int       `tlv:"0506 1 4 int" json:"type_period_id"`
	DailyWorkPeriodCountry int       `tlv:"0506 1 5 int" json:"daily_work_period_country"`
	DailyWorkPeriodRegion  int       `tlv:"0506 1 6 int" json:"daily_work_period_region"`
	VehicleOdometerValue   int       `tlv:"0506 3 7 int" json:"vehicle_odometer_value"`
}

type placeRecords []placeRecord

type cardEventRecord struct {
	EventTypeId               int       `tlv:"0502 1 0 int" json:"event_type_id"`
	EventBeginTime            time.Time `tlv:"0502 4 1 date" json:"event_begin_time"`
	EventEndTime              time.Time `tlv:"0502 4 5 date" json:"event_end_time"`
	VehicleRegistrationNation int       `tlv:"0502 1 9 int" json:"vehicle_registration_nation"`
	VehicleRegistrationNumber string    `tlv:"0502 14 10 string" json:"vehicle_registration_number"`
}

type cardEventRecords []cardEventRecord

type cardFaultRecord struct {
	FaultTypeId               int       `tlv:"0503 1 0 int" json:"fault_type_id"`
	FaultBeginTime            time.Time `tlv:"0503 4 1 date" json:"fault_begin_time"`
	FaultEndTime              time.Time `tlv:"0503 4 5 date" json:"fault_end_time"`
	VehicleRegistrationNation int       `tlv:"0503 1 9 int" json:"vehicle_registration_nation"`
	VehicleRegistrationNumber string    `tlv:"0503 14 10 string" json:"vehicle_registration_number"`
}

type cardFaultRecords []cardFaultRecord

type cardControlActivityDataRecord struct {
	ControlTypeId              int       `tlv:"0508 1 0 int" json:"control_type_id"`
	ControlTime                time.Time `tlv:"0508 4 1 date" json:"control_time"`
	CardTypeId                 int       `tlv:"0508 1 5 int" json:"card_type_id"`
	CardIssuingMemberState     int       `tlv:"0508 1 6 int" json:"card_issuing_member_state"`
	ControlCardNumber          string    `tlv:"0508 16 7 string" json:"control_card_number"`
	ControlDownloadPeriodBegin time.Time `tlv:"0508 4 38 date" json:"control_download_period_begin"`
	ControlDownloadPeriodEnd   time.Time `tlv:"0508 4 42 date" json:"control_download_period_end"`
	VehicleRegistrationNation  int       `tlv:"0508 1 46 int" json:"vehicle_registration_nation"`
	VehicleRegistrationNumber  string    `tlv:"0508 14 47 string" json:"vehicle_registration_number"`
}

type cardControlActivityDataRecords []cardControlActivityDataRecord

type specificConditionRecord struct {
	SpecificConditionTypeId int       `tlv:"0522 1 4 int" json:"specific_condition_type_id" db:"specific_condition_type_id"`
	EntryTime               time.Time `tlv:"0522 4 0 date" json:"entry_time" db:"entry_time"`
}

type specificConditionRecords []specificConditionRecord

type driver struct {
	HolderSurname               string    `tlv:"0520 36 65 string" json:"holder_surname"`
	HolderFirstNames            string    `tlv:"0520 36 101 string" json:"holder_first_names"`
	CardHolderBirthDate         time.Time `tlv:"0520 5 137 birthday" json:"card_holder_birth_date"`
	CardHolderPreferredLanguage string    `tlv:"0520 2 141 string" json:"card_holder_preferred_language"`
}

type dlicense struct {
	DrivingLicenceIssuingAuthority string `tlv:"0521 36 0 string" json:"driving_licence_issuing_authority"`
	DrivingLicenceIssuingNation    int    `tlv:"0521 1 36 int" json:"driving_licence_issuing_nation"`
	DrivingLicenceNumber           string `tlv:"0521 16 37 string" json:"driving_licence_number"`
}

type sessionOpen struct {
	SessionOpenTime          time.Time `tlv:"0507 4 0 date" json:"session_open_time"`
	SessionOpenVehicleNation int       `tlv:"0507 1 4 int" json:"vehicle_registration_nation"`
	SessionOpenVehicleNumber string    `tlv:"0507 14 5 string" json:"vehicle_registration_number"`
}

type cardInfo struct {
	IcSerialNumber            string    `tlv:"0005 4 0 hexadecimal" json:"ic_serial_number"`
	IcManufacturingReferences string    `tlv:"0005 4 4 hexadecimal" json:"ic_manufacturing_references"`
	CardExtendedSerialNumber  string    `tlv:"0002 8 1 string" json:"card_extended_serial_number"`
	CardApprovalNumber        string    `tlv:"0002 8 9 string" json:"card_approval_number"`
	CardPersonalizerId        int       `tlv:"0002 1 17 int" json:"card_personalizer_id"`
	EmbeddericAssemblerId     string    `tlv:"0002 5 18 string" json:"embedderic_assembler_id"`
	IcIdentifier              int       `tlv:"0002 2 23 int" json:"ic_identifier"`
	CardNumber                string    `tlv:"0520 16 1 string" json:"card_number"`
	CardIssuingMemberState    int       `tlv:"0520 1 0 int" json:"card_issuing_member_state"`
	CardIssuingAuthorityName  string    `tlv:"0520 36 17 string" json:"card_issuing_authority_name"`
	CardIssueDate             time.Time `tlv:"0520 4 53 date" json:"card_issue_date"`
	CardValidityBegin         time.Time `tlv:"0520 4 57 date" json:"card_validity_begin"`
	CardExpiryDate            time.Time `tlv:"0520 4 61 date" json:"card_expiry_date"`
	LastCardDownload          time.Time `tlv:"050E 4 0 date 0" json:"last_card_download"`
	TypeOfTachographCardId    int       `tlv:"0501 1 0 int" json:"type_of_tachograph_card_id"`
	CardStructureVersion      string    `tlv:"0501 2 1 hexadecimal" json:"card_structure_version"`
	CardCertificateGost       string    `tlv:"C200 1000 0 hexadecimal 0" json:"card_certificate_gost,omitempty"`
	CACertificateGost         string    `tlv:"C208 1000 0 hexadecimal 0" json:"ca_certificate_gost,omitempty"`
	CardCertificateESTR       string    `tlv:"C100 194 0 hexadecimal 0" json:"card_certificate_estr,omitempty"`
	CACertificateESTR         string    `tlv:"C108 194 0 hexadecimal 0" json:"ca_certificate_estr,omitempty"`
}

type card struct {
	Card                          cardInfo
	SessionOpen                   sessionOpen
	Driver                        driver
	DLicense                      dlicense
	CardVehicleRecords            сardVehicleRecords
	ActivityDailyRecords          activityDailyRecords
	PlaceRecords                  placeRecords
	CardEventRecords              cardEventRecords
	CardFaultRecords              cardFaultRecords
	CardControlActivityDataRecord cardControlActivityDataRecords
	SpecificConditionRecord       specificConditionRecords
}

func (c *card) ParseFromDDD(ddd []byte) error {
	TlvCardMap, err := extractFieldVals(ddd)
	if err != nil {
		return fmt.Errorf("Parse field error: %v", err)
	}

	if err = loadFields(&c.Card, TlvCardMap); err != nil {
		return fmt.Errorf("Error card info load: %v", err)
	}

	if err = loadFields(&c.SessionOpen, TlvCardMap); err != nil {
		return fmt.Errorf("Error sesion info load: %v", err)
	}

	if err = loadFields(&c.Driver, TlvCardMap); err != nil {
		return fmt.Errorf("Error driver info load: %v", err)
	}

	if err = loadFields(&c.DLicense, TlvCardMap); err != nil {
		return fmt.Errorf("Error dlicense info load: %v", err)
	}

	if err = loadFields(&c.CardEventRecords, TlvCardMap); err != nil {
		return fmt.Errorf("Event record load error: %v", err)
	}

	if err = loadFields(&c.CardFaultRecords, TlvCardMap); err != nil {
		return fmt.Errorf("Fault record load error: %v", err)
	}

	if err = loadFields(&c.CardVehicleRecords, TlvCardMap); err != nil {
		return fmt.Errorf("Vehicle record load error: %v", err)
	}

	if err = loadFields(&c.ActivityDailyRecords, TlvCardMap); err != nil {
		return fmt.Errorf("Activity daily record load error: %v", err)
	}

	if err = loadFields(&c.PlaceRecords, TlvCardMap); err != nil {
		return fmt.Errorf("Place record load error: %v", err)
	}

	if err = loadFields(&c.CardControlActivityDataRecord, TlvCardMap); err != nil {
		return fmt.Errorf("Control activity daily record load error: %v", err)
	}

	if err = loadFields(&c.SpecificConditionRecord, TlvCardMap); err != nil {
		return fmt.Errorf("Specific condition record load error: %v", err)
	}

	return err
}

// метод для экспорта объекта ddd
func (c *card) ExportToJson() (string, error) {
	ddd_json, err := json.Marshal(c)

	return string(ddd_json), err
}
