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
  source = "../../../modules/model_random"
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

  expose {
    cidrs = "0.0.0.0/0"
  }

  config = {
    "local-storage-enabled" = false
  }

  constraints       = "arch=${var.arch} tags=${var.tags} cores=2 mem=4G ${var.extra-constraints}"
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

  constraints       = "arch=${var.arch} tags=${var.tags} cores=1 mem=2G ${var.extra-constraints}"
  units             = 2
}

resource "juju_integration" "k8s-worker-cluster" {
  model_uuid = resource.juju_model.this.uuid

  application {
    name = juju_application.k8s.name
  }

  application {
    name     = juju_application.k8s-worker.name
    endpoint = "cluster"
  }
}

resource "juju_integration" "k8s-worker-containerd" {
  model_uuid = resource.juju_model.this.uuid

  application {
    name = juju_application.k8s.name
  }

  application {
    name     = juju_application.k8s-worker.name
    endpoint = "containerd"
  }
}

resource "juju_integration" "k8s-worker-cos-tokens" {
  model_uuid = resource.juju_model.this.uuid

  application {
    name = juju_application.k8s.name
  }

  application {
    name     = juju_application.k8s-worker.name
    endpoint = "cos-tokens"
  }
}
