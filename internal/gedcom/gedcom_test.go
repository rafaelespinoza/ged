package gedcom_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/rafaelespinoza/ged/internal/entity/date"
	"github.com/rafaelespinoza/ged/internal/gedcom"
)

func TestReadRecordsSanityCheck(t *testing.T) {
	// check that it can read data and that outputs are non-empty
	tests := []struct {
		Filename string
		// Unlike Individuals and Families, we're not necessarily expecting for the testdata to have a
		// non-zero number of Sources. Most of this data is fictional. Hence, no need to fail the test if
		// there aren't any Sources.
		ExpectSources bool
	}{
		{"kennedy.ged", true},
		{"game_of_thrones.ged", false},
		{"simpsons.ged", false},
	}

	for _, test := range tests {
		t.Run(test.Filename, func(t *testing.T) {
			pathToFile := filepath.Join("..", "..", "testdata", test.Filename)
			file, err := os.Open(filepath.Clean(pathToFile))
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()

			records, err := gedcom.ReadRecords(context.Background(), file)
			if err != nil {
				t.Fatal(err)
			}

			if len(records.Families) < 1 {
				t.Fatal("expected some families but got 0")
			}

			for _, got := range records.Families {
				if got.Xref == "" {
					t.Fatalf("Xref should be non-empty, %#v", got)
				}

				for j, xref := range got.ParentXrefs {
					if xref == "" {
						t.Fatalf("ParentXrefs[%d] should be non-empty, %#v", j, got)
					}
				}

				for j, xref := range got.ChildXrefs {
					if xref == "" {
						t.Fatalf("ChildXrefs[%d] should be non-empty, %#v", j, got)
					}
				}
			}

			if len(records.Individuals) < 1 {
				t.Fatal("expected some individuals but got 0")
			}

			for i, got := range records.Individuals {
				if got.Xref == "" {
					t.Fatalf("Xref should be non-empty, %#v", got)
				}

				for j, name := range got.Names {
					if name.Payload == "" {
						// If you see a lot of these but can see that the input
						// data does have Name values, then it's a bug. But if
						// the input data doesn't have any Name value in the
						// first place, then it's just incomplete input data. So
						// consider this message a warning, rather than an error.
						t.Logf("[%d].Names[%d].Payload should be non-empty (unless it's empty in data), %#v", i, j, got)
					}
				}

				for j, xref := range got.FamiliesAsPartner {
					if xref == "" {
						t.Fatalf("FamiliesAsPartner[%d] should be non-empty, %#v", j, got)
					}
				}

				for j, xref := range got.FamiliesAsChild {
					if xref == "" {
						t.Fatalf("FamiliesAsChild[%d] should be non-empty, %#v", j, got)
					}
				}
			}

			if test.ExpectSources && len(records.Sources) < 1 {
				t.Fatal("expected some sources but got 0")
			}

			for _, got := range records.Sources {
				if got.Xref == "" {
					t.Fatalf("Xref should be non-empty, %#v", got)
				}
			}
		})
	}
}

