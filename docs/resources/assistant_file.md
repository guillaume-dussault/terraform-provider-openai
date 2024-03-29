---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "openai_assistant_file Resource - terraform-provider-openai"
subcategory: ""
description: |-
  Provides an OpenAI assistant file resource.
---

# openai_assistant_file (Resource)

Provides an OpenAI assistant file resource.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `assistant_id` (String) The ID of the assistant to which this file will be included.
- `filename` (String) Path to the file within the local filesystem.

### Read-Only

- `id` (String) ID of the file.
- `last_updated` (String) Timestamp of the last Terraform update of the assistant.
