terraform {
  required_providers {
    openai = {
      source = "registry.terraform.io/guillaume-dussault/openai"
    }
  }
  required_version = ">= 1.1.0"
}

provider "openai" {}

data "openai_assistant" "example" {
  id = "your-openai-assistant-id"
}

output "assistant_name" {
  value = data.openai_assistant.example.name
}
