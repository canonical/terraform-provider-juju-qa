variable "arch" {
    type        = string
    description = "CPU architecture for app constraints"
    default     = "amd64"
}

variable "cloud" {
    type        = string
    description = "Cloud to deploy to"
    default     = "tfqa-k8s"
}
