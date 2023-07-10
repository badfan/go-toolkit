package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/spf13/viper"
)

type FirestoreProvider struct {
	client *firestore.Client
	ctx    context.Context
}

func NewFirestoreProvider(ctx context.Context, projectId string) (*FirestoreProvider, error) {
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		return nil, err
	}

	return &FirestoreProvider{client: client, ctx: ctx}, nil
}

// ReadFirestoreConfig retrieves the configuration data from a remote Firestore source. Requires a path with
// the following formatting `<SERVICE_NAME>/<ENV>`
func (f *FirestoreProvider) ReadFirestoreConfig(path string) error {
	snap, _ := f.client.Doc(path).Get(f.ctx)

	jsonData, err := json.Marshal(snap.Data())
	if err != nil {
		return err
	}

	err = viper.ReadConfig(bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	return nil
}

// WatchFirestoreConfig is listening to changes in remote Firestore source. Requires a path with
// the following formatting `<SERVICE_NAME>/<ENV>`
func (f *FirestoreProvider) WatchFirestoreConfig(path string) {
	streamChanges := f.client.Doc(path).Snapshots(f.ctx)
	defer streamChanges.Stop()
	for {
		snap, err := streamChanges.Next()
		if err != nil {
			log.Fatalln(err)
		}

		jsonData, err := json.Marshal(snap.Data())
		if err != nil {
			log.Fatalln(err)
		}

		err = viper.ReadConfig(bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatalln(err)
		}
	}
}
