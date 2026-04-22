.PHONY: test clean

run ?= ".*"
arch ?= "amd64"

test:
	TF_VAR_arch="$(arch)" go test -v -count=1 -timeout=1h -run "TestQA_$(run)" ./...

clean:
	find . -name "terraform.tfstate*" -delete
	find . -name ".terraform.lock.hcl" -delete
	find . -name "kubeconfig" -delete
	find . -type d -name ".terraform" -exec rm -rf {} +
	rm -f test_results.xml test_results.raw
