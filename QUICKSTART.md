# Quick Start - ECS Deployment

## 🚀 快速部署指南

### 1. 准备工作
```bash
# 克隆项目
git clone <your-repo>
cd dynamic-rebalance-fund

# 配置生产环境
cp config/production.env.example config/production.env
# 编辑 config/production.env，填入你的 ECS 服务器信息
```

### 2. 本地构建测试
```bash
# 本地测试
docker-compose up --build

# 访问 http://localhost:80 确认应用正常运行
```

### 3. 构建生产镜像
```bash
# 构建并推送到阿里云容器镜像仓库
./scripts/build.sh latest

# 登录阿里云容器镜像服务
docker login registry.aliyuncs.com --username=your-username

# 推送镜像
docker push registry.aliyuncs.com/your-namespace/fund-rebalance:latest
```

### 4. ECS 服务器准备
```bash
# SSH 到 ECS 服务器
ssh root@your-ecs-server

# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 创建应用目录
mkdir -p /opt/fund-rebalance
```

### 5. 部署到 ECS
```bash
# 方法1：自动化部署（推荐）
./scripts/deploy-ecs.sh production latest

# 方法2：手动部署
scp -r . root@your-ecs-server:/opt/fund-rebalance/
ssh root@your-ecs-server
cd /opt/fund-rebalance
./scripts/start.sh
```

### 6. 验证部署
```bash
# 检查服务状态
curl http://your-ecs-server/health

# 访问应用
open http://your-ecs-server
```

### 7. 设置系统服务（可选）
```bash
# 在 ECS 服务器上
sudo cp fund-rebalance.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable fund-rebalance
sudo systemctl start fund-rebalance
```

## 📋 配置检查清单

- [ ] 更新 `config/production.env` 中的 ECS 服务器信息
- [ ] 配置阿里云容器镜像仓库
- [ ] 确保 ECS 服务器安装了 Docker
- [ ] 开放 ECS 安全组端口 80 和 443
- [ ] 配置域名解析（如需要）
- [ ] 设置 SSL 证书（生产环境推荐）

## 🔧 常用命令

```bash
# 查看日志
docker-compose -f docker-compose.prod.yml logs -f

# 重启服务
./scripts/stop.sh && ./scripts/start.sh

# 备份数据
./scripts/backup.sh

# 更新应用
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

## 🆘 问题排查

如果遇到问题，请按以下顺序检查：

1. **服务状态**: `docker-compose ps`
2. **应用日志**: `docker-compose logs fund-rebalance`
3. **网络连接**: `curl http://localhost/health`
4. **磁盘空间**: `df -h`
5. **内存使用**: `free -m`

更详细的部署说明请参考 [DEPLOYMENT.md](./DEPLOYMENT.md)。