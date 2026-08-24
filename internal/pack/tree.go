package pack

import (
	"os"
	"path"
	"sort"
	"strings"
)

// Node is one entry in a pack's folder tree.
type Node struct {
	// Path is relative to the pack folder, with forward slashes, which is
	// how the index spells its own entries.
	Path string
	// Name is the last element of the path.
	Name string
	// Dir reports a folder.
	Dir bool
	// Indexed reports a file packwiz has in its index, and so a file the
	// exported pack will carry. A file that is not indexed is either new
	// or ignored, and needs a refresh either way.
	Indexed bool
	// Metafile reports one of the per-mod TOML files.
	Metafile bool
}

// Tree is a pack folder scanned into parent and children lists.
//
// It is a snapshot: the file list is read once and then answered from
// memory, because the tree widget asks for a node's children many times
// per redraw and walking the disk for each of those would be wasteful.
type Tree struct {
	children map[string][]Node
}

// skipDirs are folders never walked into. The repository's own plumbing
// is large, changes constantly and is nothing a pack author edits by
// hand.
var skipDirs = map[string]bool{
	".git": true, ".idea": true, ".vscode": true, "node_modules": true,
}

// ScanTree walks the pack folder.
func (p Pack) ScanTree() (*Tree, error) {
	indexed := map[string]bool{}
	metafiles := map[string]bool{}

	// The pack's own two files are part of the pack by definition, even
	// though packwiz does not list them in its own index.
	indexed[FileName] = true
	if p.IndexFile != "" {
		indexed[p.IndexFile] = true
	}

	if files, err := p.Files(); err == nil {
		for _, f := range files {
			indexed[f] = true
		}
	}
	if metas, err := p.metafiles(); err == nil {
		for _, f := range metas {
			indexed[f] = true
			metafiles[f] = true
		}
	}

	t := &Tree{children: map[string][]Node{}}
	if err := t.walk(p.Dir, "", indexed, metafiles); err != nil {
		return nil, err
	}
	return t, nil
}

// walk fills in one folder and recurses into the ones under it.
func (t *Tree) walk(root, rel string, indexed, metafiles map[string]bool) error {
	dir := root
	if rel != "" {
		dir = path.Join(root, rel)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var nodes []Node
	for _, e := range entries {
		if e.IsDir() && skipDirs[e.Name()] {
			continue
		}

		child := e.Name()
		if rel != "" {
			child = rel + "/" + e.Name()
		}

		nodes = append(nodes, Node{
			Path:     child,
			Name:     e.Name(),
			Dir:      e.IsDir(),
			Indexed:  indexed[child],
			Metafile: metafiles[child],
		})

		if e.IsDir() {
			// A folder that cannot be read is left empty rather than
			// failing the whole scan: a permission problem in one corner
			// of a pack should not hide the rest of it.
			_ = t.walk(root, child, indexed, metafiles)
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Dir != nodes[j].Dir {
			return nodes[i].Dir
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})

	t.children[rel] = nodes
	return nil
}

// Children lists what is directly inside a folder.
func (t *Tree) Children(dir string) []Node {
	if t == nil {
		return nil
	}
	return t.children[dir]
}

// Find looks up one node by path.
func (t *Tree) Find(rel string) (Node, bool) {
	if t == nil {
		return Node{}, false
	}

	parent := path.Dir(rel)
	if parent == "." {
		parent = ""
	}
	for _, n := range t.children[parent] {
		if n.Path == rel {
			return n, true
		}
	}
	return Node{}, false
}

// IsDir reports a path the scan found as a folder.
func (t *Tree) IsDir(rel string) bool {
	if rel == "" {
		return true
	}
	n, ok := t.Find(rel)
	return ok && n.Dir
}

// Paths lists every file in the tree, folders excluded, which is what a
// search or a save-all walks.
func (t *Tree) Paths() []string {
	if t == nil {
		return nil
	}

	var out []string
	for _, nodes := range t.children {
		for _, n := range nodes {
			if !n.Dir {
				out = append(out, n.Path)
			}
		}
	}
	sort.Strings(out)
	return out
}
