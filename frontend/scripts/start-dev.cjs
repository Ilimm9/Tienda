const { spawn } = require('node:child_process');
const { existsSync } = require('node:fs');
const { resolve } = require('node:path');

const frontendRoot = resolve(__dirname, '..');
const envPath = resolve(frontendRoot, '..', 'backend', '.env');

if (existsSync(envPath)) {
  process.loadEnvFile(envPath);
}

function port(name, fallback) {
  const rawValue = process.env[name] || fallback;
  const value = Number(rawValue);
  if (!Number.isInteger(value) || value < 1 || value > 65535) {
    console.error(`${name} debe ser un puerto entre 1 y 65535; se recibió "${rawValue}".`);
    process.exit(1);
  }
  return String(value);
}

const frontendPort = port('FRONTEND_PORT', '4200');
const backendPort = port('APP_PORT', '8080');
const angularCLI = require.resolve('@angular/cli/bin/ng.js');
const extraArguments = process.argv.slice(2);

const child = spawn(
  process.execPath,
  [angularCLI, 'serve', '--port', frontendPort, '--proxy-config', 'proxy.conf.cjs', ...extraArguments],
  {
    cwd: frontendRoot,
    env: { ...process.env, APP_PORT: backendPort },
    stdio: 'inherit',
  },
);

child.on('error', (error) => {
  console.error(`No fue posible iniciar Angular: ${error.message}`);
  process.exitCode = 1;
});

child.on('exit', (code) => {
  process.exitCode = code ?? 1;
});
