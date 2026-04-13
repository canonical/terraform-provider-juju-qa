terraform {
  required_providers {
    juju = {
      source = "juju/juju"
      version = "> 1.0.0"
    }
  }
}

provider "juju" {
}

module "model" {
  topic = "canonical-k8s"
  source = "../../modules/model_random"
}

output "model_name" {
  value = module.model.name
}

resource "juju_model" "this" {
  name = module.model.name
}

resource "juju_application" "k8s" {
  name  = "k8s"
  model_uuid = juju_model.this.uuid

  charm {
    name     = "k8s"
    channel  = "1.35/stable"
    base     = "ubuntu@24.04"
  }

  config = {
    "dns-enabled" = true
    "dns-cluster-domain" = "cluster.local"
  }

  expose {
    cidrs = "0.0.0.0/0"
  }

  constraints       = "arch=${var.arch} tags=${var.tags}"
  units             = 1
}


resource "juju_application" "k8s-worker" {
  name  = "k8s-worker"
  model_uuid = juju_model.this.uuid

  charm {
    name     = "k8s-worker"
    channel  = "1.35/stable"
    base     = "ubuntu@24.04"
  }

  expose {
    cidrs = "0.0.0.0/0"
  }

  constraints       = "arch=${var.arch} tags=${var.tags}"
  units             = 2
}

resource "juju_integration" "k8s-worker-cluster" {
  model_uuid = resource.juju_model.this.uuid

  application {
    name = juju_application.k8s.name
    endpoint = "k8s-cluster"
  }

  application {
    name     = juju_application.k8s-worker.name
    endpoint = "cluster"
  }
}
