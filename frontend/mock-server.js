import express from 'express';
import cors from 'cors';
import jwt from 'jsonwebtoken';

const app = express();
const PORT = 3001;
const JWT_SECRET = 'lab1_mock_secret_key';

// ================== 内存数据库重构 (支持 Lab2) ==================
const users = new Map(); 
// 将原先的 Map 改为扁平数组，以完美支持团队跨用户共享任务
const tasks = []; 
const teams = new Map(); // team_id -> { team_id, name, owner_id }
const teamMembers = [];  // { team_id, user_id, role }

let taskIdCounter = 1;
let teamIdCounter = 1;

app.use(cors());
app.use(express.json());

/**
 * 验证 JWT Token 的中间件
 */
const verifyToken = (req, res, next) => {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ code: 1, message: '未授权：缺少 Token' });
  }

  const token = authHeader.substring(7);
  try {
    const decoded = jwt.verify(token, JWT_SECRET);
    req.user = decoded;
    next();
  } catch (error) {
    return res.status(401).json({ code: 1, message: '未授权：Token 无效或已过期' });
  }
};

// 辅助函数：获取用户在团队中的角色
const getUserRole = (teamId, username) => {
  const member = teamMembers.find(tm => tm.team_id === teamId && tm.user_id === username);
  return member ? member.role : null;
};

// ================== Lab 1 身份认证 (保留原有的严谨校验) ==================

app.post('/api/v1/auth/register', (req, res) => {
  const { username, password } = req.body;
  if (!username || typeof username !== 'string') return res.status(400).json({ code: 1, message: 'username required' });
  if (!password || typeof password !== 'string') return res.status(400).json({ code: 1, message: 'password required' });
  if (!/^[a-zA-Z0-9_]{4,20}$/.test(username)) return res.status(400).json({ code: 1, message: '用户名需为 4~20 位，仅允许字母、数字、下划线' });
  if (password.length < 6) return res.status(400).json({ code: 1, message: '密码长度不能少于 6 位' });
  
  const hasLetter = /[a-zA-Z]/.test(password);
  const hasNumber = /\d/.test(password);
  if (!hasLetter || !hasNumber) return res.status(400).json({ code: 1, message: '密码必须同时包含字母和数字' });
  if (users.has(username)) return res.status(409).json({ code: 1, message: '用户名已存在' });

  users.set(username, { password, level: 0 });
  return res.status(201).json({ code: 0, message: '注册成功' });
});

app.post('/api/v1/auth/login', (req, res) => {
  const { username, password } = req.body;
  if (!username || !password) return res.status(400).json({ code: 1, message: '用户名和密码不能为空' });

  const user = users.get(username);
  if (!user || user.password !== password) return res.status(401).json({ code: 1, message: '用户名或密码错误' });

  const token = jwt.sign({ username }, JWT_SECRET, { expiresIn: '7d' });
  return res.status(200).json({ code: 0, data: { token, username, level: user.level ?? 0 } });
});

app.put('/api/v1/auth/password', verifyToken, (req, res) => {
  const { username } = req.user;
  const { old_password, new_password } = req.body;
  const user = users.get(username);
  
  if (user.password !== old_password) return res.status(401).json({ code: 1, message: '旧密码错误' });
  if (new_password.length < 6) return res.status(400).json({ code: 1, message: '密码长度不能少于 6 位' });
  
  user.password = new_password;
  return res.status(200).json({ code: 0, message: '密码修改成功' });
});

// ================== Lab 2 任务模块扩展 ==================

/**
 * 获取任务列表 (支持个人视图与团队视图)
 */
app.get('/api/v1/tasks', verifyToken, (req, res) => {
  const { username } = req.user;
  const teamId = req.query.team_id ? parseInt(req.query.team_id) : null;
  const page = parseInt(req.query.page) || 1;
  const size = parseInt(req.query.size) || 10;

  let filteredTasks = [];

  if (teamId) {
    // 团队视图：必须是团队成员才能看
    if (!getUserRole(teamId, username)) return res.status(403).json({ code: 1, message: '无权访问该团队' });
    filteredTasks = tasks.filter(t => t.team_id === teamId);
  } else {
    // 个人视图：我创建的个人任务 + 分配给我的团队任务
    filteredTasks = tasks.filter(t => (t.user_id === username && !t.team_id) || t.assignee_id === username);
  }

  const start = (page - 1) * size;
  const paginatedTasks = filteredTasks.slice(start, start + size);

  return res.status(200).json({ code: 0, data: { list: paginatedTasks, total: filteredTasks.length }, page, size });
});

/**
 * 创建任务 (支持挂载到团队)
 */
