package gedcom_test

import (
	"testing"

	"github.com/rafaelespinoza/ged/internal/entity/date"
	"github.com/rafaelespinoza/ged/internal/gedcom"
)

func TestIndividualRecordEventLog(t *testing.T) {
	indRecord := gedcom.IndividualRecord{
		Birth: []*gedcom.Event{
			{DateRange: &date.Range{Lo: &date.Date{Year: 1900}, Hi: &date.Date{Year: 1903}}, Type: "Birth"},
		},
		Residences: []*gedcom.Event{
			{DateRange: &date.Range{Lo: &date.Date{Year: 1910}, Hi: &date.Date{Year: 1940}}, Place: "Springfield", Type: "Residence"},
			{DateRange: &date.Range{Lo: &date.Date{Year: 1945}, Hi: &date.Date{Year: 1970}}, Place: "Shelbyville", Type: "Residence"},
		},
		Naturalizations: []*gedcom.Event{
			{Date: mustParseDate(t, "1925-04-30"), Type: "Naturalization"},
			{Date: mustParseDate(t, "1931-02-10"), Type: "Naturalization"},
		},
		Death: []*gedcom.Event{
			{Date: mustParseDate(t, "1999-01-01"), Type: "Death"},
		},
		Events: []*gedcom.Event{
			{Date: mustParseDate(t, "1919-04-30"), Type: "Misc"},
			{DateRange: &date.Range{Lo: &date.Date{Year: 1942}, Hi: &date.Date{Year: 1945}}, Type: "Misc"},
			{DateRange: &date.Range{Lo: &date.Date{Year: 1958}, Hi: &date.Date{Year: 1973}}, Type: "Misc"},
			{Date: mustParseDate(t, "1980-12-10"), Type: "Misc"},
		},
	}

	results := indRecord.EventLog()
	if len(results) != 10 {
		t.Fatalf("wrong number of results, got %d, exp %d", len(results), 10)
	}
}
