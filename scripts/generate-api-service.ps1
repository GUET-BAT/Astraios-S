<#
.SYNOPSIS
    生成 HTTP API 服务端代码（统一管理）

.DESCRIPTION
    使用 goctl 从 .api 文件生成 HTTP API 服务端代码

.PARAMETER ServiceName
    服务名称，例如 gateway-service

.EXAMPLE
    .\scripts\generate-api-service.ps1 -ServiceName gateway-service
    .\scripts\generate-api-service.ps1 gateway-service
#>

param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$ServiceName
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Get-Item $ScriptDir).Parent.FullName
$ServiceDir = Join-Path $RootDir $ServiceName
$ApiDir = Join-Path $RootDir "proto\api\$ServiceName"

Write-Host "正在生成 HTTP API 服务端代码..."
Write-Host "服务名称: $ServiceName"
Write-Host "服务目录: $ServiceDir"
Write-Host "API 目录: $ApiDir"
Write-Host ""

if (-not (Test-Path $ApiDir)) {
    Write-Host "错误: API 目录不存在: $ApiDir" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $ServiceDir)) {
    Write-Host "错误: 服务目录不存在: $ServiceDir" -ForegroundColor Red
    Write-Host "提示: 请先创建服务目录，或使用 goctl api new 创建"
    exit 1
}

$ApiFiles = Get-ChildItem -Path $ApiDir -Filter "*.api" -Recurse | Sort-Object Name

if ($ApiFiles.Count -eq 0) {
    Write-Host "错误: 在 $ApiDir 中未找到 .api 文件" -ForegroundColor Red
    exit 1
}

$ApiFile = $ApiFiles[0].FullName

if ($ApiFiles.Count -gt 1) {
    Write-Host "发现多个 .api 文件，默认使用: $ApiFile" -ForegroundColor Yellow
}

$Arguments = @(
    "api", "go",
    "--api=$ApiFile",
    "--dir=$ServiceDir"
)

Write-Host "执行命令: goctl $Arguments"
Write-Host ""

& goctl $Arguments

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "HTTP API 服务端代码生成完成！" -ForegroundColor Green
} else {
    Write-Host "错误: goctl 执行失败" -ForegroundColor Red
    exit 1
}
