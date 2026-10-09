import { randomBytes, randomUUID } from 'node:crypto'
import { spawnSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, copyFileSync, rmSync, readFileSync, writeFileSync, existsSync, chmodSync } from 'node:fs'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { dirname, resolve, join } from 'node:path'

export const scopes = [
  'access', 'recovery', 'shop', 'policy', 'staff', 'catalog', 'assignments',
  'schedules', 'agenda', 'manual', 'detail', 'public-entry', 'public-catalog',
  'public-barber', 'public-slots', 'public-confirm', 'customer', 'journeys',
  'isolation-a', 'isolation-b',
]
const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
export function databaseEnvironment(dsn, environment, role = 'barberia_migrator') {
  if (!['local', 'test'].includes(environment)) throw new Error('QA exige ambiente local/test')
  let url
  try { url = new URL(dsn) } catch { throw new Error('Conexión QA inválida') }
  if (!['postgres:', 'postgresql:'].includes(url.protocol) ||
      !['localhost', '127.0.0.1', '[::1]'].includes(url.hostname) ||
      !/^\/[a-zA-Z0-9_]+_(test|qa)$/.test(url.pathname) ||
      decodeURIComponent(url.username) !== role) {
    throw new Error('QA exige PostgreSQL loopback, base *_test/*_qa y el rol requerido')
  }
  return { ...process.env, PGHOST: url.hostname.replace(/^\[|\]$/g, ''),
    PGPORT: url.port || '5432', PGDATABASE: url.pathname.slice(1),
    PGUSER: decodeURIComponent(url.username), PGPASSWORD: decodeURIComponent(url.password),
    PGSSLMODE: url.searchParams.get('sslmode') || 'prefer', PGCONNECT_TIMEOUT: '5' }
}
export function newManifest() {
  const suffix = randomBytes(5).toString('hex')
  const phoneCampaign = String(parseInt(suffix, 16) % 10000000).padStart(7, '0')
  return { version: 1, campaign: 'qa-' + suffix, createdAt: new Date().toISOString(),
    accounts: scopes.map((scope, index) => ({
      scope, shopId: randomUUID(), userId: randomUUID(),
      email: 'qa.' + scope + '.' + suffix + '@example.test',
      password: randomBytes(24).toString('base64url') + '!aA1',
      phone: '+999' + phoneCampaign + String(index + 1).padStart(2, '0'),
      slug: 'qa-' + scope + '-' + suffix,
      barberIds: [randomUUID(), randomUUID()], serviceIds: [randomUUID(), randomUUID(), randomUUID()],
    })),
    inactive: { userId: randomUUID(), email: 'qa.inactive.' + suffix + '@example.test',
      password: randomBytes(24).toString('base64url') + '!aA1' },
  }
}
export function validateManifest(manifest) {
  const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/
  if (manifest.version !== 1 || !/^qa-[0-9a-f]{10}$/.test(manifest.campaign) ||
      !Array.isArray(manifest.accounts) || manifest.accounts.length !== scopes.length) throw new Error('Manifiesto QA inválido')
  const ids = []
  const emails = []
  for (const scope of scopes) {
    const a = manifest.accounts.find(a => a.scope === scope)
    if (!a || !a.email.endsWith('@example.test') || !a.email.startsWith('qa.' + scope + '.') ||
        !a.slug.startsWith('qa-' + scope + '-') || !/^\+999\d{9}$/.test(a.phone) ||
        typeof a.password !== 'string' || a.password.length < 10 ||
        a.barberIds?.length !== 2 || a.serviceIds?.length !== 3) throw new Error('Cuenta QA inválida')
    ids.push(a.shopId, a.userId, ...a.barberIds, ...a.serviceIds)
    emails.push(a.email)
  }
  if (!manifest.inactive?.email?.startsWith('qa.inactive.') ||
      !manifest.inactive.email.endsWith('@example.test') ||
      typeof manifest.inactive.password !== 'string' || manifest.inactive.password.length < 10) throw new Error('Cuenta inactiva QA inválida')
  ids.push(manifest.inactive.userId)
  emails.push(manifest.inactive.email)
  if (ids.some(id => !uuid.test(id)) || new Set(ids).size !== ids.length ||
      new Set(emails).size !== emails.length) throw new Error('Identidades QA inválidas o duplicadas')
}
const q = value => "'" + String(value).replaceAll("'", "''") + "'"
export function fixtureSQL(manifest, hashes) {
  validateManifest(manifest)
  if (hashes.length !== scopes.length + 1 || hashes.some(h => !h.startsWith('$argon2id$'))) throw new Error('Hashes QA inválidos')
  const admin = ["BEGIN; SET LOCAL ROLE barberia_owner;",
    "DO $$ BEGIN IF current_database() !~ '_(test|qa)$' THEN RAISE EXCEPTION 'Solo base QA'; END IF; END $$;"]
  const sql = ['BEGIN;']
  for (const [i, a] of manifest.accounts.entries()) {
    sql.push("SELECT set_config('app.barbershop_id'," + q(a.shopId) + ',true);')
    // Identidades aleatorias propias: las reejecuciones no restauran ni sobrescriben estados.
    const name = 'QA ' + a.scope + ' ' + manifest.campaign
    admin.push('INSERT INTO barbershop(id,name,timezone,public_slug) VALUES(' +
      [a.shopId, name, 'America/Bogota', a.slug].map(q).join(',') + ') ON CONFLICT(id) DO NOTHING;')
    admin.push('INSERT INTO staff_user(id,barbershop_id,email,full_name,phone,phone_verified_at) VALUES(' +
      [a.userId, a.shopId, a.email, 'Operador sintético ' + a.scope, a.phone].map(q).join(',') +
      ',now()) ON CONFLICT(id) DO NOTHING;')
    sql.push('INSERT INTO staff_credential(staff_user_id,barbershop_id,password_hash,password_algorithm) VALUES(' +
      [a.userId, a.shopId, hashes[i], 'argon2id'].map(q).join(',') + ') ON CONFLICT DO NOTHING;')
    for (const [j, id] of a.barberIds.entries()) {
      sql.push('INSERT INTO barber(id,barbershop_id,full_name) VALUES(' +
        [id, a.shopId, 'QA ' + a.scope + ' profesional ' + (j + 1)].map(q).join(',') + ') ON CONFLICT(id) DO NOTHING;')
      sql.push('INSERT INTO working_hour(barbershop_id,barber_id,iso_weekday,starts_time,duration_minutes) SELECT ' +
        q(a.shopId) + ',' + q(id) + ",d,'08:00'::time,720 FROM generate_series(1,7) d WHERE NOT EXISTS(SELECT 1 FROM working_hour WHERE barber_id=" +
        q(id) + ' AND iso_weekday=d);')
    }
    for (const [j, id] of a.serviceIds.entries()) {
      sql.push('INSERT INTO service(id,barbershop_id,name,duration_minutes,price_amount) VALUES(' +
        [id, a.shopId, 'QA ' + a.scope + ' servicio ' + (j + 1)].map(q).join(',') +
        ',' + [30, 45, 60][j] + ',' + [20000, 30000, 10000][j] + ') ON CONFLICT(id) DO NOTHING;')
      if (j < 2) for (const barberId of a.barberIds) sql.push('INSERT INTO barber_service(barbershop_id,barber_id,service_id) VALUES(' +
        [a.shopId, barberId, id].map(q).join(',') + ') ON CONFLICT DO NOTHING;')
    }
  }
  const a = manifest.accounts.find(a => a.scope === 'access')
  admin.push('INSERT INTO staff_user(id,barbershop_id,email,full_name,is_active) VALUES(' +
    [manifest.inactive.userId, a.shopId, manifest.inactive.email, 'Operador sintético inactivo'].map(q).join(',') +
    ',false) ON CONFLICT(id) DO NOTHING;')
  sql.push("SELECT set_config('app.barbershop_id'," + q(a.shopId) + ',true);')
  sql.push('INSERT INTO staff_credential(staff_user_id,barbershop_id,password_hash,password_algorithm) VALUES(' +
    [manifest.inactive.userId, a.shopId, hashes.at(-1), 'argon2id'].map(q).join(',') + ') ON CONFLICT DO NOTHING;')
  admin.push('COMMIT;')
  sql.push('COMMIT;')
  return { admin: admin.join('\n'), tenant: sql.join('\n') }
}
function run() {
  const pgEnv = databaseEnvironment(process.env.NAVA_QA_DATABASE_URL, process.env.APP_ENVIRONMENT)
  const appEnv = databaseEnvironment(process.env.APP_DATABASE_URL, process.env.APP_ENVIRONMENT, 'barberia_app')
  if (['PGHOST', 'PGPORT', 'PGDATABASE'].some(k => appEnv[k] !== pgEnv[k])) throw new Error('App y migrador deben usar la misma base QA')
  const directory = join(root, 'apps/web/.auth/nava-qa')
  mkdirSync(directory, { recursive: true, mode: 0o700 })
  chmodSync(directory, 0o700)
  const manifestPath = join(directory, 'current.json')
  const fresh = process.argv.includes('--new-run') || !existsSync(manifestPath)
  const manifest = fresh ? newManifest() : JSON.parse(readFileSync(manifestPath, 'utf8'))
  validateManifest(manifest)
  const completionPath = join(directory, manifest.campaign + '.provisioned')
  if (!fresh && existsSync(completionPath)) {
    console.log('Campaña ' + manifest.campaign + ': fixture existente conservado; sin restaurar datos ni contraseñas.')
    return
  }
  // Conserva el manifiesto antes de aplicar SQL para reintentar sin perder credenciales.
  const manifestText = JSON.stringify(manifest, null, 2) + '\n'
  writeFileSync(join(directory, manifest.campaign + '.json'), manifestText, { mode: 0o600 })
  writeFileSync(manifestPath, manifestText, { mode: 0o600 })
  chmodSync(manifestPath, 0o600)
  const runner = mkdtempSync(join(root, 'apps/api/.qa-hash-'))
  try {
    copyFileSync(join(root, 'tools/qa/password-hash.go'), join(runner, 'main.go'))
    const hash = spawnSync('go', ['run', join(runner, 'main.go')], {
      cwd: join(root, 'apps/api'), input: JSON.stringify([...manifest.accounts.map(a => a.password), manifest.inactive.password]),
      encoding: 'utf8', maxBuffer: 1024 * 1024 })
    if (hash.status !== 0) throw new Error('No se pudo generar Argon2 con el módulo Go existente')
    const sql = fixtureSQL(manifest, JSON.parse(hash.stdout))
    for (const [input, env] of [[sql.admin, pgEnv], [sql.tenant, appEnv]]) {
      const result = spawnSync('psql', ['-X', '-q', '-v', 'ON_ERROR_STOP=1'], {
      env, input, encoding: 'utf8', maxBuffer: 1024 * 1024 })
      if (result.status !== 0) {
        writeFileSync(join(directory, 'provision-error.local'), result.stderr || 'psql no disponible', { mode: 0o600 })
        throw new Error('Aprovisionamiento SQL falló; diagnóstico privado en apps/web/.auth/nava-qa/provision-error.local')
      }
    }
    writeFileSync(completionPath, new Date().toISOString() + '\n', { mode: 0o600 })
    rmSync(join(directory, 'provision-error.local'), { force: true })
    console.log('Campaña ' + manifest.campaign + ': ' + scopes.length + ' cuentas activas y 1 inactiva. Manifiesto privado: apps/web/.auth/nava-qa/current.json')
  } finally { rmSync(runner, { recursive: true, force: true }) }
}
if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try { run() } catch (error) { console.error(error.message); process.exitCode = 1 }
}
