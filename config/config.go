package config

import (
	"context"

	"github.com/badfan/go-toolkit/config/providers"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// Config defines the data structure of the possible configurations applied to a µ-service
type Config struct {
	ServiceName string `yaml:"service_name" json:"service_name"`
}

// Options defines the data required to retrieve and use a config file
type Options struct {
	ProjectID   string `json:"project_id"`
	ServiceName string `json:"service_name"`
	Environment string `default:"local" json:"environment"`
	Persistent  bool   `json:"persistent"`
}

// FirestorePath creates a path to a firestore doc using a string interpolation between service name and environment
func (co *Options) FirestorePath() string {
	return co.ServiceName + "/" + co.Environment
}

// NewConfig creates a new Viper instance that holds the µ-service's configurations.
func NewConfig(options Options) error {
	viper.SetConfigType("json")
	viper.AutomaticEnv()

	ctx := context.Background()
	path := options.FirestorePath()
	fs, err := providers.NewFirestoreProvider(ctx, options.ProjectID)
	err = fs.ReadFirestoreConfig(path)
	if err != nil {
		return err
	}

	go fs.WatchFirestoreConfig(path)

	return nil
}
