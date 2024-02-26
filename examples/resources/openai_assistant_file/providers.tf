terraform {
  required_providers {
    openai = {
      source  = "registry.terraform.io/guillaume-dussault/openai"
      version = "1.0.0-pre.2"
    }
  }
  required_version = ">= 1.1.0"
}

provider "openai" {}
