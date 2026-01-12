import json
import os
import sys
import re
import requests
# from openai import OpenAI
#from openai import APIError, AuthenticationError, RateLimitError

# -------------------------- 初始化配置 & 入参读取 --------------------------
def init_config():
    # 校验入参：必须传入diff文件路径
    if len(sys.argv) < 2:
        print("❌ 错误: 请传入diff文件路径作为参数")
        sys.exit(0)
    diff_path = sys.argv[1]
    
    # 校验核心环境变量是否存在，缺失则直接退出
    required_envs = ["OPENAI_API_KEY", "REPO", "PR_NUMBER", "GH_TOKEN", "GITHUB_SHA"]
    for env in required_envs:
        if not os.environ.get(env):
            print(f"❌ 错误: 环境变量 {env} 未配置")
            sys.exit(0)

    # 初始化OpenAI客户端
    # client = OpenAI(api_key=os.environ["OPENAI_API_KEY"])
    return {
        "diff_path": diff_path,
        #"client": client,
        "openai_api_key": os.environ["OPENAI_API_KEY"],
        "repo": os.environ["REPO"],
        "pr_number": os.environ["PR_NUMBER"],
        "gh_token": os.environ["GH_TOKEN"],
        "github_sha": os.environ["GITHUB_SHA"]
    }

# -------------------------- 读取diff文件（修复编码问题） --------------------------
def read_diff_file(diff_path):
    try:
        # ✅ 修复BUG1：指定UTF-8编码，兼容中文/特殊字符
        with open(diff_path, 'r', encoding='utf-8') as f:
            diff_content = f.read().strip()
        
        # ✅ 修复BUG4：判断diff为空，直接退出，无需评审
        if not diff_content:
            print("ℹ️ 本次PR无代码变更，跳过AI评审")
            sys.exit(0)
        
        print(f"✅ 成功读取diff文件，内容长度: {len(diff_content)} 字符")
        return diff_content
    except FileNotFoundError:
        print(f"❌ 错误: diff文件 {diff_path} 不存在")
        sys.exit(0)
    except Exception as e:
        print(f"❌ 读取diff文件失败: {str(e)}")
        sys.exit(0)

