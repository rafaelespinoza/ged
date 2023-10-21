package srv

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseGedcom(t *testing.T) {
	// check that it can read data and that outputs are non-empty
	for _, testFilename := range []string{"kennedy.ged", "game_of_thrones.ged", "simpsons.ged"} {
		pathToFile := filepath.Join("..", "..", "testdata", testFilename)
		file, err := os.Open(filepath.Clean(pathToFile))
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = file.Close() }()

		people, unions, err := ParseGedcom(context.Background(), file)
		if err != nil {
			t.Fatal(err)
		}

		if len(people) < 1 {
			t.Fatal("expected some people but got 0")
		}
		for _, person := range people {
			if person.ID == "" {
				t.Errorf("empty ID for person %q", person.Name.Full())
			}

			for _, parent := range person.Parents {
				if parent.ID == "" {
					t.Errorf("empty ID for parent %q of person %q", parent.Name.Full(), person.ID)
				}
			}

			for _, child := range person.Children {
				if child.ID == "" {
					t.Errorf("empty ID for child %q of person %q", child.Name.Full(), person.ID)
				}
			}
		}

		if len(unions) < 1 {
			t.Fatal("expected some unions but got 0")
		}
		for _, union := range unions {
			if union.ID == "" {
				t.Error("empty ID for union")
			}

			if union.Person1 != nil && union.Person1.ID == "" {
				t.Error("person 1 in union has empty ID")
			}
			if union.Person2 != nil && union.Person2.ID == "" {
				t.Error("person 2 in union has empty ID")
			}
		}
	}
}
