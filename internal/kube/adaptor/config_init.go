package adaptor

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	paths "path"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/cenkalti/backoff/v4"
	internalclient "github.com/skupperproject/skupper/internal/kube/client"
	"github.com/skupperproject/skupper/internal/kube/secrets"
	"github.com/skupperproject/skupper/internal/kube/watchers"
	"github.com/skupperproject/skupper/internal/qdr"
)

func InitialiseConfig(cli internalclient.Clients, namespace string, path string, routerConfigMap string) error {
	ctxt := context.Background()
	controller := watchers.NewEventProcessor("config-init", cli)
	secretsSync := secrets.NewSync(
		sslSecretsWatcher(namespace, controller),
		nil,
		slog.New(slog.Default().Handler()).With(slog.String("component", "kube.secrets")),
	)
	stop := make(chan struct{})
	defer close(stop)
	log.Println("Starting secret watcher")
	controller.StartWatchers(stop)
	configMaps := cli.GetKubeClient().CoreV1().ConfigMaps(namespace)
	log.Println("Waiting for secret watcher cache")
	controller.WaitForCacheSync(stop)
	secretsSync.Recover()
	controller.Start(stop)
	var (
		routerConfiguration *qdr.RouterConfig
		err                 error
	)
	retryErr := backoff.Retry(func() error {
		log.Println("Synchroninzing Secrets with router configuration")
		routerConfiguration, err = getRouterConfig(ctxt, configMaps, routerConfigMap)
		if err != nil {
			return err
		}
		if routerConfiguration == nil {
			return fmt.Errorf("empty router configuration in ConfigMap %q", routerConfigMap)
		}
		delta := secretsSync.Expect(routerConfiguration.SslProfiles)
		if len(delta.Missing) > 0 {
			log.Printf("Waiting for Secrets to be created for SslProfiles %v", delta.Missing)
		}
		for name, diff := range delta.PendingOrdinals {
			log.Printf("Secret %q has outdated ordinal %d. Profile %q wants %d", diff.SecretName, diff.Current, name, diff.Expect)
		}
		return delta.Error()
	}, backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(time.Second*60)))
	if retryErr != nil {
		return retryErr
	}
	log.Printf("Finished synchronizing Secrets with router configuration")
	value, err := qdr.MarshalRouterConfig(*routerConfiguration)
	if err != nil {
		return err
	}
	configFile := paths.Join(path, "skrouterd.json")
	if err := os.WriteFile(configFile, []byte(value), 0777); err != nil {
		return err
	}
	log.Printf("Router configuration written to %s", configFile)
	return nil
}

func getRouterConfig(ctx context.Context, configMaps v1.ConfigMapInterface, name string) (*qdr.RouterConfig, error) {
	current, err := configMaps.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return qdr.GetRouterConfigFromConfigMap(current)
}
