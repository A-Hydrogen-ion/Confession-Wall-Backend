# Confession-Wall-Backend <br>
this is a ~~4 group mates~~ 4 bammers' homework   <br>

## 项目结构： <br>

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

## 功能实现状态：
- [x] 个人资料页：可以修改昵称、上传并设置头像（avatar 字段，UploadAvatar 已实现）
- [x] 密码管理：注册/登录与密码哈希（BeforeSave 钩子与 CheckPassword 已实现）
- [x] 发布表白：支持发布、可选匿名与私有（CreateConfession）
- [x] 管理表白：支持修改/删除（UpdateConfession、相关 service 实现）
- [x] 社区查看：可以查看他人表白，匿名处理已实现（ListPublicConfessions）
- [x] 拉黑功能：支持拉黑/取消拉黑及查看拉黑列表（BlockController）
- [x] 图片存储：图片上传与存储由后端处理（UploadAvatar/UploadImages），未使用第三方图床
- [x] 评论功能：表白的评论（AddComment、ListComments、DeleteComment）
- [x] 表白带图：支持多图上传，限制最多 9 张（CreateConfession/UpdateConfession 中有校验）

注：以上为代码当前可见并实现的功能；部分安全/校验/边界场景（例如更严格的并发冲突处理、输入验证增强、测试覆盖）仍建议进一步完善。
