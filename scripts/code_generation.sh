#!/bin/bash
vendor/k8s.io/code-generator/generate-groups.sh all go.smartmachine.io/crdb-operator/pkg/client go.smartmachine.io/crdb-operator/pkg/apis "db:v1alpha1"
