import express from 'express';
import cors from 'cors';
import jwt from 'jsonwebtoken';

const app = express();
const PORT = 3001;
const JWT_SECRET = 'lab1_mock_secret_key';

// ================== 内存数据库 ==================
const users = new Map(); 
const tasks = []; 
const teams = new Map(); 
const teamMembers = [];  
const invitations = [];  
const assignmentRequests = []; // 任务分配通知表

// 👇 Lab3 新增：任务依赖关联表 (二元边表)
const taskDependencies = [];

let taskIdCounter = 1;
let teamIdCounter = 1;
let inviteIdCounter = 1;
let assignReqIdCounter = 1;
let depIdCounter = 1;

app.use(cors());
app.use(express.json());

// ================== 全自动 Token 校验 ==================
const verifyToken = (req, res, next) => {
  let token = '';
  const authHeader = req.headers.authorization;
  if (authHeader && authHeader.startsWith('Bearer ')) token = authHeader.substring(7);
  else if (req.headers.cookie) {
    const match = req.headers.cookie.match(/token=([^;]+)/);
    if (match) token = match[1];
  }
  if (!token) return res.status(401).json({ code: 1, message: '未授权：缺少 Token' });

  try {
    req.user = jwt.verify(token, JWT_SECRET);
    next();
  } catch (error) {
    return res.status(401).json({ code: 1, message: '未授权：Token 无效或已过期' });
  }
};

const getUserRole = (teamId, username) => {
  const member = teamMembers.find(tm => tm.team_id === teamId && tm.user_id === username);
  return member ? member.role : null;
};

// ================== 基础认证 ==================
app.post('/api/v1/auth/register', (req, res) => {
  const { username, password } = req.body;
  if (users.has(username)) return res.status(409).json({ code: 1, message: '用户名已存在' });
  users.set(username, { password, level: 0 });
  return res.status(201).json({ code: 0, message: '注册成功' });
});

app.post('/api/v1/auth/login', (req, res) => {
  const { username, password } = req.body;
  const user = users.get(username);
  if (!user || user.password !== password) return res.status(401).json({ code: 1, message: '用户名或密码错误' });
  const token = jwt.sign({ username }, JWT_SECRET, { expiresIn: '7d' });
  res.cookie('token', token, { path: '/' });
  return res.status(200).json({ code: 0, data: { token, username, level: user.level ?? 0 } });
});

app.put('/api/v1/auth/password', verifyToken, (req, res) => {
  const { username } = req.user;
  const { old_password, new_password } = req.body;
  const user = users.get(username);
  if (user.password !== old_password) return res.status(401).json({ code: 1, message: '旧密码错误' });
  user.password = new_password;
  return res.status(200).json({ code: 0, message: '密码修改成功' });
});

app.get('/api/v1/auth/session', verifyToken, (req, res) => {
  const { username } = req.user;
  const user = users.get(username);
  if (!user) return res.status(401).json({ code: 1, message: '用户已注销' });
  return res.status(200).json({ code: 0, data: { username: username, level: user.level ?? 0 } });
});

// ================== 任务模块与自动通知 ==================
const createAssignmentNotification = (task, reqUser) => {
  const assignee = task.assignee_id;
  if (!assignee || assignee === reqUser) return; 

  const exists = assignmentRequests.find(r => r.task_id === task.id && r.assignee_id === assignee && r.status === 'PENDING');
  if (!exists) {
    assignmentRequests.push({
      id: assignReqIdCounter++,
      task_id: task.id,
      task_title: task.title,
      team_id: task.team_id,
      requester_id: reqUser,
      assignee_id: assignee,
      status: 'PENDING', 
      created_at: new Date().toISOString()
    });
  }
};

app.get('/api/v1/tasks', verifyToken, (req, res) => {
  const { username } = req.user;
  const teamId = req.query.team_id ? parseInt(req.query.team_id) : null;
  const page = parseInt(req.query.page) || 1;
  const size = parseInt(req.query.size) || 10;

  let filteredTasks = teamId 
    ? tasks.filter(t => t.team_id === teamId)
    : tasks.filter(t => (t.user_id === username && !t.team_id) || t.assignee_id === username);

  const start = (page - 1) * size;
  return res.status(200).json({ code: 0, data: { list: filteredTasks.slice(start, start + size), total: filteredTasks.length }, page, size });
});

app.post('/api/v1/tasks', verifyToken, (req, res) => {
  const newTask = { 
    id: taskIdCounter++, 
    ...req.body, 
    user_id: req.user.username, 
    created_at: new Date().toISOString() 
  };
  tasks.unshift(newTask); 
  createAssignmentNotification(newTask, req.user.username);
  return res.status(201).json({ code: 0, message: '任务创建成功', data: newTask });
});

