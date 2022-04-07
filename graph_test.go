package main

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"testing"
)

func TestModuleGraph_Parse(t *testing.T) {
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "full",
			args: args{
				bytes.NewReader([]byte(`github.com/poloxue/testmod golang.org/x/text@v0.3.2
github.com/poloxue/testmod rsc.io/quote/v3@v3.1.0
github.com/poloxue/testmod rsc.io/sampler@v1.3.1
golang.org/x/text@v0.3.2 golang.org/x/tools@v0.0.0-20180917221912-90fa682c2a6e
rsc.io/quote/v3@v3.1.0 rsc.io/sampler@v1.3.0
rsc.io/sampler@v1.3.1 golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c
rsc.io/sampler@v1.3.0 golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c`))},
			want: []byte(`digraph {
node [shape=box];
1 [label="github.com/poloxue/testmod"];
2 [label="golang.org/x/text@v0.3.2"];
3 [label="rsc.io/quote/v3@v3.1.0"];
4 [label="rsc.io/sampler@v1.3.1"];
5 [label="golang.org/x/tools@v0.0.0-20180917221912-90fa682c2a6e"];
6 [label="rsc.io/sampler@v1.3.0"];
7 [label="golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"]
1 -> 2;
1 -> 3;
1 -> 4;
2 -> 5;
3 -> 6;
4 -> 7;
6 -> 7;
}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moduleGraph := NewModuleGraph(tt.args.reader, "")
			moduleGraph.Parse()
			for k, v := range moduleGraph.Mods {
				fmt.Println(v, k)
			}

			for k, v := range moduleGraph.Dependencies {
				fmt.Println(k)
				fmt.Println(v)
				fmt.Println()
			}
		})
	}
}

const modSample = `github.com/poloxue/testmod golang.org/x/text@v0.3.2
github.com/poloxue/testmod rsc.io/quote/v3@v3.1.0
github.com/poloxue/testmod rsc.io/sampler@v1.3.1
golang.org/x/text@v0.3.2 golang.org/x/tools@v0.0.0-20180917221912-90fa682c2a6e
rsc.io/quote/v3@v3.1.0 rsc.io/sampler@v1.3.0
rsc.io/sampler@v1.3.1 golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c
rsc.io/sampler@v1.3.0 golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c`

func TestModuleGraph_GetDepPairs(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    []*DepPair
		wantErr bool
	}{
		{
			"basic",
			args{bytes.NewBufferString(modSample)},
			[]*DepPair{
				{"github.com/poloxue/testmod", "golang.org/x/text@v0.3.2"},
				{"github.com/poloxue/testmod", "rsc.io/quote/v3@v3.1.0"},
				{"github.com/poloxue/testmod", "rsc.io/sampler@v1.3.1"},
				{"golang.org/x/text@v0.3.2", "golang.org/x/tools@v0.0.0-20180917221912-90fa682c2a6e"},
				{"rsc.io/quote/v3@v3.1.0", "rsc.io/sampler@v1.3.0"},
				{"rsc.io/sampler@v1.3.1", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
				{"rsc.io/sampler@v1.3.0", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDepPairs(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModuleGraph.GetDepPairs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModuleGraph.GetDepPairs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleGraph_GetUsers(t *testing.T) {
	type args struct {
		pairs   []*DepPair
		findMod string
	}
	tests := []struct {
		name string
		args args
		want []*DepPair
	}{
		{
			"basic",
			args{
				[]*DepPair{
					{"github.com/poloxue/testmod", "golang.org/x/text@v0.3.2"},
					{"github.com/poloxue/testmod", "rsc.io/quote/v3@v3.1.0"},
					{"github.com/poloxue/testmod", "rsc.io/sampler@v1.3.1"},
					{"golang.org/x/text@v0.3.2", "golang.org/x/tools@v0.0.0-20180917221912-90fa682c2a6e"},
					{"rsc.io/quote/v3@v3.1.0", "rsc.io/sampler@v1.3.0"},
					{"rsc.io/sampler@v1.3.1", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
					{"rsc.io/sampler@v1.3.0", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
				},
				"golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c",
			},
			[]*DepPair{
				{"rsc.io/sampler@v1.3.1", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
				{"rsc.io/sampler@v1.3.0", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUsers(tt.args.pairs, tt.args.findMod); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModuleGraph.GetUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	type args struct {
		pairs   []*DepPair
		findMod string
	}
	tests := []struct {
		name string
		args args
		want []*DepPair
	}{
		{
			"simple",
			args{
				[]*DepPair{
					{"github.com/poloxue/testmod", "golang.org/x/text@v0.3.2"},
					{"github.com/poloxue/testmod", "rsc.io/quote/v3@v3.1.0"},
					{"github.com/poloxue/testmod", "rsc.io/sampler@v1.3.1"},
					{"golang.org/x/text@v0.3.2", "golang.org/x/tools@v0.0.0-20180917221912-90fa682c2a6e"},
					{"rsc.io/quote/v3@v3.1.0", "rsc.io/sampler@v1.3.0"},
					{"rsc.io/sampler@v1.3.1", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
					{"rsc.io/sampler@v1.3.0", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
				},
				"golang.org/x/text@v0.3.2",
			},
			[]*DepPair{
				{"github.com/poloxue/testmod", "golang.org/x/text@v0.3.2"},
			},
		},
		{
			"two layers",
			args{
				[]*DepPair{
					{"github.com/poloxue/testmod", "golang.org/x/text@v0.3.2"},
					{"github.com/poloxue/testmod", "rsc.io/quote/v3@v3.1.0"},
					{"github.com/poloxue/testmod", "rsc.io/sampler@v1.3.1"},
					{"golang.org/x/text@v0.3.2", "golang.org/x/tools@v0.0.0-20180917221912-90fa682c2a6e"},
					{"rsc.io/quote/v3@v3.1.0", "rsc.io/sampler@v1.3.0"},
					{"rsc.io/sampler@v1.3.1", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
					{"rsc.io/sampler@v1.3.0", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
				},
				"rsc.io/sampler@v1.3.0",
			},
			[]*DepPair{
				{"github.com/poloxue/testmod", "rsc.io/quote/v3@v3.1.0"},
				{"rsc.io/quote/v3@v3.1.0", "rsc.io/sampler@v1.3.0"},
			},
		},
		{
			"deep",
			args{
				[]*DepPair{
					{"github.com/poloxue/testmod", "golang.org/x/text@v0.3.2"},
					{"github.com/poloxue/testmod", "rsc.io/quote/v3@v3.1.0"},
					{"github.com/poloxue/testmod", "rsc.io/sampler@v1.3.1"},
					{"golang.org/x/text@v0.3.2", "golang.org/x/tools@v0.0.0-20180917221912-90fa682c2a6e"},
					{"rsc.io/quote/v3@v3.1.0", "rsc.io/sampler@v1.3.0"},
					{"rsc.io/sampler@v1.3.1", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
					{"rsc.io/sampler@v1.3.0", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
				},
				"golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c",
			},
			[]*DepPair{
				{"github.com/poloxue/testmod", "rsc.io/quote/v3@v3.1.0"},
				{"github.com/poloxue/testmod", "rsc.io/sampler@v1.3.1"},
				{"rsc.io/quote/v3@v3.1.0", "rsc.io/sampler@v1.3.0"},
				{"rsc.io/sampler@v1.3.1", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
				{"rsc.io/sampler@v1.3.0", "golang.org/x/text@v0.0.0-20170915032832-14c0d48ead0c"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSearcher()
			got := s.Filter(tt.args.pairs, tt.args.findMod)

			// sort got by string
			sort.Slice(got, func(i, j int) bool {
				return got[i].String() < got[j].String()
			})

			// sort tt.want by string
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].String() < tt.want[j].String()
			})

			if !reflect.DeepEqual(got, tt.want) {
				// print all the differences
				for i, g := range got {
					t.Log(i, g.String())
				}
				for i, w := range tt.want {
					t.Log(i, w.String())
				}
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
