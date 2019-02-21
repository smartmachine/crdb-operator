#!/bin/bash
kubectl exec -n crdb -it ${1}-client -- ./cockroach sql --certs-dir=/cockroach-certs --host=${1}-public
