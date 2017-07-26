package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/Knetic/govaluate"
)

var functions = map[string]govaluate.ExpressionFunction{
	"listContains": func(args ...interface{}) (interface{}, error) {
		list, found := args[0].([]string)
		if found == false {
			return nil, fmt.Errorf("First argument of listContains is not a list")
		}
		for _, b := range list {
			if b == args[1].(string) {
				return true, nil
			}
		}
		return false, nil
	},
	"strlen": func(args ...interface{}) (interface{}, error) {
		log.Infof("%v", args[0])
		str, valid := args[0].(string)
		if valid == false {
			return nil, fmt.Errorf("%v is not a string", str)
		}
		length := len(str)
		return (float64)(length), nil
	},
}
