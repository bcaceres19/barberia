import test from 'node:test'
import assert from 'node:assert/strict'
import { databaseEnvironment, newManifest, validateManifest, fixtureSQL, scopes } from './provision-ui.mjs'

test('solo permite destino local de pruebas y rol migrador', () => {
  assert.equal(databaseEnvironment('postgres://barberia_migrator:p@127.0.0.1:5432/nava_test', 'local').PGDATABASE, 'nava_test')
  for (const dsn of ['postgres://barberia_migrator:p@remote.test/nava_test',
    'postgres://barberia_migrator:p@localhost/nava', 'postgres://postgres:p@localhost/nava_test',
    'https://barberia_migrator:p@localhost/nava_test']) {
    assert.throws(() => databaseEnvironment(dsn, 'local'))
  }
  assert.throws(() => databaseEnvironment('postgres://barberia_migrator:p@localhost/nava_test', 'production'))
})
test('cada tarea y cada campaña tienen identidades independientes', () => {
  const first = newManifest(), second = newManifest()
  validateManifest(first); validateManifest(second)
  assert.equal(new Set(first.accounts.map(a => a.shopId)).size, scopes.length)
  assert.equal(new Set(first.accounts.map(a => a.password)).size, scopes.length)
  assert.notEqual(first.campaign, second.campaign)
  assert.notEqual(first.accounts[0].phone, second.accounts[0].phone)
  second.accounts[0].password = "Synthetic10!"
  validateManifest(second)
  assert.notEqual(first.accounts[0].shopId, second.accounts[0].shopId)
  first.accounts[1].shopId = first.accounts[0].shopId
  assert.throws(() => validateManifest(first))
})
test('el fixture no borra ni restaura cuentas y no incluye contraseñas en claro', () => {
  const m = newManifest()
  const parts = fixtureSQL(m, Array(scopes.length + 1).fill('$argon2id$synthetic-test-only'))
  const sql = parts.admin + parts.tenant
  assert.doesNotMatch(sql, /\b(DELETE|TRUNCATE|UPDATE)\b/)
  assert.doesNotMatch(sql, /CASCADE/)
  for (const a of m.accounts) assert.equal(sql.includes(a.password), false)
  assert.match(sql, /SET LOCAL ROLE barberia_owner/)
  assert.throws(() => fixtureSQL(m, []))
})
