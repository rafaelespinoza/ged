package srv

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/rafaelespinoza/ged/internal/entity"
	"github.com/rafaelespinoza/ged/internal/log"
)

type Relator interface {
	Relate(ctx context.Context, person1ID, person2ID string) (out entity.MutualRelationship, err error)
}

func NewRelator(people []*entity.Person) Relator {
	out := relator{
		peopleByID:       make(map[string]*entity.Person, len(people)),
		childToParentIDs: make(map[string]idSet, len(people)),
		spousePartners:   make(map[string]idSet, len(people)),
	}

	for _, person := range people {
		out.peopleByID[person.ID] = person
		childParentIDs := make(idSet)
		spousePartners := make(idSet)

		for _, parent := range person.Parents {
			childParentIDs.add(parent.ID)
		}

		for _, spouse := range person.Spouses {
			spousePartners.add(spouse.ID)
		}

		out.childToParentIDs[person.ID] = childParentIDs
		out.spousePartners[person.ID] = spousePartners
	}

	return &out
}

type relator struct {
	peopleByID       map[string]*entity.Person
	childToParentIDs map[string]idSet
	spousePartners   map[string]idSet // key is ID, value is set of IDs for spouses
}

const maxGenerationsToRelate = 100

var errUnrelated = errors.New("it appears that these people are unrelated")

func (r *relator) Relate(ctx context.Context, p1ID, p2ID string) (out entity.MutualRelationship, err error) {
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

	r1, r2, ancestor, err := r.relate(ctx, p1ID, p2ID)
	if errors.Is(err, errUnrelated) {
		// now see if they are related via marriage
	} else if err != nil {
		return
	} else {
		out = entity.MutualRelationship{
			CommonPerson: ancestor,
			R1:           r1,
			R2:           r2,
		}
		return
	}

	r1, r2, union, err := r.affiniate(ctx, p1ID, p2ID)
	if err != nil {
		return
	}
	r1.SourceID, r1.TargetID = p1ID, p2ID
	r2.SourceID, r2.TargetID = p2ID, p1ID

	out = entity.MutualRelationship{
		Union: union,
		R1:    r1,
		R2:    r2,
	}

	return
}

