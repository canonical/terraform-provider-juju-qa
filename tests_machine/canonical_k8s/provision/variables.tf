variable "tags" {
    type        = string
    description = "Tags for all app constraints, including for physical machines"
    default    = ""
}

variable "arch" {
    type        = string
    description = "CPU architecture for app constraints"
    default     = "amd64"
}

variable "cloud" {
    type        = string
    description = "Cloud to deploy to"
    default     = ""
}

variable "extra-constraints" {
    type        = string
    description = "Extra constraints to add to all apps"
    default     = ""
}
