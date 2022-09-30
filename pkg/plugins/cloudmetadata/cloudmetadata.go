package cloudmetadata

import (
	"errors"
	"net/http"
	"time"

	"github.com/quarkslab/kdigger/pkg/bucket"
)

const (
	bucketName        = "cloudmetadata"
	bucketDescription = "Cloudmetadata scans the usual metadata endpoints in public clouds."
)

var bucketAliases = []string{"cloud", "meta"}

// Thanks to:
// - https://gist.github.com/jhaddix/78cece26c91c6263653f31ba453e273b
// - https://gist.github.com/BuffaloWill/fa96693af67e3a3dd3fb
// - https://github.com/Prinzhorn/cloud-metadata-services
// I selected only one endpoint for each, trying to be exclusive, maybe it would
// be more robust to have multiples, with the ones that needs headers to be
// added (a security mesure)
//
//	type CloudEndpoint struct {
//		URLs    []string
//		Headers map[string]string
//	}
var endpoints = map[string]string{
	"DigitalOcean": "http://169.254.169.254/metadata/v1.json",
	"AWS":          "http://169.254.169.254/latest",
	"OracleCloud":  "http://192.0.0.192/latest/",
	"Alibaba":      "http://100.100.100.200/latest/meta-data/",
	"GoogleCloud":  "http://metadata.google.internal/computeMetadata/", // metadata.google.internal = 169.254.169.254
	"PacketCloud":  "https://metadata.packet.net/userdata",
	"Azure":        "http://169.254.169.254/metadata/v1/maintenance",
	"OpenStack":    "http://169.254.169.254/openstack",
}

type Bucket struct{}

// wait for 100ms maximum, request should be quick
const networkTimeout = 100 * time.Millisecond

// This plugin is "slow" because it has a network timeout on scan
func (n Bucket) Run() (bucket.Results, error) {
	res := bucket.NewResults(bucketName)

	scanResult := scanEndpoints(endpoints)

	res.SetHeaders([]string{"cloudProvider", "success", "url", "error"})
	for _, resp := range scanResult {
		if resp.Error != nil {
			res.AddContent([]interface{}{resp.Platform, resp.Success, resp.URL, resp.Error.Error()})
		} else {
			res.AddContent([]interface{}{resp.Platform, resp.Success, resp.URL, ""})
		}
	}

	return *res, nil
}

type Response struct {
	Platform string
	URL      string
	Success  bool
	Error    error
}

func scanEndpoints(endpoints map[string]string) []Response {
	client := http.Client{
		Timeout: networkTimeout,
	}

	chResponses := make(chan Response, len(endpoints))

	for platform, url := range endpoints {
		go func(ch chan Response, platform string, url string) {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				ch <- Response{
					Platform: platform,
					URL:      url,
					Success:  false,
					Error:    err,
				}
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				// chances of timeout here
				// serr, ok := err.(*urlpkg.Error)
				// if ok && serr.Timeout() {
				// }
				ch <- Response{
					Platform: platform,
					URL:      url,
					Success:  false,
					Error:    err,
				}
				return
			}

			// defer resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound {
				ch <- Response{
					Platform: platform,
					URL:      url,
					Success:  false,
					Error:    errors.New("not found"),
				}
				return
			}
			ch <- Response{
				Platform: platform,
				URL:      url,
				Success:  resp.StatusCode == http.StatusOK,
			}
		}(chResponses, platform, url)
	}

	var results []Response
	for i := 0; i < cap(chResponses); i++ {
		results = append(results, <-chResponses)
	}

	return results
}

// Register registers a plugin
func Register(b *bucket.Buckets) {
	b.Register(bucket.Bucket{
		Name:        bucketName,
		Description: bucketDescription,
		Aliases:     bucketAliases,
		Factory: func(config bucket.Config) (bucket.Interface, error) {
			return NewCloudMetadataBucket(config)
		},
		SideEffects:   false,
		RequireClient: false,
	})
}

func NewCloudMetadataBucket(config bucket.Config) (*Bucket, error) {
	return &Bucket{}, nil
}
