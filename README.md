# QA
 
## Locally

Bootstrap two new controllers:
```shell
juju bootstrap microk8s tfqa
juju bootstrap microk8s tfqa-offering
```

Run all tests:
```shell
make test
```

Or run only a specific test:
```shell
make run=PrivateRegistry test
```

## SolQA cluster

- Go to [SolQA TOR3 MAAS workflows](https://github.com/canonical/terragrunt-deployment-pipelines/actions/workflows/maas_physical.yaml)
- Click `Run workflow`.
- Pick a cluster, ideally one that doesn't already have a workflow currently running.
    - Also check that the cluster isn't used in another pipeline, [locks are issues in this repo](https://github.com/canonical/workflow-coordinator/issues)
- Pick `solution`, since we're bootstrapping our own controllers.
    - `composite` bootstraps a controller for us, but doesn't allow for a second one.
- Pass in `no_product` as the product, since we're not testing a deployed product.
- Pass in SQA test injection parameters including the repo and branch you want to run.
    - Like `{"repo": "canonical/terraform-provider-juju-qa", "ref": "main"}`
- Click `Run workflow`.

### Debugging

You can choose in the `Which step to sleep after` dropdown the `Run SQA tests` tests.
Once that test runs, the logs will contain an IP you can SSH into as the `ubuntu` user,
as long as you have the VPN on.

You can run `gtr` in that shell to be taken to working directory of the tests. Inside that,
`tests/sqa_tests_repo` will be the repository you pointed the pipeline to.


# Constraints

## Tags

Every resource (Juju controller, application, etc.) must have tags attached to ensure it runs against the correct MAAS cluster.

Tags look like `category,cluster`, as an example: `juju_upgrade,sqa-dh1_j8_1`.

Tag inventory summary:
- `juju` 3 machines (virtual, for a controller)
- `juju_upgrade` 3 machines (virtual, for a controller)
- `microk8s` 3 machines (virtual)
- `vault` 3 machines (virtual)
- `foundation-nodes` 9 machines (metal)

## Arch

Arch is also required to get resources scheduled, it's always `amd64

In practice, all TF plans have this constraint on all resources where it can be set:
```
    constraints = "arch=${var.arch} tags=${var.tags}"
```

## Canonical k8s on LXD

You have to add `virt-type=virtual-machine` to constraints for Canonical k8s to run on LXD.

The Canonical k8s plan accepts the `extra-constraints` var, where you can add this.