# -------------------------- 调用OpenAI AI评审核心逻辑 --------------------------
def call_ai_review(config, diff_content):

    OPENAI_API_KEY = config["openai_api_key"]
    API_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
    # 评审提示词
    PROMPT = f"""
You are a senior Golang engineer and java engineer, expert in go-zero framework and java springboot framework, performing code review for go-zero backend projects and java springboot backend projects.
Focus on go-zero best practices: rpc/api layer separation, viper config usage, gorm sql security, error handling, concurrency safety.

Review the following git diff.
Classify issues into:
- CRITICAL: bugs, crashes, security, data loss
- MAJOR: logic issues, race conditions
- MINOR: style, refactor suggestions

Rules:
- If any CRITICAL exists -> approval = false
- Otherwise -> approval = true

Output JSON ONLY, NO OTHER TEXT, NO EXPLANATION:
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
{diff_content}
"""
    # try:
    #     print("ℹ️ 开始调用GPT-4.1-mini进行AI代码评审...")
    #     resp = client.chat.completions.create(
    #         model="gpt-4.1-mini",
    #         messages=[{"role": "user", "content": PROMPT}],
    #         temperature=0.2,  # 更低的随机性，评审更严谨，必加
    #         timeout=60         # 设置超时时间，避免卡住
    #     )
    #     ai_content = resp.choices[0].message.content.strip()
    #     return ai_content
    # except AuthenticationError:
    #     print("❌ OpenAI认证失败: API-KEY无效，请检查配置")
    #     return None
    # except RateLimitError:
    #     print("❌ OpenAI调用超限: API额度不足，请充值或更换KEY")
    #     return None
    # except APIError as e:
    #     print(f"❌ OpenAI接口错误: {str(e)}")
    #     return None
    # except Exception as e:
    #     print(f"❌ AI评审调用失败: {str(e)}")
    #     return None

        # 通义千问请求体，固定格式，可改model字段切换模型
    payload = {
        "model": "qwen-max",  # ✅ 可替换为 qwen-plus/qwen-turbo/qwen2-7b-instruct
        "input": {
            "messages": [
                {"role": "user", "content": PROMPT}
            ]
        },
        "parameters": {
            "result_format": "text",  # 返回文本格式
            "temperature": 0.2,       # 评审严谨度，和你原配置一致
            "top_p": 0.9,
            "max_tokens": 2048        # 足够容纳评审结果+JSON
        }
    }
    # 请求头，固定格式
    headers = {
        "Authorization": f"Bearer {OPENAI_API_KEY}",
        "Content-Type": "application/json"
    }

    try:
        print("开始调用【通义千问 qwen-max】进行AI代码评审...")
        resp = requests.post(API_URL, headers=headers, json=payload, timeout=60)
        resp.raise_for_status()
        resp_json = resp.json()
        
        # 解析通义千问返回结果
        if resp_json.get("output", {}).get("text"):
            ai_content = resp_json["output"]["text"].strip()
            print("✅ 通义千问调用成功，获取评审结果")
            return ai_content
        else:
            print(f"❌ 通义千问返回异常: {resp_json}")
            return None
    except requests.exceptions.HTTPError as e:
        if resp.status_code == 401:
            print("❌ 通义千问认证失败: API-KEY无效，请检查Secrets配置")
        elif resp.status_code == 429:
            print("❌ 通义千问调用超限: API额度不足或频率过高")
        else:
            print(f"❌ 通义千问接口错误: {str(e)}")
        return None
    except Exception as e:
        print(f"❌ AI评审调用失败: {str(e)}")
        return None



# -------------------------- 解析AI返回的JSON（修复核心BUG：JSON解析容错） --------------------------
def parse_ai_json(ai_content):
    if not ai_content:
        return {"approval": False, "issues": [{"severity": "CRITICAL", "file": "system", "line": 0, "message": "AI评审调用失败", "suggestion": "请检查OpenAI配置或稍后重试"}]}
    
    try:
        # ✅ 修复BUG2：最强JSON容错处理，移除首尾所有非JSON字符、```标记、空格换行
        # 正则匹配JSON大括号首尾，只提取中间的纯净JSON内容，解决99%的解析失败问题
        json_match = re.search(r'\{[\s\S]*\}', ai_content)
        if not json_match:
            raise ValueError("未匹配到有效的JSON内容")
        
        pure_json = json_match.group(0)
        result = json.loads(pure_json)
        
        # 校验JSON结构是否合规
        if "approval" not in result or "issues" not in result:
            raise ValueError("AI返回的JSON缺少必要字段")
        
        print(f"✅ AI评审完成: 发现 {len(result['issues'])} 个问题, Approval = {result['approval']}")
        return result
    except Exception as e:
        print(f"❌ JSON解析失败: {str(e)} | AI原始返回: {ai_content[:200]}")
        # 解析失败时，返回兜底结果：阻断合并+提示错误
        return {"approval": False, "issues": [{"severity": "CRITICAL", "file": "system", "line": 0, "message": "AI评审结果解析失败", "suggestion": "请查看Action日志，确认AI返回格式"}]}

# -------------------------- 构建GitHub请求头 --------------------------
def get_github_headers(gh_token):
    return {
        "Authorization": f"Bearer {gh_token}",
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28" # 指定API版本，避免兼容性问题
    }

