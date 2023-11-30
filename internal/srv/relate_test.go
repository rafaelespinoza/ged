package srv_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/srv"
)

// These IDs correspond to the GEDCOM testdata file. Each stanza denotes
// a generation.
const (
	jfkGrandfather = "@I44@"
	jfkGrandmother = "@I45@"

	jfkFather = "@I1@"
	jfkMother = "@I2@"
	jfkAunt   = "@I56@"

	jfk       = "@I0@"  // John F Kennedy
	rfk       = "@I21@" // Robert F Kennedy
	jfkWife   = "@I52@"
	jfkSister = "@I8@" // Eunice Mary Kennedy (b 1915, d 2011)

	jfkJr  = "@I54@" // JFK's son
	rfkJr  = "@I25@" // RFK's son
	jpk2   = "@I24@" // Joseph Patrick Kennedy II (b 1952)
	mariaS = "@I11@" // Maria Owings Shriver (b 1955)
	arnold = "@I10@" // The Terminator

	jpk3           = "@I70@" // Joseph Patrick Kennedy III (b 1980)
	arnoldDaughter = "@I72@" // Katherine Schwarzenegger (b 1989)
)

func buildKennedyFamily(t *testing.T) []*entity.Person {
	var kennedys []*entity.Person

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

	return kennedys
}

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
				Name:     "unrelated",
				InPeople: []*entity.Person{{ID: "10"}, {ID: "20"}},
				InP1:     "10", InP2: "20",
				ExpErrMsg: "unrelated",
			},
		}

		for _, test := range tests {
			t.Run(test.Name, func(t *testing.T) {
				_, err := srv.NewRelator(test.InPeople).Relate(context.Background(), test.InP1, test.InP2)
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
			Exp        entity.MutualRelationship
		}

		runTest := func(t *testing.T, test Testcase) {
			actual, err := srv.NewRelator(test.InPeople).Relate(context.Background(), test.InP1, test.InP2)
			if err != nil {
				t.Fatalf("expected empty error, got %v", err)
			}

			testPerson(t, ".CommonPerson", actual.CommonPerson, test.Exp.CommonPerson)
			testUnion(t, ".Union", actual.Union, test.Exp.Union)

			testRelationship(t, ".R1", actual.R1, test.Exp.R1)
			testRelationship(t, ".R2", actual.R2, test.Exp.R2)
		}

		kennedys := buildKennedyFamily(t)

		t.Run(entity.Self.String(), func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     "self",
					InPeople: kennedys,
					InP1:     jfk,
					InP2:     jfk,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfk},
						R1: entity.Relationship{
							Description:        "self",
							Type:               entity.Self,
							SourceID:           jfk,
							TargetID:           jfk,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jfk}},
						},
						R2: entity.Relationship{
							Description:        "self",
							Type:               entity.Self,
							SourceID:           jfk,
							TargetID:           jfk,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jfk}},
						},
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})

		t.Run(entity.Sibling.String(), func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     "siblings",
					InPeople: kennedys,
					InP1:     jfk,
					InP2:     rfk,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkMother},
						R1: entity.Relationship{
							Description:        "sibling",
							Type:               entity.Sibling,
							SourceID:           jfk,
							TargetID:           rfk,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jfk}, {ID: jfkMother}},
						},
						R2: entity.Relationship{
							Description:        "sibling",
							Type:               entity.Sibling,
							SourceID:           rfk,
							TargetID:           jfk,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: rfk}, {ID: jfkMother}},
						},
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
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkFather},
						R1: entity.Relationship{
							Description:        "child",
							Type:               entity.Child,
							SourceID:           jfk,
							TargetID:           jfkFather,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: jfk}, {ID: jfkFather}},
						},
						R2: entity.Relationship{
							Description:        "parent",
							Type:               entity.Parent,
							SourceID:           jfkFather,
							TargetID:           jfk,
							GenerationsRemoved: -1,
							Path:               []entity.Person{{ID: jfkFather}},
						},
					},
				},
				{
					Name:     "grand child",
					InPeople: kennedys,
					InP1:     jfk,
					InP2:     jfkGrandfather,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkGrandfather},
						R1: entity.Relationship{
							Description:        "grand child",
							Type:               entity.Child,
							SourceID:           jfk,
							TargetID:           jfkGrandfather,
							GenerationsRemoved: 2,
							Path:               []entity.Person{{ID: jfk}, {ID: jfkFather}, {ID: jfkGrandfather}},
						},
						R2: entity.Relationship{
							Description:        "grand parent",
							Type:               entity.Parent,
							SourceID:           jfkGrandfather,
							TargetID:           jfk,
							GenerationsRemoved: -2,
							Path:               []entity.Person{{ID: jfkGrandfather}},
						},
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
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkFather},
						R1: entity.Relationship{
							Description:        "parent",
							Type:               entity.Parent,
							SourceID:           jfkFather,
							TargetID:           jfk,
							GenerationsRemoved: -1,
							Path:               []entity.Person{{ID: jfkFather}},
						},
						R2: entity.Relationship{
							Description:        "child",
							Type:               entity.Child,
							SourceID:           jfk,
							TargetID:           jfkFather,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: jfk}, {ID: jfkFather}},
						},
					},
				},
				{
					Name:     "grand parent",
					InPeople: kennedys,
					InP1:     jfkGrandfather,
					InP2:     jfk,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkGrandfather},
						R1: entity.Relationship{
							Description:        "grand parent",
							Type:               entity.Parent,
							SourceID:           jfkGrandfather,
							TargetID:           jfk,
							GenerationsRemoved: -2,
							Path:               []entity.Person{{ID: jfkGrandfather}},
						},
						R2: entity.Relationship{
							Description:        "grand child",
							Type:               entity.Child,
							SourceID:           jfk,
							TargetID:           jfkGrandfather,
							GenerationsRemoved: 2,
							Path:               []entity.Person{{ID: jfk}, {ID: jfkFather}, {ID: jfkGrandfather}},
						},
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
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkMother},
						R1: entity.Relationship{
							Description:        "niece/nephew",
							Type:               entity.NieceNephew,
							SourceID:           jfkJr,
							TargetID:           rfk,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: jfkJr}, {ID: jfk}, {ID: jfkMother}},
						},
						R2: entity.Relationship{
							Description:        "aunt/uncle",
							Type:               entity.AuntUncle,
							SourceID:           rfk,
							TargetID:           jfkJr,
							GenerationsRemoved: -1,
							Path:               []entity.Person{{ID: rfk}, {ID: jfkMother}},
						},
					},
				},
				{
					Name:     "grand niece/nephew",
					InPeople: kennedys,
					InP1:     jfkJr,
					InP2:     jfkAunt,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkGrandmother},
						R1: entity.Relationship{
							Description:        "grand niece/nephew",
							Type:               entity.NieceNephew,
							SourceID:           jfkJr,
							TargetID:           jfkAunt,
							GenerationsRemoved: 2,
							Path:               []entity.Person{{ID: jfkJr}, {ID: jfk}, {ID: jfkFather}, {ID: jfkGrandmother}},
						},
						R2: entity.Relationship{
							Description:        "great aunt/uncle",
							Type:               entity.AuntUncle,
							SourceID:           jfkAunt,
							TargetID:           jfkJr,
							GenerationsRemoved: -2,
							Path:               []entity.Person{{ID: jfkAunt}, {ID: jfkGrandmother}},
						},
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
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkMother},
						R1: entity.Relationship{
							Description:        "aunt/uncle",
							Type:               entity.AuntUncle,
							SourceID:           rfk,
							TargetID:           jfkJr,
							GenerationsRemoved: -1,
							Path:               []entity.Person{{ID: rfk}, {ID: jfkMother}},
						},
						R2: entity.Relationship{
							Description:        "niece/nephew",
							Type:               entity.NieceNephew,
							SourceID:           jfkJr,
							TargetID:           rfk,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: jfkJr}, {ID: jfk}, {ID: jfkMother}},
						},
					},
				},
				{
					Name:     "grand niece/nephew",
					InPeople: kennedys,
					InP1:     jfkAunt,
					InP2:     jfkJr,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkGrandmother},
						R1: entity.Relationship{
							Description:        "great aunt/uncle",
							Type:               entity.AuntUncle,
							SourceID:           jfkAunt,
							TargetID:           jfkJr,
							GenerationsRemoved: -2,
							Path:               []entity.Person{{ID: jfkAunt}, {ID: jfkGrandmother}},
						},
						R2: entity.Relationship{
							Description:        "grand niece/nephew",
							Type:               entity.NieceNephew,
							SourceID:           jfkJr,
							TargetID:           jfkAunt,
							GenerationsRemoved: 2,
							Path:               []entity.Person{{ID: jfkJr}, {ID: jfk}, {ID: jfkFather}, {ID: jfkGrandmother}},
						},
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
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkMother},
						R1: entity.Relationship{
							Description:        "1st cousin",
							Type:               entity.Cousin,
							SourceID:           jfkJr,
							TargetID:           rfkJr,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jfkJr}, {ID: jfk}, {ID: jfkMother}},
						},
						R2: entity.Relationship{
							Description:        "1st cousin",
							Type:               entity.Cousin,
							SourceID:           rfkJr,
							TargetID:           jfkJr,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: rfkJr}, {ID: rfk}, {ID: jfkMother}},
						},
					},
				},
				{
					Name:     "1st cousin 1x removed, p1 is older",
					InPeople: kennedys,
					InP1:     jfkJr,
					InP2:     jpk3,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkMother},
						R1: entity.Relationship{
							Description:        "1st cousin 1x removed",
							Type:               entity.Cousin,
							SourceID:           jfkJr,
							TargetID:           jpk3,
							GenerationsRemoved: -1,
							Path:               []entity.Person{{ID: jfkJr}, {ID: jfk}, {ID: jfkMother}},
						},
						R2: entity.Relationship{
							Description:        "1st cousin 1x removed",
							Type:               entity.Cousin,
							SourceID:           jpk3,
							TargetID:           jfkJr,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: jpk3}, {ID: jpk2}, {ID: rfk}, {ID: jfkMother}},
						},
					},
				},
				{
					Name:     "1st cousin 1x removed, p1 is younger",
					InPeople: kennedys,
					InP1:     jpk3,
					InP2:     jfkJr,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkMother},
						R1: entity.Relationship{
							Description:        "1st cousin 1x removed",
							Type:               entity.Cousin,
							SourceID:           jpk3,
							TargetID:           jfkJr,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: jpk3}, {ID: jpk2}, {ID: rfk}, {ID: jfkMother}},
						},
						R2: entity.Relationship{
							Description:        "1st cousin 1x removed",
							Type:               entity.Cousin,
							SourceID:           jfkJr,
							TargetID:           jpk3,
							GenerationsRemoved: -1,
							Path:               []entity.Person{{ID: jfkJr}, {ID: jfk}, {ID: jfkMother}},
						},
					},
				},
				{
					Name:     "2nd cousin",
					InPeople: kennedys,
					InP1:     jpk3,
					InP2:     arnoldDaughter,
					Exp: entity.MutualRelationship{
						CommonPerson: &entity.Person{ID: jfkMother},
						R1: entity.Relationship{
							Description:        "2nd cousin",
							Type:               entity.Cousin,
							SourceID:           jpk3,
							TargetID:           arnoldDaughter,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jpk3}, {ID: jpk2}, {ID: rfk}, {ID: jfkMother}},
						},
						R2: entity.Relationship{
							Description:        "2nd cousin",
							Type:               entity.Cousin,
							SourceID:           arnoldDaughter,
							TargetID:           jpk3,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: arnoldDaughter}, {ID: mariaS}, {ID: jfkSister}, {ID: jfkMother}},
						},
					},
				},
			}

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})

		t.Run("affinal", func(t *testing.T) {
			tests := []Testcase{
				{
					Name:     entity.Spouse.String(),
					InPeople: kennedys,
					InP1:     jfk,
					InP2:     jfkWife,
					Exp: entity.MutualRelationship{
						Union: &entity.Union{
							Person1: &entity.Person{ID: jfk},
							Person2: &entity.Person{ID: jfkWife},
						},
						R1: entity.Relationship{
							Description:        "spouse",
							Type:               entity.Spouse,
							SourceID:           jfk,
							TargetID:           jfkWife,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jfk}},
						},
						R2: entity.Relationship{
							Description:        "spouse",
							Type:               entity.Spouse,
							SourceID:           jfkWife,
							TargetID:           jfk,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jfkWife}},
						},
					},
				},
				{
					Name:     entity.SiblingInLaw.String(),
					InPeople: kennedys,
					InP1:     jfkWife,
					InP2:     rfk,
					Exp: entity.MutualRelationship{
						Union: &entity.Union{
							Person1: &entity.Person{ID: jfk},
							Person2: &entity.Person{ID: jfkWife},
						},
						R1: entity.Relationship{
							Description:        "sibling in-law",
							Type:               entity.SiblingInLaw,
							SourceID:           jfkWife,
							TargetID:           rfk,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jfkWife}, {ID: jfk}},
						},
						R2: entity.Relationship{
							Description:        "sibling in-law",
							Type:               entity.SiblingInLaw,
							SourceID:           rfk,
							TargetID:           jfkWife,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: rfk}, {ID: jfkMother}, {ID: jfk}},
						},
					},
				},
				{
					Name:     entity.SiblingInLaw.String() + " reverse order",
					InPeople: kennedys,
					InP1:     rfk,
					InP2:     jfkWife,
					Exp: entity.MutualRelationship{
						Union: &entity.Union{
							Person1: &entity.Person{ID: jfk},
							Person2: &entity.Person{ID: jfkWife},
						},
						R1: entity.Relationship{
							Description:        "sibling in-law",
							Type:               entity.SiblingInLaw,
							SourceID:           rfk,
							TargetID:           jfkWife,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: rfk}, {ID: jfkMother}, {ID: jfk}},
						},
						R2: entity.Relationship{
							Description:        "sibling in-law",
							Type:               entity.SiblingInLaw,
							SourceID:           jfkWife,
							TargetID:           rfk,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: jfkWife}, {ID: jfk}},
						},
					},
				},
				{
					Name:     entity.ChildInLaw.String(),
					InPeople: kennedys,
					InP1:     jfkWife,
					InP2:     jfkFather,
					Exp: entity.MutualRelationship{
						Union: &entity.Union{
							Person1: &entity.Person{ID: jfk},
							Person2: &entity.Person{ID: jfkWife},
						},
						R1: entity.Relationship{
							Description:        "child in-law",
							Type:               entity.ChildInLaw,
							SourceID:           jfkWife,
							TargetID:           jfkFather,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: jfkWife}, {ID: jfk}},
						},
						R2: entity.Relationship{
							Description:        "parent in-law",
							Type:               entity.ParentInLaw,
							SourceID:           jfkFather,
							TargetID:           jfkWife,
							GenerationsRemoved: -1,
							Path:               []entity.Person{{ID: jfkFather}, {ID: jfk}},
						},
					},
				},
				{
					Name:     entity.ParentInLaw.String(),
					InPeople: kennedys,
					InP1:     jfkFather,
					InP2:     jfkWife,
					Exp: entity.MutualRelationship{
						Union: &entity.Union{
							Person1: &entity.Person{ID: jfk},
							Person2: &entity.Person{ID: jfkWife},
						},
						R1: entity.Relationship{
							Description:        "parent in-law",
							Type:               entity.ParentInLaw,
							SourceID:           jfkFather,
							TargetID:           jfkWife,
							GenerationsRemoved: -1,
							Path:               []entity.Person{{ID: jfkFather}, {ID: jfk}},
						},
						R2: entity.Relationship{
							Description:        "child in-law",
							Type:               entity.ChildInLaw,
							SourceID:           jfkWife,
							TargetID:           jfkFather,
							Path:               []entity.Person{{ID: jfkWife}, {ID: jfk}},
							GenerationsRemoved: 1,
						},
					},
				},
				{
					Name:     entity.AuntUncleInLaw.String(),
					InPeople: kennedys,
					InP1:     jfk,
					InP2:     arnold,
					Exp: entity.MutualRelationship{
						Union: &entity.Union{
							Person1: &entity.Person{ID: mariaS},
							Person2: &entity.Person{ID: arnold},
						},
						R1: entity.Relationship{
							Description:        "aunt/uncle in-law",
							Type:               entity.AuntUncleInLaw,
							SourceID:           jfk,
							TargetID:           arnold,
							Path:               []entity.Person{{ID: jfk}, {ID: jfkMother}, {ID: jfkSister}, {ID: mariaS}},
							GenerationsRemoved: -1,
						},
						R2: entity.Relationship{
							Description:        "niece/nephew in-law",
							Type:               entity.NieceNephewInLaw,
							SourceID:           arnold,
							TargetID:           jfk,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: arnold}, {ID: mariaS}},
						},
					},
				},
				{
					Name:     entity.CousinInLaw.String(),
					InPeople: kennedys,
					InP1:     arnold,
					InP2:     jfkJr,
					Exp: entity.MutualRelationship{
						Union: &entity.Union{
							Person1: &entity.Person{ID: mariaS},
							Person2: &entity.Person{ID: arnold},
						},
						R1: entity.Relationship{
							Description:        "1st cousin in-law",
							Type:               entity.CousinInLaw,
							SourceID:           arnold,
							TargetID:           jfkJr,
							GenerationsRemoved: 0,
							Path:               []entity.Person{{ID: arnold}, {ID: mariaS}},
						},
						R2: entity.Relationship{
							Description:        "1st cousin in-law",
							Type:               entity.CousinInLaw,
							SourceID:           jfkJr,
							TargetID:           arnold,
							Path:               []entity.Person{{ID: jfkJr}, {ID: jfk}, {ID: jfkMother}, {ID: jfkSister}, {ID: mariaS}},
							GenerationsRemoved: 0,
						},
					},
				},
				{
					Name:     entity.NieceNephewInLaw.String(),
					InPeople: kennedys,
					InP1:     arnold,
					InP2:     jfk,
					Exp: entity.MutualRelationship{
						Union: &entity.Union{
							Person1: &entity.Person{ID: mariaS},
							Person2: &entity.Person{ID: arnold},
						},
						R1: entity.Relationship{
							Description:        "niece/nephew in-law",
							Type:               entity.NieceNephewInLaw,
							SourceID:           arnold,
							TargetID:           jfk,
							GenerationsRemoved: 1,
							Path:               []entity.Person{{ID: arnold}, {ID: mariaS}},
						},
						R2: entity.Relationship{
							Description:        "aunt/uncle in-law",
							Type:               entity.AuntUncleInLaw,
							SourceID:           jfk,
							TargetID:           arnold,
							Path:               []entity.Person{{ID: jfk}, {ID: jfkMother}, {ID: jfkSister}, {ID: mariaS}},
							GenerationsRemoved: -1,
						},
					},
				},
			}
			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) { runTest(t, test) })
			}
		})
	})
}

