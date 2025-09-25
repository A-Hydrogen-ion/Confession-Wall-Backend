# Confession Wall — Backend
this is a ~~4 group mates~~ 4 bammers' homework    <br>
这是一个用 Go（Gin + GORM）实现的匿名表白墙后端服务，提供用户注册/登录、发表/管理表白、评论、图片上传、拉黑等功能。

## 项目结构：

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

## 功能实现状态（当前）

- [x] 个人资料：修改昵称、上传并设置头像（`app/controller/userController.go`）
- [x] 注册/登录与密码哈希（`app/model/model.go` 的 `BeforeSave` 与 `CheckPassword`）
- [x] 发布表白：支持发布、匿名和私有（`app/controller/confessionController.go`）
- [x] 管理表白：修改、删除（`UpdateConfession` 等）
- [x] 图片上传与存储：UploadAvatar / UploadImages，存放在 `uploads/` 目录
- [x] 评论功能：添加/查看/删除评论（`app/controller/confessionController.go`）
- [x] 拉黑功能：拉黑/取消拉黑/查看拉黑列表（`app/controller/blockController.go`）
- [x] 表白带图：支持多图上传，限制最多 9 张

> 注：以上为代码中可见并已实现的功能，仍建议进一步完善输入校验、并发冲突处理与测试覆盖。

## 本地运行

1. 复制并编辑配置文件：

```powershell
copy config\config.example.yaml config\config.yaml
# 编辑 config\config.yaml，填写数据库连接等信息
```

2. 安装依赖并运行：

```powershell
cd d:/wall2025
go mod tidy
go run main.go
```

服务默认监听 `:8080`，可在配置文件中修改。
