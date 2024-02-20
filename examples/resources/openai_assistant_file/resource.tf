terraform {
  required_providers {
    openai = {
      source = "registry.terraform.io/guillaume-dussault/openai"
    }
  }
  required_version = ">= 1.1.0"
}

provider "openai" {}

resource "openai_assistant" "example" {
  name             = "Test provider"
  model            = "gpt-4-turbo-preview"
  instructions     = "Answer every questions with a Chuck Norris joke. Be super friendly and casual."
  description      = "A friendly bot that tells jokes."
  enable_retrieval = true
}

resource "openai_assistant_file" "example" {
  assistant_id = openai_assistant.example.id
  filename     = "important content.txt"
}

output "assistant_file_id" {
  value = openai_assistant_file.example.id
}
