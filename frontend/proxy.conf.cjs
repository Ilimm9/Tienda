module.exports = {
  '/api': {
    target: process.env.BACKEND_URL ?? `http://localhost:${process.env.APP_PORT ?? '8080'}`,
    secure: false,
  },
};
