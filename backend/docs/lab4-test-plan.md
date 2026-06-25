# Lab4 测试用例与验收说明

本文档用于实验报告和现场验收，覆盖 Lab1-Lab4 回归、权限、异常和操作日志场景。自动化接口验收脚本为 `scripts/prepare_api_data.sh` 与 `scripts/run_api_tests.sh`。完整 API 与单元测试用例清单、参数说明见 `backend/docs/lab4-test-cases-and-usage.md`。

本轮新增功能识别：除任务回收站外，还新增了任务分配通知隐藏接口 `POST /api/v1/assignment-requests/:request_id/dismiss`；前端独立任务日志页复用既有日志接口，不新增后端业务 API。

## 测试用例表

| 编号 | 来源 Lab | 测试目标 | 前置条件 | 步骤 | 预期结果 | 实际结果 | 证据编号 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| TC01 | Lab1 | 用户注册与重复注册校验 | 后端服务启动 | 注册唯一用户，再用同一用户名注册 | 首次 `201`，重复 `409` | 由脚本输出填写 | E1 |
| TC02 | Lab1 | 登录成功与错误密码拦截 | 已有测试用户 | 使用正确密码登录，再使用错误密码登录 | 正确密码 `200` 且返回会话，错误密码 `401` | 由脚本输出填写 | E1 |
| TC03 | Lab1 | 个人任务 CRUD 回归 | 已登录用户 | 创建、列表查询、更新、删除个人任务 | 创建 `201`，列表可见，更新 `200`，删除 `200` | 由脚本输出填写 | E2 |
| TC04 | Lab2 | 团队邀请与成员访问 | Owner 创建团队并邀请成员 | 成员接受邀请后查询团队列表 | 成员能看到该团队 | 由脚本输出填写 | E3 |
| TC05 | Lab2 | 非团队成员访问拦截 | 存在未加入团队的用户 | 非成员查询团队任务列表 | 后端返回 `403` | 由脚本输出填写 | E4 |
| TC06 | Lab2 | 普通成员任务权限 | 成员被分配团队任务 | 成员修改任务状态，再尝试修改标题 | 状态修改成功；非状态字段修改 `403` | 由脚本输出填写 | E3 |
| TC07 | Lab3 | 同作用域任务依赖与重复依赖 | 同一用户有两个个人任务 | 创建依赖，再重复创建 | 首次 `201`，重复 `409` | 由脚本输出填写 | E5 |
| TC08 | Lab3 | 跨用户依赖与循环依赖拦截 | 已存在依赖链和其他用户任务 | 尝试跨用户依赖、反向依赖 | 跨用户被拒绝，允许 `403` 或资源隐藏式 `404`；循环依赖 `400` | 由脚本输出填写 | E5 |
| TC09 | Lab3 | 依赖状态约束 | 后继任务存在未完成前置任务；另有已完成后继链 | 尝试完成后继；尝试回退有已完成后继依赖的前驱 | 未完成前置返回 `409`；回退返回 `409` | 由脚本输出填写 | E5 |
| TC10 | Lab3/Lab4 | 成员离开后的任务处理与日志 | 成员仍在团队且有开放分配任务 | 成员离开团队，Owner 查询该任务日志，离开者再查询日志 | 任务归属变化生成日志；离开者查询 `403` | 由脚本输出填写 | E6 |
| TC11 | Lab4 | 创建/修改任务生成可读日志 | 已有个人任务 | 修改任务标题、状态、优先级后查询日志 | 日志按倒序返回，含操作人、类型、内容、时间 | 由脚本输出填写 | E7 |
| TC12 | Lab4 | 无实际变更不写日志 | 已有个人任务及创建日志 | 查询日志数量，提交相同标题，再查询数量 | 操作成功但日志总数不增加 | 由脚本输出填写 | E7 |
| TC13 | Lab4 | 删除任务后日志保留 | 已有个人任务 | 删除任务后查询该任务日志 | Owner 仍可查询，且包含 `DELETE_TASK` | 由脚本输出填写 | E8 |
| TC14 | Lab4 | 日志接口权限异常 | 已有个人任务、团队任务和非成员用户 | 未登录、非所属用户、非团队成员分别查询日志 | 分别返回 `401/403/403` | 由脚本输出填写 | E4 |
| TC15 | Lab4 | 任务回收站查询与权限 | 已有个人任务、团队和非成员用户 | 删除个人任务后查询 `/tasks/trash`；传非法 `team_id`；非成员查询团队回收站 | 删除任务出现在回收站；非法参数 `400`；非成员 `403` | 由脚本输出填写 | E9 |
| TC16 | Lab4 | 任务分配通知隐藏 | 团队 Owner 给成员分配任务并生成通知 | 成员查询通知、隐藏通知、再次查询；非本人尝试隐藏 | 通知隐藏返回 `dismiss_at`；列表不再展示；非本人 `403` 或 `404` | 由脚本输出填写 | E10 |

## 自动化验证命令

后端单元测试与接口 handler 测试：

```bash
cd backend
go test ./...
```

如需将 Service 层单元测试结果放入实验报告，可单独运行并保存输出：

```bash
cd backend
go test ./internal/service -v 2>&1 | tee ../lab4-service-test-result.txt
go test ./internal/handler -v 2>&1 | tee ../lab4-handler-test-result.txt
go test ./... -v 2>&1 | tee ../lab4-backend-test-result.txt
```

报告正文建议填写命令、覆盖范围和结果摘要，例如 AuthService、TaskService、TeamService 的业务规则、权限、日志写入、日志查询，以及 Handler 的参数解析和错误映射均通过自动化断言验证；完整 verbose 输出可作为附录或截图证据补充。

### 测试文件组织

Go 单元测试和 Handler 测试放在 `backend` 对应包内，例如 `backend/internal/service/*_test.go`、`backend/internal/handler/*_test.go`，这样可以直接复用包内类型和测试替身，也符合 Go 项目惯例。跨 Lab 的黑盒 HTTP 验收脚本放在项目根目录 `scripts/`，因为它们需要从系统外部调用真实后端接口，并且延续了 Lab3 脚本的组织方式。

Lab4 API 验收脚本：

```bash
# 需要后端服务、PostgreSQL、curl、jq 已可用
bash scripts/prepare_api_data.sh
bash scripts/run_api_tests.sh
```

默认不带参数时会执行全部用例；也可以按 Lab 或用例编号筛选：

```bash
bash scripts/run_api_tests.sh --list
bash scripts/run_api_tests.sh --lab lab3
bash scripts/run_api_tests.sh --case TC08
bash scripts/run_api_tests.sh --case TC08,TC14
bash scripts/run_api_tests.sh --case TC15,TC16
```

脚本通过标准：末尾输出 `failed=0`。如果需要重复执行完整验收，应先重新运行 `prepare_api_data.sh`，因为 `run_api_tests.sh` 会删除任务、隐藏分配通知、让成员离队并修改状态。

前端验证：

```bash
cd frontend
npm install
npm run build
npm run dev
```

人工截图建议至少保留以下 4 类证据：

- E2：个人任务 CRUD 的接口结果或前端页面。
- E5：依赖规则拦截结果，例如循环依赖或未完成前置任务。
- E7：任务详情页“操作日志”列表，能看到操作人、时间、动作内容。
- E4：无权限查询日志返回 `401` 或 `403` 的接口结果。
- E9：任务回收站 `/tasks/trash` 返回已删除任务，或非法/无权限回收站查询结果。
- E10：分配通知隐藏接口返回 `dismiss_at`，隐藏后通知列表不再包含该 request。
