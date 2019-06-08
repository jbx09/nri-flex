package load

import (
	"sync"
	"time"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/integration"
	logrus "github.com/sirupsen/logrus"
)

// ArgumentList Available Arguments
type ArgumentList struct {
	sdkArgs.DefaultArgumentList
	ForceLogEvent         bool   `default:"false" help:"Force create an event for everything - useful for testing"`
	OverrideIPMode        string `default:"" help:"Force override ipMode used for container discovery set as private or public - useful for testing"`
	Local                 bool   `default:"true" help:"Collect local entity info"`
	ConfigFile            string `default:"" help:"Set a specific config file - not usable for container discovery"`
	ConfigDir             string `default:"flexConfigs/" help:"Set directory of config files"`
	ContainerDiscoveryDir string `default:"flexContainerDiscovery/" help:"Set directory of auto discovery config files"`
	ContainerDiscovery    bool   `default:"false" help:"Enable container auto discovery"`
	DockerAPIVersion      string `default:"" help:"Force Docker client API version"`
	EventLimit            int    `default:"500" help:"Event limiter - max amount of events per execution"`
	Entity                string `default:"" help:"Manually set a remote entity name"`
	InsightsURL           string `default:"" help:"Set Insights URL"`
	InsightsAPIKey        string `default:"" help:"Set Insights API key"`
	InsightsOutput        bool   `default:"false" help:"Output the events generated to standard out"`
	MetricAPIUrl          string `default:"https://metric-api.newrelic.com/metric/v1" help:"Set Metric API URL"`
	MetricAPIKey          string `default:"" help:"Set Metric API key"`

	// not implemented yet
	// InsightsInterval      int    `default:"0" help:"Run Insights mode periodically at this set interval"`
	// ClusterModeKey string `default:"" help:"Set key used for cluster mode identification"`
	// ClusterModeExp string `default:"60s" help:"Set cluster mode key identifier expiration"`
}

// Args Infrastructure SDK Arguments List
var Args ArgumentList

// Integration Infrastructure SDK Integration
var Integration *integration.Integration

// Entity Infrastructure SDK Entity
var Entity *integration.Entity

// Hostname current host
var Hostname string

// ContainerID current container id
var ContainerID string

// Logrus create instance of the logger
var Logrus = logrus.New()

var IntegrationName = "com.newrelic.nri-flex" // IntegrationName Name
var IntegrationNameShort = "nri-flex"         // IntegrationNameShort Short Name
var IntegrationVersion = "Unknown-SNAPSHOT"   // IntegrationVersion Version

const (
	DefaultSplitBy     = ":"                      // unused currently
	DefaultTimeout     = 10000 * time.Millisecond // 10 seconds, used for raw commands
	DefaultDialTimeout = 1000                     // 1 seconds, used for dial
	DefaultPingTimeout = 5000                     // 5 seconds
	DefaultPostgres    = "postgres"
	DefaultMSSQLServer = "sqlserver"
	DefaultMySQL       = "mysql"
	DefaultOracle      = "ora"
	DefaultJmxPath     = "./nrjmx/"
	DefaultJmxHost     = "127.0.0.1"
	DefaultJmxPort     = "9999"
	DefaultJmxUser     = "admin"
	DefaultJmxPass     = "admin"
	DefaultShell       = "/bin/sh"
	DefaultLineLimit   = 255
	Public             = "public"
	Private            = "private"
	Jmx                = "jmx"
	Img                = "img"
	TypeJSON           = "json"
	TypeColumns        = "columns"
)

// FlexStatusCounter count internal metrics
var FlexStatusCounter = struct {
	sync.RWMutex
	M map[string]int
}{M: make(map[string]int)}

// StatusCounterIncrement increment the status counter for a particular key
func StatusCounterIncrement(key string) {
	FlexStatusCounter.Lock()
	FlexStatusCounter.M[key]++
	FlexStatusCounter.Unlock()
}

// StatusCounterRead the status counter for a particular key
func StatusCounterRead(key string) int {
	FlexStatusCounter.Lock()
	value := FlexStatusCounter.M[key]
	FlexStatusCounter.Unlock()
	return value
}

// MetricsStore for Dimensional Metrics to store data and lock and unlock when needed
var MetricsStore = struct {
	sync.RWMutex
	Data []Metrics
}{}

