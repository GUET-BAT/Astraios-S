<#
.SYNOPSIS
    生成 RPC 服务端代码和客户端代码（统一管理）

.DESCRIPTION
    使用 goctl 从 proto 文件生成 gRPC 服务端代码和客户端代码

.PARAMETER ServiceName
    服务名称，例如 user-service

.EXAMPLE
    .\scripts\generate-rpc-service.ps1 -ServiceName user-service
    .\scripts\generate-rpc-service.ps1 user-service
#>

param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$ServiceName
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = (Get-Item $ScriptDir).Parent.FullName
$ServiceDir = Join-Path $RootDir $ServiceName
$ProtoDir = Join-Path $RootDir "proto\rpc\$ServiceName"

Write-Host "正在生成 RPC 服务端代码和客户端代码..."
Write-Host "服务名称: $ServiceName"
Write-Host "服务目录: $ServiceDir"
Write-Host "Proto 目录: $ProtoDir"
Write-Host ""

if (-not (Test-Path $ProtoDir)) {
    Write-Host "错误: Proto 目录不存在: $ProtoDir" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $ServiceDir)) {
    Write-Host "错误: 服务目录不存在: $ServiceDir" -ForegroundColor Red
    Write-Host "提示: 请先创建服务目录，或使用 goctl rpc new 创建"
    exit 1
}

$ProtoFiles = Get-ChildItem -Path $ProtoDir -Filter "*.proto" -Recurse

if ($ProtoFiles.Count -eq 0) {
    Write-Host "错误: 在 $ProtoDir 中未找到 proto 文件" -ForegroundColor Red
    exit 1
}

$PbDir = Join-Path $ServiceDir "pb"
if (-not (Test-Path $PbDir)) {
    New-Item -ItemType Directory -Path $PbDir -Force | Out-Null
}

$ProtoFilesList = $ProtoFiles | ForEach-Object { $_.FullName }

$GoModule = "github.com/GUET-BAT/Astraios-S/$ServiceName"

$Arguments = @(
    "rpc", "protoc",
    "--proto_path=$ProtoDir"
)

foreach ($ProtoFile in $ProtoFilesList) {
    $Arguments += $ProtoFile
}

$Arguments += @(
    "--go_out=$ServiceDir",
    "--go_opt=module=$GoModule",
    "--go-grpc_out=$ServiceDir",
    "--go-grpc_opt=module=$GoModule",
    "--zrpc_out=$ServiceDir"
)

Write-Host "执行命令: goctl $Arguments"
Write-Host ""

& goctl $Arguments

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "RPC 服务端代码生成完成！" -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host "错误: goctl 执行失败" -ForegroundColor Red
    exit 1
}
