版本：v1.0
认证方式：HTTP Bearer Token (JWT)
数据格式：Content-Type: application/json
基础路径：/api/v1

---
1. 通用响应格式 (Common Response)
接口返回统一结构：
字段	类型	说明
code	Integer	业务状态码 (0: 成功, 401: 未登录, 400: 参数错误, 409: 冲突, 500: 服务器错误)
message	String	给前端用户的提示信息
data	Object/Array	具体的业务数据内容 (可选)
---
1. 身份认证模块 (Auth)
2.1 用户注册
- URL: /auth/register
- Method: POST
- Request Body:
- JSON
{
  "username": "user_123", // 约束：[a-zA-Z0-9_]{4,20}
  "password": "Password123" // 约束：长度>=6, 包含字母+数字
}
- Response: 201 Created / 409 Conflict (用户名已存在)
2.2 用户登录
- URL: /auth/login
- Method: POST
- Request Body: { "username": "...", "password": "..." }
- Success Response (200 OK):
- JSON
{
  "code": 0,
  "data": {
    "token": "eyJhbGci...", // JWT Token
    "username": "user_123",
    "level": 0 // 0: 普通用户, 1: 管理员
  }
}

2.3 用户修改自己的密码
- URL: /auth/password
- Method: PUT
- Header: Authorization: Bearer <token>
- Request Body:
- JSON
{
  "old_password": "OldPassword123",
  "new_password": "NewPassword123" // 约束：长度>=6, 包含字母+数字
}
- Response: 200 OK / 400 Bad Request / 401 Unauthorized

---
3. 任务管理模块 (Tasks)
后端需从 Authorization Header 解析出用户信息（user_id + level）。
- 普通用户（level=0）：仅操作自己的任务。
- 管理员用户（level=1）：可查看并管理所有用户任务。
3.1 获取任务列表
- URL: /tasks
- Method: GET
- Query Params: page=1&size=10 (支持分页展示?)
- Response (200 OK):
- JSON
{
  "code": 0,
  "message": "query success",
  "data": {
    "list": [
      {
        "id": 1,
        "user_id": 1,
        "title": "任务标题",
        "description": "任务描述",
        "status": "TODO", // 枚举：TODO, DOING, DONE, ERROR
        "priority": 1,      // 1:低, 2:中, 3:高
        "deadline": "2026-04-01 12:00:00",
        "created_at": "2026-03-28 10:00:00",
        "updated_at": "2026-03-28 10:00:00"
      }
    ],
    "page": 1,
    "size": 10,
    "total": 35
  }
}
3.2 创建任务
- URL: /tasks
- Method: POST
- Request Body:
- JSON
{
  // 取决于具体任务的数据库字段
  "title": "string",
  "description": "string",
  "status": "TODO",
  "priority": 1,
  "deadline": "YYYY-MM-DD HH:mm:ss"
  ……
}
3.3 修改任务
- URL: /tasks/{id}
- Method: PUT/ PATCH
- Request Body: 同创建任务字段。
3.4 删除任务
- URL: /tasks/{id}
- Method: DELETE
- Response: 200 OK (成功) / 403 Forbidden (试图删除他人任务)

---
4. 数据库字段
字段名	类型	约束	说明
id	BigInt	Primary Key, Auto Increment	唯一标识
level	SmallInt	Default 0, Check(0/1)	用户等级：0普通用户，1管理员
title	Varchar(100)	Not Null	任务标题
description	Text	Nullable	详细描述
status	Enum	Default 'TODO'	状态：TODO, DOING, DONE
priority	TinyInt	Default 1	1:低, 2:中, 3:高
user_id	BigInt	Foreign Key	关联所属用户
deadline	DateTime	Nullable	截止日期
created_at	Timestamp	Default CURRENT_TIMESTAMP	创建时间
updated_at	Timestamp	On Update CURRENT_TIMESTAMP	自动更新时间