func TestReadRecordsFields(t *testing.T) {
	// check actual field values for correct interpretation
	data := strings.NewReader(`0 HEAD
1 CHAR UTF-8
0 @I1@ INDI
1 NAME Charlie /Foxtrot/
2 TYPE Birth
2 GIVN Charlie
2 NICK Chuck
2 SURN Foxtrot
1 BIRT
2 DATE 1 JAN 1970
1 DEAT
2 DATE 19 JAN 2038
1 BURI
2 DATE 13 DEC 1901
2 PLAC The internet
1 EVEN
2 TYPE OOF
2 DATE 13 MAY 2006
2 PLAC AOL
2 NOTE According to the Wikipedia article on the Year 2038 Problem, AOL had a bug related to 2038-01-19.
1 FAMS @F1@
0 @I2@ INDI
1 NAME Charlene /Foxtrot/
2 TYPE Birth
2 GIVN Charlene
2 NICK Y2K22
2 SURN Foxtrot
1 BIRT
2 DATE 1 JAN 1970
1 CHR
2 DATE 2 JAN 1970
1 NATU
2 DATE 9 SEP 1999
3 _APID 1,1629::6961470
1 DEAT
2 DATE 1 JAN 2022
1 FAMS @F1@
0 @I3@ INDI
1 NAME Mike /Foxtrot/
2 TYPE Birth
2 GIVN Mike
2 NICK Millennium Bug
2 SURN Foxtrot
1 BIRT
2 DATE 12 JUN 1995
1 BAPM
2 DATE 13 JUN 1995
2 PLAC The media
1 RESI
2 DATE BETWEEN 1996 AND 2000
2 PLAC The mainstream media
3 _APID 1,6061::66023688
1 DEAT
2 DATE 1 JAN 2000
1 NOTE The year 2000 problem, also commonly known as the Y2K problem, Y2K scare, millennium bug, Y2K bug, Y2K glitch, Y2K error, or simply Y2K,
2 CONT refers to potential computer errors related to the formatting and storage of calendar data for dates in and after the year 2000.
2 LANG en
1 FAMC @F1@
0 @F1@ FAM
1 HUSB @I1@
1 WIFE @I2@
1 CHIL @I3@
1 MARR
2 TYPE marriage
2 DATE 18 JUN 1985
2 SOUR @S1@
3 PAGE front page
1 DIV
2 TYPE divorce
2 DATE 2000
1 ANUL
2 TYPE annulment
2 DATE 2001
1 NOTE Test that the parser can also read th
2 CONC e tag, CONC.
0 @S1@ SOUR
1 _UID 046A3AD191FF4DD3B0693F406E0A7FB87012
1 DATA
1 TITL New York Times, March 4, 1946, pp. 1,3.
1 REPO @R0@
1 CHAN
2 DATE 29 MAY 2018
3 TIME 07:21:10.114
1 NOTE Geneanet Community Trees Index
1 TEXT yes
0 @S2@ SOUR
1 TITL Texas, U.S., Death Certificates, 1903-1982
1 AUTH Ancestry.com
1 PUBL Ancestry.com Operations, Inc.
2 DATE 2013
2 PLAC Lehi, UT, USA
1 _APID 1,2272::0
1 REPO @R1@
`)

	records, err := gedcom.ReadRecords(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Families", func(t *testing.T) {
		expected := []*gedcom.FamilyRecord{
			{
				Xref:        "@F1@",
				ParentXrefs: []string{"@I1@", "@I2@"},
				ChildXrefs:  []string{"@I3@"},
				MarriedAt:   &gedcom.Event{Date: mustParseDate(t, "1985-01-18")},
				DivorcedAt:  &gedcom.Event{Date: mustParseDate(t, "2000-01-01")},
				AnnulledAt:  &gedcom.Event{Date: mustParseDate(t, "2001-01-01")},
				Notes:       []*gedcom.Note{{Payload: "Test that the parser can also read the tag, CONC."}},
			},
		}
		if len(records.Families) != len(expected) {
			t.Fatalf("got %d record(s) but expected %d", len(records.Families), len(expected))
		}

		for i, got := range records.Families {
			errMsgPrefix := fmt.Sprintf("item[%d]", i)
			exp := expected[i]

			if got.Xref != exp.Xref {
				t.Fatalf("%swrong Xref; got %q, exp %q", errMsgPrefix+"; ", got.Xref, exp.Xref)
			}

			cmpStringSlices(t, errMsgPrefix+".ParentXrefs", got.ParentXrefs, exp.ParentXrefs)
			cmpStringSlices(t, errMsgPrefix+".ChildXrefs", got.ChildXrefs, exp.ChildXrefs)
			testNotes(t, errMsgPrefix+".Notes", got.Notes, exp.Notes)
		}
	})

	t.Run("Individuals", func(t *testing.T) {
		expected := []*gedcom.IndividualRecord{
			{
				Xref: "@I1@",
				Names: []gedcom.PersonalName{
					{Payload: "Charlie /Foxtrot/", Given: "Charlie", Nickname: "Chuck", Surname: "Foxtrot"},
				},
				Birth:  []*gedcom.Event{{Date: mustParseDate(t, "1970-01-01"), Type: "Birth"}},
				Death:  []*gedcom.Event{{Date: mustParseDate(t, "2038-01-19"), Type: "Death"}},
				Burial: []*gedcom.Event{{Date: mustParseDate(t, "1901-12-13"), Place: "The internet", Type: "Burial"}},
				Events: []*gedcom.Event{
					{
						Type:  "OOF",
						Date:  mustParseDate(t, "2006-05-13"),
						Place: "AOL",
						Notes: []*gedcom.Note{{Payload: "According to the Wikipedia article on the Year 2038 Problem, AOL had a bug related to 2038-01-19."}},
					},
				},
				FamiliesAsPartner: []string{"@F1@"},
			},
			{
				Xref: "@I2@",
				Names: []gedcom.PersonalName{
					{Payload: "Charlene /Foxtrot/", Given: "Charlene", Nickname: "Y2K22", Surname: "Foxtrot"},
				},
				Birth:             []*gedcom.Event{{Date: mustParseDate(t, "1970-01-01"), Type: "Birth"}},
				Christening:       []*gedcom.Event{{Date: mustParseDate(t, "1970-01-02"), Type: "Christening"}},
				Naturalizations:   []*gedcom.Event{{Date: mustParseDate(t, "1999-09-09"), Type: "Naturalization"}},
				Death:             []*gedcom.Event{{Date: mustParseDate(t, "2022-01-01"), Type: "Death"}},
				FamiliesAsPartner: []string{"@F1@"},
			},
			{
				Xref: "@I3@",
				Names: []gedcom.PersonalName{
					{Payload: "Mike /Foxtrot/", Given: "Mike", Nickname: "Millennium Bug", Surname: "Foxtrot"},
				},
				Birth:      []*gedcom.Event{{Date: mustParseDate(t, "1995-06-12"), Type: "Birth"}},
				Baptism:    []*gedcom.Event{{Date: mustParseDate(t, "1995-06-13"), Place: "The media", Type: "Baptism"}},
				Residences: []*gedcom.Event{{DateRange: &date.Range{Lo: &date.Date{Year: 1996}, Hi: &date.Date{Year: 2000}}, Place: "The mainstream media", Type: "Residence"}},
				Death:      []*gedcom.Event{{Date: mustParseDate(t, "2000-01-01"), Type: "Death"}},
				Notes: []*gedcom.Note{
					{Payload: `The year 2000 problem, also commonly known as the Y2K problem, Y2K scare, millennium bug, Y2K bug, Y2K glitch, Y2K error, or simply Y2K,
refers to potential computer errors related to the formatting and storage of calendar data for dates in and after the year 2000.`,
						Lang: "en"},
				},
				FamiliesAsChild: []string{"@F1@"},
			},
		}
		if len(records.Individuals) != len(expected) {
			t.Fatalf("got %d record(s) but expected %d", len(records.Individuals), len(expected))
		}

		for i, got := range records.Individuals {
			errMsgPrefix := fmt.Sprintf("item[%d]", i)
			exp := expected[i]

			if got.Xref != exp.Xref {
				t.Fatalf("%s; wrong Xref; got %q, exp %q", errMsgPrefix, got.Xref, exp.Xref)
			}

			if len(got.Names) != len(exp.Names) {
				t.Errorf("%s; wrong number of Names; got %d, exp %d", errMsgPrefix, len(got.Names), len(exp.Names))
			} else {
				for j, name := range got.Names {
					errMsgPrefix := fmt.Sprintf("%s.Names[%d]", errMsgPrefix, j)
					cmpPersonalName(t, errMsgPrefix, name, exp.Names[j])
				}
			}

			testEvents(t, errMsgPrefix+".Birth", got.Birth, exp.Birth)
			testEvents(t, errMsgPrefix+".Baptism", got.Baptism, exp.Baptism)
			testEvents(t, errMsgPrefix+".Christening", got.Christening, exp.Christening)
			testEvents(t, errMsgPrefix+".Residences", got.Residences, exp.Residences)
			testEvents(t, errMsgPrefix+".Naturalizations", got.Naturalizations, exp.Naturalizations)
			testEvents(t, errMsgPrefix+".Death", got.Death, exp.Death)
			testEvents(t, errMsgPrefix+".Burial", got.Burial, exp.Burial)
			testEvents(t, errMsgPrefix+".Events", got.Events, exp.Events)
			cmpStringSlices(t, errMsgPrefix+".FamiliesAsPartner", got.FamiliesAsPartner, exp.FamiliesAsPartner)
			cmpStringSlices(t, errMsgPrefix+".FamiliesAsChild", got.FamiliesAsChild, exp.FamiliesAsChild)
			testNotes(t, errMsgPrefix+".Notes", got.Notes, exp.Notes)
		}
	})

	t.Run("Sources", func(t *testing.T) {
		expected := []*gedcom.SourceRecord{
			{
				Xref:          "@S1@",
				Title:         "New York Times, March 4, 1946, pp. 1,3.",
				Notes:         []*gedcom.Note{{Payload: "Geneanet Community Trees Index"}},
				Text:          "yes",
				RepositoryIDs: []string{"@R0@"},
			},
			{
				Xref:          "@S2@",
				Title:         "Texas, U.S., Death Certificates, 1903-1982",
				Author:        "Ancestry.com",
				Publication:   "Ancestry.com Operations, Inc.",
				RepositoryIDs: []string{"@R1@"},
			},
		}
		if len(records.Sources) != len(expected) {
			t.Fatalf("got %d record(s) but expected %d", len(records.Sources), len(expected))
		}

		for i, got := range records.Sources {
			errMsgPrefix := fmt.Sprintf("item[%d]", i)
			exp := expected[i]

			if got.Xref != exp.Xref {
				t.Fatalf("%s; wrong Xref; got %q, exp %q", errMsgPrefix, got.Xref, exp.Xref)
			}

			if got.Title != exp.Title {
				t.Errorf("%s; wrong Title, got %q, exp %q", errMsgPrefix, got.Title, exp.Title)
			}
			if got.Author != exp.Author {
				t.Errorf("%s; wrong Author, got %q, exp %q", errMsgPrefix, got.Author, exp.Author)
			}
			if got.Abbreviation != exp.Abbreviation {
				t.Errorf("%s; wrong Abbreviation, got %q, exp %q", errMsgPrefix, got.Abbreviation, exp.Abbreviation)
			}
			if got.Publication != exp.Publication {
				t.Errorf("%s; wrong Publication, got %q, exp %q", errMsgPrefix, got.Publication, exp.Publication)
			}
			if got.Text != exp.Text {
				t.Errorf("%s; wrong Text, got %q, exp %q", errMsgPrefix, got.Text, exp.Text)
			}

			cmpStringSlices(t, errMsgPrefix+".RepositoryIDs", got.RepositoryIDs, exp.RepositoryIDs)
			testNotes(t, errMsgPrefix+".Notes", got.Notes, exp.Notes)
		}
	})
}

