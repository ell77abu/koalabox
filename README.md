# Koalabox 考拉盒子 多人任务管理系统
- 复旦大学 软件工程课程 CS30018.01 Lab1实验实验项目
- 本 README 说明项目的部署步骤，产品最终在linux云服务器上部署，调试或本地验收时可以选择本地linux虚拟环境。
- 建议部署坏境版本尽量使用最新稳定版，避免出现冲突

## 本地部署
### 依赖准备
#### 安装Postgreasql数据库
安装Postgresql
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo service postgresql start
```
创建数据库
```bash
# 以 postgres 管理员身份进入数据库控制台
sudo -u postgres psql

# 在控制台内执行以下 SQL：（名称密码是例子）
CREATE USER admin1 WITH PASSWORD '0123456';
CREATE DATABASE debugsql OWNER admin1;
GRANT ALL PRIVILEGES ON DATABASE debugsql TO admin1;
\c debugsql
GRANT ALL ON SCHEMA public TO admin1;
ALTER SCHEMA public OWNER TO admin1;
\q
```
创建对应表结构  
使用`backend/db/schema_postgres.sql`中语句进行创建。
#### 安装`Go`环境
```bash
sudo apt install golang-go
```
#### 安装`Node.js`
```
sudo apt install nodejs npm
```
### 后端配置
#### 修改配置文件
打开文件`backend/configs/config.yaml`
修改`database`的`dsn`为本地dsn：
```
"host=127.0.0.1 port=5432 user=admin1 password=0123456 dbname=debugsql sslmode=disable TimeZone=Asia/Shanghai"
```
#### 启动
在后端`backend`目录下，直接进行控制台指令：
```bash
go run main.go
```
看到 `Listening and serving HTTP on 0.0.0.0:8080` 就说明后端已成功启动，不要关闭这个终端窗口。

### 前端配置
#### 配置文件修改
新开一个 WSL 终端，进入前端目录（`cd /frontend`）。
打开`/.env.production`,确保有
```
VITE_API_BASE_URL=http://127.0.0.1:8080
```
这一行。  
然后在`/frontend`目录中使用控制台指令:
```
npm install   # 安装依赖
npm run dev   # 启动本地开发服务器
```
完成后，终端会提示有监听地址，浏览器打开该地址即可。

### 前端 Mock 联调
- Mock server 是 `frontend/mock-server.js`，用于前端页面联调，不等价于真实 Go 后端。
- 启动 Mock server：
  ```bash
  cd /mnt/g/code/SE/group19-base/frontend
  npm install
  npm run dev:mock
  ```
- 另开一个 WSL 终端启动前端，并把 Vite 代理指向 Mock server：
  ```bash
  cd /mnt/g/code/SE/group19-base/frontend
  VITE_DEV_API_TARGET=http://127.0.0.1:3001 npm run dev
  ```
- 打开 Vite 输出的地址，一般是 `http://localhost:5173`。

## Lab4 测试与验收
Lab4 的测试用例表和截图证据清单见 `backend/docs/lab4-test-plan.md`。自动化测试分为后端单元/API handler 测试和真实 HTTP 接口验收脚本两部分。

### 后端自动化测试
在安装 Go 的环境中执行：
```bash
cd backend
go test ./...
```
该命令覆盖业务层与 Handler 层测试，包括任务操作日志写入、日志查询权限、Lab1/Lab2/Lab3 回归场景。

### Lab4 API 验收脚本
先启动 PostgreSQL 和 Go 后端，确认 `http://127.0.0.1:8080/api/v1/ping` 可访问，然后在项目根目录执行：
```bash
bash scripts/prepare_api_data.sh
bash scripts/run_api_tests.sh
```
脚本依赖 `curl` 和 `jq`，会自动创建隔离测试用户、团队和任务，并断言注册登录、个人任务 CRUD、团队权限、依赖规则、操作日志生成、日志倒序查询、删除后日志保留和无权限日志访问。通过标准是脚本末尾输出 `failed=0`。
脚本同时覆盖任务回收站查询、回收站权限拦截，以及任务分配通知隐藏。

