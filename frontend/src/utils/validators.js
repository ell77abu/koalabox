const USERNAME_REGEX = /^[a-zA-Z0-9_]{4,20}$/;

export const validateUsername = (username) => {
  if (!USERNAME_REGEX.test(username)) {
    return '用户名需为 4~20 位，仅允许字母、数字、下划线';
  }
  return '';
};

export const validatePassword = (password) => {
  if (password.length < 6) {
    return '密码长度不能少于 6 位';
  }
  const hasLetter = /[a-zA-Z]/.test(password);
  const hasNumber = /\d/.test(password);
  if (!hasLetter || !hasNumber) {
    return '密码必须同时包含字母和数字';
  }
  return '';
};
