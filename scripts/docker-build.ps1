param(
    [Parameter(Position=0, Mandatory=$false)]
    [ValidateSet("user-service", "gateway-service", "common-service", "auth-service")]
    [string]$Service,

    [Parameter(Mandatory=$false)]
    [Alias("t")]
    [string]$Tag = "latest",

    [Parameter(Mandatory=$false)]
    [Alias("p")]
    [switch]$Push,

    [Parameter(Mandatory=$false)]
    [Alias("h")]
    [switch]$Help
)

$HARBOR_REGISTRY = "harbor.g-oss.top/astraios"
$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$PROJECT_ROOT = Split-Path -Parent $SCRIPT_DIR

function Print-Usage {
    Write-Host "Usage: .\docker-build.ps1 -Service <service-name> [options]"
    Write-Host ""
    Write-Host "Services:"
    Write-Host "  user-service      Go service"
    Write-Host "  gateway-service   Go service"
    Write-Host "  common-service    Go service"
    Write-Host "  auth-service      Java service"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -t, -Tag          Image tag (default: latest)"
    Write-Host "  -p, -Push         Push image after build"
    Write-Host "  -h, -Help         Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\docker-build.ps1 -Service user-service                    # Build user-service with 'latest' tag"
    Write-Host "  .\docker-build.ps1 -Service user-service -Tag v1.0.0 -Push  # Build and push with specific tag"
    Write-Host "  .\docker-build.ps1 -Service auth-service -Push              # Build and push auth-service"
}

function Build-GoService {
    param([string]$Service, [string]$Tag)
    
    $serviceDir = Join-Path $PROJECT_ROOT $Service
    $imageTag = "${HARBOR_REGISTRY}/${Service}:${Tag}"
    
    Write-Host "=========================================="
    Write-Host "Building Go service: $Service"
    Write-Host "=========================================="
    
    Push-Location $PROJECT_ROOT
    try {
        docker build -f "$serviceDir/Dockerfile" -t $imageTag .
        Write-Host "Built: $imageTag"
    }
    finally {
        Pop-Location
    }
}

function Build-JavaService {
    param([string]$Service, [string]$Tag)
    
    $serviceDir = Join-Path $PROJECT_ROOT $Service
    $imageTag = "${HARBOR_REGISTRY}/${Service}:${Tag}"
    
    Write-Host "=========================================="
    Write-Host "Building Java service: $Service"
    Write-Host "=========================================="
    
    Push-Location $PROJECT_ROOT
    try {
        docker build -f "$serviceDir/Dockerfile" -t $imageTag .
        Write-Host "Built: $imageTag"
    }
    finally {
        Pop-Location
    }
}

function Push-Image {
    param([string]$Service, [string]$Tag)
    
    $imageTag = "${HARBOR_REGISTRY}/${Service}:${Tag}"
    
    Write-Host "=========================================="
    Write-Host "Pushing image: $imageTag"
    Write-Host "=========================================="
    
    docker push $imageTag
    Write-Host "Pushed: $imageTag"
}

if ($Help) {
    Print-Usage
    exit 0
}

if ([string]::IsNullOrEmpty($Service)) {
    Write-Host "Error: Service name is required" -ForegroundColor Red
    Print-Usage
    exit 1
}

switch ($Service) {
    { $_ -in @("user-service", "gateway-service", "common-service") } {
        Build-GoService -Service $Service -Tag $Tag
    }
    "auth-service" {
        Build-JavaService -Service $Service -Tag $Tag
    }
}

if ($Push) {
    Push-Image -Service $Service -Tag $Tag
}

Write-Host ""
Write-Host "=========================================="
Write-Host "Done!"
Write-Host "=========================================="
if ($Push) {
    Write-Host "Image: ${HARBOR_REGISTRY}/${Service}:${Tag}"
    Write-Host ""
    Write-Host "To deploy, run:"
    Write-Host "  kubectl rollout restart deployment $Service -n astraios"
} else {
    Write-Host "Image: ${HARBOR_REGISTRY}/${Service}:${Tag} (local only)"
    Write-Host ""
    Write-Host "To push, run:"
    Write-Host "  .\docker-build.ps1 -Service $Service -Tag $Tag -Push"
}
