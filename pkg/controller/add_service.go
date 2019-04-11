package controller

import (
	"gitlab.corp.redhat.com/ykoer/serving-cert-keystore-operator/pkg/controller/service"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, service.Add)
}