app.put('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const taskId = parseInt(req.params.id);
  const task = tasks.find(t => t.id === taskId);
  if (!task) return res.status(404).json({ code: 1, message: '任务不存在' });

  // 👇 Lab3：状态变更 409 拦截检查
  if (req.body.status === 'DONE' && task.status !== 'DONE') {
    const uncompletedPreds = taskDependencies
      .filter(d => d.successor_id === taskId)
      .map(d => tasks.find(t => t.id === d.predecessor_id))
      .filter(t => t && t.status !== 'DONE');

    if (uncompletedPreds.length > 0) {
      return res.status(409).json({ 
        code: 409, 
        message: `操作失败：尚有 ${uncompletedPreds.length} 个前置任务未完成`,
        data: { unresolved_predecessor_ids: uncompletedPreds.map(t => t.id) }
      });
    }
  }

  Object.assign(task, req.body);
  createAssignmentNotification(task, req.user.username);
  return res.status(200).json({ code: 0, message: '任务更新成功' });
});

app.delete('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const taskId = parseInt(req.params.id);
  const index = tasks.findIndex(t => t.id === taskId);
  if (index > -1) {
    tasks.splice(index, 1);
    // 👇 Lab3：级联删除相关的依赖关系 (模拟 ON DELETE CASCADE)
    for (let i = taskDependencies.length - 1; i >= 0; i--) {
      if (taskDependencies[i].predecessor_id === taskId || taskDependencies[i].successor_id === taskId) {
        taskDependencies.splice(i, 1);
      }
    }
  }
  return res.status(200).json({ code: 0, message: '任务删除成功' });
});

// ================== Lab3 任务依赖管理模块 ==================

// 辅助函数：DFS 去环检测
const checkCycle = (startNodeId, newPredecessorId) => {
  const visited = new Set();
  const dfs = (currentId) => {
    if (currentId === newPredecessorId) return true; // 发现环！
    visited.add(currentId);
    const successors = taskDependencies.filter(d => d.predecessor_id === currentId).map(d => d.successor_id);
    for (const succ of successors) {
      if (!visited.has(succ)) {
        if (dfs(succ)) return true;
      }
    }
    return false;
  };
  return dfs(startNodeId);
};

// 1. 获取任务的直接依赖
app.get('/api/v1/tasks/:id/dependencies', verifyToken, (req, res) => {
  const taskId = parseInt(req.params.id);
  const preds = taskDependencies.filter(d => d.successor_id === taskId).map(d => tasks.find(t => t.id === d.predecessor_id)).filter(Boolean);
  const succs = taskDependencies.filter(d => d.predecessor_id === taskId).map(d => tasks.find(t => t.id === d.successor_id)).filter(Boolean);
  
  res.status(200).json({ code: 0, data: { predecessors: preds, successors: succs } });
});

// 2. 获取用于画图的全量依赖数据 (DAG)
app.get('/api/v1/tasks/:id/dependency-graph', verifyToken, (req, res) => {
  const startTaskId = parseInt(req.params.id);
  const nodesMap = new Map();
  const linksMap = new Map();

  const explore = (currentId) => {
    if (nodesMap.has(currentId)) return;
    const task = tasks.find(t => t.id === currentId);
    if (!task) return;
    nodesMap.set(currentId, { id: task.id, title: task.title, status: task.status });

    taskDependencies.filter(d => d.predecessor_id === currentId || d.successor_id === currentId).forEach(d => {
      const linkKey = `${d.predecessor_id}-${d.successor_id}`;
      if (!linksMap.has(linkKey)) {
        linksMap.set(linkKey, { source: d.predecessor_id, target: d.successor_id });
        explore(d.predecessor_id);
        explore(d.successor_id);
      }
    });
  };

  explore(startTaskId);
  res.status(200).json({ code: 0, data: { nodes: Array.from(nodesMap.values()), links: Array.from(linksMap.values()) } });
});

// 3. 添加前置依赖 (带成环检测)
app.post('/api/v1/tasks/:id/dependencies', verifyToken, (req, res) => {
  const taskId = parseInt(req.params.id);
  const predId = parseInt(req.body.predecessor_id);

  if (taskId === predId) return res.status(400).json({ code: 400, message: '任务不能依赖自己' });
  
  const existing = taskDependencies.find(d => d.successor_id === taskId && d.predecessor_id === predId);
  if (existing) return res.status(200).json({ code: 0, message: '依赖已存在' });

  // 🪄 DFS 成环检测
  if (checkCycle(taskId, predId)) {
    return res.status(400).json({ code: 400, message: '无法添加依赖：检测到循环依赖 (Cycle Detected)' });
  }

  const newDep = { id: depIdCounter++, predecessor_id: predId, successor_id: taskId, created_at: new Date().toISOString() };
  taskDependencies.push(newDep);
  res.status(201).json({ code: 0, message: '依赖添加成功', data: newDep });
});

