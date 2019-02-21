package controller

import (
	"go.smartmachine.io/crdb-operator/pkg/controller/certsigner"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, certsigner.Add)
}
