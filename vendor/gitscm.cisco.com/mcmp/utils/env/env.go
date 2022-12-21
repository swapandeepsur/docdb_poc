/*
Package env defines the supported runtime environment configured metadata for a deployed microservice within Kubernetes.
*/
package env

import (
	"github.com/spf13/viper"
)

// all configuration keys.
const (
	// Environment Variable: "SERVICE_NAME".
	SvcName = "svcname"
	// Environment Variable: "APP_REGION".
	SvcRegion = "app.region"
	// Environment Variable: "APP_LIFECYCLE".
	Lifecycle = "app.lifecycle"

	// Environment Variable: "NODE_NAME".
	NodeName = "node.name"
	// Environment Variable: "POD_IP".
	PodIP = "pod.ip"
	// Environment Variable: "POD_NAME".
	PodName = "pod.name"
	// Environment Variable: "POD_NAMESPACE".
	PodNS = "pod.namespace"
)

func init() {
	initialize()
}

func initialize() {
	_ = viper.BindEnv(Lifecycle, "APP_LIFECYCLE")
	_ = viper.BindEnv(SvcRegion, "APP_REGION")
	_ = viper.BindEnv(SvcName, "SERVICE_NAME")

	_ = viper.BindEnv(NodeName, "NODE_NAME")
	_ = viper.BindEnv(PodIP, "POD_IP")
	_ = viper.BindEnv(PodName, "POD_NAME")
	_ = viper.BindEnv(PodNS, "POD_NAMESPACE")
}
