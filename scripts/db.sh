#!/bin/bash

# Script to manage PostgreSQL database using Docker Compose

set -eo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/deployments/dockers/docker-compose.db.yml"

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_header() {
    echo -e "${BLUE}[DATABASE]${NC} $1"
}

# Check dependencies
check_dependencies() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is required but not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose is required but not installed"
        exit 1
    fi
    
    # Determine which command to use
    if command -v docker-compose &> /dev/null; then
        DOCKER_COMPOSE_CMD="docker-compose"
    else
        DOCKER_COMPOSE_CMD="docker compose"
    fi
}

# Start database
start() {
    log_header "Starting PostgreSQL database..."
    
    cd "$PROJECT_ROOT"
    
    if $DOCKER_COMPOSE_CMD -f "$DOCKER_COMPOSE_FILE" ps | grep -q "Up"; then
        log_warn "Database is already running"
        return 0
    fi
    
    # Create .env file if it doesn't exist
    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        log_info "Creating .env file from example..."
        cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"
        log_info "Please review and update .env file if needed"
    fi
    
    $DOCKER_COMPOSE_CMD -f "$DOCKER_COMPOSE_FILE" up -d postgres
    
    log_info "Waiting for database to be ready..."
    sleep 5
    
    # Wait for database to be healthy
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if $DOCKER_COMPOSE_CMD -f "$DOCKER_COMPOSE_FILE" ps postgres | grep -q "healthy"; then
            log_info "Database is healthy and ready!"
            return 0
        fi
        
        log_info "Waiting for database to be healthy... (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    log_error "Database failed to become healthy after $max_attempts attempts"
    return 1
}

# Stop database
stop() {
    log_header "Stopping PostgreSQL database..."
    
    cd "$PROJECT_ROOT"
    $DOCKER_COMPOSE_CMD -f "$DOCKER_COMPOSE_FILE" down
    
    log_info "Database stopped"
}

# Restart database
restart() {
    stop
    start
}

# Show database status
status() {
    log_header "Database Status:"
    
    cd "$PROJECT_ROOT"
    $DOCKER_COMPOSE_CMD -f "$DOCKER_COMPOSE_FILE" ps postgres
}

# Show database logs
logs() {
    log_header "Database Logs:"
    
    cd "$PROJECT_ROOT"
    $DOCKER_COMPOSE_CMD -f "$DOCKER_COMPOSE_FILE" logs -f postgres
}

# Connect to database using psql
connect() {
    log_header "Connecting to PostgreSQL database..."
    
    cd "$PROJECT_ROOT"
    
    # Load environment variables
    if [ -f "$PROJECT_ROOT/.env" ]; then
        export $(grep -v '^#' "$PROJECT_ROOT/.env" | xargs)
    fi
    
    # Default values if not set
    POSTGRES_USER=${POSTGRES_USER:-electTN}
    POSTGRES_DB=${POSTGRES_DB:-electTN}
    
    $DOCKER_COMPOSE_CMD -f "$DOCKER_COMPOSE_FILE" exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"
}

# Reset database (dangerous!)
reset() {
    log_warn "This will DELETE all data in the database!"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Database reset cancelled"
        return 0
    fi
    
    log_header "Resetting PostgreSQL database..."
    
    cd "$PROJECT_ROOT"
    
    # Stop and remove volume
    $DOCKER_COMPOSE_CMD -f "$DOCKER_COMPOSE_FILE" down -v
    
    # Start fresh
    start
    
    log_info "Database has been reset"
}

# Show help
help() {
    echo "PostgreSQL Database Management Script"
    echo ""
    echo "Usage: $0 {start|stop|restart|status|logs|connect|reset|help}"
    echo ""
    echo "Commands:"
    echo "  start    Start the PostgreSQL database"
    echo "  stop     Stop the PostgreSQL database"
    echo "  restart  Restart the PostgreSQL database"
    echo "  status   Show database status"
    echo "  logs     Show database logs"
    echo "  connect  Connect to database using psql"
    echo "  reset    Reset database (DELETE ALL DATA)"
    echo "  help     Show this help message"
    echo ""
    echo "Environment variables can be set in .env file:"
    echo "  POSTGRES_DB       Database name (default: electTN)"
    echo "  POSTGRES_USER     Database user (default: electTN)"
    echo "  POSTGRES_PASSWORD Database password (default: devpassword)"
}

# Main script logic
main() {
    check_dependencies
    
    case "${1:-help}" in
        start)
            start
            ;;
        stop)
            stop
            ;;
        restart)
            restart
            ;;
        status)
            status
            ;;
        logs)
            logs
            ;;
        connect)
            connect
            ;;
        reset)
            reset
            ;;
        help|--help|-h)
            help
            ;;
        *)
            log_error "Unknown command: $1"
            help
            exit 1
            ;;
    esac
}

main "$@"