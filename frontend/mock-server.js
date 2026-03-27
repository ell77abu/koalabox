import express from 'express';
import cors from 'cors';
import jwt from 'jsonwebtoken';

const app = express();
const PORT = 3001;
const JWT_SECRET = 'lab1_mock_secret_key';

// 内存数据库
const users = new Map();
const tasks = new Map(); // { username: [task1, task2, ...] }
let taskIdCounter = 1;

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

/**
 * 用户注册
 * POST /api/v1/auth/register
 */
app.post('/api/v1/auth/register', (req, res) => {
  const { username, password } = req.body;

  // 基础验证
  if (!username || typeof username !== 'string') {
    return res.status(400).json({ code: 1, message: 'username required' });
  }

  if (!password || typeof password !== 'string') {
    return res.status(400).json({ code: 1, message: 'password required' });
  }

  // 用户名格式检查
  if (!/^[a-zA-Z0-9_]{4,20}$/.test(username)) {
    return res.status(400).json({ code: 1, message: '用户名需为 4~20 位，仅允许字母、数字、下划线' });
  }

  // 密码长度检查
  if (password.length < 6) {
    return res.status(400).json({ code: 1, message: '密码长度不能少于 6 位' });
  }

  // 检查密码是否包含字母和数字
  const hasLetter = /[a-zA-Z]/.test(password);
  const hasNumber = /\d/.test(password);
  if (!hasLetter || !hasNumber) {
    return res.status(400).json({ code: 1, message: '密码必须同时包含字母和数字' });
  }

  // 检查用户是否已存在
  if (users.has(username)) {
    return res.status(409).json({ code: 1, message: '用户名已存在' });
  }

  // 存储用户（实际应该加盐哈希密码，这里 mock 简化处理）
  users.set(username, { password });
  // 初始化用户的任务列表
  tasks.set(username, []);

  return res.status(201).json({ code: 0, message: '注册成功' });
});

/**
 * 用户登录
 * POST /api/v1/auth/login
 */
app.post('/api/v1/auth/login', (req, res) => {
  const { username, password } = req.body;

  if (!username || !password) {
    return res.status(400).json({ code: 1, message: '用户名和密码不能为空' });
  }

  const user = users.get(username);
  if (!user || user.password !== password) {
    return res.status(401).json({ code: 1, message: '用户名或密码错误' });
  }

  // 确保用户有任务列表
  if (!tasks.has(username)) {
    tasks.set(username, []);
  }

  // 生成 JWT Token
  const token = jwt.sign({ username }, JWT_SECRET, { expiresIn: '7d' });

  return res.status(200).json({
    code: 0,
    data: {
      token,
      username,
    },
  });
});

/**
 * 获取任务列表
 * GET /api/v1/tasks?page=1&size=10
 */
app.get('/api/v1/tasks', verifyToken, (req, res) => {
  const { username } = req.user;
  const page = parseInt(req.query.page) || 1;
  const size = parseInt(req.query.size) || 10;

  const userTasks = tasks.get(username) || [];
  // 简单分页
  const start = (page - 1) * size;
  const end = start + size;
  const paginatedTasks = userTasks.slice(start, end);

  return res.status(200).json({
    code: 0,
    data: paginatedTasks,
    total: userTasks.length,
    page,
    size,
  });
});

/**
 * 创建任务
 * POST /api/v1/tasks
 */
app.post('/api/v1/tasks', verifyToken, (req, res) => {
  const { username } = req.user;
  const { title, description, status, priority, deadline } = req.body;

  // 验证必填字段
  if (!title || typeof title !== 'string' || title.trim().length === 0) {
    return res.status(400).json({ code: 1, message: '任务标题不能为空' });
  }

  if (title.length > 100) {
    return res.status(400).json({ code: 1, message: '任务标题不能超过 100 字符' });
  }

  if (priority && ![1, 2, 3].includes(parseInt(priority))) {
    return res.status(400).json({ code: 1, message: '优先级值无效' });
  }

  // 构建新任务
  const newTask = {
    id: taskIdCounter++,
    title: title.trim(),
    description: description || null,
    status: status || 'TODO',
    priority: priority ? parseInt(priority) : 1,
    deadline: deadline || null,
    user_id: username, // 标记为当前用户
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  // 添加到用户的任务列表
  if (!tasks.has(username)) {
    tasks.set(username, []);
  }
  tasks.get(username).push(newTask);

  return res.status(201).json({
    code: 0,
    message: '任务创建成功',
    data: newTask,
  });
});

/**
 * 更新任务
 * PUT /api/v1/tasks/:id
 */
app.put('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const { username } = req.user;
  const taskId = parseInt(req.params.id);
  const { title, description, status, priority, deadline } = req.body;

  const userTasks = tasks.get(username);
  if (!userTasks) {
    return res.status(404).json({ code: 1, message: '任务不存在' });
  }

  const task = userTasks.find((t) => t.id === taskId);
  if (!task) {
    return res.status(404).json({ code: 1, message: '任务不存在' });
  }

  // 验证权限（用户只能修改自己的任务）
  if (task.user_id !== username) {
    return res.status(403).json({ code: 1, message: '无权修改他人任务' });
  }

  // 更新字段
  if (title !== undefined) {
    if (!title || typeof title !== 'string' || title.trim().length === 0) {
      return res.status(400).json({ code: 1, message: '任务标题不能为空' });
    }
    task.title = title.trim();
  }

  if (description !== undefined) {
    task.description = description;
  }

  if (status !== undefined) {
    if (!['TODO', 'DOING', 'DONE'].includes(status)) {
      return res.status(400).json({ code: 1, message: '状态值无效' });
    }
    task.status = status;
  }

  if (priority !== undefined) {
    if (![1, 2, 3].includes(parseInt(priority))) {
      return res.status(400).json({ code: 1, message: '优先级值无效' });
    }
    task.priority = parseInt(priority);
  }

  if (deadline !== undefined) {
    task.deadline = deadline;
  }

  task.updated_at = new Date().toISOString();

  return res.status(200).json({
    code: 0,
    message: '任务更新成功',
    data: task,
  });
});

/**
 * 删除任务
 * DELETE /api/v1/tasks/:id
 */
app.delete('/api/v1/tasks/:id', verifyToken, (req, res) => {
  const { username } = req.user;
  const taskId = parseInt(req.params.id);

  const userTasks = tasks.get(username);
  if (!userTasks) {
    return res.status(404).json({ code: 1, message: '任务不存在' });
  }

  const taskIndex = userTasks.findIndex((t) => t.id === taskId);
  if (taskIndex === -1) {
    return res.status(404).json({ code: 1, message: '任务不存在' });
  }

  const task = userTasks[taskIndex];

  // 验证权限（用户只能删除自己的任务）
  if (task.user_id !== username) {
    return res.status(403).json({ code: 1, message: '无权删除他人任务' });
  }

  // 删除任务
  userTasks.splice(taskIndex, 1);

  return res.status(200).json({
    code: 0,
    message: '任务删除成功',
  });
});


// 健康检查
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.listen(PORT, () => {
  console.log(`Mock backend running on http://localhost:${PORT}`);
});
