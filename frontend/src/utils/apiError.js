const HTTP_ERROR_MESSAGES = {
  400: '请求参数不正确，请检查提交内容',
  401: '登录状态已失效，请重新登录',
  403: '权限不足，无法执行该操作',
  404: '请求的资源不存在或已被删除',
  409: '操作冲突，请刷新后重试',
  422: '数据校验失败，请检查提交内容',
  429: '请求过于频繁，请稍后再试',
  500: '服务器内部错误，请稍后再试',
  502: '网关错误，服务暂时不可用',
  503: '服务暂时不可用，请稍后再试',
  504: '服务响应超时，请稍后再试',
};

const duplicateMessageByContext = (context, message) => {
  const contextText = String(context || '');
  const messageText = String(message || '').toLowerCase();
  const isDuplicate =
    /already exists|duplicate|重复|已存在/.test(messageText) ||
    messageText.includes('member already exists or invitation pending');

  if (!isDuplicate) return '';

  if (contextText.includes('注册')) {
    return '用户名已存在，请更换用户名';
  }
  if (contextText.includes('邀请成员') || contextText.includes('添加成员')) {
    return '该用户已在团队中，或已有待处理邀请，请勿重复邀请';
  }
  if (contextText.includes('创建团队') || contextText.includes('修改团队名称')) {
    return '团队名称已存在，请更换团队名称';
  }
  if (contextText.includes('添加前置依赖') || contextText.includes('关联前置任务')) {
    return '该前置依赖已存在，请勿重复添加';
  }

  return '数据已存在，请勿重复提交';
};

const MESSAGE_TRANSLATIONS = [
  [/invalid username or password/i, '用户名或密码错误'],
  [/invalid old password|old password/i, '旧密码错误，请重新输入'],
  [/unauthorized|invalid token|token expired/i, '登录状态已失效，请重新登录'],
  [/forbidden|permission denied|not authorized/i, '权限不足，无法执行该操作'],
  [/not found|record not found/i, '请求的资源不存在或已被删除'],
  [/member already exists or invitation pending/i, '该用户已在团队中，或已有待处理邀请，请勿重复邀请'],
  [/already exists|duplicate/i, '数据已存在，请勿重复提交'],
  [/invalid task title/i, '任务标题不合法'],
  [/invalid task status/i, '任务状态不合法'],
  [/invalid task priority/i, '任务优先级不合法'],
  [/invalid team id/i, '团队不存在或团队 ID 不合法'],
  [/invalid user id/i, '用户不存在或用户 ID 不合法'],
  [/invalid assignee id/i, '执行者不存在或执行者 ID 不合法'],
  [/assignee is not in team/i, '执行者不在该团队中'],
  [/invalid team role/i, '团队角色不合法'],
  [/team not found/i, '团队不存在或已被删除'],
  [/team or member not found/i, '团队或成员不存在'],
  [/task not found/i, '任务不存在或已被删除'],
  [/invitation is not pending/i, '邀请已处理，不能重复操作'],
  [/user_id or username is required/i, '请提供用户 ID 或用户名'],
  [/member can only update status|status only/i, 'Member 只能修改分配给自己的任务状态'],
  [/predecessor tasks are not completed|dependencies? not done/i, '尚有前置任务未完成'],
  [/deadline/i, '截止时间格式不正确'],
  [/validation|invalid/i, '提交的数据格式不正确'],
];

const hasChinese = (message) => /[\u4e00-\u9fff]/.test(String(message || ''));

const translateMessage = (message, context = '') => {
  const text = String(message || '').trim();
  if (!text) return '';

  const duplicateMessage = duplicateMessageByContext(context, text);
  if (duplicateMessage) return duplicateMessage;
  if (hasChinese(text)) return text;

  const matched = MESSAGE_TRANSLATIONS.find(([pattern]) => pattern.test(text));
  return matched ? matched[1] : '';
};

const extractServerMessage = (data) => {
  if (!data) return '';
  if (typeof data === 'string') return data;
  return data.message || data.error || data.msg || data.detail || '';
};

const withContext = (context, detail) => {
  const safeContext = String(context || '操作失败').trim();
  const safeDetail = String(detail || '请求失败').trim();
  return `${safeContext}：${safeDetail}`;
};

export const getApiErrorMessage = (error, context = '操作失败', fallback = '请求失败') => {
  const status = error?.response?.status;
  const serverMessage = extractServerMessage(error?.response?.data);
  const translatedMessage = translateMessage(serverMessage, context);
  const statusMessage = status ? HTTP_ERROR_MESSAGES[status] : '';

  if (error?.code === 'ECONNABORTED') {
    return withContext(context, '请求超时，请稍后重试');
  }

  if (!error?.response && error?.request) {
    return withContext(context, '无法连接服务器，请检查网络或后端服务是否启动');
  }

  const detail =
    translatedMessage ||
    statusMessage ||
    translateMessage(error?.message, context) ||
    fallback;
  return withContext(context, detail);
};

export const getApiFailureMessage = (response, context = '操作失败', fallback = '请求失败') => {
  const message = extractServerMessage(response?.data);
  const detail = translateMessage(message, context) || (hasChinese(message) ? message : '') || fallback;
  return withContext(context, detail);
};

export const getUnresolvedDependenciesMessage = (error, findTaskTitleById, context = '操作失败') => {
  const responseData = error?.response?.data || {};
  const data = responseData.data || {};
  const predecessors = Array.isArray(data.unresolved_predecessors)
    ? data.unresolved_predecessors
    : [];
  const ids = Array.isArray(data.unresolved_predecessor_ids)
    ? data.unresolved_predecessor_ids
    : [];

  const names = predecessors
    .map((task) => task.title || findTaskTitleById?.(task.id) || `ID ${task.id}`)
    .filter(Boolean);

  if (names.length === 0) {
    ids.forEach((id) => {
      names.push(findTaskTitleById?.(id) || `ID ${id}`);
    });
  }

  const uniqueNames = [...new Set(names)];
  if (uniqueNames.length > 0) {
    return withContext(context, `前置任务「${uniqueNames.join('」、「')}」尚未完成`);
  }

  return getApiErrorMessage(error, context, '尚有前置任务未完成');
};
