# Elevator Control System - Docker Setup

This guide shows you how to run the Elevator Control System using Docker Compose with a web interface to monitor and control elevators.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose installed
- Ports 6660 and 8080 available on your system

### Running the System

1. **Start the services:**
   ```bash
   docker-compose up -d
   ```

2. **Access the web interface:**
   Open your browser and go to: **http://localhost:8080**

3. **View logs (optional):**
   ```bash
   docker-compose logs -f elevator-service
   ```

4. **Stop the services:**
   ```bash
   docker-compose down
   ```

## 📊 What's Running

### Services Overview
- **Elevator Service**: Go application running on port 6660
- **Web Client**: Nginx serving the HTML interface on port 8080

### Service Architecture
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Web Browser   │───▶│  Nginx (8080)    │───▶│ Elevator Service│
│  localhost:8080 │    │  Proxy + Static  │    │   (Go - 6660)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │   Static Files   │
                       │client/client-docker.html│
                       └──────────────────┘
```

## 🛠️ Features Available

### 1. **API Operations**
- View API information and endpoints
- Test all available REST endpoints
- Monitor response times and status codes

### 2. **Elevator Management**
- Create new elevators with custom configurations
- Set floor ranges (min/max floors)
- Name elevators for easy identification

### 3. **Floor Requests**
- Request elevators from any floor to any floor
- Support for both v1 API and legacy endpoints
- Real-time request tracking

### 4. **Real-time Monitoring**
- WebSocket connection for live status updates
- Monitor elevator positions and states
- View system metrics and health status

### 5. **Health & Metrics**
- Health check endpoints
- System performance metrics
- Prometheus metrics endpoint

## 🔗 Available Endpoints

### Web Interface
- **Main Interface**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### Direct API Access (through nginx proxy)
- **API Info**: http://localhost:8080/v1
- **Health**: http://localhost:8080/v1/health
- **Metrics**: http://localhost:8080/v1/metrics
- **Request Elevator**: POST http://localhost:8080/v1/floors/request
- **Create Elevator**: POST http://localhost:8080/v1/elevators
- **WebSocket**: ws://localhost:8080/ws/status

### Direct Backend Access (bypass proxy)
- **Backend Direct**: http://localhost:6660
- **Backend Health**: http://localhost:6660/v1/health

## 📝 Usage Examples

### 1. **Request an Elevator**
In the web interface:
- Go to "Elevator Operations" tab
- Set "From Floor": 1
- Set "To Floor": 10
- Click "Request Elevator"

### 2. **Create a New Elevator**
In the web interface:
- Go to "Management" tab
- Set "Elevator Name": "My-Elevator"
- Set "Minimum Floor": 0
- Set "Maximum Floor": 20
- Click "Create Elevator"

### 3. **Monitor Real-time Status**
In the web interface:
- Go to "Real-time Status" tab
- Click "Connect WebSocket"
- Watch live updates as elevators move

### 4. **Using curl commands**
```bash
# Get API info
curl http://localhost:8080/v1

# Request elevator
curl -X POST http://localhost:8080/v1/floors/request \
  -H "Content-Type: application/json" \
  -d '{"from": 1, "to": 5}'

# Create elevator
curl -X POST http://localhost:8080/v1/elevators \
  -H "Content-Type: application/json" \
  -d '{"name": "Test-Elevator", "min_floor": 0, "max_floor": 10}'

# Check health
curl http://localhost:8080/v1/health
```

## 🔧 Configuration

### Environment Variables
The elevator service uses these key environment variables:
- `ENV=development` - Running in development mode
- `LOG_LEVEL=DEBUG` - Detailed logging
- `DEFAULT_ELEVATOR_COUNT=3` - Starts with 3 elevators
- `DEFAULT_MAX_FLOOR=15` - Default maximum floor
- `WEBSOCKET_ENABLED=true` - Real-time updates enabled

### Default Elevator Configuration
- **Floor Range**: 0 to 15
- **Initial Elevators**: 3 (named "Elevator-1", "Elevator-2", "Elevator-3")
- **Floor Travel Time**: 2 seconds per floor
- **Door Open Time**: 3 seconds
- **Status Updates**: Every 1 second

## 🐛 Troubleshooting

### Services won't start
```bash
# Check if ports are already in use
netstat -tulpn | grep -E "(6660|8080)"

# View service logs
docker-compose logs elevator-service
docker-compose logs web-client
```

### Can't connect to WebSocket
- Ensure the elevator service is running: `curl http://localhost:8080/v1/health`
- Check browser console for WebSocket errors
- Verify nginx proxy is working: `curl http://localhost:8080/ws/status`

### API requests failing
- Check if services are healthy: `docker-compose ps`
- Test direct backend: `curl http://localhost:6660/v1/health`
- Check nginx logs: `docker-compose logs web-client`

### Performance Issues
- Monitor resource usage: `docker stats`
- Check elevator service logs for errors
- Reduce update frequency if needed

## 🔄 Development

### Rebuild after code changes
```bash
# Rebuild and restart services
docker-compose up -d --build

# Or rebuild just the elevator service
docker-compose build elevator-service
docker-compose up -d elevator-service
```

### View detailed logs
```bash
# Follow all logs
docker-compose logs -f

# Follow specific service
docker-compose logs -f elevator-service
docker-compose logs -f web-client
```

## 📈 Monitoring Tips

1. **Use the Real-time Status tab** to see elevators moving in real-time
2. **Monitor the Metrics endpoint** for performance data
3. **Check Health status** regularly to ensure system stability
4. **Watch the logs** for detailed operation information

## 🎯 Next Steps

- Explore the different API versions (v1 vs legacy)
- Create multiple elevators and test concurrent requests
- Monitor system performance under load
- Experiment with different floor configurations
- Test error scenarios (invalid floors, etc.)

Enjoy exploring your Elevator Control System! 🚀 