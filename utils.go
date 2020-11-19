package main

import (
	"fmt"
	"net/url"
	"strconv"
)

func getParam(values url.Values, name string) (string, error) {
	if val, ok := values[name]; ok && len(val) > 0 {
		return val[0], nil
	} else {
		return "", fmt.Errorf("Missing parameter: %s", name)
	}
}

func getUrlParam(values url.Values, name string) (*url.URL, error) {
	if val, err := getParam(values, name); err != nil {
		return nil, err
	} else {
		return url.Parse(val)
	}
}

func getIntParam(values url.Values, name string) (int, error) {
	if val, err := getParam(values, name); err != nil {
		return 0, err
	} else {
		return strconv.Atoi(val)
	}
}
