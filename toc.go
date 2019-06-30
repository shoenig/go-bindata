// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package petrify

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type assetTree struct {
	Asset    Asset
	Children map[string]*assetTree
}

func newAssetTree() *assetTree {
	tree := &assetTree{}
	tree.Children = make(map[string]*assetTree)
	return tree
}

func (node *assetTree) child(name string) *assetTree {
	rv, ok := node.Children[name]
	if !ok {
		rv = newAssetTree()
		node.Children[name] = rv
	}
	return rv
}

// Add route to node.
func (node *assetTree) Add(route []string, asset Asset) {
	for _, name := range route {
		node = node.child(name)
	}
	node.Asset = asset
}

func ident(w io.Writer, n int) {
	for i := 0; i < n; i++ {
		_, _ = w.Write([]byte{'\t'})
	}
}

func (node *assetTree) funcOrNil() string {
	if node.Asset.Func == "" {
		return "nil"
	}
	return node.Asset.Func
}

func (node *assetTree) writeGoMap(w io.Writer, nident int) {
	_, _ = fmt.Fprintf(w, "&bintree{%s, map[string]*bintree{", node.funcOrNil())

	if len(node.Children) > 0 {
		_, _ = io.WriteString(w, "\n")

		// Sort to make output stable between invocations
		filenames := make([]string, len(node.Children))
		i := 0
		for filename := range node.Children {
			filenames[i] = filename
			i++
		}
		sort.Strings(filenames)

		for _, p := range filenames {
			ident(w, nident+1)
			_, _ = fmt.Fprintf(w, `"%s": `, p)
			node.Children[p].writeGoMap(w, nident+1)
		}
		ident(w, nident)
	}

	_, _ = io.WriteString(w, "}}")
	if nident > 0 {
		_, _ = io.WriteString(w, ",")
	}
	_, _ = io.WriteString(w, "\n")
}

func (node *assetTree) WriteAsGoMap(w io.Writer) error {
	_, err := fmt.Fprint(w, `type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = `)
	node.writeGoMap(w, 0)
	return err
}

func writeTOCTree(w io.Writer, toc []Asset) error {
	_, err := fmt.Fprintf(w, `// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %%s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %%s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

`)
	if err != nil {
		return err
	}
	tree := newAssetTree()
	for i := range toc {
		pathList := strings.Split(toc[i].Name, "/")
		tree.Add(pathList, toc[i])
	}
	return tree.WriteAsGoMap(w)
}

// writeTOC writes the table of contents file.
func writeTOC(w io.Writer, toc []Asset) error {
	err := writeTOCHeader(w)
	if err != nil {
		return err
	}

	for i := range toc {
		err = writeTOCAsset(w, &toc[i])
		if err != nil {
			return err
		}
	}

	return writeTOCFooter(w)
}

// writeTOCHeader writes the table of contents file header.
func writeTOCHeader(w io.Writer) error {
	_, err := fmt.Fprintf(w, `// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %%s can't read by error: %%v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %%s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %%s can't read by error: %%v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %%s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

var defaultPageFuncMap = template.FuncMap{
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

// MustParseTemplates is like MustAsset but assumes the input files
// are html templates.
func MustParseTemplates(files ...string) *template.Template {
	if len(files) == 0 {
		panic("MustParseTemplates requires at least one filename")
	}

	// the first file is the root template
	root := template.New(path.Base(files[0]))
	if b, e := Asset(string(files[0])); e != nil {
		panic("failed to load " + files[0] + ": " + e.Error())
	} else if _, e = root.Parse(string(b)); e != nil {
		panic("failed to parse " + files[0] + ": " + e.Error())
	}

	// the remaining files are children templates
	for _, file := range files[1:] {
		child := root.New(path.Base(file))
		if b, e := Asset(string(file)); e != nil {
			panic("failed to load " + string(file) + ": " + e.Error())
		} else if _, e = child.Parse(string(b)); e != nil {
			panic("failed to parse " + string(file) + ": " + e.Error())
		}
	}

	root.Funcs(defaultPageFuncMap)

	return root
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
`)
	return err
}

// writeTOCAsset write a TOC entry for the given asset.
func writeTOCAsset(w io.Writer, asset *Asset) error {
	_, err := fmt.Fprintf(w, "\t%q: %s,\n", asset.Name, asset.Func)
	return err
}

// writeTOCFooter writes the table of contents file footer.
func writeTOCFooter(w io.Writer) error {
	_, err := fmt.Fprintf(w, `}

`)
	return err
}