func mustParseDate(t *testing.T, d string) *date.Date {
	t.Helper()

	dat, err := time.Parse(time.DateOnly, d)
	if err != nil {
		t.Fatal(err)
	}
	year, month, day := dat.Date()
	return &date.Date{Year: year, Month: month, Day: day}
}

func cmpStringSlices(t *testing.T, errMsgPrefix string, actual, expected []string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf("%swrong length; got %d, exp %d", errMsgPrefix, len(actual), len(expected))
	} else {
		for i, got := range actual {
			exp := expected[i]

			if got != exp {
				// Q: why isn't this test helper function going full ham on generics?
				// A: 1) there aren't any types to compare other than []string
				// right now. 2) error message formatting is compromised b/c you
				// cannot use the print directive %q unless the underlying type
				// is string. That directive is valuable for showing any extra
				// whitespace on either end of the value.
				t.Errorf("%s[%d]; got %q, exp %q", errMsgPrefix, i, exp, exp)
			}
		}
	}
}

func cmpStringMaps(t *testing.T, errMsgPrefix string, actual, expected map[string]string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf("%swrong number of keys; got %d, exp %d", errMsgPrefix, len(actual), len(expected))
	}

	var ok bool

	// compare keys
	missingExpectedKeys := make([]string, 0, len(expected))
	for key := range expected {
		if _, ok = actual[key]; !ok {
			missingExpectedKeys = append(missingExpectedKeys, key)
		}
	}
	missingExpectedKeys = slices.Clip(missingExpectedKeys)

	if len(missingExpectedKeys) > 0 {
		slices.Sort(missingExpectedKeys)
		t.Errorf("%smissing %d expected keys %q", errMsgPrefix, len(missingExpectedKeys), missingExpectedKeys)
	}

	unexpectedKeys := make([]string, 0, len(actual))
	for key := range actual {
		_, ok = expected[key]
		if !ok {
			unexpectedKeys = append(unexpectedKeys, key)
		}
	}
	unexpectedKeys = slices.Clip(unexpectedKeys)

	if len(unexpectedKeys) > 0 {
		slices.Sort(unexpectedKeys)
		t.Errorf("%sgot %d expected key(s) %q", errMsgPrefix, len(unexpectedKeys), unexpectedKeys)
	}

	// compare values
	for key, got := range actual {
		exp, ok := expected[key]
		if !ok {
			continue
		}

		if got != exp {
			t.Errorf("%s[%v]; got %q, exp %q", errMsgPrefix, key, got, exp)
		}
	}
}

