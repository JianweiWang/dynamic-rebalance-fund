# Fund Rebalance Application - ECS Deployment Guide

## 📋 Table of Contents
1. [Prerequisites](#prerequisites)
2. [Local Development](#local-development)
3. [Building for Production](#building-for-production)
4. [ECS Server Setup](#ecs-server-setup)
5. [Deployment Process](#deployment-process)
6. [Monitoring & Maintenance](#monitoring--maintenance)
7. [Troubleshooting](#troubleshooting)

## 🔧 Prerequisites

### Local Environment
- Docker & Docker Compose
- Go 1.21+
- Git
- SSH access to ECS server

### ECS Server Requirements
- Ubuntu 20.04 LTS or CentOS 8+
- Docker & Docker Compose
- At least 2GB RAM
- At least 10GB disk space
- Open ports: 80, 443

## 🚀 Local Development

### 1. Clone and Setup
```bash
git clone <repository-url>
cd dynamic-rebalance-fund

# Create necessary directories
mkdir -p data logs config
```

### 2. Local Testing with Docker
```bash
# Build and start services
docker-compose up --build

# Access application
open http://localhost:80
```

### 3. Development Mode
```bash
# Run Go application directly
go run .

# Access application
open http://localhost:8080
```

## 📦 Building for Production

### 1. Configure Environment
```bash
# Edit production configuration
vim config/production.env

# Update the following variables:
# ECS_HOST=your-ecs-server.aliyun.com
# ECS_USER=root
# ECS_SSH_KEY=~/.ssh/your-key
# REGISTRY=registry.aliyuncs.com/your-namespace
```

### 2. Build Docker Image
```bash
# Make build script executable
chmod +x scripts/build.sh

# Build image
./scripts/build.sh latest

# Or with custom tag
./scripts/build.sh v1.0.0
```

### 3. Push to Registry (Optional)
```bash
# Login to Aliyun Container Registry
docker login registry.aliyuncs.com

# Push image
docker push registry.aliyuncs.com/your-namespace/fund-rebalance:latest
```

## 🖥️ ECS Server Setup

### 1. Install Docker
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install docker.io docker-compose

# CentOS/RHEL
sudo yum install docker docker-compose

# Start Docker service
sudo systemctl enable docker
sudo systemctl start docker
```

### 2. Create Application Directory
```bash
# Create directory structure
sudo mkdir -p /opt/fund-rebalance/{data,logs,nginx,backups}
sudo chown -R $USER:$USER /opt/fund-rebalance
```

### 3. Configure Firewall
```bash
# Ubuntu/Debian (UFW)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# CentOS/RHEL (firewalld)
sudo firewall-cmd --add-port=80/tcp --permanent
sudo firewall-cmd --add-port=443/tcp --permanent
sudo firewall-cmd --reload
```

## 🚀 Deployment Process

### Method 1: Automated Deployment
```bash
# Make deployment script executable
chmod +x scripts/deploy-ecs.sh

# Deploy to production
./scripts/deploy-ecs.sh production latest

# Deploy to staging
./scripts/deploy-ecs.sh staging v1.0.0
```

### Method 2: Manual Deployment

#### Step 1: Copy Files to Server
```bash
# Copy deployment files
scp -r docker-compose.yml nginx scripts/ config/ user@your-server:/opt/fund-rebalance/

# SSH to server
ssh user@your-server
cd /opt/fund-rebalance
```

#### Step 2: Setup Production Environment
```bash
# Copy production docker-compose
cp docker-compose.yml docker-compose.prod.yml

# Edit for production (update image tags, etc.)
vim docker-compose.prod.yml

# Make scripts executable
chmod +x scripts/*.sh
```

#### Step 3: Start Services
```bash
# Start application
./scripts/start.sh

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

### Method 3: Systemd Service
```bash
# Copy service file
sudo cp fund-rebalance.service /etc/systemd/system/

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable fund-rebalance
sudo systemctl start fund-rebalance

# Check status
sudo systemctl status fund-rebalance
```

## 📊 Monitoring & Maintenance

### Service Management
```bash
# Start application
sudo systemctl start fund-rebalance
# or
./scripts/start.sh

# Stop application
sudo systemctl stop fund-rebalance
# or
./scripts/stop.sh

# Restart application
sudo systemctl restart fund-rebalance

# View status
sudo systemctl status fund-rebalance
docker-compose -f docker-compose.prod.yml ps
```

### Log Monitoring
```bash
# Application logs
docker-compose -f docker-compose.prod.yml logs -f fund-rebalance

# Nginx logs
docker-compose -f docker-compose.prod.yml logs -f nginx

# System logs
sudo journalctl -u fund-rebalance -f
```

### Health Checks
```bash
# Check application health
curl http://localhost/health

# Check service status
curl -I http://localhost/

# Check Docker containers
docker ps
```

### Database Backup
```bash
# Manual backup
./scripts/backup.sh

# Setup automated backup (cron)
sudo crontab -e
# Add line: 0 2 * * * /opt/fund-rebalance/scripts/backup.sh
```

### Updates and Rollbacks
```bash
# Pull new image
docker-compose -f docker-compose.prod.yml pull

# Rolling update (zero downtime)
docker-compose -f docker-compose.prod.yml up -d --no-deps fund-rebalance

# Rollback to previous version
# Edit docker-compose.prod.yml with previous image tag
docker-compose -f docker-compose.prod.yml up -d --no-deps fund-rebalance
```

## 🔍 Troubleshooting

### Common Issues

#### Application Won't Start
```bash
# Check container logs
docker-compose -f docker-compose.prod.yml logs fund-rebalance

# Check disk space
df -h

# Check memory usage
free -m

# Check port conflicts
sudo netstat -tlnp | grep :80
```

#### Database Issues
```bash
# Check database file permissions
ls -la data/fund_data.db

# Backup and restore database
cp data/fund_data.db data/fund_data.db.backup

# Reset database (caution: data loss)
rm data/fund_data.db
docker-compose -f docker-compose.prod.yml restart fund-rebalance
```

#### Network Issues
```bash
# Check nginx configuration
docker-compose -f docker-compose.prod.yml exec nginx nginx -t

# Restart nginx
docker-compose -f docker-compose.prod.yml restart nginx

# Check connectivity
curl -v http://localhost/
```

#### Performance Issues
```bash
# Check resource usage
docker stats

# Check application metrics
curl http://localhost/api/portfolios | jq .

# Scale services if needed
docker-compose -f docker-compose.prod.yml up -d --scale fund-rebalance=2
```

### Log Analysis
```bash
# Application errors
docker-compose -f docker-compose.prod.yml logs fund-rebalance | grep -i error

# Nginx access logs
docker-compose -f docker-compose.prod.yml logs nginx | grep -E "4[0-9][0-9]|5[0-9][0-9]"

# System resource usage
top -p $(pgrep -f fund-rebalance)
```

## 🔒 Security Best Practices

### 1. SSL/TLS Setup
```bash
# Generate SSL certificate (Let's Encrypt)
sudo certbot certonly --standalone -d your-domain.com

# Update nginx configuration for HTTPS
# Uncomment HTTPS server block in nginx/conf.d/fund-rebalance.conf
```

### 2. Firewall Configuration
```bash
# Restrict SSH access
sudo ufw limit ssh

# Only allow HTTP/HTTPS
sudo ufw deny 8080
```

### 3. Regular Updates
```bash
# Update system packages
sudo apt update && sudo apt upgrade

# Update Docker images
docker-compose -f docker-compose.prod.yml pull
```

## 📞 Support

For issues and questions:
1. Check application logs
2. Review this documentation
3. Check GitHub issues
4. Contact system administrator

---

**Note**: Replace placeholder values (your-server, your-namespace, etc.) with actual values for your deployment.