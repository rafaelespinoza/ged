package gedcom_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rafaelespinoza/ged/internal/gedcom"
)

func TestReadRecordsSanityCheck(t *testing.T) {
	// check that it can read data and that outputs are non-empty
	for _, testFilename := range []string{"kennedy.ged", "game_of_thrones.ged"} {
		t.Run(testFilename, func(t *testing.T) {
			pathToFile := filepath.Join("..", "..", "testdata", testFilename)
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
1 FAMS @F1@
0 @I2@ INDI
1 NAME Charlene /Foxtrot/
2 TYPE Birth
2 GIVN Charlene
2 NICK Y2K22
2 SURN Foxtrot
1 BIRT
2 DATE 1 JAN 1970
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
1 DEAT
2 DATE 1 JAN 2000
1 FAMC @F1@
0 @F1@ FAM
1 HUSB @I1@
1 WIFE @I2@
1 CHIL @I3@
1 MARR
2 TYPE marriage
2 DATE 18 JUN 1985
1 DIV
2 TYPE divorce
2 DATE 2000
1 ANUL
2 TYPE annulment
2 DATE 2001
`)

	records, err := gedcom.ReadRecords(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}

	expFamilies := []*gedcom.FamilyRecord{
		{
			Xref:        "@F1@",
			ParentXrefs: []string{"@I1@", "@I2@"},
			ChildXrefs:  []string{"@I3@"},
			MarriedAt:   &gedcom.Event{Date: mustParseDate(t, "1985-01-18")},
			DivorcedAt:  &gedcom.Event{Date: mustParseDate(t, "2000-01-01")},
			AnnulledAt:  &gedcom.Event{Date: mustParseDate(t, "2001-01-01")},
		},
	}
	if len(records.Families) != len(expFamilies) {
		t.Fatalf("got %d record(s) but expected %d", len(records.Families), len(expFamilies))
	}

	for i, got := range records.Families {
		exp := expFamilies[i]
		if got.Xref != exp.Xref {
			t.Fatalf("item[%d]; wrong Xref; got %q, exp %q", i, got.Xref, exp.Xref)
		}

		if len(got.ParentXrefs) != len(exp.ParentXrefs) {
			t.Errorf("item[%d]; wrong number of ParentXrefs; got %d, exp %d", i, len(got.ParentXrefs), len(exp.ParentXrefs))
		} else {
			for j, xref := range got.ParentXrefs {
				if xref != exp.ParentXrefs[j] {
					t.Errorf("item[%d].ParentXrefs[%d] wrong value, got %q, exp %q", i, j, xref, exp.ParentXrefs[j])
				}
			}
		}

		if len(got.ChildXrefs) != len(exp.ChildXrefs) {
			t.Errorf("item[%d]; wrong number of ChildXrefs; got %d, exp %d", i, len(got.ChildXrefs), len(exp.ChildXrefs))
		} else {
			for j, xref := range got.ChildXrefs {
				if xref != exp.ChildXrefs[j] {
					t.Errorf("item[%d].ChildXrefs[%d] wrong value, got %q, exp %q", i, j, xref, exp.ChildXrefs[j])
				}
			}
		}
	}

	expIndividuals := []*gedcom.IndividualRecord{
		{
			Xref: "@I1@",
			Names: []gedcom.PersonalName{
				{Payload: "Charlie /Foxtrot/", Given: "Charlie", Nickname: "Chuck", Surname: "Foxtrot"},
			},
			Birth:             &gedcom.Event{Date: mustParseDate(t, "1970-01-01")},
			Death:             &gedcom.Event{Date: mustParseDate(t, "2038-01-19")},
			FamiliesAsPartner: []string{"@F1@"},
		},
		{
			Xref: "@I2@",
			Names: []gedcom.PersonalName{
				{Payload: "Charlene /Foxtrot/", Given: "Charlene", Nickname: "Y2K22", Surname: "Foxtrot"},
			},
			Birth:             &gedcom.Event{Date: mustParseDate(t, "1970-01-01")},
			Death:             &gedcom.Event{Date: mustParseDate(t, "2022-01-01")},
			FamiliesAsPartner: []string{"@F1@"},
		},
		{
			Xref: "@I3@",
			Names: []gedcom.PersonalName{
				{Payload: "Mike /Foxtrot/", Given: "Mike", Nickname: "Millennium Bug", Surname: "Foxtrot"},
			},
			Birth:           &gedcom.Event{Date: mustParseDate(t, "1995-06-12")},
			Death:           &gedcom.Event{Date: mustParseDate(t, "2000-01-01")},
			FamiliesAsChild: []string{"@F1@"},
		},
	}
	if len(records.Individuals) != len(expIndividuals) {
		t.Fatalf("got %d record(s) but expected %d", len(records.Individuals), len(expIndividuals))
	}

	for i, got := range records.Individuals {
		exp := expIndividuals[i]
		if got.Xref != exp.Xref {
			t.Fatalf("item[%d]; wrong Xref; got %q, exp %q", i, got.Xref, exp.Xref)
		}

		if len(got.Names) != len(exp.Names) {
			t.Errorf("item[%d]; wrong number of Names; got %d, exp %d", i, len(got.Names), len(exp.Names))
		} else {
			for j, name := range got.Names {
				if name.Payload != exp.Names[j].Payload {
					t.Errorf("item[%d].Names[%d] wrong Payload, got %q, exp %q", i, j, name.Payload, exp.Names[j].Payload)
				}
				if name.NamePrefix != exp.Names[j].NamePrefix {
					t.Errorf("item[%d].Names[%d] wrong NamePrefix, got %q, exp %q", i, j, name.NamePrefix, exp.Names[j].NamePrefix)
				}
				if name.Given != exp.Names[j].Given {
					t.Errorf("item[%d].Names[%d] wrong Given, got %q, exp %q", i, j, name.Given, exp.Names[j].Given)
				}
				if name.Nickname != exp.Names[j].Nickname {
					t.Errorf("item[%d].Names[%d] wrong Nickname, got %q, exp %q", i, j, name.Nickname, exp.Names[j].Nickname)
				}
				if name.SurnamePrefix != exp.Names[j].SurnamePrefix {
					t.Errorf("item[%d].Names[%d] wrong SurnamePrefix, got %q, exp %q", i, j, name.SurnamePrefix, exp.Names[j].SurnamePrefix)
				}
				if name.Surname != exp.Names[j].Surname {
					t.Errorf("item[%d].Names[%d] wrong Surname, got %q, exp %q", i, j, name.Surname, exp.Names[j].Surname)
				}
				if name.NameSuffix != exp.Names[j].NameSuffix {
					t.Errorf("item[%d].Names[%d] wrong NameSuffix, got %q, exp %q", i, j, name.NameSuffix, exp.Names[j].NameSuffix)
				}
			}
		}

		if len(got.FamiliesAsPartner) != len(exp.FamiliesAsPartner) {
			t.Errorf("item[%d]; wrong number of FamiliesAsPartner; got %d, exp %d", i, len(got.FamiliesAsPartner), len(exp.FamiliesAsPartner))
		} else {
			for j, xref := range got.FamiliesAsPartner {
				if xref == "" {
					t.Errorf("item[%d].Names[%d] wrong value, got %q, exp %q", i, j, xref, exp.FamiliesAsPartner[j])
				}
			}
		}

		if len(got.FamiliesAsChild) != len(exp.FamiliesAsChild) {
			t.Errorf("item[%d]; wrong number of FamiliesAsChild; got %d, exp %d", i, len(got.FamiliesAsChild), len(exp.FamiliesAsChild))
		} else {
			for j, xref := range got.FamiliesAsChild {
				if xref == "" {
					t.Errorf("item[%d].Names[%d] wrong value, got %q, exp %q", i, j, xref, exp.FamiliesAsChild[j])
				}
			}
		}
	}
}

func mustParseDate(t *testing.T, d string) *time.Time {
	t.Helper()

	out, err := time.Parse(time.DateOnly, d)
	if err != nil {
		t.Fatal(err)
	}
	return &out
}
