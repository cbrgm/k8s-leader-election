package main

import (
	"context"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var environment struct {
	KubeConfig         string `name:"kubeconfig" default:"" help:"absolute path to the kubeconfig file"`
	NodeID             string `name:"node-id" required:"true" default:"" help:"the holder identity name"`
	LeaseLockName      string `name:"leaselock.name" required:"true" default:"" help:"the lease lock resource name"`
	LeaseLockNamespace string `name:"leaselock.namespace" default:"default"  help:"the lease lock resource namespace"`
}

func main() {

	_ = kong.Parse(&environment,
		kong.Name("leader-election-demo"),
	)

	config, err := getConfigFromPath(environment.KubeConfig)
	if err != nil {
		klog.Fatal(err)
	}
	client := clientset.NewForConfigOrDie(config)

	// use ctx so the leaderelection code is notified when we want to stop being a leader
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listen on SIGTERM signals, step down if fired!
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		klog.Info("Received sigterm event, shutdown")
		cancel()
	}()

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      environment.LeaseLockName,
			Namespace: environment.LeaseLockNamespace,
		},
		Client: client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: environment.NodeID,
		},
	}

	// start leader election
	// code here is blocking and will panic if something goes wrong
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   20 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			// this is where we do stuff when we are the leader
			OnStartedLeading: func(ctx context.Context) {
				klog.Infof("I am the leader, will do management stuff now: %s", environment.NodeID)
			},
			// cleanup code goes here
			OnStoppedLeading: func() {
				klog.Infof("leader lost: %s", environment.NodeID)
				os.Exit(0)
			},
			// do something when a new leader is elected
			OnNewLeader: func(identity string) {
				if identity == environment.NodeID {
					return
				}
				klog.Infof("new leader elected: %s", identity)
			},
		},
	})
}

func getConfigFromPath(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
