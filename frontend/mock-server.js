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

let taskIdCounter = 1;
let teamIdCounter = 1;
let inviteIdCounter = 1;
let assignReqIdCounter = 1;

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
// 辅助函数：根据 assignee_id 生成任务指派通知
const createAssignmentNotification = (task, reqUser) => {
  const assignee = task.assignee_id;
  if (!assignee || assignee === reqUser) return; // 不发给自己，没指派人也不发

  // 查重：防止对同一个人重复发同一条任务的 PENDING 提示
  const exists = assignmentRequests.find(r => r.task_id === task.id && r.assignee_id === assignee && r.status === 'PENDING');
  if (!exists) {
    assignmentRequests.push({
      id: assignReqIdCounter++,
      task_id: task.id,
      task_title: task.title,
      team_id: task.team_id,
      requester_id: reqUser,
      assignee_id: assignee,
      status: 'PENDING', // 虽然前端不做处理，但后台存为 PENDING
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
  // 发送通知
  createAssignmentNotification(newTask, req.user.username);
  return res.status(201).json({ code: 0, message: '任务创建成功', data: newTask });
});

app.put('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const task = tasks.find(t => t.id === parseInt(req.params.id));
  if (task) {
    Object.assign(task, req.body);
    // 变更指派人时，发送通知
    createAssignmentNotification(task, req.user.username);
  }
  return res.status(200).json({ code: 0, message: '任务更新成功' });
});

app.delete('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const index = tasks.findIndex(t => t.id === parseInt(req.params.id));
  if (index > -1) tasks.splice(index, 1);
  return res.status(200).json({ code: 0, message: '任务删除成功' });
});

// ================== 团队模块 ==================
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

app.get('/api/v1/invitations', verifyToken, (req, res) => {
  const myInvites = invitations.filter(inv => inv.invitee_id === req.user.username);
  res.status(200).json(myInvites); 
});

app.post('/api/v1/invitations/:id/accept', verifyToken, (req, res) => {
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

app.post('/api/v1/invitations/:id/decline', verifyToken, (req, res) => {
  const invite = invitations.find(i => i.id === parseInt(req.params.id));
  if (invite && invite.status === 'PENDING') {
    invite.status = 'DECLINED';
    res.status(200).json({ code: 0, message: '已拒绝邀请', data: { invitation_id: invite.id, status: 'DECLINED' } });
  } else res.status(400).json({ code: 1, message: '邀请无效或已处理' });
});

// API：查询当前用户的任务指派通知
app.get('/api/v1/assignment-requests', verifyToken, (req, res) => {
  const myRequests = assignmentRequests.filter(reqItem => reqItem.assignee_id === req.user.username);
  res.status(200).json(myRequests); 
});

// 配合前端的“静默已读”机制，消除小红点
app.post('/api/v1/assignment-requests/:id/accept', verifyToken, (req, res) => {
  const reqId = parseInt(req.params.id);
  const assignmentReq = assignmentRequests.find(r => r.id === reqId);

  if (assignmentReq && assignmentReq.status === 'PENDING') {
    assignmentReq.status = 'ACCEPTED'; // 将状态标记为已读，App.vue 里的红点雷达就不会再扫到它了
    res.status(200).json({ code: 0, message: '静默已读成功' });
  } else {
    res.status(400).json({ code: 1, message: '请求无效' });
  }
});

app.listen(PORT, () => {
  console.log(`🚀 Lab2 Mock Backend 运行中: http://localhost:${PORT}`);
});