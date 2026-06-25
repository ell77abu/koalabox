# Lab4 API 与单元测试用例汇总

本文档用于实验报告和现场验收，汇总当前自动化 API 验收脚本和后端 Go 单元/Handler 测试，并说明运行方式、参数和输出文件。

## 新增功能识别

本轮提交中除回收站功能外，还新增了任务分配通知隐藏能力：

- 任务回收站：后端新增 `GET /api/v1/tasks/trash`，前端新增个人/团队回收站入口，用于查询当前用户可见的已删除任务。
- 分配通知隐藏：后端新增 `POST /api/v1/assignment-requests/:request_id/dismiss`，数据库新增 `task_assignment_requests.dismiss_at`，前端通知页可隐藏已处理或不想继续展示的任务分配通知。
- 独立任务日志页：前端新增 `TaskLogView.vue`，复用既有 `GET /api/v1/tasks/:id/logs`，不属于新的后端业务接口。

## API 验收测试

API 测试由两个脚本组成：

- `scripts/prepare_api_data.sh`：通过真实 HTTP API 创建隔离测试数据，输出 `scripts/lab4_api_test_data.json`。
- `scripts/run_api_tests.sh`：读取数据文件，执行 Lab1-Lab4 回归、权限、依赖和操作日志验收断言。

### API Test Cases

| 编号 | 来源 Lab | 覆盖目标 | 主要接口/动作 | 自动化断言 |
| --- | --- | --- | --- | --- |
| TC01 | Lab1 | 注册成功与重复注册拦截 | `POST /auth/register` 两次 | 首次 `201` 且 `register success`；重复 `409` |
| TC02 | Lab1 | 登录成功与错误密码拦截 | `POST /auth/login` | 正确密码 `200`；错误密码 `401` |
| TC03 | Lab1 | 个人任务 CRUD | 创建、列表、更新、删除个人任务 | 创建 `201`，列表包含任务，更新字段正确，删除 `200` |
| TC04 | Lab2 | 团队邀请后成员可见 | 成员查询 `GET /teams` | 团队列表包含准备阶段创建的团队 |
| TC05 | Lab2 | 非团队成员访问拦截 | outsider 查询团队任务 | 返回 `403 forbidden` |
| TC06 | Lab2 | 普通成员只可修改分配任务状态 | 成员更新团队任务状态和标题 | 状态更新 `200`；标题更新 `403` |
| TC07 | Lab3 | 同作用域依赖与重复依赖 | 创建个人任务依赖两次 | 首次 `201`；重复 `409` |
| TC08 | Lab3 | 跨用户依赖与循环依赖拦截 | 跨用户前置、反向依赖 | 跨用户返回 `403` 或资源隐藏式 `404`；循环依赖 `400` |
| TC09 | Lab3 | 依赖状态约束 | 未完成前置时完成后继；有完成后继时回退前驱 | 两类业务冲突均返回 `409`，响应包含相关任务 ID |
| TC10 | Lab3/Lab4 | 成员离队后的任务处理与日志 | 成员离队、Owner 查日志、离队者查日志 | 离队 `200`；Owner 可见归属变化日志；离队者 `403` |
| TC11 | Lab4 | 修改任务生成可读日志 | 更新标题、状态、优先级后查询日志 | 最新日志为 `UPDATE_TASK_INFO`，包含 actor、content、created_at，并仍包含创建日志 |
| TC12 | Lab4 | 无实际变更不写日志 | 查询日志数，提交相同标题，再查询 | 更新成功，但日志 total 不增加 |
| TC13 | Lab4 | 删除任务后日志保留 | 删除个人任务后查询日志 | Owner 仍可查询，日志包含 `DELETE_TASK` |
| TC14 | Lab4 | 日志接口权限异常 | 未登录、他人个人任务、非团队成员查日志 | 分别返回 `401/403/403` |
| TC15 | Lab4 | 任务回收站查询与权限 | 删除个人任务后查回收站；非法 `team_id`；非成员查团队回收站 | 已删除任务出现在 `/tasks/trash`；非法参数 `400`；非成员 `403` |
| TC16 | Lab4 | 任务分配通知隐藏 | 成员查询分配通知、隐藏指定通知、再次查询、outsider 尝试隐藏 | 通知先可见；隐藏返回 `dismiss_at`；列表不再展示；非本人 `403` 或 `404` |