func testEvents(t *testing.T, errMsgPrefix string, actual, expected []*gedcom.Event) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf("%s; wrong length; got %d, exp %d", errMsgPrefix, len(actual), len(expected))
	} else {
		for i, got := range actual {
			errMsgPrefix := fmt.Sprintf("%s[%d]", errMsgPrefix, i)
			exp := expected[i]

			if got.Type != exp.Type {
				t.Errorf("%s; wrong Type, got %q, exp %q", errMsgPrefix, got.Type, exp.Type)
			}

			testDate(t, errMsgPrefix+".Date", got.Date, exp.Date)
			testDateRange(t, errMsgPrefix+".DateRange", got.DateRange, exp.DateRange)

			if got.Place != exp.Place {
				t.Errorf("%s; wrong Place, got %q, exp %q", errMsgPrefix, got.Place, exp.Place)
			}

			if len(got.SourceCitations) != len(exp.SourceCitations) {
				t.Errorf("%s; wrong number of SourceCitations; got %d, exp %d", errMsgPrefix, len(got.SourceCitations), len(exp.SourceCitations))
			} else {
				for j, got := range got.SourceCitations {
					errMsgPrefix := fmt.Sprintf("%s[%d], ", errMsgPrefix, j)
					cmpSourceCitation(t, errMsgPrefix, got, exp.SourceCitations[j])
				}
			}
			testNotes(t, errMsgPrefix+".Notes", got.Notes, exp.Notes)
		}
	}
}