默认不带参数时会执行全部用例；需要定位单项问题时可以筛选：
```bash
bash scripts/run_api_tests.sh --list
bash scripts/run_api_tests.sh --lab lab3
bash scripts/run_api_tests.sh --case TC08
bash scripts/run_api_tests.sh --case TC08,TC14
bash scripts/run_api_tests.sh --case TC15,TC16
```

如果要重复跑完整验收，请先重新执行 `prepare_api_data.sh`，因为 `run_api_tests.sh` 会修改状态、删除任务、隐藏分配通知并让成员离队。完整 API 与单元测试用例表见 `backend/docs/lab4-test-cases-and-usage.md`。

### 前端构建与截图
在安装 Node.js 的环境中执行：
```bash
cd frontend
npm install
npm run build
npm run dev
```
打开任务详情页，确认“操作日志”区域能展示最近记录、暂无记录提示和错误提示。实验报告至少保留 4 条测试证据截图，可优先选择个人任务 CRUD、依赖规则拦截、日志列表成功返回、无权限查询日志返回 `401/403`。

## 服务器部署
- 示例服务器参数：Ubuntu Server 24.04 LTS 64bit
> 不需要在服务器上安装 Node.js 或 Go 的编译环境，服务器只负责纯粹的运行，节省系统资源。
> 且在服务器上，需要配置 Nginx 指向 /var/www/lab1/frontend/ 目录
### 前端
> 前端部署文件位置：/var/www/lab1/frontend/

#### 本地打包：
在自己电脑上，在 Vue 项目根目录运行 npm run build（或者 yarn build）。这会生成一个名为 dist 的文件夹。
> 第一次在本地电脑打包，先进入本地对应的项目文件“xxxx/xxxx/frontend”,进行命令npm install下载依赖包，再进行如上操作 

打包时注意配置文件.env.production的URL选项！具体参见文件注释
#### 使用FileZilla 上传：
- 打开 FileZilla，连接到腾讯云服务器。
- 右侧（服务器）进入目录：/var/www/lab1/frontend/。
- 左侧（本地电脑）进入那个打包好的 dist 文件夹内部。
- 全选 dist 里面的所有文件（包括 index.html 和 assets 等），直接拖拽到右侧。
- 如果提示覆盖，选择“覆盖”。
### 后端
> 后端部署文件位置：/opt/lab1/backend/
#### 本地交叉编译：
在自己电脑上，打开终端进入 Go 项目目录 `/backend`，执行以下命令，强制编译出一个 Linux 版本的可执行文件（假设命名为 lab1_server）：
-  Windows 电脑 (在 cmd 中依次执行)：
```bash
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o lab1_server ./cmd/server/main.go
```
- 如果是 Mac 电脑 (在终端执行)：
```bash
GOOS=linux GOARCH=amd64 go build -o lab1_server ./cmd/server/main.go
```
*编译需要一些时间*
#### FileZilla 上传：
- 在 FileZilla 右侧进入目录：/opt/lab1/backend/。
- 把刚才编译出来的那个没有后缀名的 lab1_server 文件拖拽过去。如果是第一次传，还可以把后端的配置文件（比如 config.yaml 或者 .env，里面写着连接 [database_name] 数据库的密码）一并拖过去。

### 运行

文件传好后，回到命令行（SSH 终端），把这个后端程序真正跑起来。
#### 赋予执行权限
通过 FileZilla 从 Windows 传过去的文件，默认会丢失 Linux 的可执行权限，必须手动加回来。 在服务器终端（授权用户下）执行：
```bash
cd /opt/lab1/backend/
sudo chmod +x lab1_server
```
#### 后台运行程序：
为了保证关掉终端后，程序依然能在后台默默提供接口服务，使用 nohup 命令：
```bash
nohup ./lab1_server > backend.log 2>&1 &
```
(这行命令的意思是：在后台运行程序，并且把所有的运行日志和报错信息都输出到 backend.log 文件里。)

nohup命令会使lab1_server程序一直占领8080端口，如果要运行其它版本的程序，先杀死旧版本，指令如下：
```
sudo pkill lab1_server
```
然后执行上面的nohup。 (`sudo ss -tulpn | grep 8080`可查看端口情况)
#### 检查联调结果
现在前后端都已经就位。可以执行以下命令，实时查看后端有没有报错
```Bash
tail -f /opt/lab1/backend/backend.log
```