### API 使用方法

先启动数据库和 Go 后端，确认 `BASE_URL` 指向的服务可访问，再执行：

```bash
bash scripts/prepare_api_data.sh
bash scripts/run_api_tests.sh
```

完整验收会修改任务状态、删除任务、隐藏一条分配通知，并让成员离队。重复跑完整验收前，应重新执行 `prepare_api_data.sh` 生成新的数据文件。

### `prepare_api_data.sh` 参数

该脚本通过环境变量传参，不使用命令行 flag。

| 参数 | 默认值 | 作用 | 示例 |
| --- | --- | --- | --- |
| `BASE_URL` | `http://127.0.0.1:8080/api/v1` | 后端 API 根地址 | `BASE_URL=http://server:8080/api/v1 bash scripts/prepare_api_data.sh` |
| `OUT_FILE` | `scripts/lab4_api_test_data.json` | 生成的数据文件路径 | `OUT_FILE=/tmp/lab4-data.json bash scripts/prepare_api_data.sh` |
| `PASSWORD` | `abc12345` | 测试用户统一密码 | `PASSWORD=test12345 bash scripts/prepare_api_data.sh` |
| `STAMP` | 当前时间戳 + `$RANDOM` | 生成唯一用户名、任务名、团队名 | `STAMP=demo001 bash scripts/prepare_api_data.sh` |
| `OWNER` | `l4o${STAMP}` | Owner 测试用户名 | `OWNER=lab4_owner bash scripts/prepare_api_data.sh` |
| `MEMBER` | `l4m${STAMP}` | Member 测试用户名 | `MEMBER=lab4_member bash scripts/prepare_api_data.sh` |
| `OUTSIDER` | `l4x${STAMP}` | 非团队成员测试用户名 | `OUTSIDER=lab4_outsider bash scripts/prepare_api_data.sh` |

生成的 `scripts/lab4_api_test_data.json` 只属于当前数据库和当前后端环境，包含 cookie、用户、团队和任务 ID，不应提交到 Git。

### `run_api_tests.sh` 参数

| 参数 | 默认值 | 作用 | 示例 |
| --- | --- | --- | --- |
| `DATA_FILE` | `scripts/lab4_api_test_data.json` | 指定由准备脚本生成的数据文件 | `DATA_FILE=/tmp/lab4-data.json bash scripts/run_api_tests.sh` |
| `--list` | 无 | 列出所有 API 用例后退出 | `bash scripts/run_api_tests.sh --list` |
| `--lab LAB` | 不筛选 | 只跑某个 Lab 的用例；支持 `lab1`-`lab4` 或 `1`-`4` | `bash scripts/run_api_tests.sh --lab lab3` |
| `--case CASE` | 不筛选 | 只跑指定用例组；支持 `TC08`、`tc8`、`8`，可逗号分隔 | `bash scripts/run_api_tests.sh --case TC08,TC14` |
| `-h`, `--help` | 无 | 显示帮助 | `bash scripts/run_api_tests.sh --help` |

常用命令：

```bash
# 默认跑全部 API 验收
bash scripts/run_api_tests.sh

# 只看可用用例
bash scripts/run_api_tests.sh --list

# 只跑某个 Lab 的回归
bash scripts/run_api_tests.sh --lab lab4

# 只跑单个或多个用例组
bash scripts/run_api_tests.sh --case TC08
bash scripts/run_api_tests.sh --case TC08,TC14
bash scripts/run_api_tests.sh --case TC15,TC16
```

## 后端单元与 Handler 测试

Go 测试位于 `backend/internal/service` 和 `backend/internal/handler`。其中 Service 测试验证业务规则、权限、依赖约束和日志写入；Handler 测试验证接口层参数解析、HTTP 状态码和错误映射。