// MetricsStoreAppend Append data to store
func MetricsStoreAppend(metrics Metrics) {
	MetricsStore.Lock()
	MetricsStore.Data = append(MetricsStore.Data, metrics)
	MetricsStore.Unlock()
}

// MetricsStoreEmpty empties stored data
func MetricsStoreEmpty() {
	MetricsStore.Lock()
	MetricsStore.Data = []Metrics{}
	MetricsStore.Unlock()
}

// Metrics struct
type Metrics struct {
	TimestampMs      int64                    `json:"timestamp.ms,omitempty"` // required for every metric at root or nested
	IntervalMs       int64                    `json:"interval.ms,omitempty"`  // required for count & summary
	CommonAttributes map[string]interface{}   `json:"commonAttributes,omitempty"`
	Metrics          []map[string]interface{} `json:"metrics"` // summaries have a different value structure then gauges or counters
}

// Config YAML Struct
type Config struct {
	FileName         string // this will be set when files are read
	Name             string
	Global           Global
	APIs             []API
	Datastore        map[string][]interface{} `yaml:"datastore"`
	LookupStore      map[string][]string      `yaml:"lookup_store"`
	LookupFile       string                   `yaml:"lookup_file"`
	VariableStore    map[string]string        `yaml:"variable_store"`
	CustomAttributes map[string]string        `yaml:"custom_attributes"` // set additional custom attributes
	MetricAPI        bool                     `yaml:"metric_api"`        // enable use of the dimensional data models metric api
}

// Global struct
type Global struct {
	BaseURL    string `yaml:"base_url"`
	User, Pass string
	Proxy      string
	Timeout    int
	Headers    map[string]string `yaml:"headers"`
	Jmx        JMX               `yaml:"jmx"`
	TLSConfig  TLSConfig         `yaml:"tls_config"`
}

