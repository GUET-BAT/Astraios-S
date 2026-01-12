import os
import sys
import json
import re
import requests

# ============================================================
# 初始化 & 配置
# ============================================================

def init_config():
    if len(sys.argv) < 2:
        print("❌ 请传入 diff 文件路径")
        sys.exit(0)

    required_envs = ["OPENAI_API_KEY", "REPO", "PR_NUMBER", "GH_TOKEN", "GITHUB_SHA"]
    for env in required_envs:
        if not os.environ.get(env):
            print(f"❌ 环境变量缺失: {env}")
            sys.exit(0)

    return {
        "diff_path": sys.argv[1],
        "openai_api_key": os.environ["OPENAI_API_KEY"],
        "repo": os.environ["REPO"],
        "pr_number": os.environ["PR_NUMBER"],
        "gh_token": os.environ["GH_TOKEN"],
        "github_sha": os.environ["GITHUB_SHA"]
    }

# ============================================================
# 读取 diff
# ============================================================

def read_diff_file(path):
    with open(path, "r", encoding="utf-8") as f:
        content = f.read().strip()

    if not content:
        print("ℹ️ 无 diff，跳过评审")
        sys.exit(0)

    print(f"✅ diff 读取成功，长度 {len(content)}")
    return content

# ============================================================
# AI Review（通义千问 OpenAI Compatible）
# ============================================================

def call_ai_review(config, diff_content):
    API_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"

    PROMPT = f"""
You are a senior backend engineer performing a STRICT git diff based code review.

IMPORTANT:
- Do NOT guess absolute line numbers
- Use diff hunks only
- For each issue, output hunk_index and offset

Issue severity:
- CRITICAL: bugs, crashes, security
- MAJOR: logic issues
- MINOR: style / refactor

Approval rule:
- ANY CRITICAL → approval = false

Output JSON ONLY.

JSON schema:
{{
  "approval": boolean,
  "issues": [
    {{
      "severity": "CRITICAL|MAJOR|MINOR",
      "file": "path",
      "hunk_index": number,
      "offset": number,
      "message": "problem",
      "suggestion": "fix"
    }}
  ]
}}

Git diff:
{diff_content}
"""

    payload = {
        "model": "qwen-max",
        "messages": [
            {"role": "system", "content": "You are a professional AI code reviewer."},
            {"role": "user", "content": PROMPT}
        ],
        "temperature": 0.2,
        "max_tokens": 2048
    }

    headers = {
        "Authorization": f"Bearer {config['openai_api_key']}",
        "Content-Type": "application/json"
    }

    try:
        resp = requests.post(API_URL, headers=headers, json=payload, timeout=60)
        resp.raise_for_status()
        return resp.json()["choices"][0]["message"]["content"].strip()
    except Exception as e:
        print(f"❌ AI 调用失败: {e}")
        return None

# ============================================================
# JSON 解析（容错）
# ============================================================

def parse_ai_json(content):
    if not content:
        return {"approval": False, "issues": []}

    try:
        m = re.search(r"\{[\s\S]*\}", content)
        return json.loads(m.group(0))
    except Exception:
        print("❌ AI JSON 解析失败")
        return {"approval": False, "issues": []}

# ============================================================
# Diff hunk 解析（RIGHT 行号）
# ============================================================

def parse_diff_hunks(diff):
    files = {}
    current_file = None
    hunks = []
    current_hunk = None
    right_line = None

    for line in diff.splitlines():
        if line.startswith("+++ b/"):
            current_file = line[6:].strip()
            hunks = []
            files[current_file] = hunks

        elif line.startswith("@@"):
            m = re.search(r"\+(\d+)", line)
            if not m:
                continue
            right_line = int(m.group(1))
            current_hunk = {"lines": []}
            hunks.append(current_hunk)

        elif current_file and current_hunk:
            if line.startswith("-"):
                continue
            if line.startswith("+") or line.startswith(" "):
                current_hunk["lines"].append({
                    "right_line": right_line,
                    "content": line[1:]
                })
                right_line += 1

    return files

def align_issues_to_diff(issues, diff_map):
    aligned = []

    for issue in issues:
        file = issue.get("file")
        hi = issue.get("hunk_index")
        off = issue.get("offset")

        try:
            hunk = diff_map[file][hi]
            line = hunk["lines"][off]["right_line"]
            issue["line"] = line
            aligned.append(issue)
        except Exception:
            continue

    return aligned

