# Docker Container Debugging Guide

## Overview
The elevator application now runs in a lightweight Alpine Linux container that includes essential debugging tools for pod investigation and troubleshooting.

## Container Details
- **Base Image**: Alpine Linux 3.21 (~15MB with tools)
- **Shell**: `/bin/ash` (compatible with bash scripts)
- **User**: Non-root user `appuser` (UID 1001)
- **Security**: No-new-privileges, non-root execution

## Available Debugging Tools

### Network Tools
- `curl` - HTTP client for API testing
- `wget` - File download and HTTP testing
- `netcat` (`nc`) - Network connectivity testing

### System Monitoring
- `ps` - Process monitoring
- `htop` - Interactive process viewer
- `top` - System monitoring

### File System
- `tree` - Directory structure visualization
- `ls`, `cat`, `head`, `tail` - File operations

### JSON Processing
- `jq` - JSON parser and formatter

### General Utilities
- `ash` shell with standard utilities
- `grep`, `awk`, `sed` - Text processing
- `find` - File searching

## Quick Debugging Commands

### 1. Shell Access
```bash
# Access container shell
docker exec -it elevator-go /bin/ash

# Or via docker-compose
docker-compose exec elevator-app /bin/ash
```

### 2. Process Monitoring
```bash
# Check running processes
docker exec elevator-go ps aux

# Interactive process monitoring
docker exec -it elevator-go htop
```

### 3. Network Debugging
```bash
# Check listening ports
docker exec elevator-go netstat -tln

# Test connectivity
docker exec elevator-go nc -z localhost 6660

# Check health endpoint
docker exec elevator-go curl -f http://localhost:6660/v1/health/live
```

### 4. API Testing
```bash
# Get health status with JSON formatting
docker exec elevator-go curl -s http://localhost:6660/v1/health/live | jq .

# Test detailed health
docker exec elevator-go curl -s http://localhost:6660/v1/health/detailed | jq .

# Get API info
docker exec elevator-go curl -s http://localhost:6660/v1 | jq .
```

### 5. Log Investigation
```bash
# View application logs
docker logs elevator-go

# Follow logs in real-time
docker logs -f elevator-go

# View last 50 lines
docker logs --tail 50 elevator-go
```

### 6. File System Investigation
```bash
# View directory structure
docker exec elevator-go tree /app

# Check file permissions
docker exec elevator-go ls -la /app

# View configuration
docker exec elevator-go cat /app/configs/production.env
```

### 7. Resource Monitoring
```bash
# Check memory usage
docker exec elevator-go cat /proc/meminfo

# Check CPU info
docker exec elevator-go cat /proc/cpuinfo

# Check disk usage
docker exec elevator-go df -h
```

## Health Checks

The container includes built-in health checks:

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:6660/v1/health/live"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

Check health status:
```bash
docker inspect elevator-go --format='{{.State.Health.Status}}'
```

## Kubernetes Debugging

When deployed in Kubernetes, you can use the same commands:

```bash
# Shell access in Kubernetes
kubectl exec -it <pod-name> -- /bin/ash

# Port forward for local testing
kubectl port-forward <pod-name> 6660:6660

# Check health in cluster
kubectl exec <pod-name> -- curl -f http://localhost:6660/v1/health/live
```

## Security Features

Despite having debugging tools, the container maintains security:
- Non-root user execution
- No-new-privileges security option
- Minimal attack surface with Alpine base
- Read-only filesystem where possible
- Proper file permissions and ownership

## Example Debugging Session

```bash
# 1. Access the container
docker exec -it elevator-go /bin/ash

# 2. Check the application is running
ps aux | grep elevator

# 3. Test the API
curl -s http://localhost:6660/v1/health/live | jq .

# 4. Check network connectivity
netstat -tln
nc -z localhost 6660 && echo "Port is open"

# 5. Monitor resources
htop

# 6. View logs (from another terminal)
docker logs -f elevator-go
```

## Image Size Comparison

- **Distroless**: ~20MB (no shell, no debugging tools)
- **Alpine with tools**: ~25MB (includes shell + debugging tools)
- **Debian slim**: ~80MB+
- **Ubuntu**: ~100MB+

The Alpine approach provides the best balance of size, security, and debugging capabilities. 