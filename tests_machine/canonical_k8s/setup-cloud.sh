#!/bin/bash
set -ex

juju run k8s/0 get-kubeconfig | yq -r '.kubeconfig' >> ./kubeconfig
KUBECONFIG=./kubeconfig juju add-k8s tfqa-k8s --cluster-name=k8s --client --controller=tfqa
rm -f ./kubeconfig
