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
  topic = "canonical-k8s-workload"
  source = "../../../modules/model_random"
}

output "model_name" {
  value = module.model.name
}

resource "juju_model" "this" {
  name = module.model.name

  cloud {
    name = var.cloud
  }
}

resource "juju_application" "this" {
  model_uuid = juju_model.this.uuid
  name       = "postgres"
  charm {
    name    = "postgresql-k8s"
    channel = "14/stable"
  }

  trust = true

  constraints = "arch=${var.arch}"
}