### Service Unit Test Cases

| 编号 | 文件 | 测试函数 | 覆盖目标 |
| --- | --- | --- | --- |
| UT-A01 | `auth_service_test.go` | `TestAuthServiceLoginReturnsInvalidCredentialsWhenUserNotFound` | 登录时用户不存在映射为无效凭据 |
| UT-A02 | `auth_service_test.go` | `TestAuthServiceChangePasswordRejectsMissingUser` | 修改密码时用户不存在映射为无效凭据 |
| UT-T01 | `task_service_test.go` | `TestTaskServiceListRejectsNonMemberTeamView` | 非团队成员不能查看团队任务列表 |
| UT-T02 | `task_service_test.go` | `TestTaskServiceListUsesTeamIDFilterWhenProvided` | 带 `team_id` 查询时使用团队任务过滤 |
| UT-T03 | `task_service_test.go` | `TestTaskServiceListTrashReturnsPersonalDeletedTasks` | 个人回收站返回当前用户已删除个人任务 |
| UT-T04 | `task_service_test.go` | `TestTaskServiceListTrashReturnsTeamDeletedTasksForMember` | 团队成员可查询团队回收站 |
| UT-T05 | `task_service_test.go` | `TestTaskServiceListTrashRejectsNonMemberTeam` | 非团队成员查询团队回收站返回 forbidden |
| UT-T06 | `task_service_test.go` | `TestTaskServiceDismissAssignmentReturnsTimestamp` | 隐藏分配通知返回 request id 和隐藏时间 |
| UT-T07 | `task_service_test.go` | `TestTaskServiceDismissAssignmentPropagatesForbidden` | 非本人隐藏分配通知时透传 forbidden |
| UT-T08 | `task_service_test.go` | `TestTaskServiceUpdateRejectsMemberEditingNonStatusFields` | 普通成员不能修改非状态字段 |
| UT-T09 | `task_service_test.go` | `TestTaskServiceUpdateRejectsTeamIDChangeForPersonalTask` | 个人任务不能被更新到其他团队 |
| UT-T10 | `task_service_test.go` | `TestTaskServiceUpdateRejectsTeamIDChangeForTeamTask` | 团队任务不能修改所属团队 |
| UT-T11 | `task_service_test.go` | `TestTaskServiceUpdateRejectsRollbackWhenDoneSuccessorsExist` | 有已完成后继依赖时前驱不能回退 |
| UT-T12 | `task_service_test.go` | `TestTaskServiceCreateWritesActivityLog` | 创建任务写入 `CREATE_TASK` 日志 |
| UT-T13 | `task_service_test.go` | `TestTaskServiceUpdateWritesActivityLogForActualChanges` | 标题、优先级等实际变化写入可读日志 |
| UT-T14 | `task_service_test.go` | `TestTaskServiceUpdateStatusByAssignedMemberWritesActivityLog` | 被分配成员修改状态写入团队范围日志 |
| UT-T15 | `task_service_test.go` | `TestTaskServiceUpdateAssigneeWritesActivityLog` | 修改执行者写入日志，并回收旧分配请求 |
| UT-T16 | `task_service_test.go` | `TestTaskServiceUpdateValidationFailureDoesNotWriteActivityLog` | 校验失败不更新任务、不写成功日志 |
| UT-T17 | `task_service_test.go` | `TestTaskServiceUpdateSkipsActivityLogWhenUnchanged` | 无实际变化时不写日志 |
| UT-T18 | `task_service_test.go` | `TestTaskServiceDeleteWritesActivityLog` | 删除任务写入 `DELETE_TASK` 日志 |
| UT-T19 | `task_service_test.go` | `TestTaskServiceListActivityLogsReturnsPagedItems` | 日志分页查询和时间格式化 |
| UT-T20 | `task_service_test.go` | `TestTaskServiceListActivityLogsAllowsOwnerAfterTaskDeleted` | 删除后 Owner 仍可查询任务日志 |
| UT-T21 | `task_service_test.go` | `TestTaskServiceListActivityLogsAllowsTeamMember` | 团队成员可查询团队任务日志 |
| UT-T22 | `task_service_test.go` | `TestTaskServiceListActivityLogsRejectsForbiddenTeamTask` | 非团队成员不能查询团队任务日志，且不调用日志仓库 |
| UT-T23 | `task_service_test.go` | `TestTaskServiceListActivityLogsRejectsForbiddenPersonalTask` | 他人不能查询个人任务日志 |
| UT-T24 | `task_service_test.go` | `TestTaskServiceListActivityLogsReturnsNotFoundForMissingTask` | 不存在任务查询日志返回 not found |
| UT-T25 | `task_service_test.go` | `TestTaskServiceAddDependencyMapsCycleError` | 仓储层循环依赖错误映射为业务错误 |
| UT-T26 | `task_service_test.go` | `TestTaskServiceAddDependencyRejectsCrossScope` | 个人/团队跨作用域依赖被拒绝 |
| UT-M01 | `team_service_test.go` | `TestTeamServiceCreateRejectsDuplicateName` | 创建团队重名返回冲突 |
| UT-M02 | `team_service_test.go` | `TestTeamServiceUpdateRejectsDuplicateName` | 更新团队名重名返回冲突且不更新 |
| UT-M03 | `team_service_test.go` | `TestTeamServiceUpdateAllowsSameTeamName` | 团队保持原名时允许更新 |
| UT-M04 | `team_service_test.go` | `TestTeamServiceAddMemberCreatesPendingInvitationByUsername` | 通过用户名邀请成员生成 pending 邀请 |
| UT-M05 | `team_service_test.go` | `TestTeamServiceAddMemberWrapsUsernameLookupError` | 邀请不存在用户时保留错误上下文 |
| UT-M06 | `team_service_test.go` | `TestTeamServiceAddMemberRejectsNonOwner` | 非 Owner 不能邀请成员 |
| UT-M07 | `team_service_test.go` | `TestTeamServiceAcceptInvitationAddsMemberAndUpdatesStatus` | 接受邀请调用原子接受逻辑 |
| UT-M08 | `team_service_test.go` | `TestTeamServiceAcceptInvitationRejectsWrongInvitee` | 非邀请对象不能接受邀请 |
| UT-M09 | `team_service_test.go` | `TestTeamServiceRemoveMemberRejectsSelfRemoval` | Owner 不能移除自己 |
| UT-M10 | `team_service_test.go` | `TestTeamServiceRemoveMemberRejectsNonOwner` | 非 Owner 不能移除成员 |
| UT-M11 | `team_service_test.go` | `TestTeamServiceRemoveMemberRejectsRemovingOwner` | 不能移除 Owner |
| UT-M12 | `team_service_test.go` | `TestTeamServiceLeaveRejectsOwner` | Owner 不能直接离开团队 |
| UT-M13 | `team_service_test.go` | `TestTeamServiceLeaveRejectsNonMember` | 非成员离队返回 forbidden |
| UT-M14 | `team_service_test.go` | `TestTeamServiceGetTeamRoleOrForbiddenSuccess` | 团队角色查询成功路径 |