func testPerson(t *testing.T, errMsgPrefix string, actual, expected *entity.Person) {
	t.Helper()

	if actual != nil && expected == nil {
		t.Errorf("%s; expected person to be nil", errMsgPrefix)
	} else if actual == nil && expected != nil {
		t.Errorf("%s; expected person to be non-nil", errMsgPrefix)
	} else if actual != nil && expected != nil {
		if actual.ID != expected.ID {
			t.Errorf(
				"%s; wrong ID; got %q, exp %q",
				errMsgPrefix, actual.ID, expected.ID,
			)
		}
	}
}

func testUnion(t *testing.T, errMsgPrefix string, actual, expected *entity.Union) {
	t.Helper()

	if actual != nil && expected == nil {
		t.Errorf("%s; expected union to be nil", errMsgPrefix)
	} else if actual == nil && expected != nil {
		t.Errorf("%s; expected union to be non-nil", errMsgPrefix)
	} else if actual != nil && expected != nil {
		testPerson(t, errMsgPrefix+".Person1", actual.Person1, expected.Person1)
		testPerson(t, errMsgPrefix+".Person2", actual.Person2, expected.Person2)
	}
}

func testRelationship(t *testing.T, errMsgPrefix string, actual, expected entity.Relationship) {
	t.Helper()

	if actual.SourceID != expected.SourceID {
		t.Errorf("%s; wrong SourceID; got %q, exp %q", errMsgPrefix, actual.SourceID, expected.SourceID)
	}
	if actual.TargetID != expected.TargetID {
		t.Errorf("%s; wrong TargetID; got %q, exp %q", errMsgPrefix, actual.TargetID, expected.TargetID)
	}
	if actual.Description != expected.Description {
		t.Errorf("%s; wrong Description; got %q, exp %q", errMsgPrefix, actual.Description, expected.Description)
	}
	if actual.Type != expected.Type {
		t.Errorf("%s; wrong Type; got %q, exp %q", errMsgPrefix, actual.Type.String(), expected.Type.String())
	}
	if actual.GenerationsRemoved != expected.GenerationsRemoved {
		t.Errorf("%s; wrong GenerationsRemoved; got %d, exp %d", errMsgPrefix, actual.GenerationsRemoved, expected.GenerationsRemoved)
	}

	// To keep things simple, the Path field only has one path, even
	// when multiple paths to the same common ancestor exist. The chosen path is
	// involves the lexically-greater ID so that results are determinstic.

	if len(actual.Path) != len(expected.Path) {
		t.Errorf("%s; wrong length for Path; got %d, exp %d", errMsgPrefix, len(actual.Path), len(expected.Path))
		for _, val := range actual.Path {
			mustJSON(t, "got: ", map[string]any{"id": val.ID, "name": val.Name.Full()})
		}
		for _, val := range expected.Path {
			mustJSON(t, "exp: ", map[string]any{"id": val.ID, "name": val.Name.Full()})
		}
	} else {
		for i, got := range actual.Path {
			exp := expected.Path[i]
			if got.ID != exp.ID {
				t.Errorf("%s; wrong ID on Path[%d]; got %q, exp %q", errMsgPrefix, i, got.ID, exp.ID)
			}
		}
	}
}

func mustJSON(t *testing.T, banner string, in any) {
	t.Helper()
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s %s", banner, raw)
}
