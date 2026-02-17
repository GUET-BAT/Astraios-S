#!/bin/bash

set -e

HARBOR_REGISTRY="harbor.g-oss.top/astraios"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

print_usage() {
    echo "Usage: $0 <service-name> [options]"
    echo ""
    echo "Services:"
    echo "  user-service      Go service"
    echo "  gateway-service   Go service"
    echo "  common-service    Go service"
    echo "  auth-service      Java service"
    echo ""
    echo "Options:"
    echo "  -t, --tag         Image tag (default: latest)"
    echo "  -p, --push        Push image after build"
    echo "  -h, --help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 user-service                    # Build user-service with 'latest' tag"
    echo "  $0 user-service -t v1.0.0 -p       # Build and push with specific tag"
    echo "  $0 auth-service -p                 # Build and push auth-service"
}

build_go_service() {
    local service=$1
    local tag=$2
    local service_dir="$PROJECT_ROOT/$service"
    
    echo "=========================================="
    echo "Building Go service: $service"
    echo "=========================================="
    
    cd "$PROJECT_ROOT"
    
    local image_tag="${HARBOR_REGISTRY}/${service}:${tag}"
    
    docker build -f "$service_dir/Dockerfile" -t "$image_tag" .
    
    echo "Built: $image_tag"
}

build_java_service() {
    local service=$1
    local tag=$2
    local service_dir="$PROJECT_ROOT/$service"
    
    echo "=========================================="
    echo "Building Java service: $service"
    echo "=========================================="
    
    cd "$PROJECT_ROOT"
    
    local image_tag="${HARBOR_REGISTRY}/${service}:${tag}"
    
    docker build -f "$service_dir/Dockerfile" -t "$image_tag" .
    
    echo "Built: $image_tag"
}

push_image() {
    local service=$1
    local tag=$2
    local image_tag="${HARBOR_REGISTRY}/${service}:${tag}"
    
    echo "=========================================="
    echo "Pushing image: $image_tag"
    echo "=========================================="
    
    docker push "$image_tag"
    
    echo "Pushed: $image_tag"
}

main() {
    if [ $# -lt 1 ]; then
        print_usage
        exit 1
    fi
    
    local service=""
    local tag="latest"
    local push=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                print_usage
                exit 0
                ;;
            -t|--tag)
                tag="$2"
                shift 2
                ;;
            -p|--push)
                push=true
                shift
                ;;
            user-service|gateway-service|common-service|auth-service)
                service="$1"
                shift
                ;;
            *)
                echo "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done
    
    if [ -z "$service" ]; then
        echo "Error: Service name is required"
        print_usage
        exit 1
    fi
    
    case $service in
        user-service|gateway-service|common-service)
            build_go_service "$service" "$tag"
            ;;
        auth-service)
            build_java_service "$service" "$tag"
            ;;
    esac
    
    if [ "$push" = true ]; then
        push_image "$service" "$tag"
    fi
    
    echo ""
    echo "=========================================="
    echo "Done!"
    echo "=========================================="
    if [ "$push" = true ]; then
        echo "Image: ${HARBOR_REGISTRY}/${service}:${tag}"
        echo ""
        echo "To deploy, run:"
        echo "  kubectl rollout restart deployment $service -n astraios"
    else
        echo "Image: ${HARBOR_REGISTRY}/${service}:${tag} (local only)"
        echo ""
        echo "To push, run:"
        echo "  $0 $service -t $tag -p"
    fi
}

main "$@"