### Handler Test Cases

| 编号 | 文件 | 测试函数 | 覆盖目标 |
| --- | --- | --- | --- |
| HT-T01 | `task_handler_activity_test.go` | `TestTaskHandlerListActivityLogsSuccess` | 查询任务日志成功返回 `200/query success` |
| HT-T02 | `task_handler_activity_test.go` | `TestTaskHandlerListActivityLogsUsesDefaultPagination` | 未传分页参数时使用 `page=1,size=10` |
| HT-T03 | `task_handler_activity_test.go` | `TestTaskHandlerListActivityLogsInvalidTaskID` | 非法 task id 返回 `400 invalid task id` |
| HT-T04 | `task_handler_activity_test.go` | `TestTaskHandlerListActivityLogsUnauthorized` | 未登录查询日志返回 `401 unauthorized` |
| HT-T05 | `task_handler_activity_test.go` | `TestTaskHandlerListActivityLogsForbidden` | 权限不足映射为 `403 forbidden` |
| HT-T06 | `task_handler_activity_test.go` | `TestTaskHandlerListActivityLogsNotFound` | 任务不存在映射为 `404 task not found` |
| HT-T07 | `task_handler_activity_test.go` | `TestTaskHandlerListTrashRejectsInvalidTeamID` | 回收站非法 `team_id` 返回 `400 invalid team_id` |
| HT-T08 | `task_handler_activity_test.go` | `TestTaskHandlerListTrashPassesTeamScope` | 团队回收站查询正确透传分页和 `team_id` |
| HT-T09 | `task_handler_activity_test.go` | `TestTaskHandlerDismissAssignmentSuccess` | 隐藏分配通知成功返回 `200 assignment notification dismissed` |
| HT-T10 | `task_handler_activity_test.go` | `TestTaskHandlerDismissAssignmentForbidden` | 隐藏他人分配通知映射为 `403 forbidden` |
| HT-T11 | `task_handler_activity_test.go` | `TestTaskHandlerDismissAssignmentInvalidRequestID` | 非法 request id 返回 `400 invalid request id` |
| HT-T12 | `task_handler_activity_test.go` | `TestTaskHandlerDismissAssignmentNotFound` | 不存在分配通知映射为 `404 assignment not found` |
| HT-M01 | `team_handler_test.go` | `TestTeamHandler_CreateDuplicateName` | 创建团队重名映射为 `409 team name already exists` |
| HT-M02 | `team_handler_test.go` | `TestTeamHandler_UpdateDuplicateName` | 更新团队名重名映射为 `409 team name already exists` |
| HT-M03 | `team_handler_test.go` | `TestTeamHandler_AddMember_Success` | 邀请成员成功返回 `201 invitation sent` |
| HT-M04 | `team_handler_test.go` | `TestTeamHandler_AddMember_InvalidBody` | 邀请成员请求体非法返回 `400 invalid request body` |
| HT-M05 | `team_handler_test.go` | `TestTeamHandler_AddMember_Forbidden` | 非 Owner 邀请成员返回 `403 forbidden` |
| HT-M06 | `team_handler_test.go` | `TestTeamHandler_Leave_OwnerForbidden` | Owner 离队映射为 `403 forbidden` |
| HT-M07 | `team_handler_test.go` | `TestTeamHandler_Leave_Success` | 普通成员离队成功返回 `200` |