app.post('/api/v1/tasks', verifyToken, (req, res) => {
  const { username } = req.user;
  const { title, description, status, priority, deadline, team_id, assignee_id } = req.body;

  if (!title) return res.status(400).json({ code: 1, message: '任务标题不能为空' });

  // 如果是团队任务，校验权限
  if (team_id) {
    const role = getUserRole(team_id, username);
    if (!role || role === 'Member') {
      return res.status(403).json({ code: 1, message: '仅团队 Owner 或 Admin 可发布团队任务' });
    }
  }

  const newTask = {
    id: taskIdCounter++,
    title: title.trim(),
    description: description || null,
    status: status || 'TODO',
    priority: priority ? parseInt(priority) : 1,
    deadline: deadline || null,
    user_id: username, 
    team_id: team_id || null,
    assignee_id: assignee_id || null,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  tasks.unshift(newTask); // 放到数组最前面，最新发布的在最上
  return res.status(201).json({ code: 0, message: '任务创建成功', data: newTask });
});

/**
 * 更新任务 (严格的 Member 权限拦截)
 */
app.put('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const { username } = req.user;
  const taskId = parseInt(req.params.id);
  const task = tasks.find(t => t.id === taskId);

  if (!task) return res.status(404).json({ code: 1, message: '任务不存在' });

  if (task.team_id) {
    // 团队任务权限校验
    const role = getUserRole(task.team_id, username);
    if (!role) return res.status(403).json({ code: 1, message: '无权操作' });
    
    // Member 只能修改状态
    if (role === 'Member') {
      if (req.body.title && req.body.title !== task.title) {
         return res.status(403).json({ code: 1, message: '普通成员仅能修改任务状态' });
      }
      if (req.body.status) task.status = req.body.status;
    } else {
      // Owner / Admin 可以修改所有
      Object.assign(task, req.body);
    }
  } else {
    // 个人任务权限校验
    if (task.user_id !== username) return res.status(403).json({ code: 1, message: '无权修改他人任务' });
    Object.assign(task, req.body);
  }

  task.updated_at = new Date().toISOString();
  return res.status(200).json({ code: 0, message: '任务更新成功' });
});

/**
 * 删除任务
 */
app.delete('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const { username } = req.user;
  const taskId = parseInt(req.params.id);
  const taskIndex = tasks.findIndex(t => t.id === taskId);

  if (taskIndex === -1) return res.status(404).json({ code: 1, message: '任务不存在' });
  const task = tasks[taskIndex];

  if (task.team_id) {
    const role = getUserRole(task.team_id, username);
    if (role !== 'Owner' && role !== 'Admin') return res.status(403).json({ code: 1, message: '无权删除团队任务' });
  } else {
    if (task.user_id !== username) return res.status(403).json({ code: 1, message: '无权删除' });
  }

  tasks.splice(taskIndex, 1);
  return res.status(200).json({ code: 0, message: '任务删除成功' });
});

// ================== Lab 2 全新团队模块 ==================

// 1. 获取我的团队
app.get('/api/v1/teams', verifyToken, (req, res) => {
  const { username } = req.user;
  const myTeams = teamMembers
    .filter(tm => tm.user_id === username)
    .map(tm => ({ ...teams.get(tm.team_id), role: tm.role }));
  res.status(200).json({ code: 0, data: { list: myTeams } });
});

// 2. 创建团队
app.post('/api/v1/teams', verifyToken, (req, res) => {
  const { username } = req.user;
  if (!req.body.name) return res.status(400).json({ code: 1, message: '团队名称不能为空' });
  
  const newTeam = { team_id: teamIdCounter++, name: req.body.name, owner_id: username };
  teams.set(newTeam.team_id, newTeam);
  teamMembers.push({ team_id: newTeam.team_id, user_id: username, role: 'Owner' });
  res.status(201).json({ code: 0, message: '团队创建成功' });
});

// 3. 获取团队成员
app.get('/api/v1/teams/:id/members', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.id);
  if (!getUserRole(teamId, req.user.username)) return res.status(403).json({ code: 1, message: '无权访问' });
  
  const members = teamMembers
    .filter(tm => tm.team_id === teamId)
    .map(tm => ({ user_id: tm.user_id, username: tm.user_id, role: tm.role }));
  res.status(200).json({ code: 0, data: members });
});

// 4. 邀请新成员
app.post('/api/v1/teams/:id/members', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.id);
  const targetUser = req.body.user_id; // 前端传过来的 user_id (在 mock 里用 username 模拟)
  
  if (getUserRole(teamId, req.user.username) !== 'Owner') return res.status(403).json({ code: 1, message: '仅 Owner 可拉人' });
  if (!users.has(targetUser)) return res.status(404).json({ code: 1, message: '被邀请的用户不存在' });
  if (getUserRole(teamId, targetUser)) return res.status(400).json({ code: 1, message: '该用户已在团队中' });

  teamMembers.push({ team_id: teamId, user_id: targetUser, role: 'Member' });
  res.status(200).json({ code: 0, message: '邀请成功' });
});

// 5. 修改权限
app.put('/api/v1/teams/:team_id/members/:user_id', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.team_id);
  const targetUser = req.params.user_id;
  
  if (getUserRole(teamId, req.user.username) !== 'Owner') return res.status(403).json({ code: 1, message: '无权操作' });
  
  const relation = teamMembers.find(tm => tm.team_id === teamId && tm.user_id === targetUser);
  if (relation) {
    relation.role = req.body.role;
    res.status(200).json({ code: 0, message: '权限更新成功' });
  } else {
    res.status(404).json({ code: 1, message: '成员不存在' });
  }
});

// 6. 踢出成员
app.delete('/api/v1/teams/:team_id/members/:user_id', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.team_id);
  const targetUser = req.params.user_id;

  if (getUserRole(teamId, req.user.username) !== 'Owner') return res.status(403).json({ code: 1, message: '无权操作' });
  if (targetUser === req.user.username) return res.status(400).json({ code: 1, message: 'Owner 不能踢出自己' });

  const index = teamMembers.findIndex(tm => tm.team_id === teamId && tm.user_id === targetUser);
  if (index > -1) {
    teamMembers.splice(index, 1);
    res.status(200).json({ code: 0, message: '移除成功' });
  } else {
    res.status(404).json({ code: 1, message: '成员不存在' });
  }
});

app.listen(PORT, () => {
  console.log(`🚀 Lab2 Mock Backend 运行中: http://localhost:${PORT}`);
});