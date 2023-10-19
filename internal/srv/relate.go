package srv

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelespinoza/reltree/internal/entity"
	"github.com/rafaelespinoza/reltree/internal/log"
)

type Relator interface {
	Relate(ctx context.Context, person1ID, person2ID string) (rel1, rel2 entity.Lineage, err error)
}

func NewRelator(people []*entity.Person) Relator {
	out := relator{
		peopleByID:       make(map[string]*entity.Person, len(people)),
		childToParentIDs: make(map[string]idSet, len(people)),
	}

	for _, person := range people {
		out.peopleByID[person.ID] = person
		childParentIDs := make(idSet)

		for _, parent := range person.Parents {
			childParentIDs.add(parent.ID)
		}

		out.childToParentIDs[person.ID] = childParentIDs
	}

	return &out
}

type relator struct {
	peopleByID       map[string]*entity.Person
	childToParentIDs map[string]idSet
}

const maxGenerationsToRelate = 100

func (r *relator) Relate(ctx context.Context, p1ID, p2ID string) (r1, r2 entity.Lineage, err error) {
	_, ok := r.lookupOne(p1ID)
	if !ok {
		err = fmt.Errorf("person with id %v not found", p1ID)
		return
	}
	_, ok = r.lookupOne(p2ID)
	if !ok {
		err = fmt.Errorf("person with id %v not found", p2ID)
		return
	}
	if p1ID == p2ID {
		err = fmt.Errorf("cannot compare same people ids (%v %v)", p1ID, p2ID)
		return
	}

	p1Paths, p2Paths := make(pathsToPersonID, maxGenerationsToRelate), make(pathsToPersonID, maxGenerationsToRelate)

	visited := make(idSet) // detect common ancestors.

	// The first pass collects as many ancestral paths as possible from the
	// originating person at p1ID. Each time a node is visited, it's marked.
	findCommonAncestorPaths(ctx, "p1", r, 0, visited, p1Paths, []string{}, p1ID)
	// The second pass finds the id of a common ancestor, or is 0 if there is no
	// relation. It relies on the status of previously-visited nodes to
	// recognize a common ancestor, which was built up in the first pass.
	findCommonAncestorPaths(ctx, "p2", r, 0, visited, p2Paths, []string{}, p2ID)

	ancestorID, shortestP1Path, shortestP2Path := getShortestCommonPaths(ctx, r, p1Paths, p2Paths)

	if ancestor, ok := r.peopleByID[ancestorID]; !ok {
		err = errors.New("it appears that these people are unrelated")
		log.Error(ctx, map[string]any{"p1": p1ID, "p2": p2ID}, err, "")
		return
	} else {
		log.Info(ctx, map[string]any{"id": ancestorID, "name": ancestor.Name.Full()}, "found most recent common ancestor")
	}

	r1, r2, err = makeLineages(r, shortestP1Path, shortestP2Path)
	return
}

func (r *relator) lookupOne(id string) (out *entity.Person, found bool) {
	out, found = r.peopleByID[id]
	return
}

func findCommonAncestorPaths(ctx context.Context, tag string, r *relator, currGeneration int, visited idSet, allPaths pathsToPersonID, prevPath []string, id string) {
	if currGeneration >= maxGenerationsToRelate {
		return
	}
	currPath := append(duplicateIDs(prevPath), id)

	if visited.has(id) {
		log.Debug(ctx, map[string]any{"id": id, "tag": tag}, "apparently already visited person")
		// ancestorID = id
		allPaths.add(id, currPath)
		// return
	}

	parentIDs := r.childToParentIDs[id].ids() // should be non-nil b/c of how the map was built earlier.

	for _, parentID := range parentIDs {
		findCommonAncestorPaths(ctx, tag, r, currGeneration+1, visited, allPaths, currPath, parentID)
		// if ancestorID > 0 {
		// 	return
		// }
	}
	allPaths.add(id, currPath)
	visited.add(id)

	return
}

func getShortestCommonPaths(ctx context.Context, r *relator, p1Paths, p2Paths pathsToPersonID) (id string, p1Path, p2Path []string) {
	commonIDs := make([]string, 0)
	for p1 := range p1Paths {
		_, ok := p2Paths[p1]
		if !ok {
			continue
		}
		commonIDs = append(commonIDs, p1)
	}

	p1Path, p2Path = nil, nil

	log.Debug(ctx, map[string]any{"common_ids": commonIDs}, "shortest common path")

	for _, commonID := range commonIDs {
		lt, rt := p1Paths.shortest(commonID), p2Paths.shortest(commonID)

		log.Debug(ctx, map[string]any{"id": commonID, "p1_path": p1Path, "p2_path": p2Path, "lt": lt, "rt": rt}, "common paths")
		if p1Path == nil {
			p1Path = lt
		}
		if p2Path == nil {
			p2Path = rt
		}

		if len(lt) <= len(p1Path) || len(rt) <= len(p2Path) {
			id = commonID
			p1Path = lt
			p2Path = rt
		}
	}

	return
}

func duplicateIDs(in []string) (out []string) {
	out = make([]string, len(in))
	copy(out, in)
	return
}

type pathsToPersonID map[string][][]string

func (p pathsToPersonID) add(id string, path []string) {
	_, ok := p[id]
	if !ok {
		p[id] = make([][]string, 0)
	}
	p[id] = append(p[id], path)
}

func (p pathsToPersonID) shortest(id string) (out []string) {
	paths, ok := p[id]
	if !ok || len(paths) < 1 {
		return
	}

	out = paths[0]
	for i := 1; i < len(paths); i++ {
		if len(paths[i]) < len(out) {
			out = paths[i]
		}
	}
	return
}

type idSet map[string]struct{}

func (s idSet) add(id string) { s[id] = struct{}{} }
func (s idSet) has(id string) (ok bool) {
	_, ok = s[id]
	return
}

func (s idSet) ids() []string {
	out := make([]string, 0, len(s))
	for id := range s {
		out = append(out, id)
	}
	return out
}
