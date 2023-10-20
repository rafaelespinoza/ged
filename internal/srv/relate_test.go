package srv_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/srv"
)

func TestRelatorRelate(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		tests := []struct {
			Name       string
			InPeople   []*entity.Person
			InP1, InP2 string
			ExpErrMsg  string
		}{
			{
				Name:     "person 1 not found",
				InPeople: []*entity.Person{{ID: "10"}, {ID: "20"}},
				InP1:     "5", InP2: "20",
				ExpErrMsg: "5 not found",
			},
			{
				Name:     "person 2 not found",
				InPeople: []*entity.Person{{ID: "10"}, {ID: "20"}},
				InP1:     "10", InP2: "15",
				ExpErrMsg: "15 not found",
			},
			{
				Name:     "same people",
				InPeople: []*entity.Person{{ID: "10"}, {ID: "20"}},
				InP1:     "10", InP2: "10",
				ExpErrMsg: "same people ids",
			},
			{
				Name:     "unrelated",
				InPeople: []*entity.Person{{ID: "10"}, {ID: "20"}},
				InP1:     "10", InP2: "20",
				ExpErrMsg: "unrelated",
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				_, _, err := srv.NewRelator(test.InPeople).Relate(context.Background(), test.InP1, test.InP2)
				if err == nil {
					t.Fatal("expected non-empty error")
				}
				if !strings.Contains(err.Error(), test.ExpErrMsg) {
					t.Errorf("expected error message (%q) to contain %q", err.Error(), test.ExpErrMsg)
				}
			})
		}
	})

	t.Run("ok", func(t *testing.T) {
		type Testcase struct {
			Name       string
			InPeople   []*entity.Person
			InP1, InP2 string
			Exp1, Exp2 entity.Lineage
			ExpErrMsg  string
		}

		runTest := func(t *testing.T, test Testcase) {
			r1, r2, err := srv.NewRelator(test.InPeople).Relate(context.Background(), test.InP1, test.InP2)
			if err != nil {
				t.Fatalf("expected empty error, got %v", err)
			}

			testRelationship(t, r1, test.Exp1)
			testRelationship(t, r2, test.Exp2)
		}

		var kennedys []*entity.Person
		{
			pathToFile := filepath.Join("..", "..", "testdata", "kennedy.ged")
			file, err := os.Open(filepath.Clean(pathToFile))
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = file.Close() }()
			kennedys, _, err = srv.ParseGedcom(context.Background(), file)
			if err != nil {
				t.Fatal(err)
			}
		}

		// These IDs correspond to the GEDCOM testdata file. Each stanza denotes
		// a generation.
		const (
			jfkGrandfather = "@I44@"

			jfkFather = "@I1@"
			jfkAunt   = "@I56@"

			jfk = "@I0@"  // John F Kennedy
			rfk = "@I21@" // Robert F Kennedy

			jfkJr  = "@I54@" // JFK's son
			rfkJr  = "@I25@" // RFK's son
			arnold = "@I10@" // The Terminator

			jpk3           = "@I70@" // Joseph Patrick Kennedy III (b 1980)
			arnoldDaughter = "@I72@" // Katherine Schwarzenegger (b 1989)
		)

		t.Run(entity.Self.String()+" or "+entity.Sibling.String(), func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     "siblings",
					InPeople: kennedys,
					InP1:     jfk,
					InP2:     rfk,
					Exp1: entity.Lineage{
						Description:        "sibling",
						Type:               entity.Sibling,
						GenerationsRemoved: 0,
					},
					Exp2: entity.Lineage{
						Description:        "sibling",
						Type:               entity.Sibling,
						GenerationsRemoved: 0,
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})

		t.Run(entity.Child.String(), func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     "child",
					InPeople: kennedys,
					InP1:     jfk,
					InP2:     jfkFather,
					Exp1: entity.Lineage{
						Description:        "child",
						Type:               entity.Child,
						GenerationsRemoved: 1,
					},
					Exp2: entity.Lineage{
						Description:        "parent",
						Type:               entity.Parent,
						GenerationsRemoved: -1,
					},
				},
				{
					Name:     "grand child",
					InPeople: kennedys,
					InP1:     jfk,
					InP2:     jfkGrandfather,
					Exp1: entity.Lineage{
						Description:        "grand child",
						Type:               entity.Child,
						GenerationsRemoved: 2,
					},
					Exp2: entity.Lineage{
						Description:        "grand parent",
						Type:               entity.Parent,
						GenerationsRemoved: -2,
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})

		t.Run(entity.Parent.String(), func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     "parent",
					InPeople: kennedys,
					InP1:     jfkFather,
					InP2:     jfk,
					Exp1: entity.Lineage{
						Description:        "parent",
						Type:               entity.Parent,
						GenerationsRemoved: -1,
					},
					Exp2: entity.Lineage{
						Description:        "child",
						Type:               entity.Child,
						GenerationsRemoved: 1,
					},
				},
				{
					Name:     "grand parent",
					InPeople: kennedys,
					InP1:     jfkGrandfather,
					InP2:     jfk,
					Exp1: entity.Lineage{
						Description:        "grand parent",
						Type:               entity.Parent,
						GenerationsRemoved: -2,
					},
					Exp2: entity.Lineage{
						Description:        "grand child",
						Type:               entity.Child,
						GenerationsRemoved: 2,
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})

		t.Run(entity.NieceNephew.String(), func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     "niece/nephew",
					InPeople: kennedys,
					InP1:     jfkJr,
					InP2:     rfk,
					Exp1: entity.Lineage{
						Description:        "niece/nephew",
						Type:               entity.NieceNephew,
						GenerationsRemoved: 1,
					},
					Exp2: entity.Lineage{
						Description:        "aunt/uncle",
						Type:               entity.AuntUncle,
						GenerationsRemoved: -1,
					},
				},
				{
					Name:     "grand niece/nephew",
					InPeople: kennedys,
					InP1:     jfkJr,
					InP2:     jfkAunt,
					Exp1: entity.Lineage{
						Description:        "grand niece/nephew",
						Type:               entity.NieceNephew,
						GenerationsRemoved: 2,
					},
					Exp2: entity.Lineage{
						Description:        "great aunt/uncle",
						Type:               entity.AuntUncle,
						GenerationsRemoved: -2,
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})

		t.Run(entity.AuntUncle.String(), func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     "aunt/uncle",
					InPeople: kennedys,
					InP1:     rfk,
					InP2:     jfkJr,
					Exp1: entity.Lineage{
						Description:        "aunt/uncle",
						Type:               entity.AuntUncle,
						GenerationsRemoved: -1,
					},
					Exp2: entity.Lineage{
						Description:        "niece/nephew",
						Type:               entity.NieceNephew,
						GenerationsRemoved: 1,
					},
				},
				{
					Name:     "grand niece/nephew",
					InPeople: kennedys,
					InP1:     jfkAunt,
					InP2:     jfkJr,
					Exp1: entity.Lineage{
						Description:        "great aunt/uncle",
						Type:               entity.AuntUncle,
						GenerationsRemoved: -2,
					},
					Exp2: entity.Lineage{
						Description:        "grand niece/nephew",
						Type:               entity.NieceNephew,
						GenerationsRemoved: 2,
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})

		t.Run(entity.Cousin.String(), func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     "1st cousin",
					InPeople: kennedys,
					InP1:     jfkJr,
					InP2:     rfkJr,
					Exp1: entity.Lineage{
						Description:        "1st cousin",
						Type:               entity.Cousin,
						GenerationsRemoved: 0,
					},
					Exp2: entity.Lineage{
						Description:        "1st cousin",
						Type:               entity.Cousin,
						GenerationsRemoved: 0,
					},
				},
				{
					Name:     "1st cousin 1x removed, p1 is older",
					InPeople: kennedys,
					InP1:     jfkJr,
					InP2:     jpk3,
					Exp1: entity.Lineage{
						Description:        "1st cousin 1x removed",
						Type:               entity.Cousin,
						GenerationsRemoved: -1,
					},
					Exp2: entity.Lineage{
						Description:        "1st cousin 1x removed",
						Type:               entity.Cousin,
						GenerationsRemoved: 1,
					},
				},
				{
					Name:     "1st cousin 1x removed, p1 is younger",
					InPeople: kennedys,
					InP1:     jpk3,
					InP2:     jfkJr,
					Exp1: entity.Lineage{
						Description:        "1st cousin 1x removed",
						Type:               entity.Cousin,
						GenerationsRemoved: 1,
					},
					Exp2: entity.Lineage{
						Description:        "1st cousin 1x removed",
						Type:               entity.Cousin,
						GenerationsRemoved: -1,
					},
				},
				{
					Name:     "2nd cousin",
					InPeople: kennedys,
					InP1:     jpk3,
					InP2:     arnoldDaughter,
					Exp1: entity.Lineage{
						Description:        "2nd cousin",
						Type:               entity.Cousin,
						GenerationsRemoved: 0,
					},
					Exp2: entity.Lineage{
						Description:        "2nd cousin",
						Type:               entity.Cousin,
						GenerationsRemoved: 0,
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})
	})
}

func testRelationship(t *testing.T, actual, expected entity.Lineage) {
	t.Helper()

	if actual.Description != expected.Description {
		t.Errorf("wrong Description; got %q, exp %q", actual.Description, expected.Description)
	}
	if actual.Type != expected.Type {
		t.Errorf("wrong Type; got %q, exp %q", actual.Type.String(), expected.Type.String())
	}
	if actual.GenerationsRemoved != expected.GenerationsRemoved {
		t.Errorf("wrong GenerationsRemoved; got %d, exp %d", actual.GenerationsRemoved, expected.GenerationsRemoved)
	}

	// The CommonAncestors field is not checked here b/c it can vary if there
	// are multiple paths to the same common ancestor. The current path finding
	// implementation just picks one, and it's a bit non-deterministic due to
	// the way some of the underlying data structures (maps) are traversed.
}

func parseYear(t *testing.T, in string) *time.Time {
	t.Helper()

	if in == "" {
		return nil
	}

	date, err := time.Parse("2006", in)
	if err != nil {
		t.Fatal(err)
	}
	return &date
}
