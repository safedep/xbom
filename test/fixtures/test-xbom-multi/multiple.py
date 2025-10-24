import anthropic
import openai
from google.cloud import aiplatform

# Anthropic
anthropic_client = anthropic.Anthropic()
response = anthropic_client.messages.create(
    model="claude-3-5-sonnet-20241022",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello"}]
)

# OpenAI
openai_client = openai.OpenAI()
completion = openai_client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello"}]
)

# Google Vertex AI
aiplatform.init(project="my-project")
model = aiplatform.Model.list()[0]
prediction = model.predict(instances=[{"text": "Hello"}])