// TLSConfig struct
type TLSConfig struct {
	Enable             bool   `yaml:"enable"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	MinVersion         uint16 `yaml:"min_version"`
	MaxVersion         uint16 `yaml:"max_version"`
	Ca                 string `yaml:"ca"` // path to ca to read
}

// SampleMerge merge multiple samples into one (will remove previous samples)
type SampleMerge struct {
	EventType string   `yaml:"event_type"` // new event_type name for the sample
	Samples   []string `yaml:"samples"`    // list of samples to be merged
}

// API YAML Struct
type API struct {
	Name              string            `yaml:"name"`
	EventType         string            `yaml:"event_type"`     // override eventType
	Entity            string            `yaml:"entity"`         // define a custom entity name
	EntityType        string            `yaml:"entity_type"`    // define a custom entity type (namespace)
	Inventory         map[string]string `yaml:"inventory"`      // set as inventory
	InventoryOnly     bool              `yaml:"inventory_only"` // only generate inventory data
	Events            map[string]string `yaml:"events"`         // set as events
	EventsOnly        bool              `yaml:"events_only"`    // only generate events
	Merge             string            `yaml:"merge"`          // merge into another eventType
	Prefix            string            `yaml:"prefix"`         // prefix attribute keys
	File              string            `yaml:"file"`
	URL               string            `yaml:"url"`
	EscapeURL         bool              `yaml:"escape_url"`
	Prometheus        Prometheus        `yaml:"prometheus"`
	Cache             string            `yaml:"cache"` // read data from datastore
	Database          string            `yaml:"database"`
	DbDriver          string            `yaml:"db_driver"`
	DbConn            string            `yaml:"db_conn"`
	Shell             string            `yaml:"shell"`
	CommandsAsync     bool              `yaml:"commands_async"` // run commands async
	Commands          []Command         `yaml:"commands"`
	DbQueries         []Command         `yaml:"db_queries"`
	Jmx               JMX               `yaml:"jmx"`
	IgnoreLines       []int             // not implemented - idea is to ignore particular lines starting from 0 of the command output
	User, Pass        string
	Proxy             string
	TLSConfig         TLSConfig `yaml:"tls_config"`
	Timeout           int
	Method            string
	Payload           string
	DisableParentAttr bool                `yaml:"disable_parent_attr"`
	Headers           map[string]string   `yaml:"headers"`
	StartKey          []string            `yaml:"start_key"`
	StoreLookups      map[string]string   `yaml:"store_lookups"`
	StoreVariables    map[string]string   `yaml:"store_variables"`
	StripKeys         []string            `yaml:"strip_keys"`
	LazyFlatten       []string            `yaml:"lazy_flatten"`
	SampleKeys        map[string]string   `yaml:"sample_keys"`
	ReplaceKeys       map[string]string   `yaml:"replace_keys"`   // uses rename_keys functionality
	RenameKeys        map[string]string   `yaml:"rename_keys"`    // use regex to find keys, then replace value
	RenameSamples     map[string]string   `yaml:"rename_samples"` // using regex if sample has a key that matches, make that a different sample
	RemoveKeys        []string            `yaml:"remove_keys"`
	KeepKeys          []string            `yaml:"keep_keys"`       // inverse of removing keys
	SkipProcessing    []string            `yaml:"skip_processing"` // skip processing particular keys using an array of regex strings
	ToLower           bool                `yaml:"to_lower"`        // convert all unicode letters mapped to their lower case.
	ConvertSpace      string              `yaml:"convert_space"`   // convert spaces to another char
	SnakeToCamel      bool                `yaml:"snake_to_camel"`
	PercToDecimal     bool                `yaml:"perc_to_decimal"` // will check strings, and perform a trimRight for the %
	PluckNumbers      bool                `yaml:"pluck_numbers"`   // plucks numbers out of the value
	Math              map[string]string   `yaml:"math"`            // perform match across processed metrics
	SubParse          []Parse             `yaml:"sub_parse"`
	CustomAttributes  map[string]string   `yaml:"custom_attributes"` // set additional custom attributes
	ValueParser       map[string]string   `yaml:"value_parser"`      // find keys with regex, and parse the value with regex
	ValueTransformer  map[string]string   `yaml:"value_transformer"` // find key(s) with regex, and modify the value
	MetricParser      MetricParser        `yaml:"metric_parser"`     // to use the MetricParser for setting deltas and gauges a namespace needs to be set
	SampleFilter      []map[string]string `yaml:"sample_filter"`     // sample filter key pair values with regex
	SplitObjects      bool                `yaml:"split_objects"`     // convert object with nested objects to array
	Split             string              `yaml:"split"`             // default vertical, can be set to horizontal (column) useful for tabular outputs
	SplitBy           string              `yaml:"split_by"`          // character to split by
	SetHeader         []string            `yaml:"set_header"`        // manually set header column names
	Regex             bool                `yaml:"regex"`             // process SplitBy as regex
	RowHeader         int                 `yaml:"row_header"`        // set the row header, to be used with SplitBy
	RowStart          int                 `yaml:"row_start"`         // start from this line, to be used with SplitBy
	Logging           struct {            // log to insights
		Open bool `yaml:"open"` // log open related errors
	}
}

// Command Struct
type Command struct {
	Name             string            `yaml:"name"`              // required for database use
	EventType        string            `yaml:"event_type"`        // override eventType (currently used for db only)
	Shell            string            `yaml:"shell"`             // command shell
	Cache            string            `yaml:"cache"`             // use content from cache instead of a run command
	Run              string            `yaml:"run"`               // runs commands, but if database is set, then this is used to run queries
	Jmx              JMX               `yaml:"jmx"`               // if wanting to run different jmx endpoints to merge
	CompressBean     bool              `yaml:"compress_bean"`     // compress bean name //unused
	IgnoreOutput     bool              `yaml:"ignore_output"`     // can be useful for chaining commands together
	MetricParser     MetricParser      `yaml:"metric_parser"`     // not used yet
	CustomAttributes map[string]string `yaml:"custom_attributes"` // set additional custom attributes
	Output           string            `yaml:"output"`            // jmx, raw, json
	LineEnd          int               `yaml:"line_end"`          // stop processing command output after a certain amount of lines
	LineStart        int               `yaml:"line_start"`        // start from this line
	Timeout          int               `yaml:"timeout"`           // command timeout
	Dial             string            `yaml:"dial"`              // eg. google.com:80
	Network          string            `yaml:"network"`           // default tcp

	// Parsing Options - Body
	Split       string `yaml:"split"`        // default vertical, can be set to horizontal (column) useful for outputs that look like a table
	SplitBy     string `yaml:"split_by"`     // character/match to split by
	SplitOutput string `yaml:"split_output"` // split output by found regex
	RegexMatch  bool   `yaml:"regex_match"`  // process SplitBy as a regex match
	GroupBy     string `yaml:"group_by"`     // group by character
	RowHeader   int    `yaml:"row_header"`   // set the row header, to be used with SplitBy
	RowStart    int    `yaml:"row_start"`    // start from this line, to be used with SplitBy

	// Parsing Options - Header
	SetHeader        []string `yaml:"set_header"`         // manually set header column names (used when split is is set to horizontal)
	HeaderSplitBy    string   `yaml:"header_split_by"`    // character/match to split header by
	HeaderRegexMatch bool     `yaml:"header_regex_match"` // process HeaderSplitBy as a regex match

	// RegexMatches
	RegexMatches []RegMatch `yaml:"regex_matches"`
}

type RegMatch struct {
	Expression string   `yaml:"expression"`
	Keys       []string `yaml:"keys"`
	KeysMulti  []string `yaml:"keys_multi"`
}

// Prometheus struct
type Prometheus struct {
	Enable           bool              `yaml:"enable"`
	Unflatten        bool              `yaml:"unflatten"`       // unflattens all counters and gauges into separate metric samples retaining all their metadata // make this map[string]string
	FlattenedEvent   string            `yaml:"flattened_event"` // name of the flattenedEvent
	KeyMerge         []string          `yaml:"key_merge"`       // list of keys to merge into the key name when flattening, not usable when unflatten set to true
	KeepLabels       bool              `yaml:"keep_labels"`     // not usable when unflatten set to true
	KeepHelp         bool              `yaml:"keep_help"`       // not usable when unflatten set to true
	CustomAttributes map[string]string `yaml:"custom_attributes"`
	SampleKeys       map[string]string `yaml:"sample_keys"`
	Histogram        bool              `yaml:"histogram"`       // if flattening by default, create a full histogram sample
	HistogramEvent   string            `yaml:"histogram_event"` // override histogram event type
	Summary          bool              `yaml:"summary"`         // if flattening by default, create a full summary sample
	SummaryEvent     string            `yaml:"summaryevent"`    // override summary event type
	GoMetrics        bool              `yaml:"go_metrics"`      // enable go metrics
}

// JMX struct
type JMX struct {
	Domain         string `yaml:"domain"`
	User           string `yaml:"user"`
	Pass           string `yaml:"pass"`
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	KeyStore       string `yaml:"key_store"`
	KeyStorePass   string `yaml:"key_store_pass"`
	TrustStore     string `yaml:"trust_store"`
	TrustStorePass string `yaml:"trust_store_pass"`
}

// Parse struct
type Parse struct {
	Type    string   `yaml:"type"` // perform a contains, match, hasPrefix or regex for specified key
	Key     string   `yaml:"key"`
	SplitBy []string `yaml:"split_by"`
}

// MetricParser Struct
type MetricParser struct {
	Namespace Namespace                         `yaml:"namespace"`
	Metrics   map[string]string                 `yaml:"metrics"`  // inputBytesPerSecond: RATE
	AutoSet   bool                              `yaml:"auto_set"` // if set to true, will attempt to do a contains instead of a direct key match, this is useful for setting multiple metrics
	Counts    map[string]int64                  `yaml:"counts"`
	Summaries map[string]map[string]interface{} `yaml:"summaries"`
}

// Namespace Struct
type Namespace struct {
	// if neither of the below are set and the MetricParser is used, the namespace will default to the "Name" attribute
	CustomAttr   string   `yaml:"custom_attr"`   // set your own custom namespace attribute
	ExistingAttr []string `yaml:"existing_attr"` // utilise existing attributes and chain together to create a custom namespace
}

// Refresh Helper function used for testing
func Refresh() {
	FlexStatusCounter.M = make(map[string]int)
	FlexStatusCounter.M["EventCount"] = 0
	FlexStatusCounter.M["EventDropCount"] = 0
	FlexStatusCounter.M["ConfigsProcessed"] = 0
	Args.ConfigDir = ""
	Args.ConfigFile = ""
	Args.ContainerDiscovery = false
	Args.ContainerDiscoveryDir = ""
}
