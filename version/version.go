// Package version is a convenience utility that provides SDK
// consumers with a ready-to-use version command that
// produces apps versioning information based on flags
// passed at compile time.
//
// # Configure the version command
//
// The version command can be just added to your cobra root command.
// At build time, the variables Name, Version, Commit, and BuildTags
// can be passed as build flags as shown in the following example:
//
//	go build -X github.com/cosmos/cosmos-sdk/version.Name=gaia \
//	 -X github.com/cosmos/cosmos-sdk/version.ServerName=gaiad \
//	 -X github.com/cosmos/cosmos-sdk/version.ClientName=gaiacli \
//	 -X github.com/cosmos/cosmos-sdk/version.Version=1.0 \
//	 -X github.com/cosmos/cosmos-sdk/version.Commit=f0f7b7dab7e36c20b757cebce0e8f4fc5b95de60 \
//	 -X "github.com/cosmos/cosmos-sdk/version.BuildTags=linux darwin amd64"
package version

import (
	"fmt"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// application's name
	Name = ""
	// server binary name
	ServerName = "<appd>"
	// client binary name
	ClientName = "<appcli>"
	// application's version string
	Version = ""
	// commit
	Commit = ""
)

// irisdInfoGauge stores Iris git commit and version details.
var irisdInfoGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "irisd",
	Name:      "info",
	Help:      "Iris commit and version details",
}, []string{"commit", "version"})

// UpdateIrisdInfo updates the iris_info metric with the current git commit and version details.
func UpdateIrisdInfo() {
	irisdInfoGauge.WithLabelValues(
		Commit,
		Version,
	).Set(1)
}

// Info defines the application version information.
type Info struct {
	Name       string `json:"name" yaml:"name"`
	ServerName string `json:"server_name" yaml:"server_name"`
	ClientName string `json:"client_name" yaml:"client_name"`
	Version    string `json:"version" yaml:"version"`
	GitCommit  string `json:"commit" yaml:"commit"`
	GoVersion  string `json:"go" yaml:"go"`
}

func NewInfo() Info {
	return Info{
		Name:       Name,
		ServerName: ServerName,
		ClientName: ClientName,
		Version:    Version,
		GitCommit:  Commit,
		GoVersion:  fmt.Sprintf("go version %s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
	}
}

func (vi Info) String() string {
	return fmt.Sprintf(`%s: %s
git commit: %s
%s`,
		vi.Name, vi.Version, vi.GitCommit, vi.GoVersion,
	)
}
