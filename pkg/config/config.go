package config

import "os"

type Config struct {
	OpenShiftHost string

	MDCContainerName        string
	OauthProxyContainerName string

	MDCImageStreamName        string
	MDCImageStreamTag         string
	OauthProxyImageStreamName string
	OauthProxyImageStreamTag  string

	MDCImageStreamInitialImage        string
	OauthProxyImageStreamInitialImage string
}

func New() Config {
	return Config{
		OpenShiftHost: getReqEnv("OPENSHIFT_HOST"),

		MDCContainerName:        getEnv("MDC_CONTAINER_NAME", "mdc"),
		OauthProxyContainerName: getEnv("OAUTH_PROXY_CONTAINER_NAME", "mdc-oauth-proxy"),

		MDCImageStreamName:        getEnv("MDC_IMAGE_STREAM_NAME", "mdc-imagestream"),
		MDCImageStreamTag:         getEnv("MDC_IMAGE_STREAM_TAG", "latest"),
		OauthProxyImageStreamName: getEnv("OAUTH_PROXY_IMAGE_STREAM_NAME", "mdc-oauth-proxy-imagestream"),
		OauthProxyImageStreamTag:  getEnv("OAUTH_PROXY_IMAGE_STREAM_TAG", "latest"),

		// these are used when the image stream does not exist and created for the first time by the operator
		MDCImageStreamInitialImage:        getEnv("MDC_IMAGE_STREAM_INITIAL_IMAGE", "quay.io/aerogear/mobile-developer-console:latest"),
		OauthProxyImageStreamInitialImage: getEnv("OAUTH_PROXY_IMAGE_STREAM_INITIAL_IMAGE", "docker.io/openshift/oauth-proxy:v1.1.0"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getReqEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	panic("Required env var is missing: " + key)
}
