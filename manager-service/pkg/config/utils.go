package config

import (
	"fmt"
	"github.com/pkg/errors"
	"net"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

func getIntOrPanic(key string) int {
	intValue, err := strconv.Atoi(getConfigStringValue(key))
	if err != nil {
		panicForkey(key, err)
	}
	return intValue
}

func getBoolOrPanic(key string) bool {
	boolValue, err := strconv.ParseBool(getConfigStringValue(key))
	if err != nil {
		panicForkey(key, err)
	}
	return boolValue
}

func getStringOrPanic(key string) string {
	value := getConfigStringValue(key)
	if value == "" {
		panicForkey(key, errors.New("config is not set"))
	}
	return value
}

func getConfigStringValue(key string) string {
	if !viper.IsSet(key) && os.Getenv(key) == "" {
		fmt.Printf("config %s is not set\n", key)
	}
	value := os.Getenv(key)
	if value == "" {
		value = viper.GetString(key)
	}
	return value
}

func panicForkey(key string, err error) {
	panic(fmt.Sprintf("Error %v occured while reading config %s", err.Error(), key))
}

func WithConfig(key string, value string, block func()) {
	originalValue := viper.GetString(key)
	os.Setenv(key, value)
	Load()
	block()
	os.Setenv(key, originalValue)
	Load()
}

func resolveHostIp() (string,error) {
	netInterfaceAddresses, err := net.InterfaceAddrs()
	if err != nil {
		return "",errors.Wrap(err,"error in finding network interfaces")
	}
	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if ok && !networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil {
			ip := networkIp.IP.String()
			return ip,nil
		}
	}
	return "",errors.New("unable to get ip address")
}
