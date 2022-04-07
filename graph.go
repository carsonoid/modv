package main

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"
)

var graphTemplate = `digraph {
{{- if eq .direction "horizontal" -}}
rankdir=LR;
{{ end -}}
node [shape=box];
{{ range $mod, $modId := .mods -}}
{{ $modId }} [label="{{ $mod }}"];
{{ end -}}
{{- range $modId, $depModIds := .dependencies -}}
{{- range $_, $depModId := $depModIds -}}
{{ $modId }} -> {{ $depModId }};
{{  end -}}
{{- end -}}
}
`

type ModuleGraph struct {
	Reader io.Reader

	Mods         map[string]int
	Dependencies map[int][]int
	Filter       string
}

func NewModuleGraph(r io.Reader, filter string) *ModuleGraph {
	return &ModuleGraph{
		Reader: r,

		Mods:         make(map[string]int),
		Dependencies: make(map[int][]int),
		Filter:       filter,
	}
}

func (m *ModuleGraph) Parse() error {
	deps, err := GetDepPairs(m.Reader)
	if err != nil {
		return err
	}

	if m.Filter != "" {
		s := NewSearcher()
		deps = s.Filter(deps, m.Filter)

		// sort deps by number of @ signs
		sort.Slice(deps, func(i, j int) bool {
			return strings.Count(deps[i].mod, "@") < strings.Count(deps[j].mod, "@")
		})
	}

	serialID := 1
	for _, dep := range deps {
		modId, ok := m.Mods[dep.mod]
		if !ok {
			modId = serialID
			m.Mods[dep.mod] = modId
			serialID += 1
		}

		depModId, ok := m.Mods[dep.dep]
		if !ok {
			depModId = serialID
			m.Mods[dep.dep] = depModId
			serialID += 1
		}

		m.Dependencies[modId] = append(m.Dependencies[modId], depModId)
	}

	return nil
}

type DepPair struct {
	mod, dep string
}

// String func for deppair
func (d *DepPair) String() string {
	return fmt.Sprintf("%s -> %s", d.mod, d.dep)
}

// GetDepPairs returns a list of all modules and their dependencies
func GetDepPairs(r io.Reader) ([]*DepPair, error) {
	var pairs []*DepPair

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		relation := strings.Split(line, " ")
		mod, depMod := strings.TrimSpace(string(relation[0])), strings.TrimSpace(string(relation[1]))

		pairs = append(pairs, &DepPair{mod, depMod})
	}

	return pairs, scanner.Err()
}

type Searcher struct {
	searched map[string]struct{}
}

// new searcher
func NewSearcher() *Searcher {
	return &Searcher{
		searched: make(map[string]struct{}),
	}
}

// Filter out modules that are not used by the given modules
// order is not preserved
func (s *Searcher) Filter(pairs []*DepPair, findMod string) []*DepPair {
	var ret []*DepPair

	for _, user := range GetUsers(pairs, findMod) {
		if _, ok := s.searched[user.mod]; !ok {
			s.searched[user.mod] = struct{}{}
			ret = append(ret, user)
			ret = append(ret, s.Filter(pairs, user.mod)...)
		}
	}

	return ret
}

// GetUsers returns all modules that depend on the given module
func GetUsers(pairs []*DepPair, findMod string) []*DepPair {
	var ret []*DepPair
	for _, pair := range pairs {
		if pair.dep == findMod {
			ret = append(ret, pair)
		}
	}
	return ret
}

func (m *ModuleGraph) Render(w io.Writer) error {
	templ, err := template.New("graph").Parse(graphTemplate)
	if err != nil {
		return fmt.Errorf("templ.Parse: %v", err)
	}

	var direction string
	if len(m.Dependencies) > 15 {
		direction = "horizontal"
	}

	if err := templ.Execute(w, map[string]interface{}{
		"mods":         m.Mods,
		"dependencies": m.Dependencies,
		"direction":    direction,
	}); err != nil {
		return fmt.Errorf("templ.Execute: %v", err)
	}

	return nil
}