func (r *relator) relate(ctx context.Context, p1ID, p2ID string) (r1, r2 entity.Relationship, p *entity.Person, err error) {
	if p1ID == p2ID {
		r1, r2, err = makeRelationships(r, []string{p1ID}, []string{p2ID})
		if err == nil {
			p, _ = r.lookupOne(p1ID)
			r1.SourceID, r1.TargetID = p1ID, p2ID
			r2.SourceID, r2.TargetID = p2ID, p1ID
		}
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

	ancestor, ok := r.peopleByID[ancestorID]
	if !ok {
		err = errUnrelated
		log.Error(ctx, map[string]any{"p1": p1ID, "p2": p2ID}, err, "")
		return
	}

	log.Info(ctx, map[string]any{"id": ancestorID, "name": ancestor.Name.Full()}, "found most recent common ancestor")

	r1, r2, err = makeRelationships(r, shortestP1Path, shortestP2Path)
	if err == nil {
		p = ancestor
		r1.SourceID, r1.TargetID = p1ID, p2ID
		r2.SourceID, r2.TargetID = p2ID, p1ID
	}
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
	slices.Sort(commonIDs) // ensure deterministic results.

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

func (r *relator) affiniate(ctx context.Context, p1ID, p2ID string) (r1, r2 entity.Relationship, u *entity.Union, err error) {
	if spouses(r, p1ID, p2ID) {
		r1.Type, r2.Type = entity.Spouse, entity.Spouse
		r1.Description, r2.Description = r1.Type.String(), r2.Type.String()

		p1, _ := r.lookupOne(p1ID)
		p2, _ := r.lookupOne(p2ID)
		r1.Path, r2.Path = []entity.Person{*p1}, []entity.Person{*p2}
		u = &entity.Union{Person1: p1, Person2: p2}
		return
	}

	// are any of P1's spouses related to P2?
	m1, m1Err := relateSpouses(ctx, r, p1ID, p2ID)
	if m1Err != nil && !errors.Is(m1Err, errUnrelated) {
		err = m1Err
		return
	}

	// are any of P2's spouses related to P1?
	m2, m2Err := relateSpouses(ctx, r, p2ID, p1ID)
	if m2Err != nil && !errors.Is(m2Err, errUnrelated) {
		err = m2Err
		return
	}

	if errors.Is(m1Err, errUnrelated) && errors.Is(m2Err, errUnrelated) {
		err = errUnrelated
		return
	}
	log.Debug(ctx, map[string]any{
		"p1_id":      p1ID,
		"p2_id":      p2ID,
		"m1 == nil?": m1 == nil,
		"m2 == nil?": m2 == nil,
	}, "# relator.affiniate: after 2 calls to relateSpouses()")

	if m1 != nil && m2 == nil {
		log.Debug(ctx, map[string]any{
			"p1_id":               p1ID,
			"p2_id":               p2ID,
			"m1.R1.Type":          m1.R1.Type.String(),
			"m1.R2.Type":          m1.R2.Type.String(),
			"m1.Union.Person1.ID": m1.Union.Person1.ID,
			"m1.Union.Person2.ID": m1.Union.Person2.ID,
		}, "# relator.affiniate: before affiniate")
		r2, r1, u = affiniate(ctx, r, p1ID, *m1, p2ID)
	} else if m1 == nil && m2 != nil {
		log.Debug(ctx, map[string]any{
			"p1_id":               p1ID,
			"p2_id":               p2ID,
			"m2.R1.Type":          m2.R1.Type.String(),
			"m2.R2.Type":          m2.R2.Type.String(),
			"m2.Union.Person1.ID": m2.Union.Person1.ID,
			"m2.Union.Person2.ID": m2.Union.Person2.ID,
		}, "# relator.affiniate: before affiniate")
		r1, r2, u = affiniate(ctx, r, p2ID, *m2, p1ID)
	}

	return
}

// relateSpouses determines if any spouses for the person at P1ID have a blood
// relationship with the person at p2.
func relateSpouses(ctx context.Context, r *relator, p1ID, p2ID string) (*entity.MutualRelationship, error) {
	for _, p1SpouseID := range r.spousePartners[p1ID].ids() {
		r1, r2, p, rerr := r.relate(ctx, p1SpouseID, p2ID)
		if errors.Is(rerr, errUnrelated) {
			continue
		} else if rerr != nil {
			return nil, rerr
		}

		var out entity.MutualRelationship
		log.Debug(ctx, map[string]any{
			"p1_id":            p1ID,
			"p2_id":            p2ID,
			"p1_spouse_id":     p1SpouseID,
			"r1.Type":          r1.Type.String(),
			"r2.Type":          r2.Type.String(),
			"common_person_id": p.ID,
		}, "# relateSpouses: people (p1_spouse_id + p2_id) seem related")
		out.R1, out.R2 = r1, r2
		out.CommonPerson = p

		p1, _ := r.lookupOne(p1SpouseID)
		p2, _ := r.lookupOne(p1ID)
		out.Union = &entity.Union{Person1: p1, Person2: p2}

		return &out, nil
	}

	return nil, errUnrelated
}

// affiniate builds the parts of an affinal relationship from a consanguinal
// relationship m between a spouse of the person at pID and the other person at
// oID.
func affiniate(ctx context.Context, r *relator, pID string, m entity.MutualRelationship, oID string) (r1, r2 entity.Relationship, u *entity.Union) {
	r1 = invertRelationship(m.R1)
	changeTypeToAffinal(ctx, &r1)

	r2 = invertRelationship(m.R2)
	changeTypeToAffinal(ctx, &r2)

	switch r1.Type {
	// If the relationship type involves siblings on some level...
	case entity.SiblingInLaw, entity.AuntUncleInLaw, entity.NieceNephewInLaw, entity.CousinInLaw:
		// but the path does not already include the person at oID...
		var includesPerson bool
		for _, person := range r1.Path {
			if person.ID == oID {
				includesPerson = true
				break
			}
		}

		if !includesPerson {
			var indexOfCommonPerson int
			for i, person := range m.R2.Path { // read from original path, in its pre-inversion state
				if (m.CommonPerson != nil && person.ID == m.CommonPerson.ID) ||
					(m.Union.Person1 != nil && person.ID == m.Union.Person1.ID) ||
					(m.Union.Person2 != nil && person.ID == m.Union.Person2.ID) {
					indexOfCommonPerson = i
					break
				}
			}
			// then ensure the relationship path includes people leading up
			// until the person in common.
			head := make([]entity.Person, indexOfCommonPerson)
			copy(head, m.R2.Path)
			r1.Path = append(head, r1.Path...)
		}
	}

	spIDs := spouseIDs(r, pID, r1.Path...)
	if len(spIDs) == 1 && spIDs[0] != pID {
		p, _ := r.lookupOne(pID)
		spouse, _ := r.lookupOne(spIDs[0])
		r2.Path = []entity.Person{*p, *spouse}
	}

	u = m.Union

	return
}

func invertRelationship(in entity.Relationship) (out entity.Relationship) {
	switch in.Type {
	case entity.Parent:
		out.Type = entity.Child
	case entity.Child:
		out.Type = entity.Parent
	case entity.AuntUncle:
		out.Type = entity.NieceNephew
	case entity.NieceNephew:
		out.Type = entity.AuntUncle
	default:
		out.Type = in.Type
	}

	out.GenerationsRemoved = in.GenerationsRemoved * -1

	out.Path = make([]entity.Person, len(in.Path))
	copy(out.Path, in.Path)
	slices.Reverse(out.Path)

	generationsSinceCommonAncestor := len(out.Path) - 1
	if in.Type == entity.Cousin {
		generationsSinceCommonAncestor *= -1
	}

	desc, _, err := describeRelationship(out.Type, out.GenerationsRemoved, generationsSinceCommonAncestor)
	if err != nil {
		log.Error(context.TODO(), map[string]any{
			"in":  in,
			"out": out,
		}, err, "srv.invertRelationship: could not compute lineage inversion")
	}
	out.Description = desc
	out.SourceID, out.TargetID = in.TargetID, in.SourceID

	return
}

func changeTypeToAffinal(ctx context.Context, r *entity.Relationship) {
	if r.Type >= entity.Spouse {
		log.Debug(ctx, map[string]any{"type": r.Type.String()}, "# changeTypeToAffinal: Type seems to already be affinal")
		return
	}

	r.Type += entity.Spouse - entity.Self
	r.Description += " in-law"
}

// spouseIDs extracts IDs of people who have been a spouse of the person at id.
func spouseIDs(r *relator, id string, people ...entity.Person) (out []string) {
	for _, person := range people {
		if spouses(r, id, person.ID) {
			out = append(out, person.ID)
		}
	}

	return
}

func spouses(r *relator, p1ID, p2ID string) bool {
	p1Spouses, ok := r.spousePartners[p1ID]
	if !ok {
		return false
	}
	p2Spouses, ok := r.spousePartners[p2ID]
	if !ok {
		return false
	}

	return p1Spouses.has(p2ID) && p2Spouses.has(p1ID)
}
