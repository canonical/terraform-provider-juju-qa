#!/bin/bash
set -ex

sleep 60 # wait for k8s action to work
juju run k8s/0 get-kubeconfig | yq -r '.kubeconfig' >> ./kubeconfig
KUBECONFIG=./kubeconfig juju add-k8s tfqa-k8s --cluster-name=k8s --client --controller=tfqa
