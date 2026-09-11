const backendPort = Number.parseInt(process.env.APP_PORT ?? '8080', 10);

module.exports = {
  '/api': {
    target: `http://localhost:${backendPort}`,
    secure: false,
  },
};
