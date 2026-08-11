# Configuración de Atlas para migraciones SQL versionadas.
# Fuente normativa: docs/05-backend/migraciones-atlas.md (DEC-036).
#
# Ninguna URL vive en este archivo: cada ambiente la recibe de una variable
# de entorno o de un secreto de CI, nunca de un valor escrito aquí.

env "local" {
  url = getenv("DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}

env "test" {
  url = getenv("DATABASE_TEST_URL")

  migration {
    dir = "file://migrations"
  }
}

env "pilot" {
  url = getenv("DATABASE_PILOT_URL")

  migration {
    dir = "file://migrations"
  }
}

env "production" {
  url = getenv("DATABASE_PRODUCTION_URL")

  migration {
    dir = "file://migrations"
  }
}
