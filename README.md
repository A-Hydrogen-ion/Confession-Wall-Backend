# Confession-Wall-Backend <br>
this is a ~~4 group mates~~ 4 bammers' homework   <br>

## 项目结构： <br>
# Confession Wall — Backend
this is a ~~4 group mates~~ 4 bammers' homework    <br>
这是一个用 Go（Gin + GORM）实现的匿名表白墙后端服务，提供用户注册/登录、发表/管理表白、评论、图片上传、拉黑等功能。

## 项目结构：

Confession-Wall-Backend/    <br>
├── main.go                 # 程序入口   <br>
├── go.mod                 # Go模块文件  <br>
├── config/                                <br>
│   └── config.yaml        # 配置文件   <br>
├── app/                   # 内部包  <br>
│   ├── model/             # 数据模型层  <br>
│   ├── service/           # 业务逻辑层  <br>
│   ├── controller/        # 控制层   <br>
│   ├── middleware/        # 中间件   <br>
│   └── utils/             # 各类工具   <br>
└── README.md                <br>

## 待完成的任务：
- [ ] 实现个人页面，能够修改昵称和设置自己上传的图片作为头像，账号、登陆密码等信息
- [ ] 可以发布一条表白，且能够选择是否匿名以及是否公开（仅自己可见）
- [ ] 能够管理自己的表白（修改、删除等）
- [ ] 实现社区功能，能看到别人发的表白（注意实名和匿名）
- [ ] 实现拉黑功能（看不到拉黑人所发的表白）
- [ ] 用户上传的图片只能在后端存储，不借助外部图床服务
- [ ] 实现表白消息的评论和回复评论的功能
- [ ] 表白带图（最高九张）
Confession-Wall-Backend/<br>
├── main.go                 # 程序入口<br>
├── go.mod                  # Go 模块文件<br>
├── config/                 # 配置与路由<br>
│   ├── config.example.yaml # 配置示例<br>
│   └── router/             # 路由配置<br>
├── app/                    # 应用代码<br>
│   ├── model/              # 数据模型<br>
│   ├── service/            # 业务逻辑<br>
│   ├── controller/         # 控制器（HTTP handler）<br>
│   ├── middleware/         # 中间件（JWT 等）<br>
│   └── jwt/                # JWT 工具<br>
└── README.md<br>

等待与前端对接中……

## 功能实现状态（当前）

- [x] 个人资料：修改昵称、上传并设置头像（`app/controller/userController.go`）
- [x] 注册/登录与密码哈希（`app/model/model.go` 的 `BeforeSave` 与 `CheckPassword`）
- [x] 发布表白：支持发布、匿名和私有（`app/controller/confessionController.go`）
- [x] 管理表白：修改、删除（`UpdateConfession` 等）
- [x] 图片上传与存储：UploadAvatar / UploadImages，存放在 `uploads/` 目录
- [x] 评论功能：添加/查看/删除评论（`app/controller/confessionController.go`）
- [x] 拉黑功能：拉黑/取消拉黑/查看拉黑列表（`app/controller/blockController.go`）
- [x] 表白带图：支持多图上传，限制最多 9 张

### 扩展的功能
- [x] 在docker环境下构建镜像运行以方便全平台部署
- [x] 成功部署到云端服务器，不依赖dokcer环境
- [ ] 使用https进行访问

> 注：以上为代码中可见并已实现的功能，仍建议进一步完善输入校验、并发冲突处理与测试覆盖。

## 扩展功能

## 本地运行

### 使用docker(推荐)

1. 安装好docker

2. 将项目文件夹`clone`到本地

```bash
git clone git@github.com:A-Hydrogen-ion/Confession-Wall-Backend.git
```
3. 转到项目文件夹根目录，执行`docker build`命令
```bash
docker build -t confession-wall:latest .
```
4. 构建完成后，将`docker-compose`文件复制到你想存储的位置，修改`docker-compose`文件
```yaml
services:
  app:
    image: confession-wall:latest
    container_name: confession-wall
    restart: unless-stopped
    ports:
      - "8080:8080"   # 宿主机 8080 映射到容器 8080
    environment:
      #JWT_SECRET: "${JWT_SECRET}"        # 可在 .env 文件或宿主机传入
      APP_DATABASE_HOST: 192.168.2.6       
      APP_DATABASE_PORT: 3306
      APP_DATABASE_USERNAME: root
      APP_DATABASE_PASSWORD: rootpassword
      APP_DATABASE_NAME: confession
    depends_on:
      - db
    volumes:
      - ./uploads:/app/uploads  # 持久化上传的图片
      - ./data:/app/data  # 配置文件和环境变量二选一

  db:
    image: mysql:latest
    container_name: confession-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: confession
    ports:
      - "3306:3306"
    volumes:
      - ./db_data:/var/lib/mysql  # 持久化数据库
```
5. 执行`docker compose up`即可
### 手动编译部署

1. 安装依赖：

```bash
cd ~/CONFESSION-WALL-BACKEND
go mod download
```
2. 执行`go build .`构建

3. 将`entrypoint.sh`和构建得到的主程序放在同一个目录下，使用chmod给予执行权限
```bash
chmod -R 755 ./path/to/yourbuild
```

4. 从原有的位置复制并编辑配置文件：
```bash
mkdir data
#保证data文件夹和你的build程序在一个目录下！
```

```bash
cp config/config.example.yaml /path/to/yourbuild/data/config.yaml
# 编辑 config\config.yaml，填写数据库连接等信息
```

5. 执行`sh ./entrypoint.sh`以拉起服务

服务默认监听 `:8080`，可在配置文件中修改。