func testDate(t *testing.T, errMsgPrefix string, actual, expected *date.Date) {
	t.Helper()

	if actual == nil && expected == nil {
		return
	} else if actual != nil && expected == nil {
		t.Errorf("%s; got non-empty value, but expected %v", errMsgPrefix, expected)
		return
	} else if actual == nil && expected != nil {
		t.Errorf("%s; expected non-empty value, but got %v", errMsgPrefix, actual)
		return
	}

	if actual.Year != expected.Year {
		t.Errorf("%s; wrong Year, got %d, exp %d", errMsgPrefix, actual.Year, expected.Year)
	}
	if actual.Month != expected.Month {
		t.Errorf("%s; wrong Month, got %d, exp %d", errMsgPrefix, actual.Month, expected.Month)
	}
	if actual.Day != expected.Day {
		t.Errorf("%s; wrong Day, got %d, exp %d", errMsgPrefix, actual.Day, expected.Day)
	}
}

func testDateRange(t *testing.T, errMsgPrefix string, actual, expected *date.Range) {
	t.Helper()

	if actual == nil && expected == nil {
		return
	} else if actual != nil && expected == nil {
		t.Errorf("%s; got non-empty value, but expected %v", errMsgPrefix, expected)
		return
	} else if actual == nil && expected != nil {
		t.Errorf("%s; expected non-empty value, but got %v", errMsgPrefix, actual)
		return
	}

	testDate(t, errMsgPrefix+".Lo", actual.Lo, expected.Lo)
	testDate(t, errMsgPrefix+".Hi", actual.Hi, expected.Hi)
}

func testNotes(t *testing.T, errMsgPrefix string, actual, expected []*gedcom.Note) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf("%s; wrong length; got %d, exp %d", errMsgPrefix, len(actual), len(expected))
	} else {
		for i, got := range actual {
			errMsgPrefix := fmt.Sprintf("%s[%d]", errMsgPrefix, i)
			exp := expected[i]

			if got.Payload != exp.Payload {
				t.Errorf("%s; wrong Payload, got %q, exp %q", errMsgPrefix, got.Payload, exp.Payload)
			}
			if got.Lang != exp.Lang {
				t.Errorf("%s; wrong Lang, got %q, exp %q", errMsgPrefix, got.Lang, exp.Lang)
			}

			if len(got.SourceCitations) != len(exp.SourceCitations) {
				t.Errorf("%s; wrong number of SourceCitations; got %d, exp %d", errMsgPrefix, len(got.SourceCitations), len(exp.SourceCitations))
			} else {
				for j, got := range got.SourceCitations {
					errMsgPrefix := fmt.Sprintf("%s[%d], ", errMsgPrefix, j)
					cmpSourceCitation(t, errMsgPrefix, got, exp.SourceCitations[j])
				}
			}
		}
	}
}
