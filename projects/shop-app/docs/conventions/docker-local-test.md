# Docker 本地测试规则

本文档规定 shop-app 本地开发与测试的 Docker 环境约定，含 MariaDB 启动、连接验证、服务联调。

---

## 1. 环境前提

| 项 | 要求 |
|----|------|
| Docker | 跑在 Windows 宿主机（Docker Desktop） |
| shop-app | 跑在 WSL2（Linux） |
| 网络 | WSL2 经 `127.0.0.1:3306` 访问宿主 MariaDB（Docker Desktop WSL 集成默认转发） |

> 若 `127.0.0.1` 不通，改用 Windows 宿主 IP：WSL 内执行 `cat /etc/resolv.conf | grep nameserver` 取地址，同步更新 `configs/shop-apiserver.yaml` 的 `mysql.addr`。

---

## 2. MariaDB 容器启动

### 2.1 启动命令

**PowerShell（推荐）：**

```powershell
docker run -d --name shop-mariadb `
  -p 3306:3306 `
  -e MARIADB_ROOT_PASSWORD="123456" `
  -e MARIADB_DATABASE=shop-app `
  -e MARIADB_USER=admin `
  -e MARIADB_PASSWORD="123456" `
  -v shop-mariadb-data:/var/lib/mysql `
  --restart unless-stopped `
  mariadb:11
```

**CMD（单行）：**

```cmd
docker run -d --name shop-mariadb -p 3306:3306 -e MARIADB_ROOT_PASSWORD=123456 -e MARIADB_DATABASE=shop-app -e MARIADB_USER=admin -e MARIADB_PASSWORD=123456 -v shop-mariadb-data:/var/lib/mysql --restart unless-stopped mariadb:11
```

### 2.2 容器参数约定

| 参数 | 值 | 与配置对应 |
|------|-----|-----------|
| 容器名 | `shop-mariadb` | — |
| 端口 | `3306:3306` | `mysql.addr: 127.0.0.1:3306` |
| root 密码 | `123456` | — |
| 业务用户 | `admin` | `mysql.username: admin` |
| 业务密码 | `123456` | `mysql.password: "123456"` |
| 数据库 | `shop-app` | `mysql.database: shop-app` |
| 数据卷 | `shop-mariadb-data` | 持久化，容器删除数据不丢 |
| 重启策略 | `unless-stopped` | 开机/异常自动重启 |
| 镜像 | `mariadb:11` | MariaDB 11 LTS |

> ⚠️ **配置与容器必须一致**：改账号密码时，`docker run` 参数与 `configs/shop-apiserver.yaml` 的 `mysql.*` 必须同步修改。

---

## 3. 验证 MariaDB 就绪

### 3.1 查看启动日志

```bash
docker logs shop-mariadb
```

看到 `ready for connections` 即就绪。

### 3.2 WSL 内连通性验证

```bash
# 端口连通
timeout 3 bash -c 'cat < /dev/null > /dev/tcp/127.0.0.1/3306' && echo "OK"

# 账号登录 + 库存在性
mariadb -h 127.0.0.1 -P 3306 -u admin -p'123456' -e "SELECT CURRENT_USER(), DATABASE(); SHOW DATABASES;" shop-app
```

预期输出 `admin@%` 且 `shop-app` 库存在。

---

## 4. shop-apiserver 联调

### 4.1 启动服务

```bash
cd projects/shop-app
./_output/platforms/linux/amd64/shop-apiserver -c configs/shop-apiserver.yaml
```

> 服务前台运行用于调试；后台运行见 [踩坑记录 - 坑 4](./pitfalls.md#坑-4后台进程被-shell-回收)。

### 4.2 端到端验证

```bash
# 健康检查
curl http://127.0.0.1:5555/healthz

# 注册用户（phone 必填）
curl -X POST http://127.0.0.1:5555/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@1234","nickname":"管理员","email":"admin@shop.app","phone":"13800000000"}'

# 登录获取 Token
curl -X POST http://127.0.0.1:5555/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@1234"}'
```

> ⚠️ 注册接口 `phone` 字段必填（校验来自 `validation/user.go`），漏填返回 `phone cannot be empty`。

### 4.3 Swagger 文档

```
http://localhost:5555/swagger/index.html
```

---

## 5. 容器管理常用命令

```bash
docker ps -a --filter name=shop-mariadb     # 查看状态
docker start shop-mariadb                   # 启动已停止的容器
docker stop shop-mariadb                    # 停止
docker restart shop-mariadb                 # 重启
docker logs -f shop-mariadb                 # 跟踪日志
docker exec -it shop-mariadb mariadb -u admin -p123456 shop-app   # 进容器连库
```

### 销毁与重建

```bash
# 仅删容器（保留数据）
docker rm -f shop-mariadb

# 连数据一起删（⚠️ 数据丢失，慎用）
docker rm -f shop-mariadb
docker volume rm shop-mariadb-data
```

---

## 6. 数据库初始化说明

- shop-apiserver 启动时**自动迁移**业务表（GORM AutoMigrate），无需手动建表。
- `casbin_rule` 表（权限策略）由 casbin 自动创建并建索引。
- 如需重置数据：删容器 + 数据卷后重建，重启服务即重新迁移。

---

## 7. 端口约定

| 端口 | 用途 |
|------|------|
| 3306 | MariaDB |
| 5555 | shop-apiserver HTTP |

端口冲突时：改容器 `-p` 映射 + 同步改 `configs/shop-apiserver.yaml` 的 `addr` / `mysql.addr`。
