package bucket

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"k8s.io/client-go/kubernetes"
)

var ErrMissingClient = errors.New("buckets need a kubernetes client for initialization")

type UnknownBucketError struct {
	name string
}

func (e UnknownBucketError) Error() string {
	return fmt.Sprintf("unknown bucket %q", e.name)
}

type Runnable interface {
	Run() (Results, error)
}

type Interface interface {
	Runnable
}

type Factory func(config Config) (Interface, error)

type Bucket struct {
	Name        string
	Description string
	Aliases     []string
	Factory     Factory
	// has side-effects on its environment or is just readonly
	SideEffects bool
	// requires a client to communicate with the API server
	RequireClient bool
}

type Buckets struct {
	lock     sync.RWMutex
	registry map[string]Bucket
	// this is a "compilation" of all aliases and names pointing to names
	aliases map[string]string
}

type Config struct {
	Client      kubernetes.Interface
	Namespace   string
	Color       bool
	OutputWidth int
	// This options is specific to the admission plugin, is it to force creation
	// even if we can't cleanup the mess with delete
	AdmForce bool
	// This options is specific to the admission plugin, is it to actually create
	// pod instead of use the dry run
	AdmCreate bool
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
		if !b.SideEffects {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// Register registers a plugin Factory by name.
// This is expected to happen during app startup.
// Register does not return an error but panic
func (bs *Buckets) Register(b Bucket) {
	// verify the inputs, aliases can be nil
	if b.Name == "" {
		panic("register: bucket name must be non empty")
	}
	if b.Description == "" {
		panic("register: bucket description must be non empty")
	}
	if b.Factory == nil {
		panic("register: bucket factory must be non nil")
	}

	bs.lock.Lock()
	defer bs.lock.Unlock()

	// check if the plugin wasn't already registered
	if bs.registry != nil {
		_, found := bs.registry[b.Name]
		if found {
			panic(fmt.Sprintf("bucket %q was registered twice", b.Name))
		}
	} else {
		bs.registry = map[string]Bucket{}
	}

	// register the plugin
	bs.registry[b.Name] = b

	// register the aliases of the plugin
	if bs.aliases == nil {
		bs.aliases = map[string]string{}
	}
	// name unicity was already checked
	bs.aliases[b.Name] = b.Name
	for _, alias := range b.Aliases {
		if _, found := bs.aliases[alias]; found {
			panic(fmt.Sprintf("bucket registration: aliases %q conflict for bucket %q", alias, b.Name))
		}
		bs.aliases[alias] = b.Name
	}
}

func (bs *Buckets) ResolveAlias(alias string) (string, bool) {
	e, found := bs.findEntryFromAlias(alias)
	return e.Name, found
}

func (bs *Buckets) findEntryFromAlias(alias string) (Bucket, bool) {
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

	ret, err := f.Factory(config)
	return ret, true, err
}

// InitBucket creates an instance of the named interface.
func (bs *Buckets) InitBucket(name string, config Config) (Interface, error) {
	if name == "" {
		return nil, errors.New("no bucket specified in initialization")
	}

	bucket, found, err := bs.getBucket(name, config)
	if errors.Is(err, ErrMissingClient) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("couldn't init bucket %q: %w", name, err)
	}
	if !found {
		err := UnknownBucketError{
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
	return e.Description
}

func (bs *Buckets) Aliases(name string) []string {
	e, found := bs.findEntryFromAlias(name)
	if !found {
		return []string{""}
	}
	return e.Aliases
}

func (bs *Buckets) HasSideEffects(name string) bool {
	e, found := bs.findEntryFromAlias(name)
	if !found {
		return false
	}
	return e.SideEffects
}

func (bs *Buckets) RequiresClient(name string) bool {
	e, found := bs.findEntryFromAlias(name)
	if !found {
		return false
	}
	return e.RequireClient
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
