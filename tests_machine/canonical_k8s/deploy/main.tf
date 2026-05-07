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
  
  credential = var.credential
}

resource "juju_application" "source" {
  model_uuid = juju_model.this.uuid
  name       = "source"

  charm {
    name     = "juju-qa-dummy-source"
  }

  config = {
    token = "abc"
  }

  constraints = "arch=${var.arch}"
}

resource "juju_application" "sink" {
  model_uuid = juju_model.this.uuid
  name       = "sink"

  charm {
    name     = "juju-qa-dummy-sink"
  }

  constraints = "arch=${var.arch}"
}

resource "juju_integration" "source-sink" {
  model_uuid = resource.juju_model.this.uuid

  application {
    name = juju_application.source.name
    endpoint = "sink"
  }

  application {
    name     = juju_application.sink.name
    endpoint = "source"
  }
}
