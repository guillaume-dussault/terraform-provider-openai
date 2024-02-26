resource "openai_assistant" "example" {
  name         = "Test provider"
  model        = "gpt-4-turbo-preview"
  instructions = "Answer every questions with a Chuck Norris joke. Be super friendly and casual."
  description  = "A friendly bot that tells jokes."
}

output "assistant_name" {
  value = openai_assistant.example.name
}