// 4. 删除前置依赖
app.delete('/api/v1/tasks/:id/dependencies/:pred_id', verifyToken, (req, res) => {
  const taskId = parseInt(req.params.id);
  const predId = parseInt(req.params.pred_id);
  const index = taskDependencies.findIndex(d => d.successor_id === taskId && d.predecessor_id === predId);
  if (index > -1) taskDependencies.splice(index, 1);
  res.status(200).json({ code: 0, message: '依赖移除成功' });
});

// ================== 团队模块与 Lab3 生命周期 ==================
app.get('/api/v1/teams', verifyToken, (req, res) => {
  const myTeams = teamMembers.filter(tm => tm.user_id === req.user.username).map(tm => ({ ...teams.get(tm.team_id), role: tm.role }));
  res.status(200).json({ code: 0, data: { list: myTeams } });
});

app.post('/api/v1/teams', verifyToken, (req, res) => {
  const newTeam = { team_id: teamIdCounter++, name: req.body.name, owner_id: req.user.username };
  teams.set(newTeam.team_id, newTeam);
  teamMembers.push({ team_id: newTeam.team_id, user_id: req.user.username, role: 'Owner' });
  res.status(201).json({ code: 0, message: '团队创建成功' });
});

app.get('/api/v1/teams/:id/members', verifyToken, (req, res) => {
  const members = teamMembers.filter(tm => tm.team_id === parseInt(req.params.id)).map(tm => ({ user_id: tm.user_id, username: tm.user_id, role: tm.role }));
  res.status(200).json({ code: 0, data: members });
});

//  Lab3 修改团队名称
app.put('/api/v1/teams/:id', verifyToken, (req, res) => {
  const team = teams.get(parseInt(req.params.id));
  if (team) { team.name = req.body.name; return res.status(200).json({ code: 0, message: '改名成功' }); }
  res.status(404).json({ code: 1, message: '团队不存在' });
});

// Lab3 修改成员权限
app.put('/api/v1/teams/:id/members/:userId', verifyToken, (req, res) => {
  const member = teamMembers.find(tm => tm.team_id === parseInt(req.params.id) && tm.user_id === req.params.userId);
  if (member) { member.role = req.body.role; return res.status(200).json({ code: 0, message: '权限修改成功' }); }
  res.status(400).json({ code: 1, message: '成员不存在' });
});

// Lab3 解散团队时，级联清空任务和依赖
app.delete('/api/v1/teams/:id', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.id);
  teams.delete(teamId);
  // 清理成员
  for (let i = teamMembers.length - 1; i >= 0; i--) { if (teamMembers[i].team_id === teamId) teamMembers.splice(i, 1); }
  // 清理任务及依赖
  for (let i = tasks.length - 1; i >= 0; i--) {
    if (tasks[i].team_id === teamId) {
      const deletedTaskId = tasks[i].id;
      tasks.splice(i, 1);
      for (let j = taskDependencies.length - 1; j >= 0; j--) {
        if (taskDependencies[j].predecessor_id === deletedTaskId || taskDependencies[j].successor_id === deletedTaskId) {
          taskDependencies.splice(j, 1);
        }
      }
    }
  }
  res.status(200).json({ code: 0, message: '团队已解散' });
});

// Lab3 移除成员，将其负责的未完成任务退回团队池
app.delete('/api/v1/teams/:id/members/:userId', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.id);
  const userId = req.params.userId;
  const index = teamMembers.findIndex(tm => tm.team_id === teamId && tm.user_id === userId);
  if (index > -1) {
    teamMembers.splice(index, 1);
    tasks.filter(t => t.team_id === teamId && t.assignee_id === userId && t.status !== 'DONE').forEach(t => t.assignee_id = null);
  }
  res.status(200).json({ code: 0, message: '移除成功' });
});

// Lab3 成员主动退出团队
app.post('/api/v1/teams/:id/leave', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.id);
  const userId = req.user.username;
  const role = getUserRole(teamId, userId);
  if (role === 'Owner') return res.status(403).json({ code: 403, message: 'Owner 必须先转让团队才能离开' });

  const index = teamMembers.findIndex(tm => tm.team_id === teamId && tm.user_id === userId);
  if (index > -1) {
    teamMembers.splice(index, 1);
    tasks.filter(t => t.team_id === teamId && t.assignee_id === userId && t.status !== 'DONE').forEach(t => t.assignee_id = null);
    return res.status(200).json({ code: 0, message: '已成功离开团队' });
  }
  res.status(400).json({ code: 1, message: '您不在该团队中' });
});

