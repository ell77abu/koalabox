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

let taskIdCounter = 1;
let teamIdCounter = 1;
let inviteIdCounter = 1;

app.use(cors());
app.use(express.json());

// ================== 全自动 Token 校验 ==================
const verifyToken = (req, res, next) => {
  let token = '';

  //  尝试从传统的 Header 拿（兼容别人未来的修改）
  const authHeader = req.headers.authorization;
  if (authHeader && authHeader.startsWith('Bearer ')) {
    token = authHeader.substring(7);
  } 
  // 黑魔法：如果前端没传 Header，就去 Cookie 里找
  else if (req.headers.cookie) {
    const match = req.headers.cookie.match(/token=([^;]+)/);
    if (match) token = match[1];
  }

  if (!token) {
    return res.status(401).json({ code: 1, message: '未授权：缺少 Token' });
  }

  try {
    const decoded = jwt.verify(token, JWT_SECRET);
    req.user = decoded;
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
  
  // 在登录成功时，由后端直接把 Token 种到浏览器的 Cookie 里
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

// Session 接口 (防止刷新掉线)
app.get('/api/v1/auth/session', verifyToken, (req, res) => {
  const { username } = req.user;
  const user = users.get(username);
  if (!user) return res.status(401).json({ code: 1, message: '用户已注销' });
  return res.status(200).json({ code: 0, data: { username: username, level: user.level ?? 0 } });
});

// ================== 原有任务模块 ==================

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
  const newTask = { id: taskIdCounter++, ...req.body, user_id: req.user.username, created_at: new Date().toISOString() };
  tasks.unshift(newTask); 
  return res.status(201).json({ code: 0, message: '任务创建成功', data: newTask });
});

app.put('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const task = tasks.find(t => t.id === parseInt(req.params.id));
  if (task) Object.assign(task, req.body);
  return res.status(200).json({ code: 0, message: '任务更新成功' });
});

app.delete('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const index = tasks.findIndex(t => t.id === parseInt(req.params.id));
  if (index > -1) tasks.splice(index, 1);
  return res.status(200).json({ code: 0, message: '任务删除成功' });
});

// ================== 原有团队模块 ==================

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
  
  // 强力兼容：不管前端传的是 user_id 还是 username，全收！并强行转字符串+去首尾空格
  const rawTarget = req.body.user_id || req.body.username || '';
  const targetUser = rawTarget.toString().trim(); 
  const team = teams.get(teamId);
  
  // 👇 调试监控探头
  console.log(`\n--- 🕵️ 收到发起邀请请求 ---`);
  console.log(`前端发来的原始数据:`, req.body);
  console.log(`提取到的目标用户 ID: [${targetUser}]`);
  console.log(`当前系统里注册的所有用户:`, Array.from(users.keys()));
  console.log(`------------------------------\n`);

  if (!targetUser) return res.status(400).json({ code: 1, message: '缺少被邀请用户的信息' });
  if (!users.has(targetUser)) return res.status(404).json({ code: 1, message: `被邀请的用户 [${targetUser}] 不存在` });
  if (getUserRole(teamId, targetUser)) return res.status(400).json({ code: 1, message: '该用户已在团队中' });

  // 🛡️ 防御1：防止重复发送邀请 (查重)
  const existingInvite = invitations.find(i => 
    i.team_id === teamId && 
    i.invitee_id === targetUser && 
    i.status === 'PENDING'
  );
  if (existingInvite) {
    return res.status(400).json({ code: 1, message: '已向该用户发送过邀请，等待对方处理' });
  }

  const newInvite = {
    id: inviteIdCounter++,
    team_id: teamId,
    team_name: team ? team.name : 'Unknown Team',
    inviter_id: req.user.username,
    invitee_id: targetUser,
    status: 'PENDING',
    created_at: new Date().toISOString()
  };
  invitations.push(newInvite);
  
  // 🛡️ 防御2：包裹 { code: 0, data: ... }，解决前端报错弹窗问题
  res.status(200).json({ code: 0, message: '邀请发送成功', data: newInvite }); 
});

app.get('/api/v1/invitations', verifyToken, (req, res) => {
  const myInvites = invitations.filter(inv => inv.invitee_id === req.user.username);
  // 注意：这个接口大部分前端写的是直接读数组，如果前端没有报错，这里保持原样返回数组即可
  res.status(200).json(myInvites); 
});

app.post('/api/v1/invitations/:id/accept', verifyToken, (req, res) => {
  const invite = invitations.find(i => i.id === parseInt(req.params.id));
  
  if (invite && invite.status === 'PENDING') {
    // 🛡️ 防御3：同意前再次检查是否已在队伍中 (防止重复加入)
    if (getUserRole(invite.team_id, req.user.username)) {
      invite.status = 'ACCEPTED'; 
      return res.status(400).json({ code: 1, message: '您已加入该团队，请勿重复操作' });
    }

    invite.status = 'ACCEPTED';
    teamMembers.push({ team_id: invite.team_id, user_id: req.user.username, role: 'Member' });
    
    // 同样包裹 code: 0
    res.status(200).json({ code: 0, message: '已加入团队', data: { invitation_id: invite.id, status: 'ACCEPTED' } }); 
  } else {
    res.status(400).json({ code: 1, message: '邀请无效或已处理' });
  }
});

app.post('/api/v1/invitations/:id/decline', verifyToken, (req, res) => {
  const invite = invitations.find(i => i.id === parseInt(req.params.id));
  if (invite && invite.status === 'PENDING') {
    invite.status = 'DECLINED';
    res.status(200).json({ code: 0, message: '已拒绝邀请', data: { invitation_id: invite.id, status: 'DECLINED' } });
  } else {
    res.status(400).json({ code: 1, message: '邀请无效或已处理' });
  }
});

app.listen(PORT, () => {
  console.log(`🚀 Lab2 Mock Backend 运行中: http://localhost:${PORT}`);
});