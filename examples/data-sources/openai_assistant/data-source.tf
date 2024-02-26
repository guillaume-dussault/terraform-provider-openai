data "openai_assistant" "example" {
  id = "your-openai-assistant-id"
}

output "assistant_name" {
  value = data.openai_assistant.example.name
}