// Lab3 团队所有权转让
app.post('/api/v1/teams/:id/transfer', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.id);
  const newOwnerId = req.body.new_owner_id;
  
  const currentOwner = teamMembers.find(tm => tm.team_id === teamId && tm.user_id === req.user.username);
  const newOwner = teamMembers.find(tm => tm.team_id === teamId && tm.user_id === newOwnerId);

  if (!currentOwner || currentOwner.role !== 'Owner') return res.status(403).json({ code: 403, message: '只有 Owner 能转让团队' });
  if (!newOwner) return res.status(400).json({ code: 400, message: '目标成员不存在' });

  currentOwner.role = 'Admin'; // 降级为 Admin
  newOwner.role = 'Owner';     // 升级为 Owner
  
  const team = teams.get(teamId);
  if (team) team.owner_id = newOwnerId;

  res.status(200).json({ code: 0, message: '所有权转让成功' });
});

// ================== 消息通知与邀请功能 ==================
app.post('/api/v1/teams/:id/invitations', verifyToken, (req, res) => {
  const teamId = parseInt(req.params.id);
  const rawTarget = req.body.user_id || req.body.username || '';
  const targetUser = rawTarget.toString().trim(); 
  const team = teams.get(teamId);

  if (!targetUser) return res.status(400).json({ code: 1, message: '缺少被邀请用户的信息' });
  if (!users.has(targetUser)) return res.status(404).json({ code: 1, message: `被邀请的用户 [${targetUser}] 不存在` });
  if (getUserRole(teamId, targetUser)) return res.status(400).json({ code: 1, message: '该用户已在团队中' });

  const existingInvite = invitations.find(i => i.team_id === teamId && i.invitee_id === targetUser && i.status === 'PENDING');
  if (existingInvite) return res.status(400).json({ code: 1, message: '已向该用户发送过邀请，等待对方处理' });

  const newInvite = {
    id: inviteIdCounter++, team_id: teamId, team_name: team ? team.name : 'Unknown Team',
    inviter_id: req.user.username, invitee_id: targetUser, status: 'PENDING', created_at: new Date().toISOString()
  };
  invitations.push(newInvite);
  res.status(200).json({ code: 0, message: '邀请发送成功', data: newInvite }); 
});

app.get('/api/v1/teams/invitations/pending', verifyToken, (req, res) => {
  const myInvites = invitations.filter(inv => inv.invitee_id === req.user.username);
  res.status(200).json({ code: 0, data: myInvites }); 
});

app.post('/api/v1/teams/invitations/:id/accept', verifyToken, (req, res) => {
  const invite = invitations.find(i => i.id === parseInt(req.params.id));
  if (invite && invite.status === 'PENDING') {
    if (getUserRole(invite.team_id, req.user.username)) {
      invite.status = 'ACCEPTED'; 
      return res.status(400).json({ code: 1, message: '您已加入该团队，请勿重复操作' });
    }
    invite.status = 'ACCEPTED';
    teamMembers.push({ team_id: invite.team_id, user_id: req.user.username, role: 'Member' });
    res.status(200).json({ code: 0, message: '已加入团队', data: { invitation_id: invite.id, status: 'ACCEPTED' } }); 
  } else res.status(400).json({ code: 1, message: '邀请无效或已处理' });
});

app.post('/api/v1/teams/invitations/:id/decline', verifyToken, (req, res) => {
  const invite = invitations.find(i => i.id === parseInt(req.params.id));
  if (invite && invite.status === 'PENDING') {
    invite.status = 'DECLINED';
    res.status(200).json({ code: 0, message: '已拒绝邀请', data: { invitation_id: invite.id, status: 'DECLINED' } });
  } else res.status(400).json({ code: 1, message: '邀请无效或已处理' });
});

app.get('/api/v1/assignment-requests', verifyToken, (req, res) => {
  const myRequests = assignmentRequests.filter(reqItem => reqItem.assignee_id === req.user.username);
  res.status(200).json(myRequests); 
});

app.post('/api/v1/assignment-requests/:id/accept', verifyToken, (req, res) => {
  const reqId = parseInt(req.params.id);
  const assignmentReq = assignmentRequests.find(r => r.id === reqId);

  if (assignmentReq && assignmentReq.status === 'PENDING') {
    assignmentReq.status = 'ACCEPTED'; 
    res.status(200).json({ code: 0, message: '静默已读成功' });
  } else {
    res.status(400).json({ code: 1, message: '请求无效' });
  }
});

app.listen(PORT, () => {
  console.log(`🚀 Lab3 Mock Backend 运行中: http://localhost:${PORT}`);
});