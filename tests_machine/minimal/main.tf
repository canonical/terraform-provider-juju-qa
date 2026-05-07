terraform {
  required_providers {
    juju = {
      source = "juju/juju"
      version = "~> 2.0.0-rc1"
    }
  }
}

provider "juju" {
}

module "model" {
  topic = "minimal"
  source = "../../modules/model_random"
}

output "model_name" {
  value = module.model.name
}

resource "juju_model" "this" {
  name = module.model.name
}

resource "juju_application" "this" {
  model_uuid = juju_model.this.uuid
  name       = "qa-test"
  charm {
    name    = "juju-qa-test"
  }

  config = {}
  constraints = "arch=${var.arch} tags=${var.tags}"
}