# ============================================================
# GitHub API
# ============================================================

def gh_headers(token):
    return {
        "Authorization": f"Bearer {token}",
        "Accept": "application/vnd.github+json"
    }

# Inline comment（MAJOR / MINOR）

def post_inline_comments(repo, pr, sha, headers, issues):
    url = f"https://api.github.com/repos/{repo}/pulls/{pr}/comments"

    for i in issues:
        if i["severity"] == "CRITICAL":
            continue

        payload = {
            "body": f"**{i['severity']}**\n{i['message']}\n\n建议: {i['suggestion']}",
            "commit_id": sha,
            "path": i["file"],
            "line": i["line"],
            "side": "RIGHT"
        }

        r = requests.post(url, headers=headers, json=payload)
        if r.status_code not in (201, 422):
            r.raise_for_status()

# PR Review（REQUEST_CHANGES / APPROVE）

def submit_review(repo, pr, headers, event, issues):
    url = f"https://api.github.com/repos/{repo}/pulls/{pr}/reviews"

    body = ""
    for i in issues:
        body += f"- **{i['severity']}** `{i['file']}:{i.get('line')}`\n  {i['message']}\n"

    payload = {"event": event, "body": body or "AI Review result"}

    r = requests.post(url, headers=headers, json=payload)
    r.raise_for_status()

# Check Run

def create_check_run(repo, sha, headers, success):
    url = f"https://api.github.com/repos/{repo}/check-runs"

    payload = {
        "name": "AI Code Review",
        "head_sha": sha,
        "status": "completed",
        "conclusion": "success" if success else "failure"
    }

    r = requests.post(url, headers=headers, json=payload)
    r.raise_for_status()


def submit_review_with_inline_comments(
    repo,
    pr,
    sha,
    headers,
    issues,
):
    url = f"https://api.github.com/repos/{repo}/pulls/{pr}/reviews"

    has_critical = any(i["severity"] == "CRITICAL" for i in issues)

    event = "REQUEST_CHANGES" if has_critical else "APPROVE"

    review_body_lines = []
    for i in issues:
        review_body_lines.append(
            f"- **{i['severity']}** `{i['file']}:{i.get('line')}`\n  {i['message']}"
        )

    review_body = "\n".join(review_body_lines) or "AI automated code review passed."

    comments = []
    for i in issues:
        if "line" not in i:
            continue

        comments.append({
            "path": i["file"],
            "line": i["line"],
            "side": "RIGHT",
            "body": (
                f"**{i['severity']}**\n"
                f"{i['message']}\n\n"
                f"建议：{i['suggestion']}"
            )
        })

    payload = {
        "event": event,
        "body": review_body,
        "comments": comments
    }

    r = requests.post(url, headers=headers, json=payload)

    if r.status_code == 422:
        print("❌ GitHub Review 422")
        print(r.text)
        r.raise_for_status()

    r.raise_for_status()

# ============================================================
# main
# ============================================================

def main():
    config = init_config()
    diff = read_diff_file(config["diff_path"])

    ai_raw = call_ai_review(config, diff)
    ai = parse_ai_json(ai_raw)

    diff_map = parse_diff_hunks(diff)
    issues = align_issues_to_diff(ai.get("issues", []), diff_map)

    headers = gh_headers(config["gh_token"])

    # post_inline_comments(
    #     config["repo"],
    #     config["pr_number"],
    #     config["github_sha"],
    #     headers,
    #     issues
    # )

    # critical = [i for i in issues if i["severity"] == "CRITICAL"]

    # if critical:
    #     submit_review(
    #         config["repo"],
    #         config["pr_number"],
    #         headers,
    #         "REQUEST_CHANGES",
    #         critical
    #     )
    # else:
    #     submit_review(
    #         config["repo"],
    #         config["pr_number"],
    #         headers,
    #         "APPROVE",
    #         []
    #     )

    submit_review_with_inline_comments(
    repo=config["repo"],
    pr=config["pr_number"],
    sha=config["github_sha"],
    headers=headers,
    issues=issues
    )

    # create_check_run(
    #     config["repo"],
    #     config["github_sha"],
    #     headers,
    #     success=not bool(critical)
    # )
    create_check_run(
    config["repo"],
    config["github_sha"],
    headers,
    success=not any(i["severity"] == "CRITICAL" for i in issues)
    )

if __name__ == "__main__":
    main()
