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

## 待完成的任务：
- [ ] 实现个人页面，能够修改昵称和设置自己上传的图片作为头像，账号、登陆密码等信息
- [ ] 可以发布一条表白，且能够选择是否匿名以及是否公开（仅自己可见）
- [ ] 能够管理自己的表白（修改、删除等）
- [ ] 实现社区功能，能看到别人发的表白（注意实名和匿名）
- [ ] 实现拉黑功能（看不到拉黑人所发的表白）
- [ ] 用户上传的图片只能在后端存储，不借助外部图床服务
- [ ] 实现表白消息的评论和回复评论的功能
- [ ] 表白带图（最高九张）
