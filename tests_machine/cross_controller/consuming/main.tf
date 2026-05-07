terraform {
  required_providers {
    juju = {
      source = "juju/juju"
      version = "~> 2.0.0-rc1"
    }
  }
}

variable "offering_controller_name" {
  type = string
}

variable "offering_controller_addresses" {
  type = string
}

variable "offering_controller_username" {
  type = string
}

variable "offering_controller_password" {
  type = string
} 

variable "offering_controller_ca_cert" {
  type = string 
}

provider "juju" {
  offering_controllers = {
      "${var.offering_controller_name}" = {
      controller_addresses = var.offering_controller_addresses
      username             = var.offering_controller_username
      password             = var.offering_controller_password
      ca_certificate       = var.offering_controller_ca_cert
    }
  }
}

module "model" {
  topic = "cross-controller"
  source = "../../../modules/model_random"
}

output "model_name" {
  value = module.model.name
}

resource "juju_model" "this" {
  name = module.model.name
}

resource "juju_application" "sink" {
  name = "dummy-sink"
  model_uuid = juju_model.this.uuid

  charm {
    name    = "juju-qa-dummy-sink"
  }

  constraints = "arch=${var.arch} tags=${var.tags}"
}

resource "juju_integration" "sink-source" {
  model_uuid = juju_model.this.uuid
  depends_on = [juju_application.sink]

  application {
    offering_controller = "${var.offering_controller_name}"
    offer_url = "admin/tfqa-cross-controller-offering.dummy-source"
    endpoint = "sink"
  }

  application {
    name     = juju_application.sink.name
    endpoint = "source"
  }
}
