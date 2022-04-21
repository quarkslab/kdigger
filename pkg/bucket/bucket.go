package bucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"k8s.io/client-go/kubernetes"
)

var ErrMissingClient = errors.New("buckets need a kubernetes client for initialization")

type ErrUnknownBucket struct {
	name string
}

func (e ErrUnknownBucket) Error() string {
	return fmt.Sprintf("unknown bucket %q", e.name)
}

type Runnable interface {
	Run() (Results, error)
}

type Interface interface {
	Runnable
}

type Factory func(config Config) (Interface, error)

type entry struct {
	name        string
	description string
	aliases     []string
	factory     Factory
	// True if the bucket has side-effects on its environment and False if it's a readonly bucket
	active bool
}

type Buckets struct {
	lock     sync.RWMutex
	registry map[string]entry
	// this is a "compilation" of all aliases and names pointing to names
	aliases map[string]string
}

type Config struct {
	Client      kubernetes.Interface
	Namespace   string
	Color       bool
	OutputWidth int
	// This options is specific to the admission plugin, is it to force creation even if we can't cleanup the mess with delete
	AdmForce bool
}

func NewBuckets() *Buckets {
	return &Buckets{}
}

func NewConfig() *Config {
	return &Config{}
}

// Registered enumerates the names of all registered plugins.
func (bs *Buckets) Registered() []string {
	bs.lock.RLock()
	defer bs.lock.RUnlock()
	keys := []string{}
	for k := range bs.registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Registered enumerates the names of all passive registered plugins.
func (bs *Buckets) RegisteredPassive() []string {
	bs.lock.RLock()
	defer bs.lock.RUnlock()
	keys := []string{}
	for k, b := range bs.registry {
		if !b.active {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// Register registers a plugin Factory by name. This is expected to happen
// during app startup.
func (bs *Buckets) Register(name string, aliases []string, description string, active bool, factory Factory) {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	if bs.registry != nil {
		_, found := bs.registry[name]
		if found {
			log.Fatalf("bucket %q was registered twice", name)
		}
	} else {
		bs.registry = map[string]entry{}
	}

	bs.registry[name] = entry{
		name:        name,
		description: description,
		aliases:     aliases,
		factory:     factory,
		active:      active,
	}

	if bs.aliases == nil {
		bs.aliases = map[string]string{}
	}

	// name unicity was already checked
	bs.aliases[name] = name
	for _, alias := range aliases {
		if _, found := bs.aliases[alias]; found {
			log.Fatalf("bucket registration: aliases %q conflict for bucket %q", alias, name)
		}
		bs.aliases[alias] = name
	}
}

func (bs *Buckets) ResolveAlias(alias string) (string, bool) {
	e, found := bs.findEntryFromAlias(alias)
	return e.name, found
}

func (bs *Buckets) findEntryFromAlias(alias string) (entry, bool) {
	bs.lock.RLock()
	defer bs.lock.RUnlock()
	// golang force us to write that
	e, found := bs.registry[bs.aliases[alias]]
	return e, found
}

// getBucket creates an instance of the named bucket.  It returns `false` if
// the the name is not known. The error is returned only when the named
// provider was known but failed to initialize.  The config parameter specifies
// the bucket configuration or nil for no configuration.
func (bs *Buckets) getBucket(name string, config Config) (Interface, bool, error) {
	f, found := bs.findEntryFromAlias(name)
	if !found {
		return nil, false, nil
	}

	ret, err := f.factory(config)
	return ret, true, err
}

// InitBucket creates an instance of the named interface.
func (bs *Buckets) InitBucket(name string, config Config) (Interface, error) {
	if name == "" {
		log.Println("No bucket specified.")
		return nil, nil
	}

	bucket, found, err := bs.getBucket(name, config)
	if err == ErrMissingClient {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("couldn't init bucket %q: %v", name, err)
	}
	if !found {
		err := ErrUnknownBucket{
			name: name,
		}
		return nil, err
	}

	return bucket, nil
}

func (bs *Buckets) Describe(name string) string {
	e, found := bs.findEntryFromAlias(name)
	if !found {
		return ""
	}
	return e.description
}

func (bs *Buckets) Aliases(name string) []string {
	e, found := bs.findEntryFromAlias(name)
	if !found {
		return []string{""}
	}
	return e.aliases
}

func (bs *Buckets) IsActive(name string) bool {
	e, found := bs.findEntryFromAlias(name)
	if !found {
		return false
	}
	return e.active
}

type Results struct {
	bucketName string
	headers    []string
	data       [][]interface{}
	comments   []string
}

// ResultsOpts uses pointers to have a default nil value that will be evaluated
// as true instead of bool defaut that is false.
type ResultsOpts struct {
	ShowName     *bool
	ShowComments *bool
	ShowData     *bool
	OutputWidth  int
}

func NewResults(name string) *Results {
	return &Results{
		bucketName: name,
		headers:    []string{},
		data:       [][]interface{}{},
	}
}

func (r *Results) SetHeaders(headers []string) {
	r.headers = headers
}

func (r *Results) AddComment(comment string) {
	r.comments = append(r.comments, comment)
}

func (r *Results) AddContent(content []interface{}) {
	r.data = append(r.data, content)
}

// checkWidthsCoherence checks if headers and data are both sets, that the
// width is consistant between them.
func (r Results) checkWidthsCoherence() bool {
	headerWidth := len(r.headers)

	// check if every width in data is the same
	var dataWidth int
	for _, d := range r.data {
		if len(r.data[0]) != len(d) {
			return false
		}
	}
	if len(r.data) > 0 {
		dataWidth = len(r.data[0])
	}

	// one is not set
	if headerWidth == 0 || dataWidth == 0 {
		return true
	} else {
		return headerWidth == dataWidth
	}

}

func nrColumnsToMaxWidth(termWidth int, n int) int {
	return (termWidth - 4 - ((n - 1) * 3)) / n
}

func (r Results) formatTable(outputWidth int) string {
	if len(r.data) == 0 || len(r.headers) == 0 {
		return ""
	}

	t := table.NewWriter()
	// Unfortunately this library does not propose to enfore a width on the
	// global table and I haven't found if you can specify of a tunning for
	// every column so I'm overshooting here.
	maxWidth := nrColumnsToMaxWidth(outputWidth, len(r.headers))
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 2, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 3, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 4, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 5, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
		{Number: 6, AlignHeader: text.AlignCenter, WidthMaxEnforcer: text.WrapSoft, WidthMax: maxWidth},
	})
	headers := make(table.Row, len(r.headers))
	for i := range r.headers {
		headers[i] = r.headers[i]
	}
	t.AppendHeader(headers)
	for _, content := range r.data {
		row := make(table.Row, len(content))
		copy(row, content)
		t.AppendRow(row)
	}

	return t.Render()
}

func (r Results) Human(opts ResultsOpts) string {
	var output strings.Builder

	if opts.ShowName == nil || *opts.ShowName {
		output.WriteString(fmt.Sprintf("### %s ###\n", strings.ToUpper(r.bucketName)))
	}
	if len(r.comments) != 0 {
		if opts.ShowComments == nil || *opts.ShowComments {
			output.WriteString("Comments:\n")
			for _, comment := range r.comments {
				output.WriteString(fmt.Sprintf("- %s\n", comment))
			}
		}
	}
	if opts.ShowData == nil || *opts.ShowData {
		// the table will be written only if the headers and the data has been
		// filled
		table := r.formatTable(opts.OutputWidth)
		if table != "" {
			output.WriteString(table)
			output.WriteString("\n")
		}
	}
	return output.String()
}

func (r Results) JSON(opts ResultsOpts) (string, error) {
	if !r.checkWidthsCoherence() {
		return "", fmt.Errorf("cannot output JSON for bucket %q, inconsistence between width of headers and data", r.bucketName)
	}

	type jsonOutput struct {
		Bucket   string                   `json:"bucket"`
		Comments []string                 `json:"comments,omitempty"`
		Data     []map[string]interface{} `json:"data"`
	}

	m := make([]map[string]interface{}, 0)

	if len(r.data) != 0 && len(r.headers) != 0 {
		for _, row := range r.data {
			rowMap := make(map[string]interface{})
			for headerNr, header := range r.headers {
				rowMap[header] = row[headerNr]
			}
			m = append(m, rowMap)
		}
	}

	var b []byte
	var err error
	// if hide name and comments, directly output an array of results
	if (opts.ShowName != nil && !*opts.ShowName) && (opts.ShowComments != nil && !*opts.ShowComments) {
		b, err = json.Marshal(m)
	} else {
		o := jsonOutput{}
		// TODO maybe add omitempty
		if opts.ShowName == nil || *opts.ShowName {
			o.Bucket = r.bucketName
		}
		if opts.ShowComments == nil || *opts.ShowComments {
			o.Comments = r.comments
		}
		if opts.ShowData == nil || *opts.ShowData {
			o.Data = m
		}

		b, err = json.Marshal(o)
	}
	if err != nil {
		panic(err)
	}

	return string(b), nil
}