### Go 测试使用方法

在安装 Go 的环境中执行：

```bash
cd backend

# 运行所有后端 Go 测试
go test ./...

# 只运行 Service 单元测试
go test ./internal/service -v

# 只运行 Handler 测试
go test ./internal/handler -v
```

常用 Go test 参数：

| 参数 | 作用 | 示例 |
| --- | --- | --- |
| `./...` | 运行当前模块下所有 package 测试 | `go test ./...` |
| `./internal/service` | 只运行 service package 测试 | `go test ./internal/service -v` |
| `./internal/handler` | 只运行 handler package 测试 | `go test ./internal/handler -v` |
| `-v` | 输出每个测试函数的详细运行结果 | `go test ./internal/service -v` |
| `-run REGEX` | 只运行匹配正则的测试 | `go test ./internal/service -run TestTaskServiceAddDependency -v` |
| `-count=1` | 禁用测试缓存，强制重新执行 | `go test ./... -count=1` |
| `-cover` | 输出覆盖率摘要 | `go test ./internal/service -cover` |
| `-coverprofile FILE` | 输出覆盖率文件 | `go test ./... -coverprofile coverage.out` |
| `-json` | 输出 JSON 格式结果，便于 CI/脚本分析 | `go test ./... -json` |

报告取证命令：

```bash
cd backend
go test ./internal/service -v 2>&1 | tee ../lab4-service-test-result.txt
go test ./internal/handler -v 2>&1 | tee ../lab4-handler-test-result.txt
go test ./... -v 2>&1 | tee ../lab4-backend-test-result.txt
```
