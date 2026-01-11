import json
import os
import sys
import requests
from openai import OpenAI

diff_path = sys.argv[1]
diff = open(diff_path).read()

client = OpenAI(api_key=os.environ["OPENAI_API_KEY"])

PROMPT = f"""
You are a senior software engineer performing code review.

Review the following git diff.

Classify issues into:
- CRITICAL: bugs, crashes, security, data loss
- MAJOR: logic issues, race conditions
- MINOR: style, refactor suggestions

Rules:
- If any CRITICAL exists -> approval = false
- Otherwise -> approval = true

Output JSON ONLY:

{{
  "approval": boolean,
  "issues": [
    {{
      "severity": "CRITICAL|MAJOR|MINOR",
      "file": "path",
      "line": number,
      "message": "description",
      "suggestion": "how to fix"
    }}
  ]
}}

Diff:
{diff}
"""

resp = client.chat.completions.create(
    model="gpt-4.1-mini",
    messages=[{"role": "user", "content": PROMPT}],
)

result = json.loads(resp.choices[0].message.content)

issues = result["issues"]
approval = result["approval"]

repo = os.environ["REPO"]
pr = os.environ["PR_NUMBER"]
gh_token = os.environ["GH_TOKEN"]

headers = {
    "Authorization": f"Bearer {gh_token}",
    "Accept": "application/vnd.github+json",
}

#Comment
body = "### 🤖 AI Code Review Result\n\n"
if not issues:
    body += "✅ No issues found.\n"
else:
    for i in issues:
        body += f"- **{i['severity']}** `{i['file']}:{i['line']}`\n"
        body += f"  - {i['message']}\n"
        body += f"  - Suggestion: {i['suggestion']}\n"

requests.post(
    f"https://api.github.com/repos/{repo}/issues/{pr}/comments",
    headers=headers,
    json={"body": body},
)

#Check Run（阻断 / 放行）
check = requests.post(
    f"https://api.github.com/repos/{repo}/check-runs",
    headers=headers,
    json={
        "name": "AI Code Review",
        "head_sha": os.environ["GITHUB_SHA"],
        "status": "completed",
        "conclusion": "success" if approval else "failure",
        "output": {
            "title": "AI Review",
            "summary": "Passed" if approval else "Critical issues found",
        },
    },
)

#Approve（只有通过才做）
if approval:
    requests.post(
        f"https://api.github.com/repos/{repo}/pulls/{pr}/reviews",
        headers=headers,
        json={
            "event": "APPROVE",
            "body": "🤖 AI review passed. No critical issues found.",
        },
    )