# -------------------------- 向PR发布格式化评审评论 --------------------------
def post_pr_comment(repo, pr_number, headers, issues):
    try:
        body = "### 🤖 AI Code Review Result (GPT-4.1-mini)\n\n"
        if not issues:
            body += "✅ **No issues found. Code is clean!** ✅\n\n"
        else:
            # 按严重程度排序：CRITICAL > MAJOR > MINOR
            issues_sorted = sorted(issues, key=lambda x: {"CRITICAL":0, "MAJOR":1, "MINOR":2}[x["severity"]])
            for idx, issue in enumerate(issues_sorted, 1):
                severity_emoji = {"CRITICAL": "❌", "MAJOR": "⚠️", "MINOR": "ℹ️"}[issue["severity"]]
                body += f"{idx}. **{severity_emoji} {issue['severity']}** `{issue['file']}:{issue['line']}`\n"
                body += f"   ➤ 问题: {issue['message']}\n"
                body += f"   ➤ 建议: {issue['suggestion']}\n\n"
        
        resp = requests.post(
            f"https://api.github.com/repos/{repo}/issues/{pr_number}/comments",
            headers=headers,
            json={"body": body},
            timeout=30
        )
        resp.raise_for_status() # 抛出HTTP错误
        print("✅ PR评审评论发布成功")
    except Exception as e:
        print(f"⚠️ PR评论发布失败: {str(e)}")

# -------------------------- 创建GitHub Check Run（核心：阻断/放行PR合并） --------------------------
def create_check_run(repo, github_sha, headers, approval, issues):
    try:
        conclusion = "success" if approval else "failure"
        title = "✅ AI Review Passed" if approval else "❌ AI Review Failed (Critical Issues)"
        critical_count = len([i for i in issues if i["severity"] == "CRITICAL"])
        major_count = len([i for i in issues if i["severity"] == "MAJOR"])
        minor_count = len([i for i in issues if i["severity"] == "MINOR"])
        
        summary = f"""
Critical: {critical_count} | Major: {major_count} | Minor: {minor_count}
{'✅ No critical issues, safe to merge.' if approval else '❌ Critical issues detected, merge blocked!'}
        """.strip()

        resp = requests.post(
            f"https://api.github.com/repos/{repo}/check-runs",
            headers=headers,
            json={
                "name": "AI Code Review",
                "head_sha": github_sha,
                "status": "completed",
                "conclusion": conclusion,
                "output": {
                    "title": title,
                    "summary": summary,
                },
            },
            timeout=30
        )
        resp.raise_for_status()
        print(f"✅ Check Run创建成功, 结果: {conclusion}")
    except Exception as e:
        print(f"⚠️ Check Run创建失败: {str(e)}")

# -------------------------- 自动审批PR（评审通过时） --------------------------
def approve_pr(repo, pr_number, headers):
    try:
        resp = requests.post(
            f"https://api.github.com/repos/{repo}/pulls/{pr_number}/reviews",
            headers=headers,
            json={
                "event": "APPROVE",
                "body": "🤖 AI Code Review Approved: No critical issues found, code quality is acceptable."
            },
            timeout=30
        )
        resp.raise_for_status()
        print("✅ PR自动审批成功 (APPROVE)")
    except Exception as e:
        print(f"⚠️ PR自动审批失败: {str(e)}")

# -------------------------- 主函数入口 --------------------------
def main():
    # 初始化配置
    config = init_config()
    # 读取diff文件
    diff_content = read_diff_file(config["diff_path"])
    # 调用AI评审函数
    ai_content = call_ai_review(config, diff_content)
    # 解析AI返回的JSON
    ai_result = parse_ai_json(ai_content)
    approval = ai_result["approval"]
    issues = ai_result["issues"]
    # 获取GitHub请求头
    headers = get_github_headers(config["gh_token"])
    
    # 执行三大核心动作
    post_pr_comment(config["repo"], config["pr_number"], headers, issues)
    create_check_run(config["repo"], config["github_sha"], headers, approval, issues)
    if approval:
        approve_pr(config["repo"], config["pr_number"], headers)

if __name__ == "__main__":
    main()