# AI Travel Agent

Driver-side AI travel reminder service for the RideHailing demo.

## Environment

PowerShell startup environment:

```powershell
$env:DEEPSEEK_API_KEY="你的key"
$env:DEEPSEEK_MODEL="deepseek-chat"
$env:AMAP_KEY="你的高德Web服务key"
$env:AI_TRAVEL_AGENT_URL="http://127.0.0.1:8011"
```

The service uses the OpenAI-compatible DeepSeek chat completions endpoint at
`https://api.deepseek.com/chat/completions` by default.
Idle-driver nearby warnings use Amap Web Service APIs through `AMAP_KEY`.

## Run

```powershell
cd ai-travel-agent
pip install -r requirements.txt
python app.py
```

Gateway default:

```powershell
$env:AI_TRAVEL_AGENT_URL="http://127.0.0.1:8011"
```